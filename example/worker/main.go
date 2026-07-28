// The worker side of the walkthrough. The actual service implementation
// (every workflow and activity) lives in example/orders; this binary only
// wires it to a Temporal server. Run it with:
//
//	go run ./example/worker
//
// then run the client (go run ./example/client) in another terminal. Keep an
// eye on this process' logs: the payment retries and the packing heartbeats
// show up here.
package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/thomas-maurice/protoc-gen-go-tmprl/example/orders"
	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
)

func main() {
	logger := slog.New(
		log.NewWithOptions(os.Stderr, log.Options{
			Level:           log.DebugLevel,
			ReportTimestamp: true,
			TimeFormat:      time.RFC3339,
			Formatter:       log.TextFormatter,
		}),
	)

	// TEMPORAL_ADDRESS overrides the target server, defaults to localhost:7233.
	c, err := client.NewLazyClient(client.Options{
		HostPort: os.Getenv("TEMPORAL_ADDRESS"),
		Logger:   tlog.NewStructuredLogger(logger),
	})
	if err != nil {
		logger.Error("could not create temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// The generated client is also usable from within workflow code: the
	// service implementation uses it to execute activities and child
	// workflows with the options declared in the proto.
	ordersClient, err := examplev1.NewOrdersClient(c)
	if err != nil {
		logger.Error("could not create orders client", "error", err)
		os.Exit(1)
	}

	// An empty task queue means the default one from the proto ("orders").
	w, err := examplev1.NewOrdersWorker(c, orders.New(ordersClient), "", worker.Options{})
	if err != nil {
		logger.Error("could not create worker", "error", err)
		os.Exit(1)
	}

	// Register registers every workflow and activity of the service with
	// their configured names, then the worker polls until interrupted.
	w.Register()

	logger.Info("worker started, waiting for orders", "task_queue", examplev1.DefaultOrdersTaskQueueName)
	if err := w.Run(worker.InterruptCh()); err != nil {
		logger.Error("could not run worker", "error", err)
		os.Exit(1)
	}
}
