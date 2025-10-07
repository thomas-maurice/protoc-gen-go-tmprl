package model

import (
	"testing"
)

// TestMethodTypeConstants: Tests that method type constants are defined
func TestMethodTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		methType MethodType
		expected int
	}{
		{"MethodTypeUnknown", MethodTypeUnknown, 0},
		{"MethodTypeWorkflow", MethodTypeWorkflow, 1},
		{"MethodTypeActivity", MethodTypeActivity, 2},
		{"MethodTypeSignal", MethodTypeSignal, 3},
		{"MethodTypeQuery", MethodTypeQuery, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.methType) != tt.expected {
				t.Errorf("expected %s to be %d, got %d", tt.name, tt.expected, int(tt.methType))
			}
		})
	}
}

// TestBaseMethodGetters: Tests BaseMethod getter methods
func TestBaseMethodGetters(t *testing.T) {
	base := BaseMethod{
		Name:           "TestMethod",
		GoName:         "TestMethodGo",
		RegisteredName: "pkg.Service.TestMethod",
		Comment:        "Test comment",
	}

	if base.GetName() != "TestMethod" {
		t.Errorf("expected Name 'TestMethod', got %s", base.GetName())
	}

	if base.GetGoName() != "TestMethodGo" {
		t.Errorf("expected GoName 'TestMethodGo', got %s", base.GetGoName())
	}

	if base.GetRegisteredName() != "pkg.Service.TestMethod" {
		t.Errorf("expected RegisteredName 'pkg.Service.TestMethod', got %s", base.GetRegisteredName())
	}

	if base.GetComment() != "Test comment" {
		t.Errorf("expected Comment 'Test comment', got %s", base.GetComment())
	}

	if base.GetProtoMethod() != nil {
		t.Error("expected nil ProtoMethod")
	}

	if base.GetService() != nil {
		t.Error("expected nil Service")
	}

	if base.GetInput() != nil {
		t.Error("expected nil Input")
	}

	if base.GetOutput() != nil {
		t.Error("expected nil Output")
	}
}

// TestWorkflowGetType: Tests Workflow.GetType() returns correct type
func TestWorkflowGetType(t *testing.T) {
	workflow := &Workflow{}
	if workflow.GetType() != MethodTypeWorkflow {
		t.Errorf("expected MethodTypeWorkflow, got %v", workflow.GetType())
	}
}

// TestActivityGetType: Tests Activity.GetType() returns correct type
func TestActivityGetType(t *testing.T) {
	activity := &Activity{}
	if activity.GetType() != MethodTypeActivity {
		t.Errorf("expected MethodTypeActivity, got %v", activity.GetType())
	}
}

// TestSignalGetType: Tests Signal.GetType() returns correct type
func TestSignalGetType(t *testing.T) {
	signal := &Signal{}
	if signal.GetType() != MethodTypeSignal {
		t.Errorf("expected MethodTypeSignal, got %v", signal.GetType())
	}
}

// TestQueryGetType: Tests Query.GetType() returns correct type
func TestQueryGetType(t *testing.T) {
	query := &Query{}
	if query.GetType() != MethodTypeQuery {
		t.Errorf("expected MethodTypeQuery, got %v", query.GetType())
	}
}

// TestMethodInterface: Tests that all method types implement Method interface
func TestMethodInterface(t *testing.T) {
	var _ Method = &Workflow{}
	var _ Method = &Activity{}
	var _ Method = &Signal{}
	var _ Method = &Query{}
}
