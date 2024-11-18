# protoc-gen-go-tmprl
An attempt at generating code for your temporal workflows

## What it does

You will find a reference proto [here](https://git.maurice.fr/thomas/protoc-gen-go-tmprl/src/branch/master/example/proto/example/v1/example.proto)
and it corresponding generated code [there](https://git.maurice.fr/thomas/protoc-gen-go-tmprl/src/branch/master/gen/example/v1/example_tprl.pb.go)
for reference purposes.

## How to use it

You need to include [temporal.v1](https://git.maurice.fr/thomas/protoc-gen-go-tmprl/src/branch/master/proto/temporal/v1/temporal.proto) in your
project so you can start defining actions and workflows in your services like so

```protobuf
syntax = "proto3";

package example.v1;

// This is where buf will generate your protobuf go code
option go_package = "git.maurice.fr/thomas/protoc-gen-go-tmprl/gen/example/v1";

import "temporal/v1/temporal.proto";
import "google/protobuf/empty.proto";

service HelloWorld  {
    option (temporal.v1.service) = {};

    // Just a simple ping
    rpc Ping(google.protobuf.Empty) returns (google.protobuf.Empty) {
        option (temporal.v1.activity) = {};
    }

    // Say hello to multiple people
    rpc SayMultipleHello(MultipleHelloRequest) returns (MultipleHelloResponse) {
        option (temporal.v1.workflow) = {};
    }
}
```

Easy as that

## Getting started

You need to install [buf](https://buf.build) to get started, it's a more pleasant experience when
generating protobufs.

## Build

```
$ make build
# then subsequently
$ make
```

Should be sufficient

## Run

You need [direnv](https://direnv.net/) to load some env variables into your shell. This is required to add the `bin` directory to the `PATH`

Good luck
