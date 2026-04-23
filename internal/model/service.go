package model

import (
	"fmt"
	"strings"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// Service Represents a temporal service with all its methods
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

	// GenerateCLI reports whether a Cobra CLI should be emitted for this
	// service. Mirrors the service_options.generate_cli proto option and
	// defaults to false when the option is unset.
	GenerateCLI bool

	// Lookup maps for signals/queries (exported for templates)
	SignalsMap map[string]*Signal
	QueriesMap map[string]*Query
}

// NewService Creates a service model from a protobuf service
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
		GenerateCLI:            serviceOpts.GetGenerateCli(),
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

// GetSignal Retrieves a signal by name
func (s *Service) GetSignal(name string) (*Signal, error) {
	signal, ok := s.SignalsMap[name]
	if !ok {
		return nil, fmt.Errorf("signal %s not found in service %s", name, s.GoName)
	}
	return signal, nil
}

// GetQuery Retrieves a query by name
func (s *Service) GetQuery(name string) (*Query, error) {
	query, ok := s.QueriesMap[name]
	if !ok {
		return nil, fmt.Errorf("query %s not found in service %s", name, s.GoName)
	}
	return query, nil
}

// GetClientName Returns the generated client name
func (s *Service) GetClientName() string {
	return fmt.Sprintf("%sClient", s.GoName)
}

// GetWorkerName Returns the generated worker name
func (s *Service) GetWorkerName() string {
	return fmt.Sprintf("%sWorker", s.GoName)
}

// GetServiceInterfaceName Returns the service interface name
func (s *Service) GetServiceInterfaceName() string {
	return s.GoName + "Service"
}

// GetDefaultTaskQueueConstName Returns the constant name for default task queue
func (s *Service) GetDefaultTaskQueueConstName() string {
	return fmt.Sprintf("Default%sTaskQueueName", s.GoName)
}

// GetDefaultActivityTimeoutConstName Returns the constant name for default activity timeout
func (s *Service) GetDefaultActivityTimeoutConstName() string {
	return fmt.Sprintf("Default%sActivityScheduleToCloseTimeout", s.GoName)
}

// GetScheduleMergeFuncName Returns the name of the per-service helper that
// merges user-supplied client.ScheduleOptions into a base struct.
func (s *Service) GetScheduleMergeFuncName() string {
	return fmt.Sprintf("mergeScheduleOptions%s", s.GoName)
}

// GetScheduleDefaultsFuncName Returns the name of the per-service helper
// that applies the default task queue to a client.ScheduleOptions if it
// wasn't explicitly set.
func (s *Service) GetScheduleDefaultsFuncName() string {
	return fmt.Sprintf("applyScheduleDefaults%s", s.GoName)
}

// CLIEnumType pairs a distinct enum Go type name with its declared values
// for use by the CLI template. Used to emit one pflag.Value wrapper per
// enum referenced by any of the service's FlagPlans.
type CLIEnumType struct {
	// GoType is the fully-qualified proto name of the enum (used to
	// derive the Go identifier via the template's enumValueTypeName
	// helper).
	GoType string

	// Values is the ordered list of UPPER_SNAKE enum value names, used
	// for validation in Set() and for the help-text format hint.
	Values []string

	// Numbers is the ordered list of integer numbers matching Values,
	// aligned index-by-index. The generated Set() method maps names to
	// these numbers without needing a reference to the enum's Go-level
	// `<Enum>_value` map.
	Numbers []int32
}

// DistinctCLIEnums returns the deduplicated list of enum types referenced
// by any FlagPlan on this service (workflows, signals, queries). Order is
// deterministic: enums are returned in first-encounter order as the
// template walks workflows, signals, then queries, then by flag position
// within each plan. The CLI template uses this to emit exactly one
// pflag.Value wrapper per enum, avoiding both duplicate declarations and
// forgotten ones.
func (s *Service) DistinctCLIEnums() []CLIEnumType {
	seen := make(map[string]bool)
	var out []CLIEnumType
	visit := func(plan *FlagPlan) {
		if plan == nil {
			return
		}
		for _, fs := range plan.Flags {
			if fs.Kind != FlagKindEnum {
				continue
			}
			if fs.EnumGoType == "" {
				continue
			}
			if seen[fs.EnumGoType] {
				continue
			}
			seen[fs.EnumGoType] = true
			out = append(out, CLIEnumType{GoType: fs.EnumGoType, Values: fs.EnumValues, Numbers: fs.EnumNumbers})
		}
	}
	for _, wf := range s.Workflows {
		visit(wf.InputFlagPlan)
	}
	for _, sig := range s.Signals {
		visit(sig.InputFlagPlan)
	}
	for _, q := range s.Queries {
		visit(q.InputFlagPlan)
	}
	return out
}

// HasGeneratedCLIWorkflows reports whether this service will emit any
// per-workflow CLI subcommand. A service may have GenerateCLI=true but
// every workflow opted out via skip_cli; in that case the CLI root
// entrypoint is still emitted but has no workflow subcommands.
func (s *Service) HasGeneratedCLIWorkflows() bool {
	if !s.GenerateCLI {
		return false
	}
	for _, wf := range s.Workflows {
		if wf.InputFlagPlan != nil {
			return true
		}
	}
	return false
}

// ServiceKebabName returns the service's kebab-case name for use as the
// Cobra root command Use: string. OrdersService -> "orders-service".
func (s *Service) ServiceKebabName() string {
	return kebabCaseServiceName(s.GoName)
}

// kebabCaseServiceName lower-kebabs a PascalCase / camelCase service name.
// Duplicated here from the template helper to keep the Service type
// self-contained (no template-package import).
func kebabCaseServiceName(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	var b []rune
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			if prev < 'A' || prev > 'Z' {
				b = append(b, '-')
			} else if i+1 < len(runes) {
				next := runes[i+1]
				if next >= 'a' && next <= 'z' {
					b = append(b, '-')
				}
			}
		}
		if r >= 'A' && r <= 'Z' {
			b = append(b, r+('a'-'A'))
		} else {
			b = append(b, r)
		}
	}
	return string(b)
}

// getServiceComment Extracts comments from protobuf service
func getServiceComment(service *protogen.Service) string {
	if service.Comments.Leading != "" {
		return strings.TrimSpace(string(service.Comments.Leading))
	}
	return ""
}
