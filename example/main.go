package main

import (
	"context"
	"fmt"
	"time"

	examplev1 "git.maurice.fr/thomas/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HelloWorldService struct {
	c *examplev1.HelloWorldClient
}

func (h *HelloWorldService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (h *HelloWorldService) SayHello(ctx context.Context, req *examplev1.HelloRequest) (*examplev1.HelloResponse, error) {
	fmt.Println("HELLO")
	return &examplev1.HelloResponse{
		Data: fmt.Sprintf("Hello %s!", req.Name),
	}, nil
}

func (h *HelloWorldService) SayMultipleHello(ctx workflow.Context, req *examplev1.MultipleHelloRequest) (*examplev1.MultipleHelloResponse, error) {
	resp := examplev1.MultipleHelloResponse{
		Data: make([]string, 0),
	}

	if req == nil {
		return &resp, nil
	}

	for _, i := range req.Names {
		f := workflow.ExecuteActivity(
			workflow.WithActivityOptions(
				ctx,
				workflow.ActivityOptions{
					StartToCloseTimeout: time.Minute * 1,
					ActivityID:          "foo",
				},
			),
			"hello.SayHello", &examplev1.HelloRequest{
				Name: fmt.Sprintf("%s", i),
			})
		var res *examplev1.HelloResponse
		err := f.Get(ctx, &res)
		if err != nil {
			return nil, err
		}
		resp.Data = append(resp.Data, res.Data)
	}
	return &resp, nil
}

func main() {
	c, err := client.NewLazyClient(client.Options{})
	if err != nil {
		panic(err)
	}

	helloClient, err := examplev1.NewHelloWorldClient(c)
	if err != nil {
		panic(err)
	}

	w, err := examplev1.NewHelloWorldWorker(c, &HelloWorldService{
		c: helloClient,
	}, "", worker.Options{})

	if err != nil {
		panic(err)
	}

	w.Register()

	go func() {
		err = w.Run()
		if err != nil {
			panic(err)
		}
	}()

	req := &examplev1.MultipleHelloRequest{
		Names: make([]string, 0),
	}

	for i := 0; i < 5; i++ {
		req.Names = append(req.Names, fmt.Sprintf("%d", i))
	}

	res, err := helloClient.ExecuteWorkflowSayMultipleHelloSync(context.Background(), req)
	if err != nil {
		panic(res)
	}

	for _, resp := range res.Data {
		fmt.Println(resp)
	}

	w.Stop()
}
