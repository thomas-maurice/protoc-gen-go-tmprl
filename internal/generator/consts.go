package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

func ServiceConstants(gf *protogen.GeneratedFile, service *protogen.Service) error {
	workflowsNames := jen.Comment("Workflows names constants").Line().Line()
	activitiesNames := jen.Comment("Activities names constants").Line().Line()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		switch t {
		case MethodTypeNone:
			continue
		case MethodTypeActivity:
			name, err := getMethodRegisteredName(method)
			if err != nil {
				panic(err)
			}
			activitiesNames.Comment(fmt.Sprintf("Name of activity %s", method.Desc.FullName())).Line().
				Id(fmt.Sprintf("Activity%s%sName", service.GoName, method.GoName)).Op("=").Lit(name).Line()
		case MethodTypeWorkflow:
			name, err := getMethodRegisteredName(method)
			if err != nil {
				panic(err)
			}
			workflowsNames.Comment(fmt.Sprintf("Name of workflow %s", method.Desc.FullName())).Line().
				Id(fmt.Sprintf("Workflow%s%sName", service.GoName, method.GoName)).Op("=").Lit(name).Line()

		default:
			return fmt.Errorf("invalid method type: %s", t)
		}
	}

	defaultTaskQueueName := jen.Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)).Op("=").Lit(getServiceTaskQueue(service)).Line()

	generated := jen.Const().Parens(
		defaultTaskQueueName.Add(workflowsNames.Add(activitiesNames)),
	)

	buf := bytes.NewBufferString("")
	if err := generated.Render(buf); err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
