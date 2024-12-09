package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/phsym/console-slog"
	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DieRollService struct {
	c *examplev1.DieRollClient
}

func (s *DieRollService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *DieRollService) ThrowDie(ctx context.Context, req *emptypb.Empty) (*examplev1.ThrowDieResponse, error) {
	return &examplev1.ThrowDieResponse{
		Result: int32(rand.Int() % 6),
	}, nil
}

func (s *DieRollService) ParentWorkflow(ctx workflow.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	var repetitions int
	repetitionsSideEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return rand.Int() % 20
	})

	err := repetitionsSideEffect.Get(&repetitions)
	if err != nil {
		return nil, err
	}

	for i := 0; i < repetitions; i++ {
		_, err = s.c.ExecuteChildChildWorkflow(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, nil
		}

		err = workflow.Sleep(ctx, time.Second*10)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (s *DieRollService) ChildWorkflow(ctx workflow.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *DieRollService) ThrowDies(ctx workflow.Context, req *examplev1.ThrowDiesRequest) (*examplev1.ThrowDiesResponse, error) {
	results := make([]int32, 0)

	for i := int32(0); i < req.Results; i++ {
		r, err := s.c.ExecuteActivityThrowDieSync(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, err
		}

		results = append(results, r.Result)

		workflow.Sleep(ctx, time.Second*5)
	}

	// This will let the workflow die
	examplev1.ReceiveSignalContinue(ctx)

	return &examplev1.ThrowDiesResponse{
		Results: results,
	}, nil
}

func (s *DieRollService) ThrowUntilValue(ctx workflow.Context, req *examplev1.ThrowUntilValueRequest) (*emptypb.Empty, error) {
	throws := int32(0)

	// Sends query updates
	examplev1.HandleQueryGetThrowsStatus(ctx, func(req *emptypb.Empty) (*examplev1.ThrowStatusResponse, error) {
		return &examplev1.ThrowStatusResponse{
			Throws: throws,
		}, nil
	})

	for {
		val, err := s.c.ExecuteActivityThrowDieSync(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, err
		}

		if val.Result == req.Value {
			break
		}

		throws++

		workflow.Sleep(ctx, time.Second*5)
	}

	return &emptypb.Empty{}, nil
}

func main() {
	logger := slog.New(
		console.NewHandler(os.Stderr, &console.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}),
	)

	c, err := client.NewLazyClient(client.Options{
		Logger: log.NewStructuredLogger(logger),
	})
	if err != nil {
		logger.Error("could not create temporal client", "error", err)
		os.Exit(1)
	}

	dieRollClient, err := examplev1.NewDieRollClient(c)
	if err != nil {
		logger.Error("could not create client", "error", err)
		os.Exit(1)
	}

	w, err := examplev1.NewDieRollWorker(
		c,
		&DieRollService{
			c: dieRollClient,
		},
		"",
		worker.Options{},
	)

	if err != nil {
		logger.Error("could not create worker", "error", err)
		os.Exit(1)
	}

	w.Register()

	err = w.Run()
	if err != nil {
		logger.Error("could not run worker", "error", err)
		os.Exit(1)
	}
}
