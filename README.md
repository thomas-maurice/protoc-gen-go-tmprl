# protoc-gen-go-tmprl

Easily generate client and worker code for [temporal](https://temporal.io) from [protobuf](https://protobuf.dev) definitions

## Show me the code !

You will find a reference proto [here](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/proto/example/v1/example.proto)
and it corresponding generated code [there](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/gen/example/v1/example_tmprl.pb.go)
for reference purposes.

For a runnable, narrated tour of everything the plugin generates (workflows, activities, signals,
queries, updates, failure handling, schedules), see [the example walkthrough](example/README.md).

## How to use it

You need to include [temporal.v1](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/proto/temporal/v1/temporal.proto) in your project. Note that this protobuf package is also published on
[buf.build](https://buf.build/thomas-maurice/temporal) so you can easily use it should you use `buf` to build your project's protos.

When the setup is done, you can start defining actions and workflows in your services like so

```protobuf
syntax = "proto3";

package example.v1;

// This is where buf will generate your protobuf go code
option go_package = "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1";

import "temporal/v1/temporal.proto";
import "google/protobuf/empty.proto";

service Orders  {
    option (temporal.v1.service) = {
        // this is not mandatory but will serve as a sane default
        task_queue: "orders"
    };
    // ChargePayment captures the money for an order
    rpc ChargePayment(ChargePaymentRequest) returns (ChargePaymentResponse) {
        option (temporal.v1.activity) = {
            // you don't have to but you can also define default activity
            // options that will be applied every time the activity
            // is called
        };
    }

    // ProcessOrder drives an order from payment to shipping
    rpc ProcessOrder(ProcessOrderRequest) returns (ProcessOrderResponse) {
        option (temporal.v1.workflow) = {
            // Similarily to the activity options you can define workflows
            // options such as retry policies and so on so you don't
            // have to specify them every run
        };
    }
}
```

> [!WARNING]
> Make sure your package name is not `temporal.*`, since it will clash with the imports from this package and generate broken code.

### Default workflow & activity setups

You can setup a service level (here a service refers to a worker) default for activities and workflows respectively in the `default_activity_options` and `default_workflow_options` fields of the
`temporal.v1.service` option.

If you don't want to do that, it's fine, yhou can define these setups at the individual workflow or activity level doing something like this

```protobuf
    // Some activity
    rpc SomeActivity(SomeRequest) returns (SomeResponse) {
        option (temporal.v1.activity) = {
            schedule_to_start_timeout: 30
            schedule_to_close_timeout: 120
            start_to_close_timeout: 120
            retry_policy: {
                initial_interval: 1
                backoff_coefficient: 1.5
                maximum_interval: 10
                maximum_attempts: 10
                non_retryable_error_types: ["FATAL"]
            }
        };
    }
```

Similarly for the workflows

```protobuf
    // Do some stuff
    rpc DoSomething(SomeRequest) returns (SomeResponse) {
        option (temporal.v1.workflow) = {
            // one day
            workflow_execution_timeout: 86400
            // one hour
            workflow_run_timeout: 7200
        };
    }
```

### Workflow Schedules

Every workflow in a service gets a generated set of schedule-management helpers — there is no proto annotation required to opt in. Scheduling cadence is a runtime concern, not part of the workflow's schema, so callers pass their own `client.ScheduleOptions` (with a `Spec` of their choice: `CronExpressions`, `Intervals`, `Calendars`, etc.) when they create or upsert a schedule. The generator only fills in the service's default task queue if the caller left it empty.

```protobuf
    // A fast workflow meant to be driven by a Temporal schedule
    rpc DailySalesReport(google.protobuf.Empty) returns (DailySalesReportResponse) {
        option (temporal.v1.workflow) = {};
    }
```

The generator creates the following schedule methods for every workflow in the service:

```golang
// CreateScheduleDailySalesReport creates a schedule for DailySalesReport. The caller supplies the cadence via
// options[0].Spec (CronExpressions / Intervals / Calendars). Non-zero fields from options[0] win
// over generator defaults; if TaskQueue is left empty the service's default task queue is used.
func (c *OrdersClient) CreateScheduleDailySalesReport(
    ctx context.Context,
    scheduleID string,
    req *emptypb.Empty,
    options ...client.ScheduleOptions,
) (client.ScheduleHandle, error)

// GetScheduleDailySalesReport gets a handle to an existing schedule for DailySalesReport.
func (c *OrdersClient) GetScheduleDailySalesReport(
    ctx context.Context,
    scheduleID string,
) client.ScheduleHandle

// DeleteScheduleDailySalesReport deletes an existing schedule for DailySalesReport.
// Returns the underlying client error if the schedule does not exist or the delete fails.
func (c *OrdersClient) DeleteScheduleDailySalesReport(
    ctx context.Context,
    scheduleID string,
) error

// ListScheduleDailySalesReport lists all schedules in the namespace whose action is the DailySalesReport
// workflow type. pageSize is forwarded to the Temporal ScheduleClient.List call.
func (c *OrdersClient) ListScheduleDailySalesReport(
    ctx context.Context,
    pageSize int,
) ([]client.ScheduleListEntry, error)

// UpsertScheduleDailySalesReport creates the schedule if it does not exist, otherwise performs an
// in-place update (Spec, Action, Overlap, CatchupWindow, PauseOnFailure, TypedSearchAttributes).
// A user-supplied Paused flag is deliberately not honoured on the update path: a bool can't
// distinguish "leave it alone" from "unpause", so run-state transitions are expressed through
// PauseScheduleDailySalesReport / UnpauseScheduleDailySalesReport instead.
func (c *OrdersClient) UpsertScheduleDailySalesReport(
    ctx context.Context,
    scheduleID string,
    req *emptypb.Empty,
    options ...client.ScheduleOptions,
) (client.ScheduleHandle, error)

// PauseScheduleDailySalesReport pauses a running schedule. The note is recorded on the schedule's
// audit trail. Describe is called first and the Pause RPC is skipped when the schedule is
// already paused, so this is safe to call on every reconcile tick.
func (c *OrdersClient) PauseScheduleDailySalesReport(
    ctx context.Context,
    scheduleID string,
    note string,
) error

// UnpauseScheduleDailySalesReport resumes a paused schedule. Describe is called first and the
// Unpause RPC is skipped when the schedule is already running.
func (c *OrdersClient) UnpauseScheduleDailySalesReport(
    ctx context.Context,
    scheduleID string,
    note string,
) error
```

Example usage:

```golang
// Create a schedule. The caller owns the cadence: supply a Spec.
scheduleHandle, err := ordersClient.CreateScheduleDailySalesReport(
    ctx,
    "daily-sales-report",
    &emptypb.Empty{},
    client.ScheduleOptions{
        Spec: client.ScheduleSpec{
            CronExpressions: []string{"* * * * *"},
        },
    },
)
if err != nil {
    log.Fatal(err)
}

// Get existing schedule handle
scheduleHandle = ordersClient.GetScheduleDailySalesReport(ctx, "daily-sales-report")

// Use the schedule handle to pause, unpause, describe, etc.
err = scheduleHandle.Pause(ctx, client.SchedulePauseOptions{
    Note: "Pausing for maintenance",
})

// Idempotent create-or-update. Safe to call on redeploys; the Paused flag is intentionally
// ignored on the update path.
_, err = ordersClient.UpsertScheduleDailySalesReport(
    ctx,
    "daily-sales-report",
    &emptypb.Empty{},
    client.ScheduleOptions{
        Spec: client.ScheduleSpec{CronExpressions: []string{"* * * * *"}},
    },
)

// Enumerate every schedule whose action is this workflow type.
entries, err := ordersClient.ListScheduleDailySalesReport(ctx, 100)

// Tear a schedule down.
err = ordersClient.DeleteScheduleDailySalesReport(ctx, "daily-sales-report")

// Toggle run state. Each of these is a read-then-write: a Describe round-trip
// is always paid, but the Pause/Unpause RPC is skipped when the schedule is
// already in the target state.
err = ordersClient.PauseScheduleDailySalesReport(ctx, "daily-sales-report", "maintenance window")
err = ordersClient.UnpauseScheduleDailySalesReport(ctx, "daily-sales-report", "maintenance over")
```

**Features:**
- **Type-safe:** Generated methods use workflow-specific request types
- **Option merging:** Runtime `client.ScheduleOptions` are merged field-by-field onto generator defaults (non-zero fields win); the merge logic is factored into a per-service `mergeScheduleOptions<Service>` helper in the generated file
- **Workflow configuration:** Automatically applies workflow timeouts and retry policies to scheduled executions
- **Unconditional generation:** Every workflow gets the schedule surface — cadence is a runtime concern, not part of the schema
- **Full CRUD surface:** `Create`, `Get`, `List`, `Upsert`, `Delete`, `Pause`, and `Unpause` are all generated per workflow

### The workflow objects

Each workflow will get assigned a dedicated object in the generated code. All the workflow objects implement the `internal.WorkflowRun`
interface from the Temporal SDK. They contain a few methods that can be useful for you. Let's take as an example the following
protobuf:
```protobuf
    // ProcessOrder drives an order from payment to shipping
    rpc ProcessOrder(ProcessOrderRequest) returns (ProcessOrderResponse) {
        option (temporal.v1.workflow) = {};
    }
```

The following methods will be generated for the workflow object:

```golang
// Cancel cancels a given workflow
func (w *OrdersProcessOrder) Cancel(ctx context.Context) error
// Returns the workflow ID
func (w *OrdersProcessOrder) GetID() string
// Returns the run ID
func (w *OrdersProcessOrder) GetRunID() string
// Terminates terminates a given workflow
func (w *OrdersProcessOrder) Terminate(ctx context.Context, reason string, details ...interface{})
// Get gets the result of a given workflow with its native type
func (w *OrdersProcessOrder) Result(ctx context.Context) (*ProcessOrderResponse, error)
// ResultWithOptions gets the result of a given workflow with its native type
func (w *OrdersProcessOrder) ResultWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (*ProcessOrderResponse, error)
// Get gets the result of a given workflow with pointers -- discouraged to use but required to implement internal.WorkflowRun
func (w *OrdersProcessOrder) Get(ctx context.Context, valuePtr interface{}) error
// Get gets the result of a given workflow with pointers -- discouraged to use but required to implement internal.WorkflowRun
func (w *OrdersProcessOrder) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error
```

You can retrieve this `OrdersProcessOrder` object from the client using one of these two methods:
```golang
func (c *OrdersClient) GetProcessOrder(ctx context.Context, workflowId string, runId string) *OrdersProcessOrder
func (c *OrdersClient) GetProcessOrderFromRun(future client.WorkflowRun) *OrdersProcessOrder
```

#### Workflow object signal, queries and updates
Additionally, if you have defined signal, queries and updates in your workflow options like in the following protobuf
```protobuf
    rpc ProcessOrder(ProcessOrderRequest) returns (ProcessOrderResponse) {
        option (temporal.v1.workflow) = {
            signals: ["CancelOrder"]
            queries: ["GetOrderStatus"]
            updates: ["ChangeShippingAddress"]
        };
    }
```

Then you will have access to the following methods:

```golang
// SignalCancelOrder sends the CancelOrder signal to the workflow
func (w *OrdersProcessOrder) SignalCancelOrder(ctx context.Context, req *CancelOrderRequest) error
// QueryGetOrderStatus queries the workflow with GetOrderStatus
func (w *OrdersProcessOrder) QueryGetOrderStatus(ctx context.Context, req *emptypb.Empty) (*GetOrderStatusResponse, error)
// UpdateChangeShippingAddress sends the ChangeShippingAddress update to the workflow and waits for it to complete
func (w *OrdersProcessOrder) UpdateChangeShippingAddress(ctx context.Context, req *ChangeShippingAddressRequest) (*ChangeShippingAddressResponse, error)
```

:warning: The name you pass to the protobuf must match the name of the generated go name for the signal, i.e. `some_func` would
become `SomeFunc`

:warning: The methods you set as signals, queries and updates MUST be defined inside the service whose workflow uses them. You cannot use
the signals/queries/updates defined in `Service2` for `Service1`, you can reuse types, not methods.

### Signals and queries

You can define signal and queries functions in your service, if they are annotated with the respective `temporal.v1.signal` and
`temporal.v1.query` options they will be treated as such. For example if we have a signal like so:

```protobuf
    rpc CancelOrder(CancelOrderRequest) returns (google.protobuf.Empty) {
        option (temporal.v1.signal) = {};
    }
```

You will have access to the three following methods:

```golang
// SendSignalCancelOrder sends the CancelOrder signal to a workflow
// This can be called from a workflow or externally
func (c *OrdersClient) SendSignalCancelOrder(ctx context.Context, workflowID string, runID string, req *CancelOrderRequest) error

// ReceiveSignalCancelOrder waits for the the CancelOrder signal
// This is called within a workflow exclusively
func ReceiveSignalCancelOrder(ctx workflow.Context) (*CancelOrderRequest, bool)

// ReceiveSignalCancelOrderAsync recieves the the CancelOrder signal asynchronously.
// It doesn't wait if there is no signal in the queue.
// This is called within a workflow exclusively
func ReceiveSignalCancelOrderAsync(ctx workflow.Context) (*CancelOrderRequest, bool)
```

:warning: Whatever you put in the response parameter of the signal does not matter at all and
will be ignored by the code generator, as you want to send and recieve the same object.

For queriees it is very similar, let's take for example the following query:
```protobuf
    rpc GetOrderStatus(google.protobuf.Empty) returns (GetOrderStatusResponse) {
        option (temporal.v1.query) = {};
    }
```

This will grant you the following two methods:

```golang
// QueryGetOrderStatus sends the GetOrderStatus query to a workflow
// This can be called from a workflow or externally
func (c *OrdersClient) QueryGetOrderStatus(ctx context.Context, workflowID string, runID string, req *emptypb.Empty) (*GetOrderStatusResponse, error)

// HandleQueryGetOrderStatus sets up the GetOrderStatus query and responds accordingly, returns an error if it failed
// This is called within your workflow to setup the handler method
func HandleQueryGetOrderStatus(ctx workflow.Context, queryFunc func(req *emptypb.Empty) (*GetOrderStatusResponse, error)) error
```

### Updates

Updates are synchronous request/response interactions with a running workflow: the caller blocks until the
workflow's update handler returns, and gets back a typed response. Unlike queries, updates are recorded in the
workflow history and their handlers can mutate workflow state and block (using `workflow.Await`, timers, etc)
before answering. Annotate a method with the `temporal.v1.update` option:

```protobuf
    rpc ChangeShippingAddress(ChangeShippingAddressRequest) returns (ChangeShippingAddressResponse) {
        option (temporal.v1.update) = {};
    }
```

This will grant you the following methods:

```golang
// UpdateChangeShippingAddress sends the ChangeShippingAddress update to a workflow and waits for it to complete
// This can be called externally
func (c *OrdersClient) UpdateChangeShippingAddress(ctx context.Context, workflowID string, runID string, req *ChangeShippingAddressRequest, options ...client.UpdateWorkflowOptions) (*ChangeShippingAddressResponse, error)

// HandleUpdateChangeShippingAddress sets up the ChangeShippingAddress update handler, returns an error if it failed
// This is called within your workflow to setup the handler method
func HandleUpdateChangeShippingAddress(ctx workflow.Context, updateFunc func(ctx workflow.Context, req *ChangeShippingAddressRequest) (*ChangeShippingAddressResponse, error)) error

// HandleUpdateChangeShippingAddressWithValidator sets up the ChangeShippingAddress update handler with a validator.
// The validator runs before the update is admitted to history; if it returns a non-nil error
// the update is rejected and never recorded
func HandleUpdateChangeShippingAddressWithValidator(ctx workflow.Context, updateFunc func(ctx workflow.Context, req *ChangeShippingAddressRequest) (*ChangeShippingAddressResponse, error), validatorFunc func(ctx workflow.Context, req *ChangeShippingAddressRequest) error) error
```

The client helper defaults `WaitForStage` to `client.WorkflowUpdateStageCompleted` so the call blocks until the
handler returns; pass a `client.UpdateWorkflowOptions` to override it or to set an `UpdateID` for idempotency.

### Child workflow executions
You get access to a similar API with the child workflows executions, something like so
```golang
func (c *OrdersClient) GetChildOrdersProcessOrderExecution(future workflow.ChildWorkflowFuture) *ChildOrdersProcessOrderExecution
```

However the API is a bit more limited (no way to query for example) because it is a wrapper around the `internal.ChildWorkflowExecution` object.
The method type you'd be interested in are the ones that allow you to signal the child workflow, for example:

```golang
func (w *ChildOrdersProcessOrderExecution) SignalCancelOrder(ctx workflow.Context, req *CancelOrderRequest) error
```

### The exposed API

The generated code exposes a lot of primitives such as (non exhaustive list):

* `client.ExecuteWorkflowX`: Executes a workflow and returns a future
* `client.ExecuteWorkflowXSync`: Executes a workflow and blocks until the result is returned
* `client.ExecuteChildX`: Executes a workflow from a workflow and returns a future
* `client.ExecuteChildXSync`: Executes a workflow from a workflow and blocks until the result is returned
* `client.ExecuteActivityX`: Executes an activity and returns a future
* `client.ExecuteActivityXSync`: Executes an activity and blocks until the result is returned
* `client.CreateScheduleX`: Creates a Temporal schedule for a workflow (caller supplies the Spec)
* `client.GetScheduleX`: Gets a handle to an existing schedule
* `client.DeleteScheduleX`: Deletes an existing schedule
* `client.ListScheduleX`: Lists every schedule whose action is workflow `X`
* `client.UpsertScheduleX`: Creates or updates a schedule idempotently (caller supplies the Spec)
* `client.PauseScheduleX`: Pauses a schedule if it's running; no-op if already paused
* `client.UnpauseScheduleX`: Resumes a paused schedule; no-op if already running
* `client.UpdateX`: Sends the X update to a workflow and blocks until the typed result is returned
* `client.GetX`: Gets an instance of a workflow
* `workflow.Cancel`: Cancels a workflow
* `workflow.Teminate`: Terminates a workflow
* `workflow.Get`: Gets the result of a workflow like you would on a normal future (you probably don't want that because no type safety)
* `workflow.Result`: Gets the result of a workflow *with type safety*

Generally a good starting point to get familiar with the generated code is to run [the example walkthrough](example/README.md): the [example client](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/client/main.go) is a narrated tour of every generated primitive and the [example worker](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/worker/main.go) implements the matching service.

## Options
* `gen-workflow-prefix`, if set to true, instead of using an UUID for workflow IDs, the worker will generate a name that looks like `<module>.v<X>.<service>.<rpcMethodName>/<uuid>`, like `example.v1.Orders.DailySalesReport/e2715d07-7bc0-495d-90c5-c396c0a17b46` for example.
* `gen-docs`, if set to true a markdown documentation file will be output along your generated protobuf code.
* `paths`, like on the protoc-gen-go, for example `paths=source_relative`
* `default-activity-schedule-to-close`, sets the default activity schedule to close timeout, this is required otherwise temporal won't run your activity at all if it is left unspecified  (default `86400` which is 24h)

You can enable it in buf using:
```yaml
version: v2
plugins:
  - local: protoc-gen-go-tmprl
    out: gen
    opt:
    - paths=source_relative
    - gen-workflow-prefix=true
```


## Hacking on it

### Requirements

- [buf](https://buf.build) - Protocol buffer generation and management
- [direnv](https://direnv.net/) - Optional, loads env variables to add `bin` directory to `PATH`
- Go 1.21 or higher

### Build

All build artifacts are placed in the `bin/` directory to keep the repository clean.

```bash
# Build the plugin only
make build
# Output: bin/protoc-gen-go-tmprl

# Run all tests, build, regenerate examples, and verify (recommended)
make
# Output: bin/protoc-gen-go-tmprl, bin/example-worker, bin/example-client

# Run only unit tests with race detection and coverage
make test-unit

# Run all tests
make test

# Generate example code after changes
make gen

# Verify generated examples compile and build executables
make verify-examples
# Output: bin/example-worker, bin/example-client

# Clean all build artifacts
make clean
# Removes: bin/
```

**Build Artifacts:**
- `bin/protoc-gen-go-tmprl` - The protoc plugin binary
- `bin/example-worker` - Example worker application
- `bin/example-client` - Example client application

**Note:** The default `make` command automatically:
1. Runs unit tests with race detection
2. Generates temporal protobuf definitions
3. Builds the plugin binary into `bin/`
4. Regenerates example code
5. Builds and verifies all examples compile successfully into `bin/`

This ensures that any code changes don't break the generated output, and all artifacts are contained in the `bin/` directory.

### Architecture

The codebase follows a clean architecture pattern with clear separation of concerns:

- **internal/model/** - Domain models representing Temporal services, workflows, activities, signals, and queries
  - Built from protobuf definitions with recursive parent/child relationships
  - Handles options merging (method-level overrides service-level defaults)

- **internal/tmpl/** - Template helper functions for code generation
  - Type conversions, timeouts, retry policies
  - Naming conventions for workflows and objects

- **internal/renderer/** - Template execution engine
  - Uses Go's `text/template` package
  - Embeds template files for easy distribution
  - Renders constants, interfaces, clients, workers, and workflow objects

- **main.go** - Protoc plugin entry point
  - Parses protobuf service definitions
  - Delegates to domain models and renderer

### Code Generation Flow

1. Protoc invokes the plugin with protobuf descriptors
2. Plugin creates domain model objects from service definitions
3. Domain models merge service-level and method-level options
4. Renderer executes templates with domain model data
5. Generated Go code is written to output files

### Testing

```bash
# Run all tests with coverage
go test -race -cover ./internal/...

# Run specific package tests
go test ./internal/model/
go test ./internal/tmpl/
go test ./internal/renderer/
```

### Contributing

When contributing, ensure:
- All tests pass (`make` runs full build and test suite)
- Code follows existing patterns and conventions
- Unit tests are added for new functionality
- Generated example code still compiles and works
