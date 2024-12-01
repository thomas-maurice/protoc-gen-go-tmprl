package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Defines the actual worker with a client to itself
type HelloWorldService struct {
	c *examplev1.HelloWorldClient
}

// The Ping activity
func (s *HelloWorldService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

// The SayHello activity
func (s *HelloWorldService) SayHello(ctx context.Context, req *examplev1.HelloRequest) (*examplev1.HelloResponse, error) {
	// Introduce random failures
	if rand.Int()%2 == 0 {
		time.Sleep(time.Second * 5)
		return nil, fmt.Errorf("randomly failed :(")
	}
	return &examplev1.HelloResponse{
		Data: fmt.Sprintf("Hello %s!", req.Name),
	}, nil
}

// A workflow that calls activities from that worker
func (s *HelloWorldService) SayMultipleHello(ctx workflow.Context, req *examplev1.MultipleHelloRequest) (*examplev1.MultipleHelloResponse, error) {
	firstTime := true
	err := examplev1.HandleQueryGetStatus(ctx, func(gsr *examplev1.GetStatusRequest) (*examplev1.GetStatusResponse, error) {
		fmt.Println("progress update requested")
		return &examplev1.GetStatusResponse{
			Progress: int64(rand.Int() % 100),
		}, nil
	})

	if err != nil {
		return nil, err
	}

	for firstTime {
		resp := examplev1.MultipleHelloResponse{
			Data: make([]string, 0),
		}

		if req == nil {
			return &resp, nil
		}

		for _, i := range req.Names {
			res, err := s.c.ExecuteActivitySayHelloSync(ctx, &examplev1.HelloRequest{Name: i})
			if err != nil {
				return nil, err
			}
			resp.Data = append(resp.Data, res.Data)
		}

		_, err := s.c.ExecuteChildSomeOtherWorkflowSync(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, err
		}

		r, ok := examplev1.ReceiveSignalContinue(ctx)
		if !ok {
			return &resp, nil
		}

		firstTime = r.Continue
	}
	return nil, nil
}

// A secondary workflow
func (s *HelloWorldService) SomeOtherWorkflow(ctx workflow.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	for i := 0; i < 10; i++ {
		_, err := s.c.ExecuteActivityPingSync(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, err
		}
	}
	return &emptypb.Empty{}, nil
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

	for i := 0; i < 10; i++ {
		req.Names = append(req.Names, fmt.Sprintf("%d", i))
	}

	future, err := helloClient.ExecuteWorkflowSayMultipleHello(context.Background(), req)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		qry, err := helloClient.QueryGetStatus(context.Background(), future.GetID(), future.GetRunID(), &examplev1.GetStatusRequest{})
		if err != nil {
			fmt.Println("failed to get status update: ", err)
		} else {
			fmt.Println("update: ", qry.Progress)
		}

		time.Sleep(time.Second)
	}

	var res *examplev1.MultipleHelloResponse
	err = future.Get(context.Background(), res)

	if err != nil {
		panic(err)
	}

	for _, resp := range res.Data {
		fmt.Println(resp)
	}

	fmt.Println("Stop")

	w.Stop()

	fmt.Println("Done")
}
