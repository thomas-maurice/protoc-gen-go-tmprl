package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"

	"github.com/charmbracelet/log"
	tlog "go.temporal.io/sdk/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

	ctx := context.Background()

	future, err := dieRollClient.ExecuteWorkflowThrowDies(ctx, &examplev1.ThrowDiesRequest{
		Results: 5,
	})

	if err != nil {
		logger.Error("could not execute workflow", "error", err)
		os.Exit(1)
	}

	time.Sleep(time.Second * 35)

	run := dieRollClient.GetThrowDiesFromRun(future)

	err = run.SignalContinue(ctx, &examplev1.ContinueSignalRequest{})

	if err != nil {
		logger.Error("could not send signal", "error", err)
		os.Exit(1)
	}

	until, err := dieRollClient.ExecuteWorkflowThrowUntilValue(ctx, &examplev1.ThrowUntilValueRequest{
		Value: 1,
	})

	if err != nil {
		logger.Error("cannot execute workflow", "error", err)
	}

	untilRun := dieRollClient.GetThrowUntilValueFromRun(until)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				quer, err := untilRun.QueryGetThrowsStatus(ctx, &emptypb.Empty{})
				if err != nil {
					logger.Error("could not query workflow", "error", err)
					os.Exit(1)
				}

				logger.Info("status update", "throws", quer.Throws)
			}
		}
	}()

	_, err = untilRun.Result(ctx)
	if err != nil {
		logger.Error("workflow failed", "error", err)
	}
}
