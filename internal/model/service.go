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
	Updates    []*Update

	// Defaults
	DefaultActivityOptions *temporalv1.ActivityOptions
	DefaultWorkflowOptions *temporalv1.WorkflowOptions

	// Lookup maps for signals/queries/updates (exported for templates)
	SignalsMap map[string]*Signal
	QueriesMap map[string]*Query
	UpdatesMap map[string]*Update
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
		SignalsMap:             make(map[string]*Signal),
		QueriesMap:             make(map[string]*Query),
		UpdatesMap:             make(map[string]*Update),
	}

	// Validate service-level default options up front so a bad default fails
	// loud rather than silently propagating into every method.
	if err := validateWorkflowOptions(serviceOpts.DefaultWorkflowOptions, fmt.Sprintf("service %s default_workflow_options", protoService.GoName)); err != nil {
		return nil, err
	}
	if err := validateActivityOptions(serviceOpts.DefaultActivityOptions, fmt.Sprintf("service %s default_activity_options", protoService.GoName)); err != nil {
		return nil, err
	}

	// First pass: create signals, queries and updates for lookup
	for _, method := range protoService.Methods {
		methodType, err := detectMethodType(method)
		if err != nil {
			return nil, err
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

		case MethodTypeUpdate:
			update, err := NewUpdate(method, service)
			if err != nil {
				return nil, err
			}
			service.Updates = append(service.Updates, update)
			service.UpdatesMap[update.GoName] = update
		}
	}

	// Second pass: create workflows and activities
	for _, method := range protoService.Methods {
		methodType, err := detectMethodType(method)
		if err != nil {
			return nil, err
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

	if err := service.validate(); err != nil {
		return nil, err
	}

	return service, nil
}

// validate Checks cross-method invariants after the model is fully built: every
// signal/query/update a workflow references must be defined in the same service
// and listed at most once, and no two workflows (or two activities) may resolve
// to the same registered name. These are all cases the generator would
// otherwise turn into silently-dropped methods, uncompilable output, or a
// worker-startup panic, so they fail loud here at generation time.
func (s *Service) validate() error {
	for _, w := range s.Workflows {
		if err := s.validateWorkflowRefs(w); err != nil {
			return err
		}
	}

	if err := checkUniqueRegisteredNames("workflow", len(s.Workflows), func(i int) (string, string) {
		return s.Workflows[i].GoName, s.Workflows[i].RegisteredName
	}); err != nil {
		return err
	}
	return checkUniqueRegisteredNames("activity", len(s.Activities), func(i int) (string, string) {
		return s.Activities[i].GoName, s.Activities[i].RegisteredName
	})
}

// validateWorkflowRefs Verifies a single workflow's signal/query/update lists:
// each name resolves to a method of the matching kind in this service and does
// not appear more than once.
func (s *Service) validateWorkflowRefs(w *Workflow) error {
	check := func(kind string, names []string, defined func(string) bool) error {
		seen := make(map[string]bool, len(names))
		for _, name := range names {
			if seen[name] {
				return fmt.Errorf("workflow %s lists %s %q more than once", w.GoName, kind, name)
			}
			seen[name] = true
			if !defined(name) {
				return fmt.Errorf("workflow %s references %s %q which is not a defined %s in service %s", w.GoName, kind, name, kind, s.GoName)
			}
		}
		return nil
	}

	if err := check("signal", w.Options.Signals, func(n string) bool { _, ok := s.SignalsMap[n]; return ok }); err != nil {
		return err
	}
	if err := check("query", w.Options.Queries, func(n string) bool { _, ok := s.QueriesMap[n]; return ok }); err != nil {
		return err
	}
	return check("update", w.Options.Updates, func(n string) bool { _, ok := s.UpdatesMap[n]; return ok })
}

// checkUniqueRegisteredNames Reports an error if two methods of the same kind
// resolve to the same Temporal registered name (e.g. via duplicate `name`
// overrides), which would panic the worker at registration time.
func checkUniqueRegisteredNames(kind string, n int, at func(int) (goName, registered string)) error {
	seen := make(map[string]string, n)
	for i := 0; i < n; i++ {
		goName, registered := at(i)
		if prev, ok := seen[registered]; ok {
			return fmt.Errorf("%ss %s and %s both register as %q; registered names must be unique", kind, prev, goName, registered)
		}
		seen[registered] = goName
	}
	return nil
}

// PackageScopedNames Returns the package-level identifiers this service emits
// that are NOT service-prefixed. The plugin uses these to detect collisions
// between two services generated into the same file (same Go package), which
// would otherwise produce a redeclaration and fail the consumer's build with no
// generator-side diagnostic.
func (s *Service) PackageScopedNames() []string {
	var names []string
	for _, w := range s.Workflows {
		names = append(names, "Workflow"+w.GoName+"Name")
	}
	for _, a := range s.Activities {
		names = append(names, "Activity"+a.GoName+"Name")
	}
	for _, sig := range s.Signals {
		names = append(names, "Signal"+sig.GoName+"Name", "ReceiveSignal"+sig.GoName, "ReceiveSignal"+sig.GoName+"Async")
	}
	for _, q := range s.Queries {
		names = append(names, "Query"+q.GoName+"Name", "HandleQuery"+q.GoName)
	}
	for _, u := range s.Updates {
		names = append(names, "Update"+u.GoName+"Name", "HandleUpdate"+u.GoName, "HandleUpdate"+u.GoName+"WithValidator")
	}
	return names
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

// GetUpdate Retrieves an update by name
func (s *Service) GetUpdate(name string) (*Update, error) {
	update, ok := s.UpdatesMap[name]
	if !ok {
		return nil, fmt.Errorf("update %s not found in service %s", name, s.GoName)
	}
	return update, nil
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

// ResolvedDefaultActivityOptions Returns the service-level default activity
// options with the proto's second-valued fields converted to time.Duration,
// or nil if the service declares none. The documentation template renders from
// this rather than reading DefaultActivityOptions directly: the raw proto
// timeouts are *int32, which FormatDuration cannot interpret and would render
// as 0s. Reusing MergeActivityOptions keeps the doc path on the same
// seconds-to-Duration conversion the code generator uses.
func (s *Service) ResolvedDefaultActivityOptions() *ActivityOptions {
	if s.DefaultActivityOptions == nil {
		return nil
	}
	return MergeActivityOptions(nil, s.DefaultActivityOptions, 0)
}

// ResolvedDefaultWorkflowOptions Returns the service-level default workflow
// options with the proto's second-valued fields converted to time.Duration,
// or nil if the service declares none. See ResolvedDefaultActivityOptions for
// why the documentation template renders from this rather than the raw proto.
func (s *Service) ResolvedDefaultWorkflowOptions() *WorkflowOptions {
	if s.DefaultWorkflowOptions == nil {
		return nil
	}
	return MergeWorkflowOptions(nil, s.DefaultWorkflowOptions)
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

// GetScheduleSpecCheckFuncName Returns the name of the per-service helper that
// reports whether a client.ScheduleSpec is empty (the caller supplied no
// scheduling information). UpsertSchedule uses it to avoid overwriting an
// existing schedule's spec with an empty one.
func (s *Service) GetScheduleSpecCheckFuncName() string {
	return fmt.Sprintf("scheduleSpecIsZero%s", s.GoName)
}

// getServiceComment Extracts comments from protobuf service
func getServiceComment(service *protogen.Service) string {
	if service.Comments.Leading != "" {
		return strings.TrimSpace(string(service.Comments.Leading))
	}
	return ""
}
