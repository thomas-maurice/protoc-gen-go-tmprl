# protoc-gen-go-tmprl

Easily generate client and worker code for [temporal](https://temporal.io) from [protobuf](https://protobuf.dev) definitions

## Show me the code !

You will find a reference proto [here](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/proto/example/v1/example.proto)
and it corresponding generated code [there](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/gen/example/v1/example_tmprl.pb.go)
for reference purposes.

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

service HelloWorld  {
    option (temporal.v1.service) = {
        // this is not mandatory but will serve as a sane default
        task_queue: "hello_world"
    };
    // Just a simple ping
    rpc Ping(google.protobuf.Empty) returns (google.protobuf.Empty) {
        option (temporal.v1.activity) = {
            // you don't have to but you can also define default activity
            // options that will be applied every time the activity
            // is called
        };
    }

    // Say hello to multiple people
    rpc SayMultipleHello(MultipleHelloRequest) returns (MultipleHelloResponse) {
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
    // Throws dies a few times and return the result
    rpc ThrowDies(ThrowDiesRequest) returns (ThrowDiesResponse) {
        option (temporal.v1.workflow) = {
            signals: ["Continue"]
        };
    }
```

The generator creates the following schedule methods for every workflow in the service:

```golang
// CreateScheduleThrowDies creates a schedule for ThrowDies. The caller supplies the cadence via
// options[0].Spec (CronExpressions / Intervals / Calendars). Non-zero fields from options[0] win
// over generator defaults; if TaskQueue is left empty the service's default task queue is used.
func (c *DieRollClient) CreateScheduleThrowDies(
    ctx context.Context,
    scheduleID string,
    req *ThrowDiesRequest,
    options ...client.ScheduleOptions,
) (client.ScheduleHandle, error)

// GetScheduleThrowDies gets a handle to an existing schedule for ThrowDies.
func (c *DieRollClient) GetScheduleThrowDies(
    ctx context.Context,
    scheduleID string,
) client.ScheduleHandle

// DeleteScheduleThrowDies deletes an existing schedule for ThrowDies.
// Returns the underlying client error if the schedule does not exist or the delete fails.
func (c *DieRollClient) DeleteScheduleThrowDies(
    ctx context.Context,
    scheduleID string,
) error

// ListScheduleThrowDies lists all schedules in the namespace whose action is the ThrowDies
// workflow type. pageSize is forwarded to the Temporal ScheduleClient.List call.
func (c *DieRollClient) ListScheduleThrowDies(
    ctx context.Context,
    pageSize int,
) ([]client.ScheduleListEntry, error)

// UpsertScheduleThrowDies creates the schedule if it does not exist, otherwise performs an
// in-place update (Spec, Action, Overlap, CatchupWindow, PauseOnFailure, TypedSearchAttributes).
// A user-supplied Paused flag is deliberately not honoured on the update path: a bool can't
// distinguish "leave it alone" from "unpause", so run-state transitions are expressed through
// PauseScheduleThrowDies / UnpauseScheduleThrowDies instead.
func (c *DieRollClient) UpsertScheduleThrowDies(
    ctx context.Context,
    scheduleID string,
    req *ThrowDiesRequest,
    options ...client.ScheduleOptions,
) (client.ScheduleHandle, error)

// PauseScheduleThrowDies pauses a running schedule. The note is recorded on the schedule's
// audit trail. Describe is called first and the Pause RPC is skipped when the schedule is
// already paused, so this is safe to call on every reconcile tick.
func (c *DieRollClient) PauseScheduleThrowDies(
    ctx context.Context,
    scheduleID string,
    note string,
) error

// UnpauseScheduleThrowDies resumes a paused schedule. Describe is called first and the
// Unpause RPC is skipped when the schedule is already running.
func (c *DieRollClient) UnpauseScheduleThrowDies(
    ctx context.Context,
    scheduleID string,
    note string,
) error
```

Example usage:

```golang
// Create a schedule. The caller owns the cadence: supply a Spec.
scheduleHandle, err := dieRollClient.CreateScheduleThrowDies(
    ctx,
    "throw-dies-schedule",
    &ThrowDiesRequest{Results: 3, Loop: false},
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
scheduleHandle = dieRollClient.GetScheduleThrowDies(ctx, "throw-dies-schedule")

// Use the schedule handle to pause, unpause, describe, etc.
err = scheduleHandle.Pause(ctx, client.SchedulePauseOptions{
    Note: "Pausing for maintenance",
})

// Idempotent create-or-update. Safe to call on redeploys; the Paused flag is intentionally
// ignored on the update path.
_, err = dieRollClient.UpsertScheduleThrowDies(
    ctx,
    "throw-dies-schedule",
    &ThrowDiesRequest{Results: 5},
    client.ScheduleOptions{
        Spec: client.ScheduleSpec{CronExpressions: []string{"* * * * *"}},
    },
)

// Enumerate every schedule whose action is this workflow type.
entries, err := dieRollClient.ListScheduleThrowDies(ctx, 100)

// Tear a schedule down.
err = dieRollClient.DeleteScheduleThrowDies(ctx, "throw-dies-schedule")

// Toggle run state. Each of these is a read-then-write: a Describe round-trip
// is always paid, but the Pause/Unpause RPC is skipped when the schedule is
// already in the target state.
err = dieRollClient.PauseScheduleThrowDies(ctx, "throw-dies-schedule", "maintenance window")
err = dieRollClient.UnpauseScheduleThrowDies(ctx, "throw-dies-schedule", "maintenance over")
```

**Features:**
- **Type-safe:** Generated methods use workflow-specific request types
- **Option merging:** Runtime `client.ScheduleOptions` are merged field-by-field onto generator defaults (non-zero fields win); the merge logic is factored into a per-service `mergeScheduleOptions<Service>` helper in the generated file
- **Workflow configuration:** Automatically applies workflow timeouts and retry policies to scheduled executions
- **Unconditional generation:** Every workflow gets the schedule surface — cadence is a runtime concern, not part of the schema
- **Full CRUD surface:** `Create`, `Get`, `List`, `Upsert`, `Delete`, `Pause`, and `Unpause` are all generated per workflow

### Generating a CLI

Opt a service into CLI generation and the plugin emits a `NewXxxCLI(client.Client) *cobra.Command` entry point alongside the client and worker code. The returned command wraps every workflow / signal / query / schedule helper the service already has, with semantic per-field flags rather than JSON blobs — so you can start, signal, query, cancel, terminate, and manage schedules from the shell without writing glue.

#### Opt in

```protobuf
service OrdersService {
    option (temporal.v1.service) = {
        task_queue: "orders"
        generate_cli: true
    };

    rpc Submit(SubmitRequest) returns (SubmitResponse) {
        option (temporal.v1.workflow) = {};
    }
}
```

Per-workflow opt-out for inputs that cannot reasonably be expressed as flags:

```protobuf
    rpc Complex(ComplexInput) returns (ComplexOutput) {
        option (temporal.v1.workflow) = {
            skip_cli: true
        };
    }
```

#### Wiring it into your binary

The generator never emits `main.go` — you own the `client.Client` lifecycle and connection flags. A minimal wrapper:

```go
package main

import (
    "fmt"
    "os"

    ordersv1 "example.com/gen/orders/v1"

    "github.com/spf13/cobra"
    "go.temporal.io/sdk/client"
)

func main() {
    root := &cobra.Command{Use: "myctl"}
    c, err := client.NewLazyClient(client.Options{})
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    defer c.Close()
    root.AddCommand(ordersv1.NewOrdersServiceCLI(c))
    _ = root.Execute()
}
```

#### Command tree

```
myctl
`-- orders-service
    |-- workflow
    |   |-- start     <workflow>
    |   |-- execute   <workflow>
    |   |-- signal    <workflow> <signal>
    |   |-- query     <workflow> <query>
    |   |-- cancel
    |   |-- terminate
    |   `-- describe
    `-- schedule
        |-- create    <workflow>
        |-- pause
        |-- unpause
        |-- delete
        `-- describe
```

#### Proto-type to CLI-flag mapping

Every flag's help text explicitly states the expected input shape — you never have to guess whether a list is comma-separated, JSON, or repeatable. Hints below match the generator's output character-for-character.

| Proto type | CLI flag kind | Example |
| --- | --- | --- |
| `string` | String | `--customer-id ACME-1` |
| `int32` / `sint32` / `sfixed32` | Int32 | `--quantity 5` |
| `int64` / `sint64` / `sfixed64` | Int64 | `--total-cents 1299` |
| `uint32` / `fixed32` | Uint32 | `--retries 3` |
| `uint64` / `fixed64` | Uint64 | `--cursor 42` |
| `float` | Float32 | `--rate 1.5` |
| `double` | Float64 | `--ratio 0.125` |
| `bool` | Bool | `--express` / `--express=false` |
| `bytes` | Bytes (base64) | `--payload aGVsbG8=` |
| `enum` | Enum (case-insensitive) | `--shipping-speed express` |
| `google.protobuf.Duration` | Duration | `--timeout 5m` / `1h30m` |
| `google.protobuf.Timestamp` | Timestamp (RFC3339) | `--scheduled-at 2025-01-02T15:04:05Z` |
| `google.protobuf.FieldMask` | StringSlice | `--update-mask name,address.city` |
| `google.protobuf.StringValue` and sibling wrappers | Underlying scalar, presence-gated | `--wrapped-string hi` |
| `google.protobuf.Any` / `Struct` / `Value` | JSON escape | `--payload '<JSON value>'` |
| `google.protobuf.Empty` | _(no flag emitted)_ | — |
| nested `message` | recurses, child flags prefixed (dot-separated per nesting level, dashes within a single snake_case name) | `--shipping-address.street ...` |
| `repeated string` | StringSlice (repeatable) | `--tag foo --tag bar` |
| `repeated int32` | Int32Slice (repeatable) | `--count 1 --count 2` |
| `repeated int64` / `uint32` / `uint64` | Int64Slice (repeatable) | `--id 10 --id 20` |
| `repeated bool` / `float` / `double` | StringSlice (repeatable) | `--flag true --flag false` |
| `repeated <Message>` | JSON escape | `--items '[{"sku":"x","qty":1}]'` |
| `map<string, string>` | StringToString (repeatable) | `--labels env=prod --labels team=a` |
| `map<string, int*>` | StringToInt (repeatable) | `--weights a=1 --weights b=2` |
| `map<string, <Message>>` / non-string key | JSON escape | `--by-id '{"x":{"qty":1}}'` |
| `oneof` branches | scalar flags, mutually exclusive | `--payment.card-number ...` XOR `--payment.bank-account ...` |
| Cyclic / self-referencing subtree | JSON escape at the seam | `--node '<JSON value of type Node>'` |

The `schedule create <workflow>` command additionally accepts the schedule flags `--schedule-id` (required), `--schedule-cron`, `--schedule-interval`, `--schedule-start-at`, `--schedule-end-at`, `--schedule-timezone`, `--schedule-jitter`, `--schedule-paused`, `--schedule-note`, `--schedule-overlap`, `--schedule-catchup-window`, and `--schedule-task-queue`.

#### Opt-out

Set `skip_cli: true` on `temporal.v1.workflow` when an input message is too tangled to flag-ify (heavy nesting with `Any`/`Struct`, very deep polymorphism, etc.). The workflow keeps its client/worker surface; only the `start`, `execute`, `signal`, `query`, and `schedule create` subcommands for it are skipped.

#### Not generated

- A `main.go` binary. Wire `NewXxxCLI` into your own Cobra root.
- Connection flags (`--address`, `--namespace`, TLS). These belong to the caller; see `example/cli/main.go` for how to add them.
- `workflow list` / generic search. There is no generated client equivalent to wrap.
- Direct activity invocation. Activities are invoked by workflows, not by end users.
- Structured calendar schedule specs. Users who need calendars build the `client.ScheduleOptions.Spec` in Go.
- Interactive prompts. Everything comes from flags.

### The workflow objects

Each workflow will get assigned a dedicated object in the generated code. All the workflow objects implement the `internal.WorkflowRun`
interface from the Temporal SDK. They contain a few methods that can be useful for you. Let's take as an example the following
protobuf:
```protobuf
    // Say hello to multiple people
    rpc SayMultipleHello(MultipleHelloRequest) returns (MultipleHelloResponse) {
        option (temporal.v1.workflow) = {};
    }
```

The following methods will be generated for the workflow object:

```golang
// Cancel cancels a given workflow
func (w *HelloWorldSayMultipleHello) Cancel(ctx context.Context) error
// Returns the workflow ID
func (w *HelloWorldSayMultipleHello) GetID() string
// Returns the run ID
func (w *HelloWorldSayMultipleHello) GetRunID() string
// Terminates terminates a given workflow
func (w *HelloWorldSayMultipleHello) Terminate(ctx context.Context, reason string, details ...interface{})
// Get gets the result of a given workflow with its native type
func (w *HelloWorldSayMultipleHello) Result(ctx context.Context) (*MultipleHelloResponse, error)
// ResultWithOptions gets the result of a given workflow with its native type
func (w *HelloWorldSayMultipleHello) ResultWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (*MultipleHelloResponse, error)
// Get gets the result of a given workflow with pointers -- discouraged to use but required to implement internal.WorkflowRun
func (w *HelloWorldSayMultipleHello) Get(ctx context.Context, valuePtr interface{}) error
// Get gets the result of a given workflow with pointers -- discouraged to use but required to implement internal.WorkflowRun
func (w *HelloWorldSayMultipleHello) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error
```

You can retrieve this `HelloWorldSayMultipleHello` object from the client using one of these two methods:
```golang
func (c *HelloWorldClient) GetSayMultipleHello(ctx context.Context, workflowId string, runId string) *HelloWorldSayMultipleHello
func (c *HelloWorldClient) GetSayMultipleHelloFromRun(future client.WorkflowRun) *HelloWorldSayMultipleHello
```

#### Workflow object signal and queries
Additionally, if you have defined signal and queries in your workflow options like in the following protobuf
```protobuf
    rpc SayMultipleHello(MultipleHelloRequest) returns (MultipleHelloResponse) {
        option (temporal.v1.workflow) = {
            signals: ["Continue"]
            queries: ["GetStatus"]
        };
    }
```

Then you will have access to the two following methods:

```golang
// SignalContinue sends the Continue signal to the workflow
func (w *HelloWorldSayMultipleHello) SignalContinue(ctx context.Context, req *ContinueSignalRequest) error
// QueryGetStatus queries the workflow with GetStatus
func (w *HelloWorldSayMultipleHello) QueryGetStatus(ctx context.Context, req *GetStatusRequest) (*GetStatusResponse, error)
```

:warning: The name you pass to the protobuf must match the name of the generated go name for the signal, i.e. `some_func` would
become `SomeFunc`

:warning: The methods you set as signals and queries MUST be defined inside the service whose workflow uses them. You cannot use
the signals/queries defined in `Service2` for `Service1`, you can reuse types, not methods.

### Signals and queries

You can define signal and queries functions in your service, if they are annotated with the respective `temporal.v1.signal` and
`temporal.v1.query` options they will be treated as such. For example if we have a signal like so:

```protobuf
    rpc Continue(ContinueSignalRequest) returns (google.protobuf.Empty) {
        option (temporal.v1.signal) = {};
    }
```

You will have access to the three following methods:

```golang
// SendSignalContinue sends the Continue signal to a workflow
// This can be called from a workflow or externally
func (c *ServiceClient) SendSignalContinue(ctx context.Context, workflowID string, runID string, req *ContinueSignalRequest) error

// ReceiveSignalContinue waits for the the Continue signal
// This is called within a workflow exclusively
func ReceiveSignalContinue(ctx workflow.Context) (*ContinueSignalRequest, bool)

// ReceiveSignalContinueAsync recieves the the Continue signal asynchronously.
// It doesn't wait if there is no signal in the queue.
// This is called within a workflow exclusively
func ReceiveSignalContinueAsync(ctx workflow.Context) (*ContinueSignalRequest, bool)
```

:warning: Whatever you put in the response parameter of the signal does not matter at all and
will be ignored by the code generator, as you want to send and recieve the same object.

For queriees it is very similar, let's take for example the following query:
```protobuf
    rpc GetStatus(GetStatusRequest) returns (GetStatusResponse) {
        option (temporal.v1.query) = {};
    }
```

This will grant you the following two methods:

```golang
// QueryGetStatus sends the GetStatus query to a workflow
// This can be called from a workflow or externally
func (c *ServiceClient) QueryGetStatus(ctx context.Context, workflowID string, runID string, req *GetStatusRequest) (*GetStatusResponse, error)

// HandleQueryGetStatus sets up the GetStatus query and responds accordingly, returns an error if it failed
// This is called within your workflow to setup the handler method
func HandleQueryGetStatus(ctx workflow.Context, queryFunc func(req *GetStatusRequest) (*GetStatusResponse, error)) error
```

### Child workflow executions
You get access to a similar API with the child workflows executions, something like so
```golang
func (c *HelloWorldClient) GetChildHelloWorldSayMultipleHelloExecution(future workflow.ChildWorkflowFuture) *ChildHelloWorldSayMultipleHelloExecution
```

However the API is a bit more limited (no way to query for example) because it is a wrapper around the `internal.ChildWorkflowExecution` object.
The method type you'd be interested in are the ones that allow you to signal the child workflow, for example:

```golang
func (w *ChildHelloWorldSayMultipleHelloExecution) SignalContinue(ctx workflow.Context, req *ContinueSignalRequest) error
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
* `client.GetX`: Gets an instance of a workflow
* `workflow.Cancel`: Cancels a workflow
* `workflow.Teminate`: Terminates a workflow
* `workflow.Get`: Gets the result of a workflow like you would on a normal future (you probably don't want that because no type safety)
* `workflow.Result`: Gets the result of a workflow *with type safety*

Generally a good starting point to get familiar with the generated code is to have a look at the [example client](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/client/main.go) and [example worker](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/worker/main.go) provided.

## Options
* `gen-workflow-prefix`, if set to true, instead of using an UUID for workflow IDs, the worker will generate a name that looks like `<module>.v<X>.<service>.<rpcMethodName>/<uuid>`, like `example.v1.DieRoll.ThrowDies/e2715d07-7bc0-495d-90c5-c396c0a17b46` for example.
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
