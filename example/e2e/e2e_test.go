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
