package model

import (
	"time"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
)

// RetryPolicy Represents a retry policy configuration
type RetryPolicy struct {
	InitialInterval        time.Duration
	BackoffCoefficient     float64
	MaximumInterval        time.Duration
	MaximumAttempts        int32
	NonRetryableErrorTypes []string
}

// ActivityOptions Configuration for activity execution
type ActivityOptions struct {
	Name                          string
	ScheduleToStartTimeout        time.Duration
	ScheduleToCloseTimeout        time.Duration
	StartToCloseTimeout           time.Duration
	HeartbeatTimeout              time.Duration
	RetryPolicy                   *RetryPolicy
	ScheduleToCloseTimeoutFromDefault bool // true if ScheduleToCloseTimeout came from default config
}

// WorkflowOptions Configuration for workflow execution
type WorkflowOptions struct {
	Name                     string
	WorkflowExecutionTimeout time.Duration
	WorkflowRunTimeout       time.Duration
	WorkflowTaskTimeout      time.Duration
	RetryPolicy              *RetryPolicy
	Signals                  []string
	Queries                  []string
	Updates                  []string
}

// MergeRetryPolicy Merges two retry policies, method takes precedence
func MergeRetryPolicy(method, service *temporalv1.RetryPolicy) *RetryPolicy {
	if method == nil && service == nil {
		return nil
	}

	result := &RetryPolicy{}

	if method != nil {
		if method.InitialInterval != nil {
			result.InitialInterval = time.Duration(*method.InitialInterval) * time.Second
		}
		if method.BackoffCoefficient != nil {
			result.BackoffCoefficient = float64(*method.BackoffCoefficient)
		}
		if method.MaximumInterval != nil {
			result.MaximumInterval = time.Duration(*method.MaximumInterval) * time.Second
		}
		if method.MaximumAttempts != nil {
			result.MaximumAttempts = *method.MaximumAttempts
		}
		if method.NonRetryableErrorTypes != nil {
			result.NonRetryableErrorTypes = method.NonRetryableErrorTypes
		}
	}

	// Fill in fields the method did not explicitly set from the service
	// defaults. Presence is decided by the upstream proto3 `optional` pointer
	// being nil, NOT by the merged value being zero: an explicit method-level
	// zero (e.g. maximum_attempts: 0 meaning unlimited retries) is a real
	// choice and must not be treated as "unset" and overwritten by the default.
	if service != nil {
		if unset(method, func(m *temporalv1.RetryPolicy) bool { return m.InitialInterval != nil }) && service.InitialInterval != nil {
			result.InitialInterval = time.Duration(*service.InitialInterval) * time.Second
		}
		if unset(method, func(m *temporalv1.RetryPolicy) bool { return m.BackoffCoefficient != nil }) && service.BackoffCoefficient != nil {
			result.BackoffCoefficient = float64(*service.BackoffCoefficient)
		}
		if unset(method, func(m *temporalv1.RetryPolicy) bool { return m.MaximumInterval != nil }) && service.MaximumInterval != nil {
			result.MaximumInterval = time.Duration(*service.MaximumInterval) * time.Second
		}
		if unset(method, func(m *temporalv1.RetryPolicy) bool { return m.MaximumAttempts != nil }) && service.MaximumAttempts != nil {
			result.MaximumAttempts = *service.MaximumAttempts
		}
		if unset(method, func(m *temporalv1.RetryPolicy) bool { return len(m.NonRetryableErrorTypes) > 0 }) && service.NonRetryableErrorTypes != nil {
			result.NonRetryableErrorTypes = service.NonRetryableErrorTypes
		}
	}

	return result
}

// unset Reports whether the method-level options did not explicitly set a field,
// so the service default should apply. present returns true when the given
// method options carry the field. A nil method (no method-level options at all)
// is always unset.
func unset[T any](method *T, present func(*T) bool) bool {
	if method == nil {
		return true
	}
	return !present(method)
}

// MergeActivityOptions Merges activity options with service defaults
func MergeActivityOptions(method, service *temporalv1.ActivityOptions, defaultScheduleToClose int) *ActivityOptions {
	if method == nil && service == nil {
		return &ActivityOptions{
			ScheduleToCloseTimeout: time.Duration(defaultScheduleToClose) * time.Second,
		}
	}

	result := &ActivityOptions{}

	// Apply method options first
	if method != nil {
		if method.Name != "" {
			result.Name = method.Name
		}
		if method.ScheduleToStartTimeout != nil {
			result.ScheduleToStartTimeout = time.Duration(*method.ScheduleToStartTimeout) * time.Second
		}
		if method.ScheduleToCloseTimeout != nil {
			result.ScheduleToCloseTimeout = time.Duration(*method.ScheduleToCloseTimeout) * time.Second
		}
		if method.StartToCloseTimeout != nil {
			result.StartToCloseTimeout = time.Duration(*method.StartToCloseTimeout) * time.Second
		}
		if method.HeartbeatTimeout != nil {
			result.HeartbeatTimeout = time.Duration(*method.HeartbeatTimeout) * time.Second
		}
	}

	// Fill in fields the method did not explicitly set from the service
	// defaults, keyed on proto presence rather than a zero value (see the note
	// in MergeRetryPolicy).
	if service != nil {
		if unset(method, func(m *temporalv1.ActivityOptions) bool { return m.ScheduleToStartTimeout != nil }) && service.ScheduleToStartTimeout != nil {
			result.ScheduleToStartTimeout = time.Duration(*service.ScheduleToStartTimeout) * time.Second
		}
		if unset(method, func(m *temporalv1.ActivityOptions) bool { return m.ScheduleToCloseTimeout != nil }) && service.ScheduleToCloseTimeout != nil {
			result.ScheduleToCloseTimeout = time.Duration(*service.ScheduleToCloseTimeout) * time.Second
		}
		if unset(method, func(m *temporalv1.ActivityOptions) bool { return m.StartToCloseTimeout != nil }) && service.StartToCloseTimeout != nil {
			result.StartToCloseTimeout = time.Duration(*service.StartToCloseTimeout) * time.Second
		}
		if unset(method, func(m *temporalv1.ActivityOptions) bool { return m.HeartbeatTimeout != nil }) && service.HeartbeatTimeout != nil {
			result.HeartbeatTimeout = time.Duration(*service.HeartbeatTimeout) * time.Second
		}
	}

	// Floor the schedule-to-close timeout to the generator default whenever it is
	// still zero. This is DELIBERATELY value-based, not presence-based: unlike
	// every other timeout, an activity with no schedule-to-close timeout will not
	// run at all in Temporal, so guaranteeing one is the entire purpose of the
	// default. An explicit `schedule_to_close_timeout: 0` is therefore floored
	// here rather than honored as "unlimited" -- the one intentional exception to
	// the presence-based merge.
	if result.ScheduleToCloseTimeout == 0 {
		result.ScheduleToCloseTimeout = time.Duration(defaultScheduleToClose) * time.Second
		result.ScheduleToCloseTimeoutFromDefault = true
	}

	// Merge retry policies
	var methodRetry, serviceRetry *temporalv1.RetryPolicy
	if method != nil {
		methodRetry = method.RetryPolicy
	}
	if service != nil {
		serviceRetry = service.RetryPolicy
	}
	result.RetryPolicy = MergeRetryPolicy(methodRetry, serviceRetry)

	return result
}

// MergeWorkflowOptions Merges workflow options with service defaults
func MergeWorkflowOptions(method, service *temporalv1.WorkflowOptions) *WorkflowOptions {
	if method == nil && service == nil {
		return &WorkflowOptions{}
	}

	result := &WorkflowOptions{}

	// Apply method options first
	if method != nil {
		if method.Name != "" {
			result.Name = method.Name
		}
		if method.WorkflowExecutionTimeout != nil {
			result.WorkflowExecutionTimeout = time.Duration(*method.WorkflowExecutionTimeout) * time.Second
		}
		if method.WorkflowRunTimeout != nil {
			result.WorkflowRunTimeout = time.Duration(*method.WorkflowRunTimeout) * time.Second
		}
		if method.WorkflowTaskTimeout != nil {
			result.WorkflowTaskTimeout = time.Duration(*method.WorkflowTaskTimeout) * time.Second
		}
		if method.Signals != nil {
			result.Signals = method.Signals
		}
		if method.Queries != nil {
			result.Queries = method.Queries
		}
		if method.Updates != nil {
			result.Updates = method.Updates
		}
	}

	// Fill in fields the method did not explicitly set from the service
	// defaults, keyed on proto presence rather than a zero value (see the note
	// in MergeRetryPolicy).
	if service != nil {
		if unset(method, func(m *temporalv1.WorkflowOptions) bool { return m.WorkflowExecutionTimeout != nil }) && service.WorkflowExecutionTimeout != nil {
			result.WorkflowExecutionTimeout = time.Duration(*service.WorkflowExecutionTimeout) * time.Second
		}
		if unset(method, func(m *temporalv1.WorkflowOptions) bool { return m.WorkflowRunTimeout != nil }) && service.WorkflowRunTimeout != nil {
			result.WorkflowRunTimeout = time.Duration(*service.WorkflowRunTimeout) * time.Second
		}
		if unset(method, func(m *temporalv1.WorkflowOptions) bool { return m.WorkflowTaskTimeout != nil }) && service.WorkflowTaskTimeout != nil {
			result.WorkflowTaskTimeout = time.Duration(*service.WorkflowTaskTimeout) * time.Second
		}
	}

	// Merge retry policies
	var methodRetry, serviceRetry *temporalv1.RetryPolicy
	if method != nil {
		methodRetry = method.RetryPolicy
	}
	if service != nil {
		serviceRetry = service.RetryPolicy
	}
	result.RetryPolicy = MergeRetryPolicy(methodRetry, serviceRetry)

	return result
}
