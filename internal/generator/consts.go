package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

func ServiceConstants(gf *protogen.GeneratedFile, service *protogen.Service) error {
	workflowsNames := jen.Line().Comment("Workflows names constants").Line().Line()
	activitiesNames := jen.Line().Comment("Activities names constants").Line().Line()
	signalsNames := jen.Line().Comment("Signals names constants").Line().Line()
	queriesNames := jen.Line().Comment("Queries names constants").Line().Line()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		name, err := getMethodRegisteredName(method)
		if err != nil {
			panic(err)
		}

		switch t {
		case MethodTypeNone:
			continue
		case MethodTypeActivity:
			activitiesNames.Comment(fmt.Sprintf("Name of activity %s", method.Desc.FullName())).Line().
				Id(fmt.Sprintf("Activity%s%sName", service.GoName, method.GoName)).Op("=").Lit(name).Line()
		case MethodTypeWorkflow:
			workflowsNames.Comment(fmt.Sprintf("Name of workflow %s", method.Desc.FullName())).Line().
				Id(fmt.Sprintf("Workflow%s%sName", service.GoName, method.GoName)).Op("=").Lit(name).Line()
		case MethodTypeSignal:
			signalsNames.Comment(fmt.Sprintf("Name of signal %s", method.Desc.FullName())).Line().
				Id(fmt.Sprintf("Signal%s%sName", service.GoName, method.GoName)).Op("=").Lit(name).Line()
		case MethodTypeQuery:
			queriesNames.Comment(fmt.Sprintf("Name of query %s", method.Desc.FullName())).Line().
				Id(fmt.Sprintf("Query%s%sName", service.GoName, method.GoName)).Op("=").Lit(name).Line()
		default:
			return fmt.Errorf("invalid method type: %s", t)
		}
	}

	defaultTaskQueueName := jen.Comment("Default task queue name for the service").Line().
		Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)).Op("=").Lit(getServiceTaskQueue(service)).Line()

	generated := jen.Const().Parens(
		defaultTaskQueueName.Add(workflowsNames.Add(activitiesNames).Add(signalsNames).Add(queriesNames)).Line().Line().
			Comment("Default timeout for activities when none is specified").Line().
			Id(fmt.Sprintf("Default%sActivityScheduleToCloseTimeout", service.GoName)).Op("=").Id(getTimeObject(gf, "Hour")).Line().
			Comment("Default timeout for activities when none is specified").Line().
			Id(fmt.Sprintf("Default%sActivityStartToCloseTimeout", service.GoName)).Op("=").Id(getTimeObject(gf, "Hour")),
	)

	buf := bytes.NewBufferString("")
	if err := generated.Render(buf); err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
