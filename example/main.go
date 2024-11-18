package main

import (
	"context"
	"fmt"

	examplev1 "git.maurice.fr/thomas/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HelloWorldService struct{}

func (h *HelloWorldService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (h *HelloWorldService) SayHello(ctx context.Context, req *examplev1.HelloRequest) (*examplev1.HelloResponse, error) {
	return &examplev1.HelloResponse{
		Data: fmt.Sprintf("Hello %s!", req.Name),
	}, nil
}

func (h *HelloWorldService) SayMultipleHello(ctx workflow.Context, req *examplev1.MultipleHelloRequest) (*examplev1.MultipleHelloResponse, error) {
	resp := &examplev1.MultipleHelloResponse{
		Data: make([]string, 0),
	}

	for _, i := range req.Names {
		resp.Data = append(resp.Data, fmt.Sprintf("Hello %s!", i))
	}
	return nil, nil
}

func main() {
	c, err := client.NewLazyClient(client.Options{})
	if err != nil {
		panic(err)
	}

	_, err = examplev1.NewHelloWorldWorker(&c, &HelloWorldService{})
	if err != nil {
		panic(err)
	}
}
