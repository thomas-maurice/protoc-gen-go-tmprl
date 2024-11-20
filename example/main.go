package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
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

	return &resp, nil
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

	res, err := helloClient.ExecuteWorkflowSayMultipleHelloSync(context.Background(), req)
	if err != nil {
		panic(res)
	}

	for _, resp := range res.Data {
		fmt.Println(resp)
	}

	futuresCount := 10

	wg := &sync.WaitGroup{}
	wg.Add(futuresCount)

	for i := 0; i < futuresCount; i++ {
		f, err := helloClient.ExecuteWorkflowSayMultipleHello(context.Background(), req)
		if err != nil {
			panic(err)
		}

		go func() {
			err := f.Get(context.Background(), nil)
			if err != nil {
				panic(err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	w.Stop()
}
