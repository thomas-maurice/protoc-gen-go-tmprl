package model

import (
	"testing"
	"time"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
)

// TestServiceGetClientName Tests client name generation
func TestServiceGetClientName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "ExampleServiceClient"
	if result := service.GetClientName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetWorkerName Tests worker name generation
func TestServiceGetWorkerName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "ExampleServiceWorker"
	if result := service.GetWorkerName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetServiceInterfaceName Tests service interface name generation
func TestServiceGetServiceInterfaceName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "ExampleServiceService"
	if result := service.GetServiceInterfaceName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetDefaultTaskQueueConstName Tests default task queue constant name generation
func TestServiceGetDefaultTaskQueueConstName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "DefaultExampleServiceTaskQueueName"
	if result := service.GetDefaultTaskQueueConstName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetDefaultActivityTimeoutConstName Tests default activity timeout constant name generation
func TestServiceGetDefaultActivityTimeoutConstName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "DefaultExampleServiceActivityScheduleToCloseTimeout"
	if result := service.GetDefaultActivityTimeoutConstName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetSignal Tests signal retrieval from map
func TestServiceGetSignal(t *testing.T) {
	signal := &Signal{
		BaseMethod: BaseMethod{
			GoName: "TestSignal",
		},
	}

	service := &Service{
		GoName: "TestService",
		SignalsMap: map[string]*Signal{
			"TestSignal": signal,
		},
	}

	t.Run("existing signal", func(t *testing.T) {
		result, err := service.GetSignal("TestSignal")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != signal {
			t.Error("expected to get the same signal instance")
		}
	})

	t.Run("non-existing signal", func(t *testing.T) {
		_, err := service.GetSignal("NonExisting")
		if err == nil {
			t.Error("expected error for non-existing signal")
		}
	})
}

// TestServiceGetQuery Tests query retrieval from map
func TestServiceGetQuery(t *testing.T) {
	query := &Query{
		BaseMethod: BaseMethod{
			GoName: "TestQuery",
		},
	}

	service := &Service{
		GoName: "TestService",
		QueriesMap: map[string]*Query{
			"TestQuery": query,
		},
	}

	t.Run("existing query", func(t *testing.T) {
		result, err := service.GetQuery("TestQuery")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != query {
			t.Error("expected to get the same query instance")
		}
	})

	t.Run("non-existing query", func(t *testing.T) {
		_, err := service.GetQuery("NonExisting")
		if err == nil {
			t.Error("expected error for non-existing query")
		}
	})
}

// TestServiceGetUpdate Tests update retrieval from map
func TestServiceGetUpdate(t *testing.T) {
	update := &Update{
		BaseMethod: BaseMethod{
			GoName: "TestUpdate",
		},
	}

	service := &Service{
		GoName: "TestService",
		UpdatesMap: map[string]*Update{
			"TestUpdate": update,
		},
	}

	t.Run("existing update", func(t *testing.T) {
		result, err := service.GetUpdate("TestUpdate")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != update {
			t.Error("expected to get the same update instance")
		}
	})

	t.Run("non-existing update", func(t *testing.T) {
		_, err := service.GetUpdate("NonExisting")
		if err == nil {
			t.Error("expected error for non-existing update")
		}
	})
}

// TestServiceScheduleHelperNames Verifies the per-service schedule helper naming
func TestServiceScheduleHelperNames(t *testing.T) {
	s := &Service{GoName: "DieRoll"}
	if got, want := s.GetScheduleMergeFuncName(), "mergeScheduleOptionsDieRoll"; got != want {
		t.Errorf("GetScheduleMergeFuncName() = %q, want %q", got, want)
	}
	if got, want := s.GetScheduleDefaultsFuncName(), "applyScheduleDefaultsDieRoll"; got != want {
		t.Errorf("GetScheduleDefaultsFuncName() = %q, want %q", got, want)
	}
}

// TestServiceResolvedDefaultActivityOptions guards the doc-renderer regression
// where the "Default activity options" table rendered every timeout as 0s: the
// proto's *int32 second fields must surface to the template as time.Duration so
// FormatDuration humanizes them, matching what the code generator bakes in.
func TestServiceResolvedDefaultActivityOptions(t *testing.T) {
	i32 := func(v int32) *int32 { return &v }

	if got := (&Service{}).ResolvedDefaultActivityOptions(); got != nil {
		t.Fatalf("ResolvedDefaultActivityOptions() with no defaults = %+v, want nil", got)
	}

	s := &Service{
		DefaultActivityOptions: &temporalv1.ActivityOptions{
			ScheduleToStartTimeout: i32(120),
			ScheduleToCloseTimeout: i32(3600),
		},
	}
	got := s.ResolvedDefaultActivityOptions()
	if got == nil {
		t.Fatal("ResolvedDefaultActivityOptions() = nil, want resolved options")
	}
	if got.ScheduleToStartTimeout != 2*time.Minute {
		t.Errorf("ScheduleToStartTimeout = %s, want 2m0s", got.ScheduleToStartTimeout)
	}
	if got.ScheduleToCloseTimeout != time.Hour {
		t.Errorf("ScheduleToCloseTimeout = %s, want 1h0m0s", got.ScheduleToCloseTimeout)
	}
	// Fields the service left unset stay zero so the template omits their rows.
	if got.StartToCloseTimeout != 0 || got.HeartbeatTimeout != 0 {
		t.Errorf("unset timeouts = (%s, %s), want (0s, 0s)", got.StartToCloseTimeout, got.HeartbeatTimeout)
	}
}

// TestServiceResolvedDefaultWorkflowOptions is the workflow-side counterpart to
// TestServiceResolvedDefaultActivityOptions: the "Default workflow options"
// table shared the same 0s bug.
func TestServiceResolvedDefaultWorkflowOptions(t *testing.T) {
	i32 := func(v int32) *int32 { return &v }

	if got := (&Service{}).ResolvedDefaultWorkflowOptions(); got != nil {
		t.Fatalf("ResolvedDefaultWorkflowOptions() with no defaults = %+v, want nil", got)
	}

	s := &Service{
		DefaultWorkflowOptions: &temporalv1.WorkflowOptions{
			WorkflowExecutionTimeout: i32(3600),
		},
	}
	got := s.ResolvedDefaultWorkflowOptions()
	if got == nil {
		t.Fatal("ResolvedDefaultWorkflowOptions() = nil, want resolved options")
	}
	if got.WorkflowExecutionTimeout != time.Hour {
		t.Errorf("WorkflowExecutionTimeout = %s, want 1h0m0s", got.WorkflowExecutionTimeout)
	}
	if got.WorkflowRunTimeout != 0 || got.WorkflowTaskTimeout != 0 {
		t.Errorf("unset timeouts = (%s, %s), want (0s, 0s)", got.WorkflowRunTimeout, got.WorkflowTaskTimeout)
	}
}
