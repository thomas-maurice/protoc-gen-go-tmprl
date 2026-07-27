// Package orders implements the example.v1.Orders service: every workflow and
// activity of the example walkthrough. It lives in its own package (rather
// than in the worker binary) so both the worker's main and the tests -- unit
// and e2e -- can register the exact same implementation.
package orders

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Card tokens the ChargePayment activity understands. The client uses them to
// trigger the different failure scenarios of the walkthrough.
const (
	CardOK       = "tok-ok"       // works first try
	CardFlaky    = "tok-flaky"    // fails twice, then works (retry policy demo)
	CardDeclined = "tok-declined" // fails the workflow with a non-retryable error
)

// Service implements the generated examplev1.OrdersService interface: three
// workflows (ProcessOrder, ShipOrder, DailySalesReport) and three activities
// (ChargePayment, PackItems, DispatchCourier). The embedded generated client
// is used to invoke activities and child workflows from within workflow code
// with their configured options applied.
type Service struct {
	orders *examplev1.OrdersClient
}

// New returns a Service wired to the given generated client.
func New(orders *examplev1.OrdersClient) *Service {
	return &Service{orders: orders}
}

// ---------------------------------------------------------------------------
// Activities
// ---------------------------------------------------------------------------

// ChargePayment captures the money for an order. Its retry policy (declared
// in the proto) retries transient failures with exponential backoff but
// gives up immediately on a CardDeclined application error, because that
// error type is listed in non_retryable_error_types.
func (s *Service) ChargePayment(ctx context.Context, req *examplev1.ChargePaymentRequest) (*examplev1.ChargePaymentResponse, error) {
	info := activity.GetInfo(ctx)
	logger := activity.GetLogger(ctx)

	switch req.CardToken {
	case CardDeclined:
		// A non-retryable failure: the error type "CardDeclined" matches the
		// retry policy's non_retryable_error_types, so Temporal will NOT
		// retry this activity and the workflow fails immediately.
		logger.Error("card declined, giving up without retries", "order_id", req.OrderId)
		return nil, temporal.NewApplicationError("card declined by issuer", "CardDeclined")
	case CardFlaky:
		// A transient failure: fail the first two attempts so the retry
		// policy (initial_interval 1s, backoff 2.0) is visible in the logs
		// and in the UI.
		if info.Attempt < 3 {
			logger.Warn("payment provider timed out, temporal will retry",
				"order_id", req.OrderId, "attempt", info.Attempt)
			return nil, fmt.Errorf("payment provider timeout (attempt %d)", info.Attempt)
		}
		logger.Info("payment went through after retries", "order_id", req.OrderId, "attempt", info.Attempt)
	}

	return &examplev1.ChargePaymentResponse{
		TransactionId: fmt.Sprintf("txn-%s-%d", req.OrderId, time.Now().UnixMilli()),
	}, nil
}

// PackItems packs one parcel per order line and records a heartbeat after
// each one. If the worker crashed mid-pack, Temporal would notice within the
// configured heartbeat_timeout and reschedule the activity elsewhere.
func (s *Service) PackItems(ctx context.Context, req *examplev1.PackItemsRequest) (*examplev1.PackItemsResponse, error) {
	logger := activity.GetLogger(ctx)

	for i, item := range req.Items {
		time.Sleep(300 * time.Millisecond) // simulated hard work
		logger.Info("packed a parcel", "sku", item.Sku, "quantity", item.Quantity)
		activity.RecordHeartbeat(ctx, i+1)
	}

	return &examplev1.PackItemsResponse{
		Parcels: int32(len(req.Items)),
	}, nil
}

// DispatchCourier books a courier. It is registered under the custom name
// "courier.Dispatch" (see the `name` option in the proto).
func (s *Service) DispatchCourier(ctx context.Context, req *examplev1.DispatchCourierRequest) (*examplev1.DispatchCourierResponse, error) {
	return &examplev1.DispatchCourierResponse{
		TrackingNumber: fmt.Sprintf("TRK-%06d", rand.Intn(1000000)),
	}, nil
}

// ---------------------------------------------------------------------------
// Workflows
// ---------------------------------------------------------------------------

// ProcessOrder drives an order from payment to shipping:
//
//	PENDING -> (ChargePayment) -> PAID -> (PackItems) -> PACKED
//	        -> carrier pickup window -> (ShipOrder child workflow) -> SHIPPED
//
// While it runs it answers the GetOrderStatus query, accepts the
// ChangeShippingAddress update (validated) and can be stopped with the
// CancelOrder signal.
func (s *Service) ProcessOrder(ctx workflow.Context, req *examplev1.ProcessOrderRequest) (*examplev1.ProcessOrderResponse, error) {
	logger := workflow.GetLogger(ctx)

	status := examplev1.OrderStatus_ORDER_STATUS_PENDING
	address := req.ShippingAddress
	cancelled := false
	cancelReason := ""

	// Query handler: a read-only view of the workflow state. Queries never
	// mutate anything and are answered even after the workflow completed.
	err := examplev1.HandleQueryGetOrderStatus(ctx, func(_ *emptypb.Empty) (*examplev1.GetOrderStatusResponse, error) {
		return &examplev1.GetOrderStatusResponse{
			Status:          status,
			ShippingAddress: address,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	// Update handler: a synchronous mutation. The validator runs first; if it
	// errors, the update is rejected before it is ever written to history.
	// The handler itself can also fail (here: when it is too late to re-route)
	// which fails that update but NOT the workflow.
	err = examplev1.HandleUpdateChangeShippingAddressWithValidator(ctx,
		func(ctx workflow.Context, u *examplev1.ChangeShippingAddressRequest) (*examplev1.ChangeShippingAddressResponse, error) {
			if status == examplev1.OrderStatus_ORDER_STATUS_SHIPPED || status == examplev1.OrderStatus_ORDER_STATUS_CANCELLED {
				return nil, fmt.Errorf("order is already %s, too late to change the address", status)
			}
			previous := address
			address = u.Address
			logger.Info("shipping address changed", "from", previous, "to", address)
			return &examplev1.ChangeShippingAddressResponse{PreviousAddress: previous}, nil
		},
		func(ctx workflow.Context, u *examplev1.ChangeShippingAddressRequest) error {
			if u.Address == "" {
				return fmt.Errorf("address cannot be empty")
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Signal handler: listen for cancellations in the background. Signals are
	// fire and forget, the sender gets no response.
	workflow.Go(ctx, func(ctx workflow.Context) {
		sig, ok := examplev1.ReceiveSignalCancelOrder(ctx)
		if !ok {
			return
		}
		cancelled = true
		cancelReason = sig.Reason
		logger.Info("cancellation requested", "reason", cancelReason)
	})

	cancelledResponse := func() *examplev1.ProcessOrderResponse {
		status = examplev1.OrderStatus_ORDER_STATUS_CANCELLED
		return &examplev1.ProcessOrderResponse{
			Status:          status,
			ShippingAddress: address,
			CancelReason:    cancelReason,
		}
	}

	// Step 1: take the money. If the card is declined this returns a
	// non-retryable CardDeclined error and the whole workflow fails: that is
	// the "what happens when things fail" part of the walkthrough.
	charge, err := s.orders.ExecuteActivityChargePaymentSync(ctx, &examplev1.ChargePaymentRequest{
		OrderId:     req.OrderId,
		AmountCents: req.AmountCents,
		CardToken:   req.CardToken,
	})
	if err != nil {
		return nil, err
	}
	status = examplev1.OrderStatus_ORDER_STATUS_PAID
	logger.Info("payment captured", "transaction_id", charge.TransactionId)

	if cancelled {
		return cancelledResponse(), nil
	}

	// Step 2: pack the parcels (the activity heartbeats after each one).
	parcels, err := s.orders.ExecuteActivityPackItemsSync(ctx, &examplev1.PackItemsRequest{
		Items: req.Items,
	})
	if err != nil {
		return nil, err
	}
	status = examplev1.OrderStatus_ORDER_STATUS_PACKED
	logger.Info("order packed", "parcels", parcels.Parcels)

	// Step 3: wait for the carrier pickup. This window is what makes the
	// order cancellable and re-routable from the client: the workflow simply
	// blocks on its own state, which the signal handler may change.
	pickupWindow := 5 * time.Second
	if req.PickupWindowSeconds > 0 {
		pickupWindow = time.Duration(req.PickupWindowSeconds) * time.Second
	}
	if _, err := workflow.AwaitWithTimeout(ctx, pickupWindow, func() bool { return cancelled }); err != nil {
		return nil, err
	}
	if cancelled {
		return cancelledResponse(), nil
	}

	// Step 4: ship, as a child workflow so it gets its own execution,
	// history and entry in the UI.
	shipped, err := s.orders.ExecuteChildShipOrderSync(ctx, &examplev1.ShipOrderRequest{
		OrderId: req.OrderId,
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	status = examplev1.OrderStatus_ORDER_STATUS_SHIPPED
	logger.Info("order shipped", "tracking_number", shipped.TrackingNumber)

	return &examplev1.ProcessOrderResponse{
		Status:          status,
		ShippingAddress: address,
		TrackingNumber:  shipped.TrackingNumber,
	}, nil
}

// ShipOrder is executed by ProcessOrder as a child workflow. It only books a
// courier, but being a workflow of its own it could retry, wait or fan out
// without bloating the parent's history.
func (s *Service) ShipOrder(ctx workflow.Context, req *examplev1.ShipOrderRequest) (*examplev1.ShipOrderResponse, error) {
	courier, err := s.orders.ExecuteActivityDispatchCourierSync(ctx, &examplev1.DispatchCourierRequest{
		OrderId: req.OrderId,
		Address: req.Address,
	})
	if err != nil {
		return nil, err
	}

	return &examplev1.ShipOrderResponse{
		TrackingNumber: courier.TrackingNumber,
	}, nil
}

// DailySalesReport exists to demonstrate the generated schedule helpers: the
// client creates a Temporal schedule that runs it every minute.
func (s *Service) DailySalesReport(ctx workflow.Context, _ *emptypb.Empty) (*examplev1.DailySalesReportResponse, error) {
	return &examplev1.DailySalesReportResponse{
		Report: fmt.Sprintf("sales report generated at %s: business is booming", workflow.Now(ctx).Format(time.RFC3339)),
	}, nil
}
