# protoc-gen-go-tmprl

Easily generate client and worker code for [temporal](https://temporal.io) from [protobuf](https://protobuf.dev) definitions

## Show me the code !

You will find a reference proto [here](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/example/proto/example/v1/example.proto)
and it corresponding generated code [there](https://github.com/thomas-maurice/protoc-gen-go-tmprl/blob/master/gen/example/v1/example_tprl.pb.go)
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

### The exposed API

The generated code exposes a lot of primitives such as (non exhaustive list):

* `client.ExecuteWorkflowX`: Executes a workflow and returns a future
* `client.ExecuteWorkflowXSync`: Executes a workflow and blocks until the result is returned
* `client.ExecuteActivityX`: Executes an activity and returns a future
* `client.ExecuteActivityXSync`: Executes an activity and blocks until the result is returned
* `client.GetWorkflowX`: Gets an instance of a workflow
* `workflow.Cancel`: Cancels a workflow
* `workflow.Teminate`: Terminates a workflow
* `workflow.Get`: Gets the result of a workflow

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
