package model

import (
	"fmt"
	"math"
	"strings"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// MethodType Type of temporal method
type MethodType int

const (
	MethodTypeUnknown MethodType = iota
	MethodTypeWorkflow
	MethodTypeActivity
	MethodTypeSignal
	MethodTypeQuery
	MethodTypeUpdate
)

// Method Base interface for all method types
type Method interface {
	GetName() string
	GetGoName() string
	GetRegisteredName() string
	GetInput() *protogen.Message
	GetOutput() *protogen.Message
	GetComment() string
	GetProtoMethod() *protogen.Method
	GetService() *Service
	GetType() MethodType
}

// BaseMethod Common fields for all method types
type BaseMethod struct {
	Name           string
	GoName         string
	RegisteredName string
	Input          *protogen.Message
	Output         *protogen.Message
	Comment        string
	ProtoMethod    *protogen.Method
	Service        *Service
}

func (m *BaseMethod) GetName() string                  { return m.Name }
func (m *BaseMethod) GetGoName() string                { return m.GoName }
func (m *BaseMethod) GetRegisteredName() string        { return m.RegisteredName }
func (m *BaseMethod) GetInput() *protogen.Message      { return m.Input }
func (m *BaseMethod) GetOutput() *protogen.Message     { return m.Output }
func (m *BaseMethod) GetComment() string               { return m.Comment }
func (m *BaseMethod) GetProtoMethod() *protogen.Method { return m.ProtoMethod }
func (m *BaseMethod) GetService() *Service             { return m.Service }

// Workflow Represents a temporal workflow
type Workflow struct {
	BaseMethod
	Options *WorkflowOptions
}

func (w *Workflow) GetType() MethodType { return MethodTypeWorkflow }

// Activity Represents a temporal activity
type Activity struct {
	BaseMethod
	Options *ActivityOptions
}

func (a *Activity) GetType() MethodType { return MethodTypeActivity }

// Signal Represents a temporal signal
type Signal struct {
	BaseMethod
	CustomName string
}

func (s *Signal) GetType() MethodType { return MethodTypeSignal }

// Query Represents a temporal query
type Query struct {
	BaseMethod
	CustomName string
}

func (q *Query) GetType() MethodType { return MethodTypeQuery }

// Update Represents a temporal update
type Update struct {
	BaseMethod
	CustomName string
}

func (u *Update) GetType() MethodType { return MethodTypeUpdate }

// detectMethodType Determines the type of a protobuf method. It returns
// MethodTypeUnknown with a nil error for a method that carries no temporal
// annotation (the caller skips it), and a non-nil error only when a method
// carries more than one temporal annotation, which is ambiguous and would
// otherwise be resolved silently to whichever annotation happened to be checked
// first.
func detectMethodType(method *protogen.Method) (MethodType, error) {
	var found []MethodType
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions); ok && opts != nil {
		found = append(found, MethodTypeWorkflow)
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions); ok && opts != nil {
		found = append(found, MethodTypeActivity)
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions); ok && opts != nil {
		found = append(found, MethodTypeSignal)
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions); ok && opts != nil {
		found = append(found, MethodTypeQuery)
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Update).(*temporalv1.UpdateOptions); ok && opts != nil {
		found = append(found, MethodTypeUpdate)
	}

	switch len(found) {
	case 0:
		return MethodTypeUnknown, nil
	case 1:
		return found[0], nil
	default:
		names := make([]string, len(found))
		for i, t := range found {
			names[i] = methodTypeName(t)
		}
		return MethodTypeUnknown, fmt.Errorf("method %s has multiple temporal annotations (%s); a method may carry at most one", method.GoName, strings.Join(names, ", "))
	}
}

// methodTypeName Returns a human-readable name for a MethodType, used in error
// messages.
func methodTypeName(t MethodType) string {
	switch t {
	case MethodTypeWorkflow:
		return "workflow"
	case MethodTypeActivity:
		return "activity"
	case MethodTypeSignal:
		return "signal"
	case MethodTypeQuery:
		return "query"
	case MethodTypeUpdate:
		return "update"
	default:
		return "unknown"
	}
}

// validateRetryPolicy Rejects retry-policy values that would produce broken or
// nonsensical generated code: a non-finite backoff coefficient renders as
// +Inf/NaN (valid Go syntax but uncompilable), and negative intervals/attempts
// are meaningless to Temporal.
func validateRetryPolicy(rp *temporalv1.RetryPolicy, context string) error {
	if rp == nil {
		return nil
	}
	if rp.BackoffCoefficient != nil {
		bc := float64(*rp.BackoffCoefficient)
		if math.IsInf(bc, 0) || math.IsNaN(bc) {
			return fmt.Errorf("%s: retry_policy.backoff_coefficient must be a finite number, got %v", context, bc)
		}
		if bc < 0 {
			return fmt.Errorf("%s: retry_policy.backoff_coefficient must be >= 0, got %v", context, bc)
		}
	}
	if rp.InitialInterval != nil && *rp.InitialInterval < 0 {
		return fmt.Errorf("%s: retry_policy.initial_interval must be >= 0, got %d", context, *rp.InitialInterval)
	}
	if rp.MaximumInterval != nil && *rp.MaximumInterval < 0 {
		return fmt.Errorf("%s: retry_policy.maximum_interval must be >= 0, got %d", context, *rp.MaximumInterval)
	}
	if rp.MaximumAttempts != nil && *rp.MaximumAttempts < 0 {
		return fmt.Errorf("%s: retry_policy.maximum_attempts must be >= 0, got %d", context, *rp.MaximumAttempts)
	}
	return nil
}

// validateWorkflowOptions Rejects negative timeouts and an invalid retry policy
// on a workflow's options.
func validateWorkflowOptions(o *temporalv1.WorkflowOptions, context string) error {
	if o == nil {
		return nil
	}
	if o.WorkflowExecutionTimeout != nil && *o.WorkflowExecutionTimeout < 0 {
		return fmt.Errorf("%s: workflow_execution_timeout must be >= 0, got %d", context, *o.WorkflowExecutionTimeout)
	}
	if o.WorkflowRunTimeout != nil && *o.WorkflowRunTimeout < 0 {
		return fmt.Errorf("%s: workflow_run_timeout must be >= 0, got %d", context, *o.WorkflowRunTimeout)
	}
	if o.WorkflowTaskTimeout != nil && *o.WorkflowTaskTimeout < 0 {
		return fmt.Errorf("%s: workflow_task_timeout must be >= 0, got %d", context, *o.WorkflowTaskTimeout)
	}
	return validateRetryPolicy(o.RetryPolicy, context)
}

// validateActivityOptions Rejects negative timeouts and an invalid retry policy
// on an activity's options.
func validateActivityOptions(o *temporalv1.ActivityOptions, context string) error {
	if o == nil {
		return nil
	}
	if o.ScheduleToStartTimeout != nil && *o.ScheduleToStartTimeout < 0 {
		return fmt.Errorf("%s: schedule_to_start_timeout must be >= 0, got %d", context, *o.ScheduleToStartTimeout)
	}
	if o.ScheduleToCloseTimeout != nil && *o.ScheduleToCloseTimeout < 0 {
		return fmt.Errorf("%s: schedule_to_close_timeout must be >= 0, got %d", context, *o.ScheduleToCloseTimeout)
	}
	if o.StartToCloseTimeout != nil && *o.StartToCloseTimeout < 0 {
		return fmt.Errorf("%s: start_to_close_timeout must be >= 0, got %d", context, *o.StartToCloseTimeout)
	}
	if o.HeartbeatTimeout != nil && *o.HeartbeatTimeout < 0 {
		return fmt.Errorf("%s: heartbeat_timeout must be >= 0, got %d", context, *o.HeartbeatTimeout)
	}
	return validateRetryPolicy(o.RetryPolicy, context)
}

// getRegisteredName Gets the fully qualified name for temporal registration
func getRegisteredName(method *protogen.Method) string {
	pkg := string(method.Parent.Desc.ParentFile().Package())
	service := string(method.Parent.Desc.Name())
	methodName := string(method.Desc.Name())
	return fmt.Sprintf("%s.%s.%s", pkg, service, methodName)
}

// getComment Extracts comments from protobuf method
func getComment(method *protogen.Method) string {
	if method.Comments.Leading != "" {
		return strings.TrimSpace(string(method.Comments.Leading))
	}
	return ""
}

// NewWorkflow Creates a workflow from a protobuf method
func NewWorkflow(protoMethod *protogen.Method, service *Service, config *Config) (*Workflow, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not a workflow", protoMethod.GoName)
	}

	if err := validateWorkflowOptions(opts, fmt.Sprintf("workflow %s", protoMethod.GoName)); err != nil {
		return nil, err
	}

	base := BaseMethod{
		Name:           string(protoMethod.Desc.Name()),
		GoName:         protoMethod.GoName,
		RegisteredName: getRegisteredName(protoMethod),
		Input:          protoMethod.Input,
		Output:         protoMethod.Output,
		Comment:        getComment(protoMethod),
		ProtoMethod:    protoMethod,
		Service:        service,
	}

	// Override name if specified
	if opts.Name != "" {
		base.RegisteredName = opts.Name
	}

	merged := MergeWorkflowOptions(opts, service.DefaultWorkflowOptions)

	return &Workflow{
		BaseMethod: base,
		Options:    merged,
	}, nil
}

// NewActivity Creates an activity from a protobuf method
func NewActivity(protoMethod *protogen.Method, service *Service, config *Config) (*Activity, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not an activity", protoMethod.GoName)
	}

	if err := validateActivityOptions(opts, fmt.Sprintf("activity %s", protoMethod.GoName)); err != nil {
		return nil, err
	}

	base := BaseMethod{
		Name:           string(protoMethod.Desc.Name()),
		GoName:         protoMethod.GoName,
		RegisteredName: getRegisteredName(protoMethod),
		Input:          protoMethod.Input,
		Output:         protoMethod.Output,
		Comment:        getComment(protoMethod),
		ProtoMethod:    protoMethod,
		Service:        service,
	}

	// Override name if specified
	if opts.Name != "" {
		base.RegisteredName = opts.Name
	}

	return &Activity{
		BaseMethod: base,
		Options:    MergeActivityOptions(opts, service.DefaultActivityOptions, config.DefaultActivityScheduleToClose),
	}, nil
}

// NewSignal Creates a signal from a protobuf method
func NewSignal(protoMethod *protogen.Method, service *Service) (*Signal, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not a signal", protoMethod.GoName)
	}

	base := BaseMethod{
		Name:           string(protoMethod.Desc.Name()),
		GoName:         protoMethod.GoName,
		RegisteredName: getRegisteredName(protoMethod),
		Input:          protoMethod.Input,
		Output:         protoMethod.Output,
		Comment:        getComment(protoMethod),
		ProtoMethod:    protoMethod,
		Service:        service,
	}

	customName := ""
	if opts.Name != "" {
		customName = opts.Name
		base.RegisteredName = opts.Name
	}

	return &Signal{
		BaseMethod: base,
		CustomName: customName,
	}, nil
}

// NewQuery Creates a query from a protobuf method
func NewQuery(protoMethod *protogen.Method, service *Service) (*Query, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not a query", protoMethod.GoName)
	}

	base := BaseMethod{
		Name:           string(protoMethod.Desc.Name()),
		GoName:         protoMethod.GoName,
		RegisteredName: getRegisteredName(protoMethod),
		Input:          protoMethod.Input,
		Output:         protoMethod.Output,
		Comment:        getComment(protoMethod),
		ProtoMethod:    protoMethod,
		Service:        service,
	}

	customName := ""
	if opts.Name != "" {
		customName = opts.Name
		base.RegisteredName = opts.Name
	}

	return &Query{
		BaseMethod: base,
		CustomName: customName,
	}, nil
}

// NewUpdate Creates an update from a protobuf method
func NewUpdate(protoMethod *protogen.Method, service *Service) (*Update, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Update).(*temporalv1.UpdateOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not an update", protoMethod.GoName)
	}

	base := BaseMethod{
		Name:           string(protoMethod.Desc.Name()),
		GoName:         protoMethod.GoName,
		RegisteredName: getRegisteredName(protoMethod),
		Input:          protoMethod.Input,
		Output:         protoMethod.Output,
		Comment:        getComment(protoMethod),
		ProtoMethod:    protoMethod,
		Service:        service,
	}

	customName := ""
	if opts.Name != "" {
		customName = opts.Name
		base.RegisteredName = opts.Name
	}

	return &Update{
		BaseMethod: base,
		CustomName: customName,
	}, nil
}
