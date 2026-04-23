package model

import (
	"testing"

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

// TestServiceGenerateCLIDefault Verifies that GenerateCLI is false when the
// proto option is absent (default) or explicitly set to false.
func TestServiceGenerateCLIDefault(t *testing.T) {
	t.Run("option absent", func(t *testing.T) {
		opts := &temporalv1.ServiceOptions{}
		if got := opts.GetGenerateCli(); got != false {
			t.Errorf("expected GetGenerateCli() false when unset, got %v", got)
		}
		s := &Service{GenerateCLI: opts.GetGenerateCli()}
		if s.GenerateCLI {
			t.Error("expected Service.GenerateCLI to default to false")
		}
	})

	t.Run("option explicitly false", func(t *testing.T) {
		v := false
		opts := &temporalv1.ServiceOptions{GenerateCli: &v}
		s := &Service{GenerateCLI: opts.GetGenerateCli()}
		if s.GenerateCLI {
			t.Error("expected Service.GenerateCLI false when option is false")
		}
	})
}

// TestServiceGenerateCLITrue Verifies that GenerateCLI is true when the
// proto option is set to true.
func TestServiceGenerateCLITrue(t *testing.T) {
	v := true
	opts := &temporalv1.ServiceOptions{GenerateCli: &v}
	s := &Service{GenerateCLI: opts.GetGenerateCli()}
	if !s.GenerateCLI {
		t.Error("expected Service.GenerateCLI true when option is true")
	}
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
