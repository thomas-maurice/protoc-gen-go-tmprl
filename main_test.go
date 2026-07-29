package main

import (
	"strings"
	"testing"

	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/model"
)

// TestCheckServiceCollisions Regression for F5: two temporal services in one
// proto file that share a workflow/activity/signal/query/update method name emit
// the same package-level identifier, which redeclares in the generated Go and
// fails the consumer's build with no diagnostic from the plugin. check
// ServiceCollisions must catch this at generation time.
func TestCheckServiceCollisions(t *testing.T) {
	svc := func(name, method string) *model.Service {
		return &model.Service{
			GoName:    name,
			Workflows: []*model.Workflow{{BaseMethod: model.BaseMethod{GoName: method}, Options: &model.WorkflowOptions{}}},
			Signals:   []*model.Signal{{BaseMethod: model.BaseMethod{GoName: "Ping"}}},
		}
	}

	t.Run("distinct method names do not collide", func(t *testing.T) {
		// Same signal name "Ping" in both -> collision expected, so use distinct
		// services with no shared package-scoped names.
		a := &model.Service{GoName: "Alpha", Workflows: []*model.Workflow{{BaseMethod: model.BaseMethod{GoName: "RunA"}, Options: &model.WorkflowOptions{}}}}
		b := &model.Service{GoName: "Beta", Workflows: []*model.Workflow{{BaseMethod: model.BaseMethod{GoName: "RunB"}, Options: &model.WorkflowOptions{}}}}
		if err := checkServiceCollisions([]*model.Service{a, b}, "x.proto"); err != nil {
			t.Fatalf("unexpected collision: %v", err)
		}
	})

	t.Run("shared workflow name collides", func(t *testing.T) {
		a := &model.Service{GoName: "Alpha", Workflows: []*model.Workflow{{BaseMethod: model.BaseMethod{GoName: "Run"}, Options: &model.WorkflowOptions{}}}}
		b := &model.Service{GoName: "Beta", Workflows: []*model.Workflow{{BaseMethod: model.BaseMethod{GoName: "Run"}, Options: &model.WorkflowOptions{}}}}
		err := checkServiceCollisions([]*model.Service{a, b}, "x.proto")
		if err == nil {
			t.Fatal("expected a collision error for shared workflow name")
		}
		if !strings.Contains(err.Error(), "WorkflowRunName") || !strings.Contains(err.Error(), "Alpha") || !strings.Contains(err.Error(), "Beta") {
			t.Fatalf("error should name the identifier and both services, got: %v", err)
		}
	})

	t.Run("shared signal name collides", func(t *testing.T) {
		err := checkServiceCollisions([]*model.Service{svc("Alpha", "RunA"), svc("Beta", "RunB")}, "x.proto")
		if err == nil {
			t.Fatal("expected a collision error for shared signal name")
		}
		if !strings.Contains(err.Error(), "SignalPingName") {
			t.Fatalf("expected SignalPingName collision, got: %v", err)
		}
	})

	t.Run("single service never collides with itself", func(t *testing.T) {
		if err := checkServiceCollisions([]*model.Service{svc("Alpha", "Run")}, "x.proto"); err != nil {
			t.Fatalf("unexpected collision for a single service: %v", err)
		}
	})
}
