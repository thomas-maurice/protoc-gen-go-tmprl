// Tests for the generated update helpers. These exercise the emitted
// HandleUpdateChangeShippingAddress / HandleUpdateChangeShippingAddressWithValidator
// functions against the Temporal in-memory test suite to prove the full
// round-trip: an update handler registered by generated code receives the
// typed request, may block on workflow state, and returns the typed
// response to the caller; a validator can reject an update before it is
// admitted. If the generator changes the shape of these helpers, these
// tests must be kept in sync.
package examplev1

import (
	"errors"
	"testing"
	"time"

	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// errEmptyAddress is returned by the validator under test to reject empty
// shipping addresses.
var errEmptyAddress = errors.New("address cannot be empty")

// updateCallback records the callbacks the test environment fires for an
// update. It satisfies the (internal) UpdateCallbacks interface expected by
// TestWorkflowEnvironment.UpdateWorkflow.
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

// TestHandleUpdateChangeShippingAddress_BlockingRoundTrip: verifies that an
// update handler registered through the generated
// HandleUpdateChangeShippingAddress can block on workflow state (here, a bool
// flipped by the CancelOrder signal) and that the caller receives the typed
// response once the handler unblocks. This is the synchronous
// request/response property updates exist for.
func TestHandleUpdateChangeShippingAddress_BlockingRoundTrip(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	wf := func(ctx workflow.Context, req *ProcessOrderRequest) (*ProcessOrderResponse, error) {
		address := req.ShippingAddress
		unblocked := false
		handled := false

		err := HandleUpdateChangeShippingAddress(ctx, func(ctx workflow.Context, u *ChangeShippingAddressRequest) (*ChangeShippingAddressResponse, error) {
			// Block until the CancelOrder signal arrives, proving update
			// handlers can wait on workflow state before responding.
			if err := workflow.Await(ctx, func() bool { return unblocked }); err != nil {
				return nil, err
			}
			previous := address
			address = u.Address
			handled = true
			return &ChangeShippingAddressResponse{PreviousAddress: previous}, nil
		})
		if err != nil {
			return nil, err
		}

		if _, ok := ReceiveSignalCancelOrder(ctx); !ok {
			return nil, workflow.ErrCanceled
		}
		unblocked = true

		// Don't complete the workflow before the update handler finished.
		if err := workflow.Await(ctx, func() bool { return handled }); err != nil {
			return nil, err
		}
		return &ProcessOrderResponse{ShippingAddress: address}, nil
	}
	env.RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{Name: WorkflowProcessOrderName})

	cb := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateChangeShippingAddressName, "update-1", cb, &ChangeShippingAddressRequest{Address: "new address"})
	}, 0)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalCancelOrderName, &CancelOrderRequest{Reason: "unblock"})
	}, time.Millisecond)

	env.ExecuteWorkflow(WorkflowProcessOrderName, &ProcessOrderRequest{ShippingAddress: "old address"})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow failed: %v", err)
	}
	if !cb.accepted {
		t.Error("update was not accepted")
	}
	if !cb.completed {
		t.Fatal("update did not complete")
	}
	if cb.err != nil {
		t.Fatalf("update completed with error: %v", cb.err)
	}
	resp, ok := cb.result.(*ChangeShippingAddressResponse)
	if !ok {
		t.Fatalf("update result has wrong type: %T", cb.result)
	}
	if resp.PreviousAddress != "old address" {
		t.Errorf("expected previous address %q, got %q", "old address", resp.PreviousAddress)
	}
}

// TestHandleUpdateChangeShippingAddressWithValidator: verifies that a
// validator registered through the generated
// HandleUpdateChangeShippingAddressWithValidator rejects invalid updates
// before they reach the handler (no state change), while valid updates still
// complete with the typed response.
func TestHandleUpdateChangeShippingAddressWithValidator(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	handlerCalls := 0
	finalAddress := ""

	wf := func(ctx workflow.Context, req *ProcessOrderRequest) (*ProcessOrderResponse, error) {
		address := req.ShippingAddress

		err := HandleUpdateChangeShippingAddressWithValidator(ctx,
			func(ctx workflow.Context, u *ChangeShippingAddressRequest) (*ChangeShippingAddressResponse, error) {
				handlerCalls++
				previous := address
				address = u.Address
				return &ChangeShippingAddressResponse{PreviousAddress: previous}, nil
			},
			func(ctx workflow.Context, u *ChangeShippingAddressRequest) error {
				if u.Address == "" {
					return errEmptyAddress
				}
				return nil
			},
		)
		if err != nil {
			return nil, err
		}

		if _, ok := ReceiveSignalCancelOrder(ctx); !ok {
			return nil, workflow.ErrCanceled
		}
		finalAddress = address
		return &ProcessOrderResponse{ShippingAddress: address}, nil
	}
	env.RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{Name: WorkflowProcessOrderName})

	rejectedCb := &updateCallback{}
	acceptedCb := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateChangeShippingAddressName, "update-invalid", rejectedCb, &ChangeShippingAddressRequest{Address: ""})
	}, 0)
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateChangeShippingAddressName, "update-valid", acceptedCb, &ChangeShippingAddressRequest{Address: "new address"})
	}, time.Millisecond)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalCancelOrderName, &CancelOrderRequest{Reason: "unblock"})
	}, 2*time.Millisecond)

	env.ExecuteWorkflow(WorkflowProcessOrderName, &ProcessOrderRequest{ShippingAddress: "old address"})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow failed: %v", err)
	}

	if rejectedCb.rejectErr == nil {
		t.Error("expected the invalid update to be rejected by the validator")
	}
	if rejectedCb.completed && rejectedCb.err == nil {
		t.Error("invalid update completed successfully, expected rejection")
	}

	if !acceptedCb.completed {
		t.Fatal("valid update did not complete")
	}
	if acceptedCb.err != nil {
		t.Fatalf("valid update completed with error: %v", acceptedCb.err)
	}
	resp, ok := acceptedCb.result.(*ChangeShippingAddressResponse)
	if !ok {
		t.Fatalf("update result has wrong type: %T", acceptedCb.result)
	}
	if resp.PreviousAddress != "old address" {
		t.Errorf("expected previous address %q, got %q", "old address", resp.PreviousAddress)
	}

	if handlerCalls != 1 {
		t.Errorf("expected handler to run exactly once, ran %d times", handlerCalls)
	}
	if finalAddress != "new address" {
		t.Errorf("expected final address %q (valid update applied, invalid one not), got %q", "new address", finalAddress)
	}
}
