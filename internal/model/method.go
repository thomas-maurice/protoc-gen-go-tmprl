package model

import (
	"fmt"
	"strings"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// MethodType: Type of temporal method
type MethodType int

const (
	MethodTypeUnknown MethodType = iota
	MethodTypeWorkflow
	MethodTypeActivity
	MethodTypeSignal
	MethodTypeQuery
)

// Method: Base interface for all method types
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

// BaseMethod: Common fields for all method types
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

// Workflow: Represents a temporal workflow
type Workflow struct {
	BaseMethod
	Options *WorkflowOptions
}

func (w *Workflow) GetType() MethodType { return MethodTypeWorkflow }

// Activity: Represents a temporal activity
type Activity struct {
	BaseMethod
	Options *ActivityOptions
}

func (a *Activity) GetType() MethodType { return MethodTypeActivity }

// Signal: Represents a temporal signal
type Signal struct {
	BaseMethod
	CustomName string
}

func (s *Signal) GetType() MethodType { return MethodTypeSignal }

// Query: Represents a temporal query
type Query struct {
	BaseMethod
	CustomName string
}

func (q *Query) GetType() MethodType { return MethodTypeQuery }

// detectMethodType: Determines the type of a protobuf method
func detectMethodType(method *protogen.Method) (MethodType, error) {
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions); ok && opts != nil {
		return MethodTypeWorkflow, nil
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions); ok && opts != nil {
		return MethodTypeActivity, nil
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions); ok && opts != nil {
		return MethodTypeSignal, nil
	}
	if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions); ok && opts != nil {
		return MethodTypeQuery, nil
	}
	return MethodTypeUnknown, fmt.Errorf("method %s has no temporal annotation", method.GoName)
}

// getRegisteredName: Gets the fully qualified name for temporal registration
func getRegisteredName(method *protogen.Method) string {
	pkg := string(method.Parent.Desc.ParentFile().Package())
	service := string(method.Parent.Desc.Name())
	methodName := string(method.Desc.Name())
	return fmt.Sprintf("%s.%s.%s", pkg, service, methodName)
}

// getComment: Extracts comments from protobuf method
func getComment(method *protogen.Method) string {
	if method.Comments.Leading != "" {
		return strings.TrimSpace(string(method.Comments.Leading))
	}
	return ""
}

// NewWorkflow: Creates a workflow from a protobuf method
func NewWorkflow(protoMethod *protogen.Method, service *Service, config *Config) (*Workflow, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not a workflow", protoMethod.GoName)
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

// NewActivity: Creates an activity from a protobuf method
func NewActivity(protoMethod *protogen.Method, service *Service, config *Config) (*Activity, error) {
	opts, ok := proto.GetExtension(protoMethod.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions)
	if !ok || opts == nil {
		return nil, fmt.Errorf("method %s is not an activity", protoMethod.GoName)
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

// NewSignal: Creates a signal from a protobuf method
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

// NewQuery: Creates a query from a protobuf method
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
