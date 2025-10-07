package model

import (
	"fmt"
	"strings"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// Service: Represents a temporal service with all its methods
type Service struct {
	Name          string
	GoName        string
	PackageName   string
	TaskQueue     string
	Comment       string
	ProtoService  *protogen.Service
	GeneratedFile *protogen.GeneratedFile
	Config        *Config

	// Method collections
	Workflows  []*Workflow
	Activities []*Activity
	Signals    []*Signal
	Queries    []*Query

	// Defaults
	DefaultActivityOptions *temporalv1.ActivityOptions
	DefaultWorkflowOptions *temporalv1.WorkflowOptions

	// Lookup maps for signals/queries (exported for templates)
	SignalsMap map[string]*Signal
	QueriesMap map[string]*Query
}

// NewService: Creates a service model from a protobuf service
func NewService(protoService *protogen.Service, gf *protogen.GeneratedFile, config *Config) (*Service, error) {
	serviceOpts, ok := proto.GetExtension(protoService.Desc.Options(), temporalv1.E_Service).(*temporalv1.ServiceOptions)
	if !ok || serviceOpts == nil {
		return nil, fmt.Errorf("service %s is not a temporal service", protoService.GoName)
	}

	service := &Service{
		Name:                   string(protoService.Desc.Name()),
		GoName:                 protoService.GoName,
		PackageName:            string(protoService.Desc.ParentFile().Package()),
		TaskQueue:              serviceOpts.TaskQueue,
		Comment:                getServiceComment(protoService),
		ProtoService:           protoService,
		GeneratedFile:          gf,
		Config:                 config,
		DefaultActivityOptions: serviceOpts.DefaultActivityOptions,
		DefaultWorkflowOptions: serviceOpts.DefaultWorkflowOptions,
		SignalsMap:             make(map[string]*Signal),
		QueriesMap:             make(map[string]*Query),
	}

	// First pass: create signals and queries for lookup
	for _, method := range protoService.Methods {
		methodType, err := detectMethodType(method)
		if err != nil {
			continue
		}

		switch methodType {
		case MethodTypeSignal:
			signal, err := NewSignal(method, service)
			if err != nil {
				return nil, err
			}
			service.Signals = append(service.Signals, signal)
			service.SignalsMap[signal.GoName] = signal

		case MethodTypeQuery:
			query, err := NewQuery(method, service)
			if err != nil {
				return nil, err
			}
			service.Queries = append(service.Queries, query)
			service.QueriesMap[query.GoName] = query
		}
	}

	// Second pass: create workflows and activities
	for _, method := range protoService.Methods {
		methodType, err := detectMethodType(method)
		if err != nil {
			continue
		}

		switch methodType {
		case MethodTypeWorkflow:
			workflow, err := NewWorkflow(method, service, config)
			if err != nil {
				return nil, err
			}
			service.Workflows = append(service.Workflows, workflow)

		case MethodTypeActivity:
			activity, err := NewActivity(method, service, config)
			if err != nil {
				return nil, err
			}
			service.Activities = append(service.Activities, activity)
		}
	}

	return service, nil
}

// GetSignal: Retrieves a signal by name
func (s *Service) GetSignal(name string) (*Signal, error) {
	signal, ok := s.SignalsMap[name]
	if !ok {
		return nil, fmt.Errorf("signal %s not found in service %s", name, s.GoName)
	}
	return signal, nil
}

// GetQuery: Retrieves a query by name
func (s *Service) GetQuery(name string) (*Query, error) {
	query, ok := s.QueriesMap[name]
	if !ok {
		return nil, fmt.Errorf("query %s not found in service %s", name, s.GoName)
	}
	return query, nil
}

// GetClientName: Returns the generated client name
func (s *Service) GetClientName() string {
	return fmt.Sprintf("%sClient", s.GoName)
}

// GetWorkerName: Returns the generated worker name
func (s *Service) GetWorkerName() string {
	return fmt.Sprintf("%sWorker", s.GoName)
}

// GetServiceInterfaceName: Returns the service interface name
func (s *Service) GetServiceInterfaceName() string {
	return s.GoName + "Service"
}

// GetDefaultTaskQueueConstName: Returns the constant name for default task queue
func (s *Service) GetDefaultTaskQueueConstName() string {
	return fmt.Sprintf("Default%sTaskQueueName", s.GoName)
}

// GetDefaultActivityTimeoutConstName: Returns the constant name for default activity timeout
func (s *Service) GetDefaultActivityTimeoutConstName() string {
	return fmt.Sprintf("Default%sActivityScheduleToCloseTimeout", s.GoName)
}

// getServiceComment: Extracts comments from protobuf service
func getServiceComment(service *protogen.Service) string {
	if service.Comments.Leading != "" {
		return strings.TrimSpace(string(service.Comments.Leading))
	}
	return ""
}
