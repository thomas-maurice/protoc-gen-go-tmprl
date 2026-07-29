//go:build e2e

// Package e2e tests the example Orders service against a REAL temporal
// server, driving everything through the generated client-side API -- the
// part the in-memory testsuite cannot reach (client.UpdateWorkflow wiring,
// WaitForStage defaulting, schedule CRUD, error unwrapping across the RPC
// boundary). The worker runs in-process.
//
// These tests are build-tagged so plain `go test ./...` skips them; run them
// with a server up (docker compose up -d --wait) via:
//
//	make test-e2e
//
// TEMPORAL_ADDRESS overrides the target server, defaults to localhost:7233.
package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/thomas-maurice/protoc-gen-go-tmprl/example/orders"
	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ordersClient *examplev1.OrdersClient

// someItems returns a single-line order to keep packing (300ms per parcel)
// fast in e2e runs.
func someItems() []*examplev1.OrderItem {
	return []*examplev1.OrderItem{{Sku: "die-d20", Quantity: 1}}
}

// orderRequest returns a ProcessOrder request with the given card token and
// pickup window.
func orderRequest(orderID, cardToken string, pickupSeconds int32) *examplev1.ProcessOrderRequest {
	return &examplev1.ProcessOrderRequest{
		OrderId:             orderID,
		Items:               someItems(),
		AmountCents:         4200,
		CardToken:           cardToken,
		ShippingAddress:     "1 rue de la Paix, Paris",
		PickupWindowSeconds: pickupSeconds,
	}
}

// waitForStatus polls the GetOrderStatus query until the workflow reaches the
// wanted status or the deadline passes.
func waitForStatus(ctx context.Context, t *testing.T, run *examplev1.OrdersProcessOrder, want examplev1.OrderStatus) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		st, err := run.QueryGetOrderStatus(ctx, &emptypb.Empty{})
		if err == nil && st.Status == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("workflow never reached %v (last error: %v)", want, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// TestMain dials the server and runs one in-process worker for the whole
// suite. No server -> hard failure: these tests only run when a stack is
// expected to be up.
func TestMain(m *testing.M) {
	c, err := client.Dial(client.Options{
		HostPort: os.Getenv("TEMPORAL_ADDRESS"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: cannot reach temporal server (start it with `docker compose up -d --wait`): %v\n", err)
		os.Exit(1)
	}

	ordersClient, err = examplev1.NewOrdersClient(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: cannot create orders client: %v\n", err)
		os.Exit(1)
	}

	w, err := examplev1.NewOrdersWorker(c, orders.New(ordersClient), "", worker.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: cannot create worker: %v\n", err)
		os.Exit(1)
	}
	w.Register()
	if err := w.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: cannot start worker: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	w.Stop()
	c.Close()
	os.Exit(code)
}

// TestHappyPathWithClientSideUpdates: the full lifecycle driven through the
// generated client: async start, query polling, one update through the
// client-level helper and one through the workflow object (both blocking,
// both returning the replaced address), a validator rejection, then the typed
// result shipped to the last address set.
func TestHappyPathWithClientSideUpdates(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowProcessOrder(ctx, orderRequest("e2e-happy", orders.CardOK, 20))
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}
	run := ordersClient.GetProcessOrderFromRun(future)
	t.Logf("started order %s: it will charge the card, pack the parcel, then wait 20s for carrier pickup", future.GetID())

	waitForStatus(ctx, t, run, examplev1.OrderStatus_ORDER_STATUS_PACKED)

	// Client-level update helper: addresses the execution by ID, blocks until
	// the handler returns, yields the typed response.
	t.Log("order is PACKED and parked in its pickup window: this is the phase where a running workflow can be talked to")
	first, err := ordersClient.UpdateChangeShippingAddress(ctx, future.GetID(), future.GetRunID(), &examplev1.ChangeShippingAddressRequest{
		Address: "99 boulevard de Test, Lille",
	})
	if err != nil {
		t.Fatalf("client-level update failed: %v", err)
	}
	if first.PreviousAddress != "1 rue de la Paix, Paris" {
		t.Errorf("first update previous address = %q, want the original", first.PreviousAddress)
	}
	t.Logf("update #1 blocked until the workflow answered, then returned the replaced address: %q", first.PreviousAddress)

	// Workflow-object update helper: same semantics, bound to the run.
	second, err := run.UpdateChangeShippingAddress(ctx, &examplev1.ChangeShippingAddressRequest{
		Address: "221B Baker Street, London",
	})
	if err != nil {
		t.Fatalf("workflow-object update failed: %v", err)
	}
	if second.PreviousAddress != "99 boulevard de Test, Lille" {
		t.Errorf("second update previous address = %q, want the first re-route", second.PreviousAddress)
	}
	t.Logf("update #2 (workflow-object flavour) chained correctly: it replaced %q", second.PreviousAddress)

	// The validator must reject an empty address and surface the error to the
	// caller without touching the workflow.
	if _, err := ordersClient.UpdateChangeShippingAddress(ctx, future.GetID(), future.GetRunID(), &examplev1.ChangeShippingAddressRequest{}); err == nil {
		t.Error("empty address update should have been rejected by the validator")
	}
	t.Log("empty-address update was rejected by the validator BEFORE reaching workflow history: the caller gets the error, the workflow never sees it")

	result, err := run.Result(ctx)
	if err != nil {
		t.Fatalf("workflow failed: %v", err)
	}
	if result.Status != examplev1.OrderStatus_ORDER_STATUS_SHIPPED {
		t.Errorf("status = %v, want SHIPPED", result.Status)
	}
	if result.ShippingAddress != "221B Baker Street, London" {
		t.Errorf("shipped to %q, want the last updated address", result.ShippingAddress)
	}
	if result.TrackingNumber == "" {
		t.Error("expected a tracking number from the ShipOrder child workflow")
	}
	t.Logf("pickup window expired: order shipped to the LAST address set (%s), tracking %s from the ShipOrder child workflow", result.ShippingAddress, result.TrackingNumber)

	// Queries keep working after completion as long as a worker runs.
	st, err := run.QueryGetOrderStatus(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("query after completion failed: %v", err)
	}
	if st.Status != examplev1.OrderStatus_ORDER_STATUS_SHIPPED {
		t.Errorf("post-completion query status = %v, want SHIPPED", st.Status)
	}
	t.Log("query still answered AFTER workflow completion: queries read history, they do not need a running workflow")
}

// TestCancelSignal: a signal sent through the generated client helper cancels
// an order parked in its pickup window.
func TestCancelSignal(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowProcessOrder(ctx, orderRequest("e2e-cancel", orders.CardOK, 60))
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}
	run := ordersClient.GetProcessOrderFromRun(future)

	waitForStatus(ctx, t, run, examplev1.OrderStatus_ORDER_STATUS_PACKED)
	t.Log("order is parked in a 60s pickup window; sending the CancelOrder signal instead of waiting")

	err = ordersClient.SendSignalCancelOrder(ctx, future.GetID(), future.GetRunID(), &examplev1.CancelOrderRequest{
		Reason: "changed my mind",
	})
	if err != nil {
		t.Fatalf("could not send signal: %v", err)
	}
	t.Log("signal sent: unlike an update it is fire-and-forget, there is no response to wait for")

	result, err := run.Result(ctx)
	if err != nil {
		t.Fatalf("workflow failed: %v", err)
	}
	if result.Status != examplev1.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Errorf("status = %v, want CANCELLED", result.Status)
	}
	if result.CancelReason != "changed my mind" {
		t.Errorf("cancel reason = %q, want the signalled one", result.CancelReason)
	}
	t.Logf("workflow noticed the flag its signal handler flipped and ended CANCELLED (reason %q) without waiting out the window", result.CancelReason)
}

// TestFlakyPaymentRetries: the retry policy from the proto rides out two
// transient ChargePayment failures; the caller only sees success.
func TestFlakyPaymentRetries(t *testing.T) {
	ctx := context.Background()

	t.Log("the tok-flaky card makes ChargePayment fail twice; the retry policy from the proto (1s initial, backoff 2.0) retries it server-side")
	result, err := ordersClient.ExecuteWorkflowProcessOrderSync(ctx, orderRequest("e2e-flaky", orders.CardFlaky, 1))
	if err != nil {
		t.Fatalf("workflow failed despite retry policy: %v", err)
	}
	if result.Status != examplev1.OrderStatus_ORDER_STATUS_SHIPPED {
		t.Errorf("status = %v, want SHIPPED", result.Status)
	}
	t.Log("caller saw none of the two failures, just a slower success: that is what a retry policy buys you")
}

// TestCardDeclinedFailsFast: a CardDeclined application error is in the retry
// policy's non-retryable list, so the workflow fails immediately and the
// typed error is recoverable with errors.As on the caller side.
func TestCardDeclinedFailsFast(t *testing.T) {
	ctx := context.Background()

	_, err := ordersClient.ExecuteWorkflowProcessOrderSync(ctx, orderRequest("e2e-declined", orders.CardDeclined, 1))
	if err == nil {
		t.Fatal("declined card should have failed the workflow")
	}
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an ApplicationError in the chain, got %T: %v", err, err)
	}
	if appErr.Type() != "CardDeclined" {
		t.Errorf("error type = %q, want CardDeclined", appErr.Type())
	}
	t.Logf("workflow failed in one attempt (CardDeclined is non-retryable) and the caller can errors.As the typed failure: type=%s message=%q", appErr.Type(), appErr.Message())
}

// TestScheduleCRUD: the generated schedule surface against a real server:
// upsert, list (visibility is eventually consistent, so poll), idempotent
// pause/unpause, delete.
func TestScheduleCRUD(t *testing.T) {
	ctx := context.Background()
	const scheduleID = "e2e-daily-sales-report"

	if _, err := ordersClient.UpsertScheduleDailySalesReport(ctx, scheduleID, &emptypb.Empty{}, client.ScheduleOptions{
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{{Every: time.Minute}},
		},
	}); err != nil {
		t.Fatalf("could not upsert schedule: %v", err)
	}
	// Best-effort cleanup even if an assertion below fails.
	defer func() { _ = ordersClient.DeleteScheduleDailySalesReport(ctx, scheduleID) }()

	found := false
	for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); {
		entries, err := ordersClient.ListScheduleDailySalesReport(ctx, 10)
		if err != nil {
			t.Fatalf("could not list schedules: %v", err)
		}
		if len(entries) > 0 {
			found = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !found {
		t.Error("schedule never showed up in the list (visibility)")
	}
	t.Log("schedule created and visible in listings (listing reads the visibility store, which lags creation by design, hence the poll)")

	if err := ordersClient.PauseScheduleDailySalesReport(ctx, scheduleID, "e2e pause"); err != nil {
		t.Fatalf("could not pause schedule: %v", err)
	}
	// Pausing again must be a no-op, not an error.
	if err := ordersClient.PauseScheduleDailySalesReport(ctx, scheduleID, "e2e pause again"); err != nil {
		t.Fatalf("second pause should be idempotent: %v", err)
	}
	if err := ordersClient.UnpauseScheduleDailySalesReport(ctx, scheduleID, "e2e unpause"); err != nil {
		t.Fatalf("could not unpause schedule: %v", err)
	}
	if err := ordersClient.DeleteScheduleDailySalesReport(ctx, scheduleID); err != nil {
		t.Fatalf("could not delete schedule: %v", err)
	}
	t.Log("pause was idempotent (second call no-ops), unpause and delete clean: the full generated schedule lifecycle works against a real server")
}


// cancelInventory requests cancellation and waits for the workflow to
// actually END, gracefully returning its final state. Regression guard: an
// earlier TrackInventory used a blocking signal receive that never observed
// cancellation, so Cancel left the workflow running forever and this
// "cleanup" silently leaked executions.
func cancelInventory(ctx context.Context, t *testing.T, inv *examplev1.OrdersTrackInventory) {
	t.Helper()
	if err := inv.Cancel(ctx); err != nil {
		t.Fatalf("could not cancel the inventory workflow: %v", err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	final, err := inv.Result(waitCtx)
	if err != nil {
		t.Fatalf("inventory workflow did not end gracefully after cancel: %v", err)
	}
	t.Logf("inventory ended gracefully after cancel: final stock %d at generation %d", final.Stock, final.Generation)
}

// TestBlockedUpdateSurvivesContinueAsNew: the entity-workflow rollover story.
// A Reserve update parks on an empty TrackInventory workflow. The Restock
// that satisfies it is ALSO the one that triggers the continue-as-new
// rollover -- and because the workflow drains its handlers
// (workflow.AllHandlersFinished) before rolling over, the parked caller gets
// its answer from the old run, then the state (post-reservation stock)
// carries into the new run. From the caller's perspective the rollover is
// invisible.
func TestBlockedUpdateSurvivesContinueAsNew(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowTrackInventory(ctx, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d20",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
	})
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}
	t.Logf("inventory %s started empty; it will roll over with continue-as-new after every restock", future.GetID())

	// Reserve 3 items from an EMPTY inventory: the update handler parks on
	// workflow.Await(stock >= 3) and this client call blocks with it.
	type reserveResult struct {
		resp *examplev1.ReserveResponse
		err  error
	}
	resultCh := make(chan reserveResult, 1)
	go func() {
		updCtx, cancelUpd := context.WithTimeout(ctx, 90*time.Second)
		defer cancelUpd()
		resp, err := ordersClient.UpdateReserve(updCtx, future.GetID(), "", &examplev1.ReserveRequest{Quantity: 3})
		resultCh <- reserveResult{resp, err}
	}()

	// Give the update time to be accepted and parked before restocking.
	time.Sleep(2 * time.Second)
	select {
	case r := <-resultCh:
		t.Fatalf("reserve should still be blocked on empty stock, got resp=%v err=%v", r.resp, r.err)
	default:
	}
	t.Log("Reserve(3) is parked: the caller is blocked, the workflow is idle, nobody is polling anything")

	// This restock satisfies the parked reservation AND triggers the rollover.
	err = ordersClient.SendSignalRestock(ctx, future.GetID(), "", &examplev1.RestockRequest{Quantity: 5})
	if err != nil {
		t.Fatalf("could not send restock signal: %v", err)
	}

	r := <-resultCh
	if r.err != nil {
		t.Fatalf("blocked reserve failed: %v", r.err)
	}
	if r.resp.RemainingStock != 2 {
		t.Errorf("remaining stock = %d, want 2 (5 restocked - 3 reserved)", r.resp.RemainingStock)
	}
	if r.resp.Generation != 1 {
		t.Errorf("generation = %d, want 1: the OLD run must answer before rolling over", r.resp.Generation)
	}
	t.Logf("parked caller got its answer (remaining=%d) from generation %d, BEFORE the rollover: that is the handler drain at work", r.resp.RemainingStock, r.resp.Generation)

	// The workflow chain lives on: poll the query (by workflow ID, latest
	// run) until the new generation is up, with the reserved stock carried.
	inv := ordersClient.GetTrackInventory(ctx, future.GetID(), "")
	deadline := time.Now().Add(15 * time.Second)
	for {
		st, err := inv.QueryGetStock(ctx, &emptypb.Empty{})
		if err == nil && st.Generation == 2 {
			if st.Stock != 2 {
				t.Errorf("new generation stock = %d, want 2 carried over", st.Stock)
			}
			t.Logf("continue-as-new done: generation %d is running with stock %d carried over from the old run", st.Generation, st.Stock)
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("new generation never came up (last: %v, err=%v)", st, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	cancelInventory(ctx, t, inv)
}

// TestBlockedUpdateAbortedWithoutHandlerDrain: the anti-pattern, pinned so
// the failure mode is documented. Note the nuance (the server taught us this
// one): a parked update that the rollover-triggering event SATISFIES still
// completes, drain or no drain, because the dispatcher finishes runnable
// coroutines before shipping the continue-as-new. The update that dies is
// one that is STILL UNSATISFIABLE when the run ends -- here Reserve(10) with
// only 5 in stock. With skip_handler_drain the workflow rolls over anyway:
// the parked Reserve belonged to the old run and is aborted, its caller gets
// an error, and the new run's stock proves the reservation never happened.
// (With the drain, the rollover would instead WAIT until some restock makes
// the reservation satisfiable -- no caller is ever abandoned.)
func TestBlockedUpdateAbortedWithoutHandlerDrain(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowTrackInventory(ctx, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d6",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
		SkipHandlerDrain:            true,
	})
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}

	type reserveResult struct {
		resp *examplev1.ReserveResponse
		err  error
	}
	resultCh := make(chan reserveResult, 1)
	go func() {
		updCtx, cancelUpd := context.WithTimeout(ctx, 90*time.Second)
		defer cancelUpd()
		resp, err := ordersClient.UpdateReserve(updCtx, future.GetID(), "", &examplev1.ReserveRequest{Quantity: 10})
		resultCh <- reserveResult{resp, err}
	}()
	time.Sleep(2 * time.Second)

	// 5 < 10: the reservation stays unsatisfiable, and the rollover happens
	// anyway because the drain is skipped. The parked update is left behind.
	if err := ordersClient.SendSignalRestock(ctx, future.GetID(), "", &examplev1.RestockRequest{Quantity: 5}); err != nil {
		t.Fatalf("could not send restock signal: %v", err)
	}

	r := <-resultCh
	if r.err == nil {
		t.Fatalf("expected the parked update to be aborted by continue-as-new, got resp=%v", r.resp)
	}
	t.Logf("parked Reserve(10) got an error instead of an answer: %v", r.err)

	// The new run is up but the reservation never happened: full stock.
	inv := ordersClient.GetTrackInventory(ctx, future.GetID(), "")
	deadline := time.Now().Add(15 * time.Second)
	for {
		st, err := inv.QueryGetStock(ctx, &emptypb.Empty{})
		if err == nil && st.Generation == 2 {
			if st.Stock != 5 {
				t.Errorf("new generation stock = %d, want 5: the aborted reservation must NOT have been applied", st.Stock)
			}
			t.Logf("generation %d carries stock %d: the 10-item reservation was lost with the old run -- this is why you drain handlers before continue-as-new", st.Generation, st.Stock)
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("new generation never came up (last: %v, err=%v)", st, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	cancelInventory(ctx, t, inv)
}

// TestNoRestockLostAcrossContinueAsNew: regression for a reproduced
// data-loss bug. Two Restock signals arrive back-to-back around a rollover:
// the first triggers continue-as-new, and the second used to sit unread in
// the closing run's signal channel. Signals recorded in a closed run's
// history do NOT carry over to its continue-as-new successor, so that stock
// was silently lost (8/8 reproductions before the fix). The workflow now
// sweeps its signal channel right before rolling over, on both the drain and
// the skip_handler_drain paths.
func TestNoRestockLostAcrossContinueAsNew(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowTrackInventory(ctx, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d12",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
	})
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Back-to-back: the first restock triggers the rollover, the second must
	// not be lost whether it lands before or after the boundary.
	if err := ordersClient.SendSignalRestock(ctx, future.GetID(), "", &examplev1.RestockRequest{Quantity: 5}); err != nil {
		t.Fatalf("could not send first restock: %v", err)
	}
	if err := ordersClient.SendSignalRestock(ctx, future.GetID(), "", &examplev1.RestockRequest{Quantity: 7}); err != nil {
		t.Fatalf("could not send second restock: %v", err)
	}

	inv := ordersClient.GetTrackInventory(ctx, future.GetID(), "")
	deadline := time.Now().Add(15 * time.Second)
	for {
		st, err := inv.QueryGetStock(ctx, &emptypb.Empty{})
		if err == nil && st.Stock == 12 {
			t.Logf("no stock lost across the rollover: %d units accounted for at generation %d", st.Stock, st.Generation)
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("stock never reached 12, a restock was lost across continue-as-new (last: %v, err=%v)", st, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	cancelInventory(ctx, t, inv)
}

// TestSkipDrainNeverDoubleCounts: regression for a reproduced data-corruption
// bug. An earlier fix ran the pre-rollover signal sweep on the
// skip_handler_drain path too; with Reserve(10) parked and Restock(5)+
// Restock(7) straddling the rollover, the sweep made the reservation
// satisfiable AFTER the continue-as-new arguments were snapshotted -- the
// caller was told "reserved, 2 remaining" while the successor started with
// all 12 units. The invariant pinned here is race-proof: the caller must
// NEVER receive success while the successor's stock does not reflect the
// decrement. On the fixed code the parked update is always aborted on this
// path (stock never reaches 10 in the closing run).
func TestSkipDrainNeverDoubleCounts(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowTrackInventory(ctx, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d4",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
		SkipHandlerDrain:            true,
	})
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}

	type reserveResult struct {
		resp *examplev1.ReserveResponse
		err  error
	}
	resultCh := make(chan reserveResult, 1)
	go func() {
		updCtx, cancelUpd := context.WithTimeout(ctx, 90*time.Second)
		defer cancelUpd()
		resp, err := ordersClient.UpdateReserve(updCtx, future.GetID(), "", &examplev1.ReserveRequest{Quantity: 10})
		resultCh <- reserveResult{resp, err}
	}()
	time.Sleep(2 * time.Second)

	// Back-to-back: the first triggers the rollover, the second may land on
	// either side of it. Neither ordering may produce a double-count.
	if err := ordersClient.SendSignalRestock(ctx, future.GetID(), "", &examplev1.RestockRequest{Quantity: 5}); err != nil {
		t.Fatalf("could not send first restock: %v", err)
	}
	if err := ordersClient.SendSignalRestock(ctx, future.GetID(), "", &examplev1.RestockRequest{Quantity: 7}); err != nil {
		t.Fatalf("could not send second restock: %v", err)
	}

	r := <-resultCh

	// Let the chain settle, then read the successor's state.
	inv := ordersClient.GetTrackInventory(ctx, future.GetID(), "")
	var st *examplev1.GetStockResponse
	deadline := time.Now().Add(15 * time.Second)
	for {
		var qErr error
		st, qErr = inv.QueryGetStock(ctx, &emptypb.Empty{})
		if qErr == nil && st.Generation >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("successor generation never came up (last: %v, err=%v)", st, qErr)
		}
		time.Sleep(200 * time.Millisecond)
	}

	if r.err != nil {
		t.Logf("parked update was aborted by the rollover (the anti-pattern's documented failure): %v", r.err)
		t.Logf("successor state: stock %d at generation %d -- the reservation was never applied", st.Stock, st.Generation)
	} else {
		// If the update somehow succeeded, the successor's stock MUST
		// reflect the decrement -- anything else is the double-count.
		t.Logf("update succeeded with remaining=%d; successor stock=%d", r.resp.RemainingStock, st.Stock)
		if st.Stock != r.resp.RemainingStock {
			t.Fatalf("DOUBLE-COUNT: caller told remaining=%d but successor starts with %d", r.resp.RemainingStock, st.Stock)
		}
	}

	cancelInventory(ctx, t, inv)
}

// TestReserveTimeoutFailsCleanly: the caller-supplied reservation timeout,
// against a real server. Reserve(10, timeout 2s) on an empty inventory: the
// handler's AwaitWithTimeout expires (a durable workflow timer, not a client
// deadline), the caller receives the handler's error as the update outcome,
// and the workflow is untouched -- still running, stock unchanged, no trace
// of the reservation.
func TestReserveTimeoutFailsCleanly(t *testing.T) {
	ctx := context.Background()

	future, err := ordersClient.ExecuteWorkflowTrackInventory(ctx, &examplev1.TrackInventoryRequest{
		Sku:          "die-d8",
		InitialStock: 4,
		// no rollovers: this test is about the handler timeout alone
	})
	if err != nil {
		t.Fatalf("could not start workflow: %v", err)
	}
	inv := ordersClient.GetTrackInventory(ctx, future.GetID(), "")

	start := time.Now()
	_, err = ordersClient.UpdateReserve(ctx, future.GetID(), "", &examplev1.ReserveRequest{
		Quantity:       10,
		TimeoutSeconds: 2,
	})
	took := time.Since(start)
	if err == nil {
		t.Fatal("expected the reservation to fail on timeout")
	}
	t.Logf("reservation failed after %s with the handler's own error: %v", took.Round(time.Millisecond), err)
	if took > 15*time.Second {
		t.Errorf("timeout took %s, expected roughly the requested 2s", took)
	}

	// The failed reservation left no trace: same stock, workflow healthy.
	st, err := inv.QueryGetStock(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("could not query after the timed-out update: %v", err)
	}
	if st.Stock != 4 {
		t.Errorf("stock = %d, want 4 untouched", st.Stock)
	}
	t.Logf("workflow unaffected: stock still %d, and a fresh reservation within means still works", st.Stock)

	// Sanity: a satisfiable bounded reservation still fills.
	resp, err := ordersClient.UpdateReserve(ctx, future.GetID(), "", &examplev1.ReserveRequest{
		Quantity:       3,
		TimeoutSeconds: 30,
	})
	if err != nil {
		t.Fatalf("satisfiable bounded reservation failed: %v", err)
	}
	if resp.RemainingStock != 1 {
		t.Errorf("remaining stock = %d, want 1", resp.RemainingStock)
	}

	cancelInventory(ctx, t, inv)
}
