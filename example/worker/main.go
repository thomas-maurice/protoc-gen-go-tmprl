package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/log"
	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DieRollService struct {
	c      *examplev1.DieRollClient
	client client.Client
}

func (s *DieRollService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *DieRollService) ThrowDie(ctx context.Context, req *emptypb.Empty) (*examplev1.ThrowDieResponse, error) {
	return &examplev1.ThrowDieResponse{
		Result: int32(rand.Int() % 6),
	}, nil
}

func (s *DieRollService) ParentWorkflow(ctx workflow.Context, req *emptypb.Empty) (*examplev1.ParentWorkflowReply, error) {
	var repetitions int
	repetitionsSideEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return rand.Int() % 10
	})

	err := repetitionsSideEffect.Get(&repetitions)
	if err != nil {
		return &examplev1.ParentWorkflowReply{Status: examplev1.Status_FAILURE}, err
	}

	workflow.Go(ctx, func(ctx workflow.Context) {
		more := true
		var sig *examplev1.ContinueSignalRequest
		for more {
			sig, more = examplev1.ReceiveSignalContinue(ctx)
			workflow.GetLogger(ctx).Info("Continue", fmt.Sprintf("%v", sig))
		}
	})

	for i := 0; i < repetitions; i++ {
		_, err := s.c.ExecuteChildChildWorkflow(ctx, &emptypb.Empty{})
		if err != nil {
			return &examplev1.ParentWorkflowReply{Status: examplev1.Status_SUCCESS}, nil
		}

		err = workflow.Sleep(ctx, time.Second*10)
		if err != nil {
			return &examplev1.ParentWorkflowReply{Status: examplev1.Status_FAILURE}, err
		}
	}

	return &examplev1.ParentWorkflowReply{Status: examplev1.Status_SUCCESS}, nil
}

func (s *DieRollService) ChildWorkflow(ctx workflow.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	info := workflow.GetInfo(ctx)

	err := s.c.SendSignalContinue(context.Background(), info.ParentWorkflowExecution.ID, info.ParentWorkflowExecution.RunID, &examplev1.ContinueSignalRequest{
		Continue: true,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *DieRollService) ThrowDies(ctx workflow.Context, req *examplev1.ThrowDiesRequest) (*examplev1.ThrowDiesResponse, error) {
	results := make([]int32, 0)

	for {
		// Throw the requested number of dice
		for i := int32(0); i < req.Results; i++ {
			r, err := s.c.ExecuteActivityThrowDieSync(ctx, &emptypb.Empty{})
			if err != nil {
				return nil, err
			}

			results = append(results, r.Result)

			if err := workflow.Sleep(ctx, time.Second*5); err != nil {
				return nil, err
			}
		}

		// If not looping, exit immediately
		if !req.Loop {
			break
		}

		// Wait for Continue signal with 1 minute timeout
		signalChan := workflow.GetSignalChannel(ctx, examplev1.SignalContinueName)
		selector := workflow.NewSelector(ctx)

		shouldContinue := false
		selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
			var sig examplev1.ContinueSignalRequest
			c.Receive(ctx, &sig)
			shouldContinue = sig.Continue
			workflow.GetLogger(ctx).Info("Received Continue signal", "continue", sig.Continue)
		})

		selector.AddFuture(workflow.NewTimer(ctx, time.Minute), func(f workflow.Future) {
			workflow.GetLogger(ctx).Info("No signal received within 1 minute, exiting workflow")
		})

		selector.Select(ctx)

		// Exit if signal said not to continue or timeout occurred
		if !shouldContinue {
			break
		}

		// Clear results for next iteration
		results = make([]int32, 0)
	}

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
		log.NewWithOptions(os.Stderr, log.Options{
			Level:           log.DebugLevel,
			ReportTimestamp: true,
			ReportCaller:    true,
			TimeFormat:      time.RFC3339,
			Formatter:       log.TextFormatter,
		}),
	)

	c, err := client.NewLazyClient(client.Options{
		Logger: tlog.NewStructuredLogger(logger),
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
			c:      dieRollClient,
			client: c,
		},
		"",
		worker.Options{},
	)
	if err != nil {
		logger.Error("could not create worker", "error", err)
		os.Exit(1)
	}

	w.Register()

	err = w.Run(worker.InterruptCh())
	if err != nil {
		logger.Error("could not run worker", "error", err)
		os.Exit(1)
	}
}
