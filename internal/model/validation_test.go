package model

import (
	"math"
	"strings"
	"testing"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
)

// f32 and i32 are pointer helpers for building proto option structs.
func f32(v float32) *float32 { return &v }
func i32(v int32) *int32     { return &v }

// TestValidateRetryPolicy Regression for F7 (non-finite backoff -> uncompilable
// Go) and F13 (negative values). A non-finite backoff coefficient renders as
// +Inf/NaN which is valid Go syntax but does not compile, so it must be rejected
// at generation time rather than surfacing as a cryptic build error downstream.
func TestValidateRetryPolicy(t *testing.T) {
	cases := []struct {
		name    string
		rp      *temporalv1.RetryPolicy
		wantErr string // substring; "" means expect success
	}{
		{"nil is ok", nil, ""},
		{"finite backoff ok", &temporalv1.RetryPolicy{BackoffCoefficient: f32(2.0)}, ""},
		{"positive values ok", &temporalv1.RetryPolicy{InitialInterval: i32(1), MaximumInterval: i32(10), MaximumAttempts: i32(5)}, ""},
		{"zero maximum attempts ok", &temporalv1.RetryPolicy{MaximumAttempts: i32(0)}, ""},
		{"inf backoff rejected", &temporalv1.RetryPolicy{BackoffCoefficient: f32(float32(math.Inf(1)))}, "finite"},
		{"nan backoff rejected", &temporalv1.RetryPolicy{BackoffCoefficient: f32(float32(math.NaN()))}, "finite"},
		{"negative backoff rejected", &temporalv1.RetryPolicy{BackoffCoefficient: f32(-1)}, ">= 0"},
		{"negative initial interval rejected", &temporalv1.RetryPolicy{InitialInterval: i32(-1)}, "initial_interval"},
		{"negative maximum interval rejected", &temporalv1.RetryPolicy{MaximumInterval: i32(-1)}, "maximum_interval"},
		{"negative maximum attempts rejected", &temporalv1.RetryPolicy{MaximumAttempts: i32(-1)}, "maximum_attempts"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRetryPolicy(tc.rp, "ctx")
			assertErr(t, err, tc.wantErr)
		})
	}
}

// TestValidateWorkflowOptions Regression for F13: negative workflow timeouts are
// rejected.
func TestValidateWorkflowOptions(t *testing.T) {
	cases := []struct {
		name    string
		o       *temporalv1.WorkflowOptions
		wantErr string
	}{
		{"nil ok", nil, ""},
		{"positive ok", &temporalv1.WorkflowOptions{WorkflowExecutionTimeout: i32(3600)}, ""},
		{"negative execution timeout rejected", &temporalv1.WorkflowOptions{WorkflowExecutionTimeout: i32(-1)}, "workflow_execution_timeout"},
		{"negative run timeout rejected", &temporalv1.WorkflowOptions{WorkflowRunTimeout: i32(-1)}, "workflow_run_timeout"},
		{"negative task timeout rejected", &temporalv1.WorkflowOptions{WorkflowTaskTimeout: i32(-1)}, "workflow_task_timeout"},
		{"invalid retry policy propagated", &temporalv1.WorkflowOptions{RetryPolicy: &temporalv1.RetryPolicy{BackoffCoefficient: f32(float32(math.Inf(1)))}}, "finite"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertErr(t, validateWorkflowOptions(tc.o, "ctx"), tc.wantErr)
		})
	}
}

// TestValidateActivityOptions Regression for F13: negative activity timeouts are
// rejected.
func TestValidateActivityOptions(t *testing.T) {
	cases := []struct {
		name    string
		o       *temporalv1.ActivityOptions
		wantErr string
	}{
		{"nil ok", nil, ""},
		{"positive ok", &temporalv1.ActivityOptions{StartToCloseTimeout: i32(30)}, ""},
		{"negative schedule-to-start rejected", &temporalv1.ActivityOptions{ScheduleToStartTimeout: i32(-1)}, "schedule_to_start_timeout"},
		{"negative schedule-to-close rejected", &temporalv1.ActivityOptions{ScheduleToCloseTimeout: i32(-1)}, "schedule_to_close_timeout"},
		{"negative start-to-close rejected", &temporalv1.ActivityOptions{StartToCloseTimeout: i32(-1)}, "start_to_close_timeout"},
		{"negative heartbeat rejected", &temporalv1.ActivityOptions{HeartbeatTimeout: i32(-1)}, "heartbeat_timeout"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertErr(t, validateActivityOptions(tc.o, "ctx"), tc.wantErr)
		})
	}
}

// newWorkflow is a test helper that builds a minimal Workflow model with the
// given signal/query/update reference lists.
func newWorkflow(goName string, signals, queries, updates []string) *Workflow {
	return &Workflow{
		BaseMethod: BaseMethod{GoName: goName, RegisteredName: "pkg.Svc." + goName},
		Options:    &WorkflowOptions{Signals: signals, Queries: queries, Updates: updates},
	}
}

// TestServiceValidateWorkflowRefs Regression for F4: a workflow that references
// a signal/query/update name that is not defined in the service, or lists one
// twice, must fail generation loud instead of silently dropping the wrapper
// method (unknown name) or emitting a duplicate method (uncompilable).
func TestServiceValidateWorkflowRefs(t *testing.T) {
	base := func() *Service {
		return &Service{
			GoName:     "Svc",
			SignalsMap: map[string]*Signal{"Cancel": {}},
			QueriesMap: map[string]*Query{"Status": {}},
			UpdatesMap: map[string]*Update{"Reserve": {}},
		}
	}

	t.Run("all references resolve", func(t *testing.T) {
		s := base()
		s.Workflows = []*Workflow{newWorkflow("Proc", []string{"Cancel"}, []string{"Status"}, []string{"Reserve"})}
		if err := s.validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unknown signal rejected", func(t *testing.T) {
		s := base()
		s.Workflows = []*Workflow{newWorkflow("Proc", []string{"CancelTypo"}, nil, nil)}
		assertErr(t, s.validate(), "not a defined signal")
	})

	t.Run("unknown query rejected", func(t *testing.T) {
		s := base()
		s.Workflows = []*Workflow{newWorkflow("Proc", nil, []string{"Nope"}, nil)}
		assertErr(t, s.validate(), "not a defined query")
	})

	t.Run("unknown update rejected", func(t *testing.T) {
		s := base()
		s.Workflows = []*Workflow{newWorkflow("Proc", nil, nil, []string{"Nope"})}
		assertErr(t, s.validate(), "not a defined update")
	})

	t.Run("query name mislisted under signals rejected", func(t *testing.T) {
		s := base()
		s.Workflows = []*Workflow{newWorkflow("Proc", []string{"Status"}, nil, nil)}
		assertErr(t, s.validate(), "not a defined signal")
	})

	t.Run("duplicate signal entry rejected", func(t *testing.T) {
		s := base()
		s.Workflows = []*Workflow{newWorkflow("Proc", []string{"Cancel", "Cancel"}, nil, nil)}
		assertErr(t, s.validate(), "more than once")
	})
}

// TestServiceValidateDuplicateRegisteredNames Regression for F8: two workflows
// (or two activities) resolving to the same registered name would panic the
// worker at registration time, so generation must reject it.
func TestServiceValidateDuplicateRegisteredNames(t *testing.T) {
	t.Run("duplicate workflow registered name rejected", func(t *testing.T) {
		s := &Service{
			GoName: "Svc",
			Workflows: []*Workflow{
				{BaseMethod: BaseMethod{GoName: "A", RegisteredName: "same"}, Options: &WorkflowOptions{}},
				{BaseMethod: BaseMethod{GoName: "B", RegisteredName: "same"}, Options: &WorkflowOptions{}},
			},
		}
		assertErr(t, s.validate(), "register as")
	})

	t.Run("duplicate activity registered name rejected", func(t *testing.T) {
		s := &Service{
			GoName: "Svc",
			Activities: []*Activity{
				{BaseMethod: BaseMethod{GoName: "A", RegisteredName: "same"}},
				{BaseMethod: BaseMethod{GoName: "B", RegisteredName: "same"}},
			},
		}
		assertErr(t, s.validate(), "register as")
	})

	t.Run("distinct names ok", func(t *testing.T) {
		s := &Service{
			GoName: "Svc",
			Workflows: []*Workflow{
				{BaseMethod: BaseMethod{GoName: "A", RegisteredName: "a"}, Options: &WorkflowOptions{}},
				{BaseMethod: BaseMethod{GoName: "B", RegisteredName: "b"}, Options: &WorkflowOptions{}},
			},
		}
		if err := s.validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestPackageScopedNames Regression for F5: PackageScopedNames must list exactly
// the package-level identifiers that are NOT service-prefixed, so the plugin can
// detect cross-service collisions. Service-prefixed identifiers (client/worker
// types, task-queue const) must NOT appear.
func TestPackageScopedNames(t *testing.T) {
	s := &Service{
		GoName:     "Svc",
		Workflows:  []*Workflow{{BaseMethod: BaseMethod{GoName: "Run"}, Options: &WorkflowOptions{}}},
		Activities: []*Activity{{BaseMethod: BaseMethod{GoName: "Do"}}},
		Signals:    []*Signal{{BaseMethod: BaseMethod{GoName: "Ping"}}},
		Queries:    []*Query{{BaseMethod: BaseMethod{GoName: "Peek"}}},
		Updates:    []*Update{{BaseMethod: BaseMethod{GoName: "Set"}}},
	}

	got := s.PackageScopedNames()
	want := map[string]bool{
		"WorkflowRunName":              true,
		"ActivityDoName":               true,
		"SignalPingName":               true,
		"ReceiveSignalPing":            true,
		"ReceiveSignalPingAsync":       true,
		"QueryPeekName":                true,
		"HandleQueryPeek":              true,
		"UpdateSetName":                true,
		"HandleUpdateSet":              true,
		"HandleUpdateSetWithValidator": true,
	}

	gotSet := map[string]bool{}
	for _, n := range got {
		gotSet[n] = true
		if strings.Contains(n, "Svc") {
			t.Errorf("service-prefixed identifier %q must not be in the collision set", n)
		}
	}
	for w := range want {
		if !gotSet[w] {
			t.Errorf("expected %q in PackageScopedNames, missing", w)
		}
	}
	if len(got) != len(want) {
		t.Errorf("PackageScopedNames returned %d names, want %d: %v", len(got), len(want), got)
	}
}

// assertErr checks err against a wanted substring; wantSub == "" expects no
// error.
func assertErr(t *testing.T, err error, wantSub string) {
	t.Helper()
	if wantSub == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantSub)
	}
	if !strings.Contains(err.Error(), wantSub) {
		t.Fatalf("expected error containing %q, got: %v", wantSub, err)
	}
}
