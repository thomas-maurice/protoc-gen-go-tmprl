// Tests the real Service implementation against the Temporal in-memory
// test suite. This pins the "parked order" scenario from the walkthrough: an
// order sits in a multi-hour pickup window, gets re-routed twice by updates
// while waiting, then the timer fires and it ships to the last address set.
// The test environment's mock clock auto-advances past timers, so hours of
// workflow time run in about a second of wall time.
package orders

import (
	"errors"
	"testing"
	"time"

	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// updateCallback satisfies the (internal) UpdateCallbacks interface expected
// by TestWorkflowEnvironment.UpdateWorkflow.
type updateCallback struct {
	accepted  bool
	rejectErr error
	completed bool
	result    interface{}
	err       error
}

func (c *updateCallback) Accept()          { c.accepted = true }
func (c *updateCallback) Reject(err error) { c.rejectErr = err }
func (c *updateCallback) Complete(success interface{}, err error) {
	c.completed = true
	c.result = success
	c.err = err
}

// newTestEnv registers the real workflows and activities of the service under
// their generated registered names, exactly like the generated worker does.
func newTestEnv(t *testing.T) *testsuite.TestWorkflowEnvironment {
	t.Helper()

	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	// The generated client only needs a temporal client for client-side calls;
	// workflow-side activity and child workflow execution goes through the
	// workflow package, so nil is fine here.
	orders, err := examplev1.NewOrdersClient(nil)
	if err != nil {
		t.Fatalf("could not create orders client: %v", err)
	}
	svc := New(orders)

	env.RegisterWorkflowWithOptions(svc.ProcessOrder, workflow.RegisterOptions{Name: examplev1.WorkflowProcessOrderName})
	env.RegisterWorkflowWithOptions(svc.ShipOrder, workflow.RegisterOptions{Name: examplev1.WorkflowShipOrderName})
	env.RegisterWorkflowWithOptions(svc.TrackInventory, workflow.RegisterOptions{Name: examplev1.WorkflowTrackInventoryName})
	env.RegisterActivityWithOptions(svc.ChargePayment, activity.RegisterOptions{Name: examplev1.ActivityChargePaymentName})
	env.RegisterActivityWithOptions(svc.PackItems, activity.RegisterOptions{Name: examplev1.ActivityPackItemsName})
	env.RegisterActivityWithOptions(svc.DispatchCourier, activity.RegisterOptions{Name: examplev1.ActivityDispatchCourierName})

	return env
}

// TestProcessOrder_ParkedOrderUpdatedThenShips: an order parked in a two hour
// pickup window accepts address updates while waiting (each returning the
// address it replaced), then ships to the LAST address set when the pickup
// timer fires. This is the property that makes updates useful on long-lived
// workflows; the mock clock means the two simulated hours cost no wall time.
func TestProcessOrder_ParkedOrderUpdatedThenShips(t *testing.T) {
	env := newTestEnv(t)
	env.SetTestTimeout(30 * time.Second)

	firstCb := &updateCallback{}
	secondCb := &updateCallback{}

	// First re-route 30 simulated minutes in, second at the one hour mark.
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(examplev1.UpdateChangeShippingAddressName, "reroute-1", firstCb,
			&examplev1.ChangeShippingAddressRequest{Address: "99 boulevard de Test, Lille"})
	}, 30*time.Minute)
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(examplev1.UpdateChangeShippingAddressName, "reroute-2", secondCb,
			&examplev1.ChangeShippingAddressRequest{Address: "221B Baker Street, London"})
	}, time.Hour)

	env.ExecuteWorkflow(examplev1.WorkflowProcessOrderName, &examplev1.ProcessOrderRequest{
		OrderId:             "order-parked",
		Items:               []*examplev1.OrderItem{{Sku: "die-d20", Quantity: 1}},
		AmountCents:         4200,
		CardToken:           CardOK,
		ShippingAddress:     "1 rue de la Paix, Paris",
		PickupWindowSeconds: 7200,
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow failed: %v", err)
	}

	// Both updates completed and each returned the address it replaced.
	if !firstCb.completed || firstCb.err != nil {
		t.Fatalf("first update did not complete cleanly: completed=%v err=%v", firstCb.completed, firstCb.err)
	}
	if resp := firstCb.result.(*examplev1.ChangeShippingAddressResponse); resp.PreviousAddress != "1 rue de la Paix, Paris" {
		t.Errorf("first update previous address = %q, want the original one", resp.PreviousAddress)
	}
	if !secondCb.completed || secondCb.err != nil {
		t.Fatalf("second update did not complete cleanly: completed=%v err=%v", secondCb.completed, secondCb.err)
	}
	if resp := secondCb.result.(*examplev1.ChangeShippingAddressResponse); resp.PreviousAddress != "99 boulevard de Test, Lille" {
		t.Errorf("second update previous address = %q, want the first re-route", resp.PreviousAddress)
	}

	// The order shipped, to the last address set, with a tracking number from
	// the ShipOrder child workflow.
	var result examplev1.ProcessOrderResponse
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("could not get workflow result: %v", err)
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
}

// TestProcessOrder_CancelledDuringPickupWindow: the same parked order, but a
// CancelOrder signal lands mid-window; the workflow must end CANCELLED with
// the reason, well before the pickup timer would have fired.
func TestProcessOrder_CancelledDuringPickupWindow(t *testing.T) {
	env := newTestEnv(t)
	env.SetTestTimeout(30 * time.Second)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(examplev1.SignalCancelOrderName,
			&examplev1.CancelOrderRequest{Reason: "changed my mind"})
	}, 45*time.Minute)

	env.ExecuteWorkflow(examplev1.WorkflowProcessOrderName, &examplev1.ProcessOrderRequest{
		OrderId:             "order-parked-cancel",
		Items:               []*examplev1.OrderItem{{Sku: "die-d6", Quantity: 1}},
		AmountCents:         4200,
		CardToken:           CardOK,
		ShippingAddress:     "1 rue de la Paix, Paris",
		PickupWindowSeconds: 7200,
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow failed: %v", err)
	}

	var result examplev1.ProcessOrderResponse
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("could not get workflow result: %v", err)
	}
	if result.Status != examplev1.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Errorf("status = %v, want CANCELLED", result.Status)
	}
	if result.CancelReason != "changed my mind" {
		t.Errorf("cancel reason = %q, want the signalled one", result.CancelReason)
	}
	if result.TrackingNumber != "" {
		t.Errorf("cancelled order must not have a tracking number, got %q", result.TrackingNumber)
	}
}

// TestTrackInventory_DrainsBlockedUpdateBeforeContinueAsNew: a Reserve update
// parked on "not enough stock" must be answered BEFORE the workflow rolls
// over with continue-as-new. The workflow drains its handlers
// (workflow.AllHandlersFinished) after the last restock of the run, so the
// order of events is: restock arrives -> parked handler fills the
// reservation -> only then does the run end with a ContinueAsNewError
// carrying the remaining stock to the next generation.
func TestTrackInventory_DrainsBlockedUpdateBeforeContinueAsNew(t *testing.T) {
	env := newTestEnv(t)
	env.SetTestTimeout(30 * time.Second)

	cb := &updateCallback{}
	// The reservation arrives while the stock is still 0: the handler parks.
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(examplev1.UpdateReserveName, "reserve-1", cb,
			&examplev1.ReserveRequest{Quantity: 3})
	}, time.Minute)
	// The restock both satisfies the parked reservation and triggers the
	// rollover (restocks_before_continue_as_new = 1).
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(examplev1.SignalRestockName, &examplev1.RestockRequest{Quantity: 5})
	}, 2*time.Minute)

	env.ExecuteWorkflow(examplev1.WorkflowTrackInventoryName, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d20",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}

	// The parked update was answered by this run, before the rollover.
	if !cb.completed || cb.err != nil {
		t.Fatalf("blocked update did not complete cleanly: completed=%v err=%v", cb.completed, cb.err)
	}
	resp, ok := cb.result.(*examplev1.ReserveResponse)
	if !ok {
		t.Fatalf("update result has wrong type: %T", cb.result)
	}
	if resp.RemainingStock != 2 {
		t.Errorf("remaining stock = %d, want 2 (5 restocked - 3 reserved)", resp.RemainingStock)
	}
	if resp.Generation != 1 {
		t.Errorf("generation = %d, want 1 (answered by the first run)", resp.Generation)
	}

	// The run ended in a continue-as-new carrying the post-reservation stock.
	wfErr := env.GetWorkflowError()
	if wfErr == nil {
		t.Fatal("expected the run to end with a ContinueAsNewError")
	}
	var canErr *workflow.ContinueAsNewError
	if !errors.As(wfErr, &canErr) {
		t.Fatalf("expected ContinueAsNewError, got %T: %v", wfErr, wfErr)
	}
}

// TestTrackInventory_DrainKeepsServicingRestocks: liveness regression test.
// A Reserve(10) is still short when the rollover-triggering restock (5)
// lands, so the drain begins with the handler parked. A LATER restock (7)
// arriving DURING the drain must still be absorbed so the reservation can
// complete -- a drain that stops consuming signals would deadlock this run
// forever (handler waits for stock, stock waits for the signal loop, signal
// loop already exited).
func TestTrackInventory_DrainKeepsServicingRestocks(t *testing.T) {
	env := newTestEnv(t)
	env.SetTestTimeout(30 * time.Second)

	cb := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(examplev1.UpdateReserveName, "reserve-big", cb,
			&examplev1.ReserveRequest{Quantity: 10})
	}, time.Minute)
	// Triggers the rollover, but 5 < 10: the handler stays parked and the
	// drain starts.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(examplev1.SignalRestockName, &examplev1.RestockRequest{Quantity: 5})
	}, 2*time.Minute)
	// Arrives mid-drain; must be absorbed for the reservation to complete.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(examplev1.SignalRestockName, &examplev1.RestockRequest{Quantity: 7})
	}, 3*time.Minute)

	env.ExecuteWorkflow(examplev1.WorkflowTrackInventoryName, &examplev1.TrackInventoryRequest{
		Sku:                         "die-d20",
		InitialStock:                0,
		RestocksBeforeContinueAsNew: 1,
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}

	if !cb.completed || cb.err != nil {
		t.Fatalf("parked update did not complete cleanly: completed=%v err=%v", cb.completed, cb.err)
	}
	resp, ok := cb.result.(*examplev1.ReserveResponse)
	if !ok {
		t.Fatalf("update result has wrong type: %T", cb.result)
	}
	if resp.RemainingStock != 2 {
		t.Errorf("remaining stock = %d, want 2 (5+7 restocked - 10 reserved)", resp.RemainingStock)
	}

	// Still ends in a rollover once the handler is done.
	wfErr := env.GetWorkflowError()
	var canErr *workflow.ContinueAsNewError
	if wfErr == nil || !errors.As(wfErr, &canErr) {
		t.Fatalf("expected ContinueAsNewError, got %v", wfErr)
	}
}
