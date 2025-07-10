package generator

import (
	"fmt"
	"strings"
	"time"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// github is stupid I suspect it doesn't allow you to put whatever you want in the anchor ids
func makeAnchor(typeName string, name string) string {
	in := fmt.Sprintf("%s:%s", typeName, name)
	replaced := ":.-"
	for _, r := range replaced {
		in = strings.ReplaceAll(in, string(r), "_")
	}

	return in
}

func trim(in protogen.Comments) string {
	return trimComments(string(in))
}

func trimComments(in string) string {
	in = strings.Replace(in, "//", "", 1)
	return strings.TrimSpace(in)
}

// addComments appends the set of comments (message, service, method) to the file
func addComments(f *protogen.GeneratedFile, comments protogen.CommentSet) {
	for _, comment := range comments.LeadingDetached {
		f.P(trim(comment))
	}

	f.P(trim(comments.Leading))
}

func addMethOptions(f *protogen.GeneratedFile, svc *protogen.Service, meth *protogen.Method) error {
	t, err := getMethodType(meth)
	if err != nil {
		return err
	}

	switch t {
	case MethodTypeWorkflow:
		opts, _ := proto.GetExtension(meth.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions)
		if opts == nil {
			return nil
		}

		if opts.WorkflowExecutionTimeout != nil {
			f.P(fmt.Sprintf("| Workflow execution timeout | %v |", time.Second*time.Duration(opts.GetWorkflowExecutionTimeout())))
		}

		if opts.WorkflowRunTimeout != nil {
			f.P(fmt.Sprintf("| Workflow run timeout | %v |", time.Second*time.Duration(opts.GetWorkflowRunTimeout())))
		}

		if opts.WorkflowTaskTimeout != nil {
			f.P(fmt.Sprintf("| Workflow task timeout | %v |", time.Second*time.Duration(opts.GetWorkflowTaskTimeout())))
		}

		if opts.RetryPolicy != nil {
			addRetryPolicy(f, opts.RetryPolicy)
		}

	case MethodTypeActivity:
		opts, _ := proto.GetExtension(meth.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions)
		if opts == nil {
			return nil
		}

		if opts.ScheduleToCloseTimeout != nil {
			f.P(fmt.Sprintf("| Schedule to close timeout | %v |", time.Second*time.Duration(opts.GetScheduleToCloseTimeout())))
		}

		if opts.ScheduleToStartTimeout != nil {
			f.P(fmt.Sprintf("| Schedule to start timeout | %v |", time.Second*time.Duration(opts.GetScheduleToStartTimeout())))

		}

		if opts.StartToCloseTimeout != nil {
			f.P(fmt.Sprintf("| Start to close timeout | %v |", time.Second*time.Duration(opts.GetStartToCloseTimeout())))

		}

		if opts.RetryPolicy != nil {
			addRetryPolicy(f, opts.RetryPolicy)
		}
	case MethodTypeSignal:
	case MethodTypeQuery:
	}

	return nil
}

func addRetryPolicy(f *protogen.GeneratedFile, rp *temporalv1.RetryPolicy) {
	if rp == nil {
		return
	}

	f.P("\nRetry policy:\n")
	f.P("| Option | Value |")
	f.P("| --- | --- |")
	f.P(fmt.Sprintf("| Initial interval | %v |", time.Second*time.Duration(rp.GetInitialInterval())))
	f.P(fmt.Sprintf("| Backoff coefficient | %f |", rp.GetBackoffCoefficient()))
	f.P(fmt.Sprintf("| Maximum attemps | %d |", rp.GetMaximumAttempts()))
	f.P(fmt.Sprintf("| Maximum interval | %v |", time.Second*time.Duration(rp.GetMaximumInterval())))
	f.P(fmt.Sprintf("| Non retryable error types | %v |", rp.GetNonRetryableErrorTypes()))
	f.P("")
}

func addMethodDocs(f *protogen.GeneratedFile, svc *protogen.Service, meth *protogen.Method) error {
	name, err := getMethodRegisteredName(meth)
	if err != nil {
		return err
	}

	f.P(fmt.Sprintf(`<a id="%s"></a>`, makeAnchor("method", string(meth.Desc.FullName()))))
	f.P("#### " + meth.Desc.FullName())
	addComments(f, meth.Comments)

	f.P("")

	if meth.Input != nil {
		f.P(fmt.Sprintf("Input : [%s](#%s)\n", meth.Input.Desc.FullName(), makeAnchor("message", string(meth.Input.Desc.FullName()))))
	}
	if meth.Output != nil {
		f.P(fmt.Sprintf("Output : [%s](#%s)\n", meth.Output.Desc.FullName(), makeAnchor("message", string(meth.Output.Desc.FullName()))))
	}

	f.P("\n| Setting | Value |")
	f.P("| ----------- | ----------------------- |")
	f.P(fmt.Sprintf("| Temporal registered method name | `%s` |", name))
	err = addMethOptions(f, svc, meth)
	if err != nil {
		return err
	}
	f.P("")

	// if we're dealing with a workflow, it might have signals and queries
	opts, _ := proto.GetExtension(meth.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions)
	if opts != nil {
		if len(opts.Signals) != 0 {
			f.P("\nSignals:")
			for _, sig := range opts.Signals {
				f.P(fmt.Sprintf(" * [%s.%s](#%s)", svc.Desc.FullName(), sig, makeAnchor("method", string(svc.Desc.FullName())+"."+sig)))
			}
		}

		if len(opts.Queries) != 0 {
			f.P("\nQueries:")
			for _, q := range opts.Signals {
				f.P(fmt.Sprintf(" * [%s.%s](#%s)", svc.Desc.FullName(), q, makeAnchor("method", string(svc.Desc.FullName())+"."+q)))
			}
		}

		f.P("")
	}

	return nil
}

func ReadmeService(f *protogen.GeneratedFile, service *protogen.Service, cfg *Config) error {
	f.P(fmt.Sprintf(`<a id="service:%s"></a>`, service.Desc.FullName()))
	f.P(fmt.Sprintf("## %s", service.Desc.FullName()))
	addComments(f, service.Comments)

	f.P("| Option | Value |")
	f.P("| --- | --- |")
	f.P(fmt.Sprintf("| Default task queue | `%s` |", getServiceTaskQueue(service)))

	f.P("")

	workflows := make([]*protogen.Method, 0)
	activities := make([]*protogen.Method, 0)
	signals := make([]*protogen.Method, 0)
	queries := make([]*protogen.Method, 0)

	for _, meth := range service.Methods {
		t, err := getMethodType(meth)
		if err != nil {
			return err
		}

		switch t {
		case MethodTypeWorkflow:
			workflows = append(workflows, meth)
		case MethodTypeActivity:
			activities = append(activities, meth)
		case MethodTypeSignal:
			signals = append(signals, meth)
		case MethodTypeQuery:
			queries = append(queries, meth)
		}
	}

	f.P("### Table of contents\n")
	if len(workflows) != 0 {
		f.P(" * Workflows")
		for _, meth := range workflows {
			f.P(fmt.Sprintf("   * [%s](#%s)", meth.Desc.FullName(), makeAnchor("method", string(meth.Desc.FullName()))))
		}
	}
	if len(activities) != 0 {
		f.P(" * Activities")
		for _, meth := range activities {
			f.P(fmt.Sprintf("   * [%s](#%s)", meth.Desc.FullName(), makeAnchor("method", string(meth.Desc.FullName()))))
		}
	}
	if len(signals) != 0 {
		f.P(" * Signals")
		for _, meth := range signals {
			f.P(fmt.Sprintf("   * [%s](#%s)", meth.Desc.FullName(), makeAnchor("method", string(meth.Desc.FullName()))))
		}
	}
	if len(queries) != 0 {
		f.P(" * Queries")
		for _, meth := range queries {
			f.P(fmt.Sprintf("   * [%s](#%s)", meth.Desc.FullName(), makeAnchor("method", string(meth.Desc.FullName()))))
		}
	}

	f.P("### Workflows")
	for _, meth := range workflows {
		err := addMethodDocs(f, service, meth)
		if err != nil {
			return err
		}
	}

	f.P("### Activities")
	for _, meth := range activities {
		err := addMethodDocs(f, service, meth)
		if err != nil {
			return err
		}
	}

	f.P("### Queries")
	for _, meth := range queries {
		err := addMethodDocs(f, service, meth)
		if err != nil {
			return err
		}
	}

	f.P("### Signals")
	for _, meth := range signals {
		err := addMethodDocs(f, service, meth)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadmeMessage(f *protogen.GeneratedFile, message *protogen.Message, cfg *Config) error {
	f.P(fmt.Sprintf(`<a id="%s"></a>`, makeAnchor("message", string(message.Desc.FullName()))))
	f.P(fmt.Sprintf("## %s", message.Desc.FullName()))
	addComments(f, message.Comments)

	return nil
}
