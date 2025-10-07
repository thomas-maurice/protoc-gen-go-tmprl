package model

import (
	"testing"
)

// TestServiceGetClientName: Tests client name generation
func TestServiceGetClientName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "ExampleServiceClient"
	if result := service.GetClientName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetWorkerName: Tests worker name generation
func TestServiceGetWorkerName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "ExampleServiceWorker"
	if result := service.GetWorkerName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetServiceInterfaceName: Tests service interface name generation
func TestServiceGetServiceInterfaceName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "ExampleServiceService"
	if result := service.GetServiceInterfaceName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetDefaultTaskQueueConstName: Tests default task queue constant name generation
func TestServiceGetDefaultTaskQueueConstName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "DefaultExampleServiceTaskQueueName"
	if result := service.GetDefaultTaskQueueConstName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetDefaultActivityTimeoutConstName: Tests default activity timeout constant name generation
func TestServiceGetDefaultActivityTimeoutConstName(t *testing.T) {
	service := &Service{
		GoName: "ExampleService",
	}

	expected := "DefaultExampleServiceActivityScheduleToCloseTimeout"
	if result := service.GetDefaultActivityTimeoutConstName(); result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TestServiceGetSignal: Tests signal retrieval from map
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

// TestServiceGetQuery: Tests query retrieval from map
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
