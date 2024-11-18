# protoc-gen-go-tmprl
An attempt at generating code for your temporal workflows

## What it does

You will find a reference proto [here](https://git.maurice.fr/thomas/protoc-gen-go-tmprl/src/branch/master/example/proto/example/v1/example.proto)
and it corresponding generated code [there](https://git.maurice.fr/thomas/protoc-gen-go-tmprl/src/branch/master/gen/example/v1/example_tprl.pb.go)
for reference purposes.

## Getting started

You need to install [buf](https://buf.build) to get started, it's a more pleasant experience when
generating protobufs.

## Build

```
$ make
```

Should be sufficient

## Run

You need [direnv](https://direnv.net/) to load some env variables into your shell. This is required to add the `bin` directory to the `PATH`

Good luck
