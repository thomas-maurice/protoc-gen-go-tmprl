// A guided tour of the code generated for the example.v1.Orders service.
//
// Start the temporal dev stack (docker compose up -d) and the worker
// (go run ./example/worker), then run this:
//
//	go run ./example/client
//
// Each step is narrated in the logs and maps to a feature of the plugin:
//
//	step 1: workflows, queries and updates on the happy path
//	step 2: rejected updates (validation)
//	step 3: cancelling a workflow with a signal
//	step 4: transient activity failures and retry policies
//	step 5: non-retryable failures (how errors surface to the caller)
//	step 6: schedules
//	step 7: blocking updates and continue-as-new (the entity pattern)
//
// The workflow executions are also visible in the temporal UI, by default on
// http://localhost:8080
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/charmbracelet/log"
	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/types/known/emptypb"
)

// step prints a visible banner so the walkthrough is easy to follow in the
// terminal.
func step(n int, title string) {
	fmt.Println()
	fmt.Printf("---[ step %d: %s ]---\n", n, title)
	fmt.Println()
}

// someItems is what every demo order contains.
func someItems() []*examplev1.OrderItem {
	return []*examplev1.OrderItem{
		{Sku: "die-d6", Quantity: 2},
		{Sku: "die-d20", Quantity: 1},
		{Sku: "dice-tray", Quantity: 1},
	}
}

func main() {
	logger := slog.New(
		log.NewWithOptions(os.Stderr, log.Options{
			Level:           log.InfoLevel,
			ReportTimestamp: true,
			TimeFormat:      time.RFC3339,
			Formatter:       log.TextFormatter,
		}),
	)

	// TEMPORAL_ADDRESS overrides the target server, defaults to localhost:7233.
	c, err := client.NewLazyClient(client.Options{
		HostPort: os.Getenv("TEMPORAL_ADDRESS"),
		Logger:   tlog.NewStructuredLogger(logger),
	})
	if err != nil {
		logger.Error("could not create temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// The generated, typed client: every workflow/activity/signal/query/update
	// of the service is a method on it. Passing no task queue uses the default
	// one declared in the proto ("orders").
	orders, err := examplev1.NewOrdersClient(c)
	if err != nil {
		logger.Error("could not create orders client", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// -----------------------------------------------------------------------
	step(1, "the happy path: start a workflow, query it, update it")
	// -----------------------------------------------------------------------

	// Start the workflow asynchronously: we get a future back immediately.
	// With gen-workflow-prefix=true the generator picks a readable workflow ID
	// (example.v1.Orders.ProcessOrder/<uuid>) for us.
	future, err := orders.ExecuteWorkflowProcessOrder(ctx, &examplev1.ProcessOrderRequest{
		OrderId:         "order-happy",
		Items:           someItems(),
		AmountCents:     4200,
		CardToken:       "tok-ok",
		ShippingAddress: "1 rue de la Paix, Paris",
	})
	if err != nil {
		logger.Error("could not start workflow", "error", err)
		os.Exit(1)
	}
	logger.Info("workflow started", "workflow_id", future.GetID(), "run_id", future.GetRunID())

	// Wrap the run in the generated workflow object: it carries typed
	// Query/Update/Signal/Result methods for this specific execution.
	order := orders.GetProcessOrderFromRun(future)

	// Watch the order move through its lifecycle with the GetOrderStatus
	// query. Queries are read only and cheap; poll away.
	watchCtx, stopWatching := context.WithCancel(ctx)
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		last := ""
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
				st, err := order.QueryGetOrderStatus(watchCtx, &emptypb.Empty{})
				if err != nil {
					continue
				}
				if st.Status.String() != last {
					last = st.Status.String()
					logger.Info("query: order status changed", "status", st.Status, "ships_to", st.ShippingAddress)
				}
			}
		}
	}()

	// Give the order a moment to get paid and packed...
	time.Sleep(2 * time.Second)

	// ...then re-route it mid-flight with an update. This call BLOCKS until
	// the workflow's update handler ran and returns its typed response:
	// request/response semantics against a running workflow.
	moved, err := order.UpdateChangeShippingAddress(ctx, &examplev1.ChangeShippingAddressRequest{
		Address: "42 quai des Orfevres, Paris",
	})
	if err != nil {
		logger.Error("could not update the shipping address", "error", err)
		os.Exit(1)
	}
	logger.Info("update: shipping address changed", "previous_address", moved.PreviousAddress)

	// -----------------------------------------------------------------------
	step(2, "a rejected update: validators run before anything is recorded")
	// -----------------------------------------------------------------------

	// An empty address does not pass the validator the worker registered.
	// The update is rejected before it ever reaches the workflow history and
	// the error comes straight back to us. This time we use the client-level
	// helper (addressing the execution by ID) instead of the workflow object.
	_, err = orders.UpdateChangeShippingAddress(ctx, future.GetID(), future.GetRunID(), &examplev1.ChangeShippingAddressRequest{
		Address: "",
	})
	if err != nil {
		logger.Info("update rejected as expected", "error", err)
	} else {
		logger.Error("the empty address should have been rejected")
		os.Exit(1)
	}

	// Wait for the order to ship and collect the typed result.
	result, err := order.Result(ctx)
	if err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}
	stopWatching()
	<-watcherDone
	logger.Info("workflow finished",
		"status", result.Status,
		"shipped_to", result.ShippingAddress,
		"tracking_number", result.TrackingNumber,
	)

	// Queries still work on completed workflows, as long as a worker runs.
	st, err := order.QueryGetOrderStatus(ctx, &emptypb.Empty{})
	if err != nil {
		logger.Error("could not query the finished workflow", "error", err)
		os.Exit(1)
	}
	logger.Info("query still answers after completion", "status", st.Status)

	// -----------------------------------------------------------------------
	step(3, "cancelling an order with a signal")
	// -----------------------------------------------------------------------

	future, err = orders.ExecuteWorkflowProcessOrder(ctx, &examplev1.ProcessOrderRequest{
		OrderId:         "order-cancelled",
		Items:           someItems(),
		AmountCents:     4200,
		CardToken:       "tok-ok",
		ShippingAddress: "1 rue de la Paix, Paris",
	})
	if err != nil {
		logger.Error("could not start workflow", "error", err)
		os.Exit(1)
	}
	logger.Info("workflow started", "workflow_id", future.GetID())

	// Let it get paid, then change our mind. Signals are fire and forget:
	// this returns as soon as the signal is recorded, there is no response.
	time.Sleep(1500 * time.Millisecond)
	err = orders.SendSignalCancelOrder(ctx, future.GetID(), future.GetRunID(), &examplev1.CancelOrderRequest{
		Reason: "changed my mind",
	})
	if err != nil {
		logger.Error("could not send signal", "error", err)
		os.Exit(1)
	}
	logger.Info("signal: cancellation sent")

	result, err = orders.GetProcessOrderFromRun(future).Result(ctx)
	if err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}
	logger.Info("workflow finished", "status", result.Status, "cancel_reason", result.CancelReason)

	// -----------------------------------------------------------------------
	step(4, "transient failures: the retry policy does its job")
	// -----------------------------------------------------------------------

	// The "tok-flaky" card makes the ChargePayment activity fail twice before
	// succeeding. The retry policy declared in the proto (1s initial interval,
	// backoff 2.0) retries it transparently: from here it just looks slow.
	// Watch the worker logs to see the two failed attempts.
	logger.Info("ordering with a flaky payment provider, watch the worker logs...")
	start := time.Now()
	result, err = orders.ExecuteWorkflowProcessOrderSync(ctx, &examplev1.ProcessOrderRequest{
		OrderId:         "order-flaky",
		Items:           someItems(),
		AmountCents:     4200,
		CardToken:       "tok-flaky",
		ShippingAddress: "1 rue de la Paix, Paris",
	})
	if err != nil {
		logger.Error("workflow failed", "error", err)
		os.Exit(1)
	}
	logger.Info("workflow survived the flaky payment provider",
		"status", result.Status,
		"took", time.Since(start).Round(time.Second),
	)

	// -----------------------------------------------------------------------
	step(5, "non-retryable failures: how errors reach the caller")
	// -----------------------------------------------------------------------

	// "tok-declined" makes ChargePayment return an application error of type
	// "CardDeclined", which the retry policy lists as non retryable: no
	// retries, the workflow fails immediately, and the typed error is
	// available to the caller through errors.As.
	_, err = orders.ExecuteWorkflowProcessOrderSync(ctx, &examplev1.ProcessOrderRequest{
		OrderId:         "order-declined",
		Items:           someItems(),
		AmountCents:     999999,
		CardToken:       "tok-declined",
		ShippingAddress: "1 rue de la Paix, Paris",
	})
	if err == nil {
		logger.Error("the declined card should have failed the workflow")
		os.Exit(1)
	}
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		logger.Info("workflow failed as expected",
			"error_type", appErr.Type(),
			"message", appErr.Message(),
		)
	} else {
		logger.Error("unexpected error shape", "error", err)
		os.Exit(1)
	}

	// -----------------------------------------------------------------------
	step(6, "schedules: run a workflow on a cadence")
	// -----------------------------------------------------------------------

	// Every workflow gets schedule helpers, no proto annotation needed. The
	// caller owns the Spec (interval, cron, calendar...); the generator fills
	// in the workflow name, arguments and default task queue.
	const scheduleID = "daily-sales-report"
	_, err = orders.UpsertScheduleDailySalesReport(ctx, scheduleID, &emptypb.Empty{}, client.ScheduleOptions{
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{{Every: time.Minute}},
		},
	})
	if err != nil {
		logger.Error("could not upsert schedule", "error", err)
		os.Exit(1)
	}
	logger.Info("schedule upserted", "schedule_id", scheduleID, "every", "1m")

	// Listing goes through the visibility store, which is eventually
	// consistent: a schedule created a millisecond ago may not show up yet,
	// so poll briefly instead of trusting a single call.
	var schedules []client.ScheduleListEntry
	for deadline := time.Now().Add(10 * time.Second); ; {
		schedules, err = orders.ListScheduleDailySalesReport(ctx, 10)
		if err != nil {
			logger.Error("could not list schedules", "error", err)
			os.Exit(1)
		}
		if len(schedules) > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	logger.Info("schedules targeting DailySalesReport", "count", len(schedules))

	// Pause and resume are idempotent: they describe first and no-op if the
	// schedule is already in the desired state.
	if err := orders.PauseScheduleDailySalesReport(ctx, scheduleID, "walkthrough pause"); err != nil {
		logger.Error("could not pause schedule", "error", err)
		os.Exit(1)
	}
	logger.Info("schedule paused")
	if err := orders.UnpauseScheduleDailySalesReport(ctx, scheduleID, "walkthrough unpause"); err != nil {
		logger.Error("could not unpause schedule", "error", err)
		os.Exit(1)
	}
	logger.Info("schedule unpaused")

	// Clean up so re-running the walkthrough starts fresh. Comment this out
	// if you want to watch the report fire every minute in the UI.
	if err := orders.DeleteScheduleDailySalesReport(ctx, scheduleID); err != nil {
		logger.Error("could not delete schedule", "error", err)
		os.Exit(1)
	}
	logger.Info("schedule deleted")

	// -----------------------------------------------------------------------
	step(7, "blocking updates and continue-as-new: the entity pattern")
	// -----------------------------------------------------------------------

	// TrackInventory is a long-lived entity workflow: it absorbs Restock
	// signals and rolls over with continue-as-new after each one. Reserve is
	// a BLOCKING update: with an empty inventory the handler parks inside the
	// workflow until stock arrives, and so does our call.
	invFuture, err := orders.ExecuteWorkflowTrackInventory(ctx, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d20",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
	})
	if err != nil {
		logger.Error("could not start inventory workflow", "error", err)
		os.Exit(1)
	}
	logger.Info("inventory workflow started with zero stock", "workflow_id", invFuture.GetID())

	type reserveOutcome struct {
		resp *examplev1.ReserveResponse
		err  error
	}
	reserveCh := make(chan reserveOutcome, 1)
	go func() {
		resp, err := orders.UpdateReserve(ctx, invFuture.GetID(), "", &examplev1.ReserveRequest{Quantity: 3})
		reserveCh <- reserveOutcome{resp, err}
	}()
	logger.Info("sent Reserve(3): the update handler is now parked on workflow.Await(stock >= 3), and so are we")

	time.Sleep(2 * time.Second)
	select {
	case <-reserveCh:
		logger.Error("the reservation should still be blocked")
		os.Exit(1)
	default:
		logger.Info("still blocked, as expected: no stock yet")
	}

	// This restock satisfies the parked reservation AND triggers the
	// continue-as-new rollover. The workflow drains its update handlers
	// before rolling over, so the parked caller is answered first.
	if err := orders.SendSignalRestock(ctx, invFuture.GetID(), "", &examplev1.RestockRequest{Quantity: 5}); err != nil {
		logger.Error("could not send restock signal", "error", err)
		os.Exit(1)
	}
	logger.Info("sent Restock(5)")

	outcome := <-reserveCh
	if outcome.err != nil {
		logger.Error("reservation failed", "error", outcome.err)
		os.Exit(1)
	}
	logger.Info("blocked update answered",
		"remaining_stock", outcome.resp.RemainingStock,
		"answered_by_generation", outcome.resp.Generation,
	)

	// The workflow has since rolled over: same workflow ID, brand-new run,
	// carrying only what the old run passed along (the remaining stock).
	inventory := orders.GetTrackInventory(ctx, invFuture.GetID(), "")
	for {
		st, err := inventory.QueryGetStock(ctx, &emptypb.Empty{})
		if err == nil && st.Generation >= 2 {
			logger.Info("continue-as-new rolled the entity over",
				"generation", st.Generation,
				"carried_stock", st.Stock,
			)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := inventory.Cancel(ctx); err != nil {
		logger.Error("could not clean up the inventory workflow", "error", err)
		os.Exit(1)
	}
	logger.Info("inventory workflow cleaned up; see both runs of it in the UI, the update completes in the FIRST one")

	fmt.Println()
	logger.Info("walkthrough complete, check the executions in the temporal UI", "url", "http://localhost:8080")
}
