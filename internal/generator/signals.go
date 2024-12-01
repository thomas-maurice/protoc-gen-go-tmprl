package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func ServiceSignals(gf *protogen.GeneratedFile, service *protogen.Service) error {
	clientName := getClientName(service)

	signals := jen.Null()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		if t != MethodTypeSignal {
			continue
		}

		methName, err := getMethodRegisteredName(method)
		if err != nil {
			return err
		}

		sigOpts, _ := proto.GetExtension(method.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions)

		signals.Comment(fmt.Sprintf("SendSignal%s sends the %s signal to a workflow", method.GoName, method.GoName)).Line().
			Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("SendSignal%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("ctx").Id(getContext(gf)))
			g.Add(jen.Id("workflowID").String())
			g.Add(jen.Id("runID").String())
			g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
		}).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Error())
		}).BlockFunc(func(g *jen.Group) {
			sigName := methName
			if sigOpts != nil && sigOpts.Name != "" {
				sigName = sigOpts.Name
			}

			g.Add(jen.Return(
				jen.Id("c").Dot("client").Dot("SignalWorkflow").CallFunc(func(g *jen.Group) {
					g.Add(jen.Id("ctx"))
					g.Add(jen.Id("workflowID"))
					g.Add(jen.Id("runID"))
					g.Add(jen.Lit(sigName))
					g.Add(jen.Id("req"))
				})))

		}).Line()

		signals.Comment(fmt.Sprintf("ReceiveSignal%s waits for the the %s signal", method.GoName, method.GoName)).Line().
			Func().Id(fmt.Sprintf("ReceiveSignal%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context")))
		}).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
			g.Add(jen.Bool())
		}).BlockFunc(func(g *jen.Group) {
			sigName := methName
			if sigOpts != nil && sigOpts.Name != "" {
				sigName = sigOpts.Name
			}

			g.Add(jen.Var().Id("result").Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))

			g.Id("ok").Op(":=").Id(getTemporalWorkflowObject(gf, "GetSignalChannel")).CallFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx"))
				g.Add(jen.Lit(sigName))
			}).Dot("Receive").CallFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx"))
				g.Add(jen.Op("&").Id("result"))
			})

			g.Add(jen.ReturnFunc(func(g *jen.Group) {
				g.Add(jen.Id("result"))
				g.Add(jen.Op("ok"))
			}))

		}).Line()
	}

	buf := bytes.NewBufferString("")
	if err := signals.Render(buf); err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
