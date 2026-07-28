<a id="top"></a>
# Services
<a id="service_example_v1_Orders"></a>
## example.v1.Orders
Service Orders is a small e-commerce style fulfillment service used as a
 guided tour of everything this plugin generates:

   - workflows:  ProcessOrder (long lived), ShipOrder (child), DailySalesReport (scheduled)
   - activities: ChargePayment (retries + non-retryable errors), PackItems (heartbeats), DispatchCourier (custom name)
   - signals:    CancelOrder
   - queries:    GetOrderStatus
   - updates:    ChangeShippingAddress (validated, synchronous request/response)

 See `example/worker` for the implementation and `example/client` for a
 runnable walkthrough of all of the above.

### Table of contents

   * [example.v1.Orders default settings](#svcoptions_example_v1_Orders)
 * Workflows
   * [example.v1.Orders.ProcessOrder](#method_example_v1_Orders_ProcessOrder)
   * [example.v1.Orders.ShipOrder](#method_example_v1_Orders_ShipOrder)
   * [example.v1.Orders.DailySalesReport](#method_example_v1_Orders_DailySalesReport)
   * [example.v1.Orders.TrackInventory](#method_example_v1_Orders_TrackInventory)
 * Activities
   * [example.v1.Orders.ChargePayment](#method_example_v1_Orders_ChargePayment)
   * [example.v1.Orders.PackItems](#method_example_v1_Orders_PackItems)
   * [example.v1.Orders.DispatchCourier](#method_example_v1_Orders_DispatchCourier)
 * Signals
   * [example.v1.Orders.CancelOrder](#method_example_v1_Orders_CancelOrder)
   * [example.v1.Orders.Restock](#method_example_v1_Orders_Restock)
 * Queries
   * [example.v1.Orders.GetOrderStatus](#method_example_v1_Orders_GetOrderStatus)
   * [example.v1.Orders.GetStock](#method_example_v1_Orders_GetStock)
 * Updates
   * [example.v1.Orders.ChangeShippingAddress](#method_example_v1_Orders_ChangeShippingAddress)
   * [example.v1.Orders.Reserve](#method_example_v1_Orders_Reserve)

<a id="svcoptions_example_v1_Orders"></a>
### Service options
| Option | Value |
| --- | --- |
| Default task queue | `orders` |

### Default workflow options
| Option | Value |
| --- | --- |
| Workflow execution timeout | 0s |

### Default activity options
| Option | Value |
| --- | --- |
| Schedule to close timeout | 0s |
| Start to close timeout | 0s |

### Workflows
<a id="method_example_v1_Orders_ProcessOrder"></a>
#### example.v1.Orders.ProcessOrder
ProcessOrder drives an order from payment to shipping. While it runs it
 can be queried (GetOrderStatus), updated (ChangeShippingAddress) and
 cancelled (CancelOrder signal)

Input: [example.v1.ProcessOrderRequest](#message_example_v1_ProcessOrderRequest)

Output: [example.v1.ProcessOrderResponse](#message_example_v1_ProcessOrderResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.ProcessOrder` |
| Workflow execution timeout | 1h0m0s |

Signals:
 * [example.v1.Orders.CancelOrder](#method_example_v1_Orders_CancelOrder)

Queries:
 * [example.v1.Orders.GetOrderStatus](#method_example_v1_Orders_GetOrderStatus)

Updates:
 * [example.v1.Orders.ChangeShippingAddress](#method_example_v1_Orders_ChangeShippingAddress)
<a id="method_example_v1_Orders_ShipOrder"></a>
#### example.v1.Orders.ShipOrder
ShipOrder hands the package over to a courier. ProcessOrder runs it as a
 child workflow so shipping shows up as its own execution in the UI

Input: [example.v1.ShipOrderRequest](#message_example_v1_ShipOrderRequest)

Output: [example.v1.ShipOrderResponse](#message_example_v1_ShipOrderResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.ShipOrder` |
| Workflow execution timeout | 1h0m0s |
<a id="method_example_v1_Orders_DailySalesReport"></a>
#### example.v1.Orders.DailySalesReport
DailySalesReport is a fast workflow meant to be driven by a Temporal
 schedule -- see the schedule part of the client walkthrough

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.DailySalesReportResponse](#message_example_v1_DailySalesReportResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.DailySalesReport` |
| Workflow execution timeout | 1h0m0s |
<a id="method_example_v1_Orders_TrackInventory"></a>
#### example.v1.Orders.TrackInventory
TrackInventory is a long-lived "entity" workflow tracking the stock of
 one SKU. It rolls over with continue-as-new after a number of restocks,
 which makes it the demo for how BLOCKED updates interact with
 continue-as-new: the workflow drains its update handlers (see
 workflow.AllHandlersFinished) before rolling over, so a Reserve update
 parked on "not enough stock" is answered before the run ends

Input: [example.v1.TrackInventoryRequest](#message_example_v1_TrackInventoryRequest)

Output: [example.v1.GetStockResponse](#message_example_v1_GetStockResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.TrackInventory` |
| Workflow execution timeout | 1h0m0s |

Signals:
 * [example.v1.Orders.Restock](#method_example_v1_Orders_Restock)

Queries:
 * [example.v1.Orders.GetStock](#method_example_v1_Orders_GetStock)

Updates:
 * [example.v1.Orders.Reserve](#method_example_v1_Orders_Reserve)

### Activities
<a id="method_example_v1_Orders_ChargePayment"></a>
#### example.v1.Orders.ChargePayment
ChargePayment captures the money. The retry policy retries transient
 payment provider hiccups with exponential backoff, but gives up
 immediately when the card is declined: "CardDeclined" is listed in
 non_retryable_error_types and the worker returns application errors of
 that type when the card is bad

Input: [example.v1.ChargePaymentRequest](#message_example_v1_ChargePaymentRequest)

Output: [example.v1.ChargePaymentResponse](#message_example_v1_ChargePaymentResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.ChargePayment` |
| Schedule to close timeout | 5m0s |
| Start to close timeout | 10s |

Retry policy:

| Option | Value |
| --- | --- |
| Initial interval | 1s |
| Backoff coefficient | 2.000000 |
| Maximum attempts | 5 |
| Maximum interval | 10s |
| Non retryable error types | `CardDeclined` |
<a id="method_example_v1_Orders_PackItems"></a>
#### example.v1.Orders.PackItems
PackItems packs the order, one parcel per item. It is slow, so it
 records a heartbeat after every parcel: if the worker dies mid-pack,
 Temporal notices within heartbeat_timeout and reschedules the activity

Input: [example.v1.PackItemsRequest](#message_example_v1_PackItemsRequest)

Output: [example.v1.PackItemsResponse](#message_example_v1_PackItemsResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.PackItems` |
| Schedule to close timeout | 5m0s |
| Start to close timeout | 1m0s |
| Heartbeat timeout | 10s |
<a id="method_example_v1_Orders_DispatchCourier"></a>
#### example.v1.Orders.DispatchCourier
DispatchCourier books a courier and returns a tracking number. The
 `name` option overrides the registered activity name, which otherwise
 defaults to <package>.<service>.<method>

Input: [example.v1.DispatchCourierRequest](#message_example_v1_DispatchCourierRequest)

Output: [example.v1.DispatchCourierResponse](#message_example_v1_DispatchCourierResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `courier.Dispatch` |
| Schedule to close timeout | 5m0s |
| Start to close timeout | 30s |

### Queries
<a id="method_example_v1_Orders_GetOrderStatus"></a>
#### example.v1.Orders.GetOrderStatus
GetOrderStatus reads the current state of an order without touching it.
 Queries are read only and are answered even after the workflow
 completed, as long as a worker is running

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.GetOrderStatusResponse](#message_example_v1_GetOrderStatusResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.GetOrderStatus` |
<a id="method_example_v1_Orders_GetStock"></a>
#### example.v1.Orders.GetStock
GetStock reads the current stock of a running (or finished)
 TrackInventory workflow

Input: [google.protobuf.Empty](#message_google_protobuf_Empty)

Output: [example.v1.GetStockResponse](#message_example_v1_GetStockResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.GetStock` |

### Signals
<a id="method_example_v1_Orders_CancelOrder"></a>
#### example.v1.Orders.CancelOrder
CancelOrder asks a running ProcessOrder workflow to stop. Signals are
 fire and forget: the response type of a signal rpc is ignored by the
 generator

Input: [example.v1.CancelOrderRequest](#message_example_v1_CancelOrderRequest)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.CancelOrder` |
<a id="method_example_v1_Orders_Restock"></a>
#### example.v1.Orders.Restock
Restock adds stock to a running TrackInventory workflow. Fire and forget

Input: [example.v1.RestockRequest](#message_example_v1_RestockRequest)

Output: [google.protobuf.Empty](#message_google_protobuf_Empty)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.Restock` |

### Updates
<a id="method_example_v1_Orders_ChangeShippingAddress"></a>
#### example.v1.Orders.ChangeShippingAddress
ChangeShippingAddress is an update: a synchronous request/response
 against the running workflow. The caller blocks until the handler
 answers, and a validator rejects garbage before it ever reaches the
 workflow history

Input: [example.v1.ChangeShippingAddressRequest](#message_example_v1_ChangeShippingAddressRequest)

Output: [example.v1.ChangeShippingAddressResponse](#message_example_v1_ChangeShippingAddressResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.ChangeShippingAddress` |
<a id="method_example_v1_Orders_Reserve"></a>
#### example.v1.Orders.Reserve
Reserve takes stock out of a TrackInventory workflow. This is a BLOCKING
 update: if there is not enough stock the handler parks on workflow.Await
 until a Restock signal makes the quantity available, and only then
 answers the caller. This is the lease/semaphore pattern

Input: [example.v1.ReserveRequest](#message_example_v1_ReserveRequest)

Output: [example.v1.ReserveResponse](#message_example_v1_ReserveResponse)


| Setting | Value |
| ----------- | ----------------------- |
| Temporal registered method name | `example.v1.Orders.Reserve` |

# Messages
<a id="message_example_v1_OrderItem"></a>
## example.v1.OrderItem
A single line of an order
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Sku | string | Optional | ✅ | <pre>Stock keeping unit, e.g. "die-d20"</pre> |
| Quantity | int32 | Optional | ✅ | <pre>How many of them</pre> |


<a id="message_example_v1_ProcessOrderRequest"></a>
## example.v1.ProcessOrderRequest
Everything needed to process an order
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| OrderId | string | Optional | ✅ | <pre>Client-chosen order identifier</pre> |
| Items | message | Repeated | ✅ | <pre>What was bought</pre> |
| AmountCents | int64 | Optional | ✅ | <pre>Total price in cents</pre> |
| CardToken | string | Optional | ✅ | <pre>Payment card token. The example worker understands three of them:
 "tok-ok" (works), "tok-flaky" (fails twice then works, demonstrating
 the retry policy) and "tok-declined" (fails the workflow with a
 non-retryable CardDeclined error)</pre> |
| ShippingAddress | string | Optional | ✅ | <pre>Where to ship. Can be changed while the order is in flight with the
 ChangeShippingAddress update</pre> |
| PickupWindowSeconds | int32 | Optional | ✅ | <pre>How long the order waits for carrier pickup once packed, in seconds.
 Defaults to 5 when unset. Set it high to keep an order parked so you
 can poke at it (query it, update it, cancel it) while watching the UI</pre> |


<a id="message_example_v1_ProcessOrderResponse"></a>
## example.v1.ProcessOrderResponse
Final state of a processed order
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Status | enum | Optional | ✅ | <pre>Terminal status, SHIPPED or CANCELLED</pre> |
| ShippingAddress | string | Optional | ✅ | <pre>The address the order actually shipped to</pre> |
| TrackingNumber | string | Optional | ✅ | <pre>Carrier tracking number, empty if the order never shipped</pre> |
| CancelReason | string | Optional | ✅ | <pre>Why the order was cancelled, empty otherwise</pre> |


<a id="message_example_v1_CancelOrderRequest"></a>
## example.v1.CancelOrderRequest
Asks a running order to stop
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Reason | string | Optional | ✅ | <pre>Human readable reason, recorded in the final result</pre> |


<a id="message_example_v1_GetOrderStatusResponse"></a>
## example.v1.GetOrderStatusResponse
Current state of a running (or finished) order
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Status | enum | Optional | ✅ | <pre></pre> |
| ShippingAddress | string | Optional | ✅ | <pre>Address the order will ship to if nobody changes it</pre> |


<a id="message_example_v1_ChangeShippingAddressRequest"></a>
## example.v1.ChangeShippingAddressRequest
Requests re-routing an order that has not shipped yet
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Address | string | Optional | ✅ | <pre>The new address, must not be empty</pre> |


<a id="message_example_v1_ChangeShippingAddressResponse"></a>
## example.v1.ChangeShippingAddressResponse
Confirms the re-routing
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| PreviousAddress | string | Optional | ✅ | <pre>The address that was replaced</pre> |


<a id="message_example_v1_ChargePaymentRequest"></a>
## example.v1.ChargePaymentRequest
Payment capture input
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| OrderId | string | Optional | ✅ | <pre></pre> |
| AmountCents | int64 | Optional | ✅ | <pre></pre> |
| CardToken | string | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_ChargePaymentResponse"></a>
## example.v1.ChargePaymentResponse
Payment capture receipt
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| TransactionId | string | Optional | ✅ | <pre>Payment provider transaction reference</pre> |


<a id="message_example_v1_PackItemsRequest"></a>
## example.v1.PackItemsRequest
Packing input
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Items | message | Repeated | ✅ | <pre></pre> |


<a id="message_example_v1_PackItemsResponse"></a>
## example.v1.PackItemsResponse
Packing result
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Parcels | int32 | Optional | ✅ | <pre>Number of parcels produced</pre> |


<a id="message_example_v1_ShipOrderRequest"></a>
## example.v1.ShipOrderRequest
Child workflow input: ship one order
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| OrderId | string | Optional | ✅ | <pre></pre> |
| Address | string | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_ShipOrderResponse"></a>
## example.v1.ShipOrderResponse
Child workflow result
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| TrackingNumber | string | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_DispatchCourierRequest"></a>
## example.v1.DispatchCourierRequest
Courier booking input
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| OrderId | string | Optional | ✅ | <pre></pre> |
| Address | string | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_DispatchCourierResponse"></a>
## example.v1.DispatchCourierResponse
Courier booking result
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| TrackingNumber | string | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_DailySalesReportResponse"></a>
## example.v1.DailySalesReportResponse
Output of the scheduled reporting workflow
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Report | string | Optional | ✅ | <pre>A very serious business report</pre> |


<a id="message_example_v1_TrackInventoryRequest"></a>
## example.v1.TrackInventoryRequest
Input of the TrackInventory entity workflow
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Sku | string | Optional | ✅ | <pre>Which SKU this inventory tracks</pre> |
| InitialStock | int32 | Optional | ✅ | <pre>Stock at the start of this run. On continue-as-new the workflow carries
 the current stock over through this field</pre> |
| Generation | int32 | Optional | ✅ | <pre>Incremented on every continue-as-new rollover; leave unset when starting.
 Exposed through GetStock so you can observe the rollovers happening</pre> |
| RestocksBeforeContinueAsNew | int32 | Optional | ✅ | <pre>Roll over with continue-as-new after this many Restock signals.
 0 disables rollovers</pre> |
| SkipHandlerDrain | bool | Optional | ✅ | <pre>DO NOT SET, demo of the anti-pattern: skip draining update handlers
 before continue-as-new. A Reserve update blocked at rollover time is
 then aborted and its caller gets an error instead of an answer</pre> |


<a id="message_example_v1_RestockRequest"></a>
## example.v1.RestockRequest
Adds stock
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Quantity | int32 | Optional | ✅ | <pre></pre> |


<a id="message_example_v1_ReserveRequest"></a>
## example.v1.ReserveRequest
Takes stock, blocking until enough is available
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Quantity | int32 | Optional | ✅ | <pre>How many items to reserve, must be > 0</pre> |
| TimeoutSeconds | int32 | Optional | ✅ | <pre>How long to wait for enough stock before failing the update, in
 seconds. 0 waits forever. The timeout is enforced INSIDE the workflow
 with a durable timer (workflow.AwaitWithTimeout), so it survives worker
 restarts and also bounds how long a parked reservation can delay a
 continue-as-new rollover</pre> |


<a id="message_example_v1_ReserveResponse"></a>
## example.v1.ReserveResponse
Answer to a successful reservation
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| RemainingStock | int32 | Optional | ✅ | <pre>Stock left after the reservation</pre> |
| Generation | int32 | Optional | ✅ | <pre>Which run answered: if the update had to wait across a rollover this is
 higher than the generation it was sent to</pre> |


<a id="message_example_v1_GetStockResponse"></a>
## example.v1.GetStockResponse
Current state of a TrackInventory workflow
| Field name | Type | Cardinality | Deprecated ? | Description |
| --- | --- | --- | --- | --- |
| Stock | int32 | Optional | ✅ | <pre></pre> |
| Generation | int32 | Optional | ✅ | <pre>Continue-as-new rollovers so far, starts at 1</pre> |




[Back to top](#top)
