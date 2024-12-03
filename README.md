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

### Default workflow & activity setups

You can setup a service level (here a service refers to a worker) default for activities and workflows respectively in the `default_activity_options` and `default_workflow_options` fields of the
`temporal.v1.service` option.

If you don't want to do that, it's fine, yhou can define these setups at the individual workflow or activity level doing something like this

```protobuf
    // Some activity
    rpc SomeActivity(SomeRequest) returns (SomeResponse) {
        option (temporal.v1.activity) = {
            schedule_to_start_timeout: { value: 30 }
            schedule_to_close_timeout: { value: 120 }
            start_to_close_timeout: { value: 120 }
            retry_policy: {
                initial_interval: { value: 1 }
                backoff_coefficient: { value: 1.5 }
                maximum_interval: { value: 10 }
                maximum_attempts: { value: 10 }
                non_retryable_error_types: [{value: "FATAL"}]
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
            workflow_execution_timeout: { value: 86400 }
            // one hour
            workflow_run_timeout: { value: 7200 }
        };
    }
```

You might have noticed that the options are wrapped in `{ value: something }` statements. While it might look ugly it is the only way for us to know if something is set to `0` (whatever `0` means, let it be numerical `0` or an empty string) on purpose, or not set.

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
* `client.GetX`: Gets an instance of a workflow
* `workflow.Cancel`: Cancels a workflow
* `workflow.Teminate`: Terminates a workflow
* `workflow.Get`: Gets the result of a workflow like you would on a normal future (you probably don't want that because no type safety)
* `workflow.Result`: Gets the result of a workflow *with type safety*

Generally a good starting point to get familiar with the generated code is to have a look at the [example](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/proto/example/main.go) provided.

## Hacking on it
### Install `buf`

You need to install [buf](https://buf.build) to get started, it's a more pleasant experience when
generating protobufs.

### Build

You need [direnv](https://direnv.net/) to load some env variables into your shell. This is required to add the `bin` directory to the `PATH`


```
$ make build
$ make
```

Should be sufficient

Good luck, have fun
