package tmpl

import (
	"testing"
	"time"
)

// TestToSeconds: Tests duration to seconds conversion
func TestToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"int32", int32(50), 50},
		{"time.Duration seconds", 30 * time.Second, 30},
		{"time.Duration minutes", 2 * time.Minute, 120},
		{"unknown type", "string", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("toSeconds(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHasTimeout: Tests timeout value checking
func TestHasTimeout(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"positive int", 10, true},
		{"zero int", 0, false},
		{"negative int", -5, false},
		{"positive duration", 30 * time.Second, true},
		{"zero duration", 0 * time.Second, false},
		{"unknown type", "string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTimeout(tt.input)
			if result != tt.expected {
				t.Errorf("hasTimeout(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHasRetryPolicy: Tests retry policy nil checking
func TestHasRetryPolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"non-nil struct", struct{}{}, true},
		{"non-nil pointer", &struct{}{}, true},
		{"nil", nil, false},
		{"non-nil string", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRetryPolicy(tt.input)
			if result != tt.expected {
				t.Errorf("hasRetryPolicy(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWorkflowObjectName: Tests workflow object name generation
func TestWorkflowObjectName(t *testing.T) {
	tests := []struct {
		serviceName string
		methodName  string
		expected    string
	}{
		{"Example", "Ping", "ExamplePing"},
		{"User", "Create", "UserCreate"},
		{"", "Test", "Test"},
		{"Service", "", "Service"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := workflowObjectName(tt.serviceName, tt.methodName)
			if result != tt.expected {
				t.Errorf("workflowObjectName(%q, %q) = %q, expected %q",
					tt.serviceName, tt.methodName, result, tt.expected)
			}
		})
	}
}

// TestChildWorkflowObjectName: Tests child workflow object name generation
func TestChildWorkflowObjectName(t *testing.T) {
	tests := []struct {
		serviceName string
		methodName  string
		expected    string
	}{
		{"Example", "Ping", "ChildExamplePingExecution"},
		{"User", "Create", "ChildUserCreateExecution"},
		{"", "Test", "ChildTestExecution"},
		{"Service", "", "ChildServiceExecution"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := childWorkflowObjectName(tt.serviceName, tt.methodName)
			if result != tt.expected {
				t.Errorf("childWorkflowObjectName(%q, %q) = %q, expected %q",
					tt.serviceName, tt.methodName, result, tt.expected)
			}
		})
	}
}

// TestDict: Tests dictionary creation from key-value pairs
func TestDict(t *testing.T) {
	t.Run("valid pairs", func(t *testing.T) {
		result, err := dict("key1", "value1", "key2", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("expected 2 entries, got %d", len(result))
		}

		if result["key1"] != "value1" {
			t.Errorf("expected key1 to be 'value1', got %v", result["key1"])
		}

		if result["key2"] != 42 {
			t.Errorf("expected key2 to be 42, got %v", result["key2"])
		}
	})

	t.Run("odd number of arguments", func(t *testing.T) {
		_, err := dict("key1", "value1", "key2")
		if err == nil {
			t.Error("expected error for odd number of arguments, got nil")
		}
	})

	t.Run("non-string key", func(t *testing.T) {
		_, err := dict(42, "value1")
		if err == nil {
			t.Error("expected error for non-string key, got nil")
		}
	})

	t.Run("empty dict", func(t *testing.T) {
		result, err := dict()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("expected empty dict, got %d entries", len(result))
		}
	})
}

// TestCommentOneLine: Tests multiline comment collapse
func TestCommentOneLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"multiline with newlines",
			"This is a comment\nwith multiple lines\nand more text",
			"This is a comment with multiple lines and more text",
		},
		{
			"single line",
			"This is a single line comment",
			"This is a single line comment",
		},
		{
			"with carriage returns",
			"Line1\r\nLine2\r\nLine3",
			"Line1 Line2 Line3",
		},
		{
			"multiple spaces",
			"This   has    multiple     spaces",
			"This has multiple spaces",
		},
		{
			"mixed newlines and spaces",
			"First line\n  Second line  \n    Third line",
			"First line Second line Third line",
		},
		{
			"empty string",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commentOneLine(tt.input)
			if result != tt.expected {
				t.Errorf("commentOneLine(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
