package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

func UnimplementedServiceInterface(gf *protogen.GeneratedFile, service *protogen.Service) error {
	workflows := jen.Comment("Workflows definitions").Line().Line()
	activities := jen.Comment("Activities definitions").Line().Line()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		switch t {
		case MethodTypeNone:
			continue
		case MethodTypeActivity:
			activities.Comment(method.Comments.Leading.String()).
				Id(method.GoName).
				ParamsFunc(func(g *jen.Group) {
					g.Add(jen.Id("ctx").Id(getContext(gf)))
					if method.Input != nil {
						g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
					}
				}).
				ParamsFunc(func(g *jen.Group) {
					if method.Output != nil {
						g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
					}
					g.Add(jen.Error())
				}).Line()
		case MethodTypeWorkflow:
			workflows.Comment(method.Comments.Leading.String()).
				Id(method.GoName).
				ParamsFunc(func(g *jen.Group) {
					g.Add(jen.Id("ctx").Id(getWorkflowContext(gf)))
					if method.Input != nil {
						g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
					}
				}).
				ParamsFunc(func(g *jen.Group) {
					if method.Output != nil {
						g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
					}
					g.Add(jen.Error())
				}).Line()
		default:
			return fmt.Errorf("invalid method type: %s", t)
		}
	}

	generated := jen.Comment(fmt.Sprintf("%s is the interface your service must implement", getSvcName(service))).Line().
		Type().Id(getSvcName(service)).InterfaceFunc(func(g *jen.Group) {
		g.Add(workflows)
		g.Add(activities)
	})

	buf := bytes.NewBufferString("")
	if err := generated.Render(buf); err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
