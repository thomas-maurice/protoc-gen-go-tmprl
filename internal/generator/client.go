package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

func Client(gf *protogen.GeneratedFile, service *protogen.Service) error {
	clientName := fmt.Sprintf("%sClient", service.GoName)

	client := jen.Comment(fmt.Sprintf("%s: Client for the %s service", clientName, service.GoName)).Line().
		Type().Id(clientName).
		StructFunc(func(g *jen.Group) {
			g.Add(jen.Id("client").Id(getTemporalClientObject(gf, "Client")))
			g.Add(jen.Id("taskQueue").String())
		}).Line().Line().
		// New client func
		Comment(fmt.Sprintf("New%s: Returns a new instance of the client.", clientName)).Line().
		Comment("If `taskQueue` stays empty the default one will be used").Line().
		Func().Id(fmt.Sprintf("New%s", clientName)).
		ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("client").Id(getTemporalClientObject(gf, "Client")))
			g.Add(jen.Id("taskQueue").Op("...").String())
		}).
		Parens(jen.ListFunc(func(g *jen.Group) {
			g.Add(jen.Op("*").Id(clientName))
			g.Add(jen.Error())
		})).
		BlockFunc(func(g *jen.Group) {
			g.Add(jen.Id("clientTaskQueue").Op(":=").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)))
			g.Add(jen.If(jen.Len(jen.Id("taskQueue")).Op(">").Lit(0).Block(
				jen.Id("clientTaskQueue").Op("=").Id("taskQueue").Index(jen.Lit(0)),
			)))

			g.Add(
				jen.ReturnFunc(func(g *jen.Group) {
					g.ListFunc(func(g *jen.Group) {
						g.Add(jen.Op("&").Id(clientName).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("client").Op(":").Id("client")).Op(",")
							g.Add(jen.Id("taskQueue").Op(":").Id("clientTaskQueue")).Op(",")
						}),
						)
						g.Add(jen.Nil())
					},
					)
				}),
			)
		}).Line().Line()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		methName, err := getMethodRegisteredName(method)
		if err != nil {
			return err
		}

		switch t {
		case MethodTypeWorkflow:
			client.Comment(fmt.Sprintf("ExecuteWorkflow%s executes the workflow and returns a future to it", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("ExecuteWorkflow%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
				g.Add(jen.Id("options").Op("...").Id(getTemporalClientObject(gf, "StartWorkflowOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id(getTemporalClientObject(gf, "WorkflowRun")))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("wOptions").Op(":=").Id(getTemporalClientObject(gf, "StartWorkflowOptions")).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("TaskQueue").Op(":").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName))).Op(",")
					}))
					g.Add(jen.If(jen.Len(jen.Id("options")).Op(">").Lit(0).Block(
						jen.Id("wOptions").Op("=").Id("options").Index(jen.Lit(0)),
					)))

					g.Add(jen.If(jen.Id("wOptions").Dot("TaskQueue").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("wOptions").Dot("TaskQueue").Op("=").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)))
					}))

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("c").Dot("client").Dot("ExecuteWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("wOptions"))
							g.Add(jen.Lit(methName))
							g.Add(jen.Id("req"))
						}))
					}))
				}).Line().Line()

			client.Comment(fmt.Sprintf("ExecuteWorkflow%sSync executes the workflow and returns the result when finished", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("ExecuteWorkflow%sSync", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
				g.Add(jen.Id("options").Op("...").Id(getTemporalClientObject(gf, "StartWorkflowOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("future").Op(",").Id("err").Op(":=").Id("c").Dot(fmt.Sprintf("ExecuteWorkflow%s", method.GoName)).CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("req"))
						g.Add(jen.Id("options").Op("..."))
					}))

					g.Add(IfErrNilDouble)

					g.Add(jen.Var().Id("resp").Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))

					g.Add(jen.Id("err").Op("=").Id("future").Dot("Get").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Op("&").Id("resp"))
					}))

					g.Add(IfErrNilDouble)

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("resp"))
						g.Add(jen.Nil())
					}))
				}).Line().Line()

			client.Comment(fmt.Sprintf("ExecuteChild%s executes the workflow as a child workflow and returns a future to it", method.GoName)).Line()

			client.Comment(fmt.Sprintf("ExecuteChild%sSync executes the workflow as a child workflow and returns the result when finished", method.GoName)).Line()
		case MethodTypeActivity:
			client.Comment(fmt.Sprintf("ExecuteActivity%s executes the activity asynchronously and returns a future to it", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("ExecuteActivity%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context")))
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
				g.Add(jen.Id("options").Op("...").Id(getTemporalWorkflowObject(gf, "ActivityOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id(getTemporalWorkflowObject(gf, "Future")))
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Var().Id("aOptions").Id(getTemporalWorkflowObject(gf, "ActivityOptions")))

					g.Add(jen.If(jen.Len(jen.Id("options")).Op(">").Lit(0).Block(
						jen.Id("aOptions").Op("=").Id("options").Index(jen.Lit(0)),
					)))

					g.Add(
						jen.If(jen.Id("aOptions").Dot("TaskQueue").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("aOptions").Dot("TaskQueue").Op("=").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)))
						},
						),
					)

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("workflow").Dot("ExecuteActivity").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id(getTemporalWorkflowObject(gf, "WithActivityOptions")).CallFunc(func(g *jen.Group) {
								g.Add(jen.Id("ctx"))
								g.Add(jen.Id("aOptions"))
							}))
							g.Add(jen.Lit(methName))
							g.Add(jen.Id("req"))
						}))
					}))
				}).Line().Line()

			client.Comment(fmt.Sprintf("ExecuteActivity%sSync executes the activity synchronously and returns the result when finished", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("ExecuteActivity%sSync", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context")))
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
				g.Add(jen.Id("options").Op("...").Id(getTemporalWorkflowObject(gf, "ActivityOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("aOptions").Op(":=").Id(getTemporalWorkflowObject(gf, "ActivityOptions")).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("TaskQueue").Op(":").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName))).Op(",")
					}))
					g.Add(jen.If(jen.Len(jen.Id("options")).Op(">").Lit(0).Block(
						jen.Id("aOptions").Op("=").Id("options").Index(jen.Lit(0)),
					)))

					g.Add(jen.Id("future").Op(":=").Id("c").Dot(fmt.Sprintf("ExecuteActivity%s", method.GoName)).CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("req"))
						g.Add(jen.Id("aOptions"))
					}))

					g.Add(jen.Var().Id("resp").Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))

					g.Add(jen.Id("err").Op(":=").Id("future").Dot("Get").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Op("&").Id("resp"))
					}))

					g.Add(IfErrNilDouble)

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("resp"))
						g.Add(jen.Nil())
					}))
				}).Line().Line()
		}
	}

	buf := bytes.NewBufferString("")

	err := client.Render(buf)
	if err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
