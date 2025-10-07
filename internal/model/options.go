package model

import (
	"time"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
)

// RetryPolicy: Represents a retry policy configuration
type RetryPolicy struct {
	InitialInterval        time.Duration
	BackoffCoefficient     float64
	MaximumInterval        time.Duration
	MaximumAttempts        int32
	NonRetryableErrorTypes []string
}

// ActivityOptions: Configuration for activity execution
type ActivityOptions struct {
	Name                   string
	ScheduleToStartTimeout time.Duration
	ScheduleToCloseTimeout time.Duration
	StartToCloseTimeout    time.Duration
	HeartbeatTimeout       time.Duration
	RetryPolicy            *RetryPolicy
}

// WorkflowOptions: Configuration for workflow execution
type WorkflowOptions struct {
	Name                     string
	WorkflowExecutionTimeout time.Duration
	WorkflowRunTimeout       time.Duration
	WorkflowTaskTimeout      time.Duration
	RetryPolicy              *RetryPolicy
	Signals                  []string
	Queries                  []string
}

// MergeRetryPolicy: Merges two retry policies, method takes precedence
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

	// Fill in missing fields from service defaults
	if service != nil {
		if result.InitialInterval == 0 && service.InitialInterval != nil {
			result.InitialInterval = time.Duration(*service.InitialInterval) * time.Second
		}
		if result.BackoffCoefficient == 0 && service.BackoffCoefficient != nil {
			result.BackoffCoefficient = float64(*service.BackoffCoefficient)
		}
		if result.MaximumInterval == 0 && service.MaximumInterval != nil {
			result.MaximumInterval = time.Duration(*service.MaximumInterval) * time.Second
		}
		if result.MaximumAttempts == 0 && service.MaximumAttempts != nil {
			result.MaximumAttempts = *service.MaximumAttempts
		}
		if len(result.NonRetryableErrorTypes) == 0 && service.NonRetryableErrorTypes != nil {
			result.NonRetryableErrorTypes = service.NonRetryableErrorTypes
		}
	}

	return result
}

// MergeActivityOptions: Merges activity options with service defaults
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

	// Fill in missing fields from service defaults
	if service != nil {
		if result.ScheduleToStartTimeout == 0 && service.ScheduleToStartTimeout != nil {
			result.ScheduleToStartTimeout = time.Duration(*service.ScheduleToStartTimeout) * time.Second
		}
		if result.ScheduleToCloseTimeout == 0 && service.ScheduleToCloseTimeout != nil {
			result.ScheduleToCloseTimeout = time.Duration(*service.ScheduleToCloseTimeout) * time.Second
		}
		if result.StartToCloseTimeout == 0 && service.StartToCloseTimeout != nil {
			result.StartToCloseTimeout = time.Duration(*service.StartToCloseTimeout) * time.Second
		}
		if result.HeartbeatTimeout == 0 && service.HeartbeatTimeout != nil {
			result.HeartbeatTimeout = time.Duration(*service.HeartbeatTimeout) * time.Second
		}
	}

	// Apply default if still not set
	if result.ScheduleToCloseTimeout == 0 {
		result.ScheduleToCloseTimeout = time.Duration(defaultScheduleToClose) * time.Second
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

// MergeWorkflowOptions: Merges workflow options with service defaults
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
	}

	// Fill in missing fields from service defaults
	if service != nil {
		if result.WorkflowExecutionTimeout == 0 && service.WorkflowExecutionTimeout != nil {
			result.WorkflowExecutionTimeout = time.Duration(*service.WorkflowExecutionTimeout) * time.Second
		}
		if result.WorkflowRunTimeout == 0 && service.WorkflowRunTimeout != nil {
			result.WorkflowRunTimeout = time.Duration(*service.WorkflowRunTimeout) * time.Second
		}
		if result.WorkflowTaskTimeout == 0 && service.WorkflowTaskTimeout != nil {
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
