package model

import (
	"testing"
	"time"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
)

// TestMergeRetryPolicy: Tests retry policy merging
func TestMergeRetryPolicy(t *testing.T) {
	t.Run("both nil returns nil", func(t *testing.T) {
		result := MergeRetryPolicy(nil, nil)
		if result != nil {
			t.Error("expected nil result for both nil inputs")
		}
	})

	t.Run("method only", func(t *testing.T) {
		maxAttempts := int32(5)
		initialInterval := int32(10)
		coeff := float32(2.0)

		method := &temporalv1.RetryPolicy{
			MaximumAttempts:    &maxAttempts,
			InitialInterval:    &initialInterval,
			BackoffCoefficient: &coeff,
		}

		result := MergeRetryPolicy(method, nil)

		if result.MaximumAttempts != 5 {
			t.Errorf("expected MaximumAttempts 5, got %d", result.MaximumAttempts)
		}
		if result.InitialInterval != 10*time.Second {
			t.Errorf("expected InitialInterval 10s, got %v", result.InitialInterval)
		}
		if result.BackoffCoefficient != 2.0 {
			t.Errorf("expected BackoffCoefficient 2.0, got %f", result.BackoffCoefficient)
		}
	})

	t.Run("service defaults used when method fields are nil", func(t *testing.T) {
		serviceMaxAttempts := int32(3)
		serviceInitialInterval := int32(5)

		service := &temporalv1.RetryPolicy{
			MaximumAttempts: &serviceMaxAttempts,
			InitialInterval: &serviceInitialInterval,
		}

		methodMaxInterval := int32(100)
		method := &temporalv1.RetryPolicy{
			MaximumInterval: &methodMaxInterval,
		}

		result := MergeRetryPolicy(method, service)

		if result.MaximumAttempts != 3 {
			t.Errorf("expected MaximumAttempts from service (3), got %d", result.MaximumAttempts)
		}
		if result.InitialInterval != 5*time.Second {
			t.Errorf("expected InitialInterval from service (5s), got %v", result.InitialInterval)
		}
		if result.MaximumInterval != 100*time.Second {
			t.Errorf("expected MaximumInterval from method (100s), got %v", result.MaximumInterval)
		}
	})

	t.Run("method overrides service", func(t *testing.T) {
		serviceMaxAttempts := int32(3)
		methodMaxAttempts := int32(10)

		service := &temporalv1.RetryPolicy{
			MaximumAttempts: &serviceMaxAttempts,
		}

		method := &temporalv1.RetryPolicy{
			MaximumAttempts: &methodMaxAttempts,
		}

		result := MergeRetryPolicy(method, service)

		if result.MaximumAttempts != 10 {
			t.Errorf("expected method MaximumAttempts (10), got %d", result.MaximumAttempts)
		}
	})

	t.Run("non-retryable error types", func(t *testing.T) {
		method := &temporalv1.RetryPolicy{
			NonRetryableErrorTypes: []string{"FATAL", "INVALID"},
		}

		result := MergeRetryPolicy(method, nil)

		if len(result.NonRetryableErrorTypes) != 2 {
			t.Errorf("expected 2 error types, got %d", len(result.NonRetryableErrorTypes))
		}
		if result.NonRetryableErrorTypes[0] != "FATAL" {
			t.Errorf("expected first error type 'FATAL', got %s", result.NonRetryableErrorTypes[0])
		}
	})
}

// TestMergeActivityOptions: Tests activity options merging
func TestMergeActivityOptions(t *testing.T) {
	t.Run("default schedule to close used when nothing specified", func(t *testing.T) {
		result := MergeActivityOptions(nil, nil, 3600)

		if result.ScheduleToCloseTimeout != 3600*time.Second {
			t.Errorf("expected default ScheduleToCloseTimeout 3600s, got %v", result.ScheduleToCloseTimeout)
		}
	})

	t.Run("method options override service defaults", func(t *testing.T) {
		serviceStart := int32(60)
		methodStart := int32(120)

		service := &temporalv1.ActivityOptions{
			StartToCloseTimeout: &serviceStart,
		}

		method := &temporalv1.ActivityOptions{
			StartToCloseTimeout: &methodStart,
		}

		result := MergeActivityOptions(method, service, 3600)

		if result.StartToCloseTimeout != 120*time.Second {
			t.Errorf("expected method StartToCloseTimeout (120s), got %v", result.StartToCloseTimeout)
		}
	})

	t.Run("service defaults used when method doesn't specify", func(t *testing.T) {
		serviceStart := int32(60)
		serviceHeartbeat := int32(30)

		service := &temporalv1.ActivityOptions{
			StartToCloseTimeout: &serviceStart,
			HeartbeatTimeout:    &serviceHeartbeat,
		}

		methodSchedule := int32(180)
		method := &temporalv1.ActivityOptions{
			ScheduleToCloseTimeout: &methodSchedule,
		}

		result := MergeActivityOptions(method, service, 3600)

		if result.StartToCloseTimeout != 60*time.Second {
			t.Errorf("expected service StartToCloseTimeout (60s), got %v", result.StartToCloseTimeout)
		}
		if result.HeartbeatTimeout != 30*time.Second {
			t.Errorf("expected service HeartbeatTimeout (30s), got %v", result.HeartbeatTimeout)
		}
		if result.ScheduleToCloseTimeout != 180*time.Second {
			t.Errorf("expected method ScheduleToCloseTimeout (180s), got %v", result.ScheduleToCloseTimeout)
		}
	})

	t.Run("custom name from method", func(t *testing.T) {
		method := &temporalv1.ActivityOptions{
			Name: "custom-activity-name",
		}

		result := MergeActivityOptions(method, nil, 3600)

		if result.Name != "custom-activity-name" {
			t.Errorf("expected custom name, got %s", result.Name)
		}
	})

	t.Run("retry policy merged", func(t *testing.T) {
		methodMaxAttempts := int32(5)
		method := &temporalv1.ActivityOptions{
			RetryPolicy: &temporalv1.RetryPolicy{
				MaximumAttempts: &methodMaxAttempts,
			},
		}

		serviceInitialInterval := int32(10)
		service := &temporalv1.ActivityOptions{
			RetryPolicy: &temporalv1.RetryPolicy{
				InitialInterval: &serviceInitialInterval,
			},
		}

		result := MergeActivityOptions(method, service, 3600)

		if result.RetryPolicy == nil {
			t.Fatal("expected retry policy to be set")
		}
		if result.RetryPolicy.MaximumAttempts != 5 {
			t.Errorf("expected MaximumAttempts from method (5), got %d", result.RetryPolicy.MaximumAttempts)
		}
		if result.RetryPolicy.InitialInterval != 10*time.Second {
			t.Errorf("expected InitialInterval from service (10s), got %v", result.RetryPolicy.InitialInterval)
		}
	})
}

// TestMergeWorkflowOptions: Tests workflow options merging
func TestMergeWorkflowOptions(t *testing.T) {
	t.Run("both nil returns empty options", func(t *testing.T) {
		result := MergeWorkflowOptions(nil, nil)

		if result.WorkflowExecutionTimeout != 0 {
			t.Errorf("expected WorkflowExecutionTimeout to be 0, got %v", result.WorkflowExecutionTimeout)
		}
	})

	t.Run("method options override service defaults", func(t *testing.T) {
		serviceExecTimeout := int32(3600)
		methodExecTimeout := int32(7200)

		service := &temporalv1.WorkflowOptions{
			WorkflowExecutionTimeout: &serviceExecTimeout,
		}

		method := &temporalv1.WorkflowOptions{
			WorkflowExecutionTimeout: &methodExecTimeout,
		}

		result := MergeWorkflowOptions(method, service)

		if result.WorkflowExecutionTimeout != 7200*time.Second {
			t.Errorf("expected method WorkflowExecutionTimeout (7200s), got %v", result.WorkflowExecutionTimeout)
		}
	})

	t.Run("service defaults used when method doesn't specify", func(t *testing.T) {
		serviceExecTimeout := int32(3600)
		serviceRunTimeout := int32(1800)

		service := &temporalv1.WorkflowOptions{
			WorkflowExecutionTimeout: &serviceExecTimeout,
			WorkflowRunTimeout:       &serviceRunTimeout,
		}

		methodTaskTimeout := int32(60)
		method := &temporalv1.WorkflowOptions{
			WorkflowTaskTimeout: &methodTaskTimeout,
		}

		result := MergeWorkflowOptions(method, service)

		if result.WorkflowExecutionTimeout != 3600*time.Second {
			t.Errorf("expected service WorkflowExecutionTimeout (3600s), got %v", result.WorkflowExecutionTimeout)
		}
		if result.WorkflowRunTimeout != 1800*time.Second {
			t.Errorf("expected service WorkflowRunTimeout (1800s), got %v", result.WorkflowRunTimeout)
		}
		if result.WorkflowTaskTimeout != 60*time.Second {
			t.Errorf("expected method WorkflowTaskTimeout (60s), got %v", result.WorkflowTaskTimeout)
		}
	})

	t.Run("custom name from method", func(t *testing.T) {
		method := &temporalv1.WorkflowOptions{
			Name: "custom-workflow-name",
		}

		result := MergeWorkflowOptions(method, nil)

		if result.Name != "custom-workflow-name" {
			t.Errorf("expected custom name, got %s", result.Name)
		}
	})

	t.Run("signals and queries from method", func(t *testing.T) {
		method := &temporalv1.WorkflowOptions{
			Signals: []string{"SignalA", "SignalB"},
			Queries: []string{"QueryX", "QueryY"},
		}

		result := MergeWorkflowOptions(method, nil)

		if len(result.Signals) != 2 {
			t.Errorf("expected 2 signals, got %d", len(result.Signals))
		}
		if result.Signals[0] != "SignalA" {
			t.Errorf("expected first signal 'SignalA', got %s", result.Signals[0])
		}

		if len(result.Queries) != 2 {
			t.Errorf("expected 2 queries, got %d", len(result.Queries))
		}
		if result.Queries[0] != "QueryX" {
			t.Errorf("expected first query 'QueryX', got %s", result.Queries[0])
		}
	})

	t.Run("retry policy merged", func(t *testing.T) {
		methodMaxAttempts := int32(5)
		method := &temporalv1.WorkflowOptions{
			RetryPolicy: &temporalv1.RetryPolicy{
				MaximumAttempts: &methodMaxAttempts,
			},
		}

		serviceInitialInterval := int32(10)
		service := &temporalv1.WorkflowOptions{
			RetryPolicy: &temporalv1.RetryPolicy{
				InitialInterval: &serviceInitialInterval,
			},
		}

		result := MergeWorkflowOptions(method, service)

		if result.RetryPolicy == nil {
			t.Fatal("expected retry policy to be set")
		}
		if result.RetryPolicy.MaximumAttempts != 5 {
			t.Errorf("expected MaximumAttempts from method (5), got %d", result.RetryPolicy.MaximumAttempts)
		}
		if result.RetryPolicy.InitialInterval != 10*time.Second {
			t.Errorf("expected InitialInterval from service (10s), got %v", result.RetryPolicy.InitialInterval)
		}
	})
}
