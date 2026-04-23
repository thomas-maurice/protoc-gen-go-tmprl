// Package main is the CLI example for protoc-gen-go-tmprl. It wires the
// generated NewDieRollCLI entry point under a small Cobra root, adding
// an --address flag so users can point the client at a non-default
// Temporal frontend. The generator does not own connection configuration
// (see PLAN.md §15); this file shows the idiomatic way to plug the
// generated CLI into a caller-owned root command.
package main

import (
	"log/slog"
	"os"
	"time"

	examplev1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/example/v1"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
)

func main() {
	var address string
	var namespace string
	var logLevel string

	root := &cobra.Command{
		Use:   "example-cli",
		Short: "Example CLI built on top of the generated Cobra commands",
	}
	root.PersistentFlags().StringVar(&address, "address", client.DefaultHostPort, "Temporal frontend host:port.")
	root.PersistentFlags().StringVar(&namespace, "namespace", client.DefaultNamespace, "Temporal namespace.")
	root.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "Log level (debug, info, warn, error).")

	// Parse flags early so --log-level takes effect before the Temporal
	// client is constructed. Cobra normally defers flag parsing until
	// Execute runs the matched command; we need the value here.
	_ = root.ParseFlags(os.Args[1:])

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.WarnLevel
	}

	logger := slog.New(
		log.NewWithOptions(os.Stderr, log.Options{
			Level:           level,
			ReportTimestamp: true,
			TimeFormat:      time.RFC3339,
			Formatter:       log.TextFormatter,
		}),
	)

	// Construct the Temporal client lazily so --help works without a
	// running Temporal cluster. NewLazyClient defers the connection until
	// the first API call.
	c, err := client.NewLazyClient(client.Options{
		HostPort:  address,
		Namespace: namespace,
		Logger:    tlog.NewStructuredLogger(logger),
	})
	if err != nil {
		logger.Error("cannot create temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// Attach the generated service CLI as a subcommand. Multiple services
	// can be added the same way.
	root.AddCommand(examplev1.NewDieRollCLI(c))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
