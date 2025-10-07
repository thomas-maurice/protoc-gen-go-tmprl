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
