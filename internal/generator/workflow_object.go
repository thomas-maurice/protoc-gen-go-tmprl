package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func WorkflowObjects(gf *protogen.GeneratedFile, service *protogen.Service) error {
	clientName := getClientName(service)

	// build a map of the queries and signals so we can do a lookup when a workflow
	// can recieve them
	signalsMap := make(map[string]*protogen.Method)
	queriesMap := make(map[string]*protogen.Method)

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		switch t {
		case MethodTypeQuery:
			queriesMap[method.GoName] = method
		case MethodTypeSignal:
			signalsMap[method.GoName] = method
		}
	}

	workflowObjects := jen.Null().Line()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		workflowOptions := getWorkflowOptions(method)
		if workflowOptions == nil {
			workflowOptions = getDefaultWorkflowOptions(service)
		}

		switch t {
		case MethodTypeWorkflow:
			/*
				Workflow result structs

				This creates a struct like so

				type SomeServiceSomeFuncWorkflow struct {
					WorkflowID string
					RunID string
					future client.WorkflowRun
					client temporal.Client
				}

				And a bunch of methods like
				func (s *SomeServiceSomeFuncWorkflow).Cancel(ctx context.Context) error {
					wf :=s.client.Cancel(ctx, s.WorkflowID, s.RunID)
				}

				and so on
			*/

			wfObjName := fmt.Sprintf("%s%s", service.GoName, method.GoName)
			workflowObjects.Comment(fmt.Sprintf("%s is a struct that wraps a workflow", wfObjName)).Line().
				Type().Id(wfObjName).StructFunc(func(g *jen.Group) {
				g.Add(jen.Id("WorkflowID").String())
				g.Add(jen.Id("RunID").String())
				g.Add(jen.Id("client").Id(getTemporalClientObject(gf, "Client")))
				g.Add(jen.Id("future").Id(getTemporalClientObject(gf, "WorkflowRun")))
			}).Line()

			// Gets an instance of a workflow
			workflowObjects.Comment(fmt.Sprintf("Get%s gets an instance of a given workflow", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("Get%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("workflowId").String())
				g.Add(jen.Id("runId").String())
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(wfObjName))
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("future").Op(":=").Id("c").Dot("client").Dot("GetWorkflow").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("workflowId"))
						g.Add(jen.Id("runId"))
					}))

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Op("&").Id(wfObjName).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("WorkflowID").Op(":").Id("future").Dot("GetID").Call(jen.Null()).Op(","))
							g.Add(jen.Id("RunID").Op(":").Id("future").Dot("GetRunID").Call(jen.Null()).Op(","))
							g.Add(jen.Id("client").Op(":").Id("c").Dot("client").Op(","))
							g.Add(jen.Id("future").Op(":").Id("future").Op(","))
						}))
					}))
				}).Line().Line()

			// Gets an instance of a workflow from a future
			workflowObjects.Comment(fmt.Sprintf("Get%sFromRun gets an instance of a given workflow from a future", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("Get%sFromRun", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("future").Id(getTemporalClientObject(gf, "WorkflowRun")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(wfObjName))
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Op("&").Id(wfObjName).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("WorkflowID").Op(":").Id("future").Dot("GetID").Call(jen.Null()).Op(","))
							g.Add(jen.Id("RunID").Op(":").Id("future").Dot("GetRunID").Call(jen.Null()).Op(","))
							g.Add(jen.Id("client").Op(":").Id("c").Dot("client").Op(","))
							g.Add(jen.Id("future").Op(":").Id("future").Op(","))
						}))
					}))
				}).Line().Line()

			// Cancels the workflow
			workflowObjects.Comment("Cancel cancels a given workflow").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("Cancel").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("w").Dot("client").Dot("CancelWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("w").Dot("WorkflowID"))
							g.Add(jen.Id("w").Dot("RunID"))
						}))
					}))
				}).Line()

			// Gets the workflow ID
			workflowObjects.Comment("Returns the workflow ID").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("GetID").Parens(jen.Null()).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.String())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("w").Dot("future").Dot("GetID").Parens(jen.Null()))
					}))
				}).Line()

			// Gets the run ID
			workflowObjects.Comment("Returns the run ID").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("GetRunID").Parens(jen.Null()).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.String())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("w").Dot("future").Dot("GetRunID").Parens(jen.Null()))
					}))
				}).Line()

			// Terminates the workflow
			workflowObjects.Comment("Terminates terminates a given workflow").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("Terminate").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("reason").String())
				g.Add(jen.Id("details").Op("...").Id("interface{}"))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("w").Dot("client").Dot("TerminateWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("w").Dot("WorkflowID"))
							g.Add(jen.Id("w").Dot("RunID"))
							g.Add(jen.Id("reason"))
							g.Add(jen.Id("details").Op("..."))
						}))
					}))
				}).Line()

			// Gets the result of a workflow
			workflowObjects.Comment("Get gets the result of a given workflow with its native type").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("Result").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Var().Id("resp").Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))

					g.Add(jen.Id("err").Op(":=").Id("w").Dot("future").Dot("Get").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Op("&").Id("resp"))
					}))

					g.Add(IfErrNilDouble)

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("resp"))
						g.Add(jen.Nil())
					}))
				}).Line().Line()

			// Gets the result of a workflow with options
			workflowObjects.Comment("ResultWithOptions gets the result of a given workflow with its native type").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("ResultWithOptions").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("options").Id(getTemporalClientObject(gf, "WorkflowRunGetOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Var().Id("resp").Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))

					g.Add(jen.Id("err").Op(":=").Id("w").Dot("future").Dot("GetWithOptions").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Op("&").Id("resp"))
						g.Add(jen.Id("options"))
					}))

					g.Add(IfErrNilDouble)

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("resp"))
						g.Add(jen.Nil())
					}))
				}).Line().Line()

			workflowObjects.Comment("Get gets the result of a given workflow with pointers -- discouraged to use but required to implement internal.WorkflowRun").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("Get").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("valuePtr").InterfaceFunc(func(g *jen.Group) {}))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Return(jen.Id("w").Dot("future").Dot("Get").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("valuePtr"))
					})))
				}).Line().Line()

			// Gets the result of a workflow with options
			workflowObjects.Comment("Get gets the result of a given workflow with pointers -- discouraged to use but required to implement internal.WorkflowRun").Line().
				Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("GetWithOptions").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("valuePtr").InterfaceFunc(func(g *jen.Group) {}))
				g.Add(jen.Id("options").Id(getTemporalClientObject(gf, "WorkflowRunGetOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Return(jen.Id("w").Dot("future").Dot("GetWithOptions").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("valuePtr"))
						g.Add(jen.Id("options"))
					})))
				}).Line().Line()

			for _, sig := range workflowOptions.Signals {
				meth, ok := signalsMap[sig]
				if !ok {
					return fmt.Errorf("no signal %s defined in service %s for workflow %s", sig, service.GoName, method.GoName)
				}

				sigName, err := getMethodRegisteredName(meth)
				if err != nil {
					return err
				}
				sigOpts, _ := proto.GetExtension(method.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions)

				if sigOpts != nil && sigOpts.Name != "" {
					sigName = sigOpts.Name
				}

				// Sends a signal to a workflow
				workflowObjects.Comment(fmt.Sprintf("Signal%s sends the %s signal to the workflow", sig, sig)).Line().
					Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("Signal" + sig).ParamsFunc(func(g *jen.Group) {
					g.Add(jen.Id("ctx").Id(getContext(gf)))
					g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(meth.Input.GoIdent)))
				}).ParamsFunc(func(g *jen.Group) {
					g.Add(jen.Error())
				}).
					BlockFunc(func(g *jen.Group) {
						g.Add(jen.Return(jen.Id("w").Dot("client").Dot("SignalWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("w").Dot("future").Dot("GetID").Parens(jen.Null()))
							g.Add(jen.Id("w").Dot("future").Dot("GetRunID").Parens(jen.Null()))
							g.Add(jen.Lit(sigName))
							g.Add(jen.Id("req"))
						})))
					}).Line().Line()
			}

			for _, query := range workflowOptions.Queries {
				meth, ok := queriesMap[query]
				if !ok {
					return fmt.Errorf("no query %s defined in service %s for workflow %s", query, service.GoName, method.GoName)
				}

				queryName, err := getMethodRegisteredName(meth)
				if err != nil {
					return err
				}
				queryOpts, _ := proto.GetExtension(method.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions)

				if queryOpts != nil && queryOpts.Name != "" {
					queryName = queryOpts.Name
				}

				// Sens a query to a workflow
				workflowObjects.Comment(fmt.Sprintf("Query%s queries the workflow with %s", query, query)).Line().
					Func().Parens(jen.Id("w").Op("*").Id(wfObjName)).Id("Query" + query).ParamsFunc(func(g *jen.Group) {
					g.Add(jen.Id("ctx").Id(getContext(gf)))
					g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(meth.Input.GoIdent)))
				}).ParamsFunc(func(g *jen.Group) {
					g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(meth.Output.GoIdent)))
					g.Add(jen.Error())
				}).
					BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("future").Op(",").Err().Op(":=").Id("w").Dot("client").Dot("QueryWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("w").Dot("future").Dot("GetID").Parens(jen.Null()))
							g.Add(jen.Id("w").Dot("future").Dot("GetRunID").Parens(jen.Null()))
							g.Add(jen.Lit(queryName))
							g.Add(jen.Id("req"))
						}))

						g.Add(IfErrNilDouble)

						g.Add(jen.Var().Id("resp").Op("*").Id(gf.QualifiedGoIdent(meth.Output.GoIdent)))

						g.Add(jen.Id("err").Op("=").Id("future").Dot("Get").CallFunc(func(g *jen.Group) {
							g.Add(jen.Op("&").Id("resp"))
						}))

						g.Add(IfErrNilDouble)

						g.Add(jen.ReturnFunc(func(g *jen.Group) {
							g.Add(jen.Id("resp"))
							g.Add(jen.Nil())
						}))
					}).Line().Line()
			}

			workflowObjects.Line()
		}
	}

	buf := bytes.NewBufferString("")

	err := workflowObjects.Render(buf)
	if err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
