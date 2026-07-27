# The guided tour

This directory is a self-contained walkthrough of everything `protoc-gen-go-tmprl`
generates, built around a small e-commerce style **Orders** service:

```
example/
  proto/example/v1/example.proto   the annotated service definition
  orders/service.go                implements every workflow and activity
  worker/main.go                   wires the service to a Temporal server
  client/main.go                   a narrated, step by step tour
  e2e/e2e_test.go                  build-tagged tests against a real server
```

The generated code lives in [`gen/example/v1/example_tmprl.pb.go`](../gen/example/v1/example_tmprl.pb.go),
with markdown docs in [`gen/example/v1/example_tmprl_doc.md`](../gen/example/v1/example_tmprl_doc.md).

## What the service does

`ProcessOrder` drives an order through its lifecycle, and every plugin feature
has a natural place in it:

```
             ChargePayment                PackItems              ShipOrder (child workflow)
  PENDING ----------------->  PAID  ----------------->  PACKED -----------------> SHIPPED
             activity with               activity with        |    runs the DispatchCourier
             retry policy +              heartbeats           |    activity (custom name)
             non-retryable                                    |
             "CardDeclined"                            pickup window:
                                                       cancellable (signal)
                                                       re-routable (update)
```

| Feature | Proto annotation | Where to look |
| --- | --- | --- |
| Workflow | `(temporal.v1.workflow)` on `ProcessOrder` | orders: `ProcessOrder` |
| Child workflow | plain workflow, executed via `ExecuteChildShipOrderSync` | orders: `ShipOrder` |
| Activity retry policy | `retry_policy` on `ChargePayment` | orders: `ChargePayment`, the `tok-flaky` card |
| Non-retryable errors | `non_retryable_error_types: ["CardDeclined"]` | orders: `ChargePayment`, the `tok-declined` card |
| Activity heartbeats | `heartbeat_timeout` on `PackItems` | orders: `PackItems` |
| Custom registered name | `name: "courier.Dispatch"` on `DispatchCourier` | proto + Temporal UI |
| Signal | `(temporal.v1.signal)` on `CancelOrder` | client step 3 |
| Query | `(temporal.v1.query)` on `GetOrderStatus` | client step 1 |
| Update (+ validator) | `(temporal.v1.update)` on `ChangeShippingAddress` | client steps 1 and 2 |
| Schedules | generated for every workflow, no annotation | client step 6 |

## Running it

You need docker (for the Temporal dev stack) and Go.

```console
# 1. bring up Temporal + its UI. Every image version has a sane default baked
#    into the compose file; .envrc pins them explicitly if you use direnv
docker compose up -d --wait

# 2. start the worker, and keep this terminal visible: payment retries and
#    packing heartbeats are logged here
go run ./example/worker

# 3. in another terminal, run the walkthrough
go run ./example/client
```

The stack follows the [temporalio/samples-server](https://github.com/temporalio/samples-server/tree/main/compose)
compose layout: a plain `temporalio/server` container plus one-shot
admin-tools jobs that create the databases and the `default` namespace
(`scripts/setup-postgres.sh`, `scripts/create-namespace.sh`). `--wait` blocks
until the server is healthy and the namespace exists.

If the default host ports are taken on your machine (another temporal stack,
for example), override them and point the binaries at the right address:

```console
TEMPORAL_HOST_PORT=27233 TEMPORAL_UI_HOST_PORT=28080 POSTGRES_HOST_PORT=25433 docker compose up -d --wait
TEMPORAL_ADDRESS=localhost:27233 go run ./example/worker
TEMPORAL_ADDRESS=localhost:27233 go run ./example/client
```

The client narrates six steps:

1. **The happy path.** Starts `ProcessOrder` asynchronously, polls the
   `GetOrderStatus` query in the background (you see the status transitions
   as they happen), then re-routes the order mid-flight with the
   `ChangeShippingAddress` update. The update call blocks until the workflow's
   handler answers and returns a typed response containing the previous
   address: synchronous request/response against a running workflow.
2. **A rejected update.** Sends an empty address. The validator the worker
   registered rejects it before it is ever recorded in history, and the error
   comes back to the caller. The workflow then ships and the client collects
   the typed result, plus one last query to show queries still answer after
   completion.
3. **Cancellation with a signal.** Starts a second order and sends the
   `CancelOrder` signal during the pickup window. Signals are fire and forget;
   the workflow ends with status `CANCELLED` and the reason in its result.
4. **Transient failures.** The `tok-flaky` card makes `ChargePayment` fail
   twice before succeeding. The retry policy from the proto (1s initial
   interval, backoff 2.0) retries transparently; watch the worker logs to see
   the failed attempts.
5. **Non-retryable failures.** The `tok-declined` card makes `ChargePayment`
   return an application error of type `CardDeclined`, which the retry policy
   lists as non-retryable. The workflow fails immediately and the client
   unwraps the typed error with `errors.As`.
6. **Schedules.** Upserts a schedule running `DailySalesReport` every minute,
   lists, pauses and unpauses it (both idempotent), then deletes it. Comment
   out the delete in the client if you want to watch it fire.

Everything is also visible in the Temporal UI at
[http://localhost:8080](http://localhost:8080) — the child `ShipOrder`
executions, the update in the `order-happy` history, the retry attempts on the
flaky payment, and the failed `order-declined` workflow.

## Tests

The example is tested at two layers:

- **Unit** (`make test-unit`, runs everywhere): `example/orders` exercises the
  real service against the SDK's in-memory test suite — the mock clock means a
  two hour pickup window costs no wall time — and `gen/example/v1` pins the
  generated helpers themselves.
- **e2e** (`make test-e2e`, needs docker): brings up the compose stack and runs
  `example/e2e` against the real server with an in-process worker, driving the
  generated client-side API end to end. Build-tagged `e2e` so plain
  `go test ./...` skips it. CI runs both layers on every PR.

### What each e2e test proves, in human terms

The tests are narrated with `t.Log`, so `go test -v` (and the CI output) reads
as a story. What each one demonstrates:

- **`TestHappyPathWithClientSideUpdates`** — the core promise of updates. An
  order is started and reaches PACKED, where it sits in its pickup window. The
  test then calls the generated update method: the call **blocks** until code
  *inside the running workflow* processes the request, and comes back with a
  typed response (the address that was replaced). It does it twice (once via
  the client, once via the workflow object) to show both flavours chain
  correctly, then sends an empty address: the validator rejects it before it
  is ever written to workflow history — the caller gets the error, the
  workflow never even sees the request. Finally the window expires and the
  order ships to the *last* address set, proving the updates actually mutated
  live workflow state. A last query after completion shows queries read
  history and keep working once the workflow is done.
- **`TestCancelSignal`** — signals vs updates. A parked order gets a
  `CancelOrder` signal: fire-and-forget, no response, but the workflow's
  signal handler flips a flag the pickup wait is watching, so it ends
  CANCELLED immediately with the reason in its result.
- **`TestFlakyPaymentRetries`** — retry policies do their job server-side.
  The `tok-flaky` card fails the payment activity twice; the retry policy
  declared in the proto retries it with backoff. The caller sees no failure
  at all, just a slower success.
- **`TestCardDeclinedFailsFast`** — failures you should NOT retry. The
  `tok-declined` card raises an application error whose type is listed in
  `non_retryable_error_types`: no retries, the workflow fails in one attempt,
  and the caller recovers the typed error with `errors.As` to react to the
  business reason (`CardDeclined`).
- **`TestScheduleCRUD`** — the whole generated schedule lifecycle against a
  real server: upsert, list (polling, because listings go through the
  eventually-consistent visibility store), idempotent pause/unpause, delete.

## Poking at it

The stack keeps running until you `docker compose down`. Some ideas:

- Kill the worker while an order is packing and restart it: the workflow picks
  up where it left off. That is Temporal doing its job, the generated code
  just wires your service into it.
- Run the client twice concurrently. Workflow IDs are generated per run
  (`gen-workflow-prefix=true`), so the orders don't collide.
- Send an update after the order shipped (e.g. from a quick Go snippet or
  `temporal workflow update`): the handler rejects it with "too late to change
  the address" — a handler error fails the update, not the workflow.
