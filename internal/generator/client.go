package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

func getClientName(service *protogen.Service) string {
	return fmt.Sprintf("%sClient", service.GoName)
}

func Client(gf *protogen.GeneratedFile, service *protogen.Service) error {
	clientName := getClientName(service)

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

		activityOptions := getActivityOptions(method)
		if activityOptions == nil {
			activityOptions = getDefaultActivityOptions(service)
		}
		workflowOptions := getWorkflowOptions(method)
		if workflowOptions == nil {
			workflowOptions = getDefaultWorkflowOptions(service)
		}

		switch t {
		case MethodTypeWorkflow:
			// Executes a new workflow from the client asynchronously
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
						g.Add(jen.Id("TaskQueue").Op(":").Id("c").Dot("taskQueue").Op(","))
					}))
					g.Add(jen.If(jen.Len(jen.Id("options")).Op(">").Lit(0).Block(
						jen.Id("wOptions").Op("=").Id("options").Index(jen.Lit(0)),
					)))

					g.Add(jen.If(jen.Id("wOptions").Dot("TaskQueue").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("wOptions").Dot("TaskQueue").Op("=").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)))
					}))

					g.Add(
						jen.If(jen.Id("wOptions").Dot("ID").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("wOptions").Dot("ID").Op("=").Id(getFmtObject(gf, "Sprintf")).CallFunc(func(g *jen.Group) {
								g.Add(jen.Lit("%s/%s"))
								g.Add(jen.Lit(methName))
								g.Add(jen.Id(getUUIDObject(gf, "NewString")).Parens(jen.Null()))
							}))
						},
						),
					)

					if workflowOptions != nil {
						if workflowOptions.WorkflowExecutionTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowExecutionTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowExecutionTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(workflowOptions.WorkflowExecutionTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowRunTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(workflowOptions.WorkflowRunTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowTaskTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(workflowOptions.WorkflowRunTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.RetryPolicy != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("RetryPolicy").Op("==").Nil())).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("RetryPolicy").Op("=").Op("&").Id(getTemporalObject(gf, "RetryPolicy")).BlockFunc(func(g *jen.Group) {
									if workflowOptions.RetryPolicy.InitialInterval != nil {
										g.Add(jen.Id("InitialInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(workflowOptions.RetryPolicy.InitialInterval.Value))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumInterval != nil {
										g.Add(jen.Id("MaximumInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(workflowOptions.RetryPolicy.MaximumInterval.Value))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.BackoffCoefficient != nil {
										g.Add(jen.Id("BackoffCoefficient").Op(":").Float64().Call(jen.Lit(workflowOptions.RetryPolicy.BackoffCoefficient.Value)).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumAttempts != nil {
										g.Add(jen.Id("MaximumAttempts").Op(":").Lit(workflowOptions.RetryPolicy.MaximumAttempts.Value).Op(","))
									}
									if workflowOptions.RetryPolicy.NonRetryableErrorTypes != nil {
										g.Add(jen.Id("NonRetryableErrorTypes").Op(":").Index(jen.Null()).String().Block(jen.ListFunc(func(g *jen.Group) {
											for i := 0; i < len(workflowOptions.RetryPolicy.NonRetryableErrorTypes); i++ {
												g.Add(jen.Lit(workflowOptions.RetryPolicy.NonRetryableErrorTypes[i].Value).Op(","))
											}
										})).Op(","))
									}
								}))
							})
						}
					}

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("c").Dot("client").Dot("ExecuteWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("wOptions"))
							g.Add(jen.Lit(methName))
							g.Add(jen.Id("req"))
						}))
					}))
				}).Line().Line()

			// Executes a new workflow from the client synchronously
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

			// Gets the result of a workflow
			client.Comment(fmt.Sprintf("GetWorkflow%sResult gets the result of a given workflow", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("GetWorkflow%sResult", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("workflowId").String())
				g.Add(jen.Id("runId").String())
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("future").Op(":=").Id("c").Dot("client").Dot("GetWorkflow").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("workflowId"))
						g.Add(jen.Id("runId"))
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

			// Executes the workflow as a child workflow asynchronously
			client.Comment(fmt.Sprintf("ExecuteChild%s executes the workflow as a child workflow and returns a future to it", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("ExecuteChild%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context")))
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
				g.Add(jen.Id("options").Op("...").Id(getTemporalWorkflowObject(gf, "ChildWorkflowOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id(getTemporalWorkflowObject(gf, "ChildWorkflowFuture")))
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("wOptions").Op(":=").Id(getTemporalWorkflowObject(gf, "ChildWorkflowOptions")).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("TaskQueue").Op(":").Id("c").Dot("taskQueue").Op(","))
					}))
					g.Add(jen.If(jen.Len(jen.Id("options")).Op(">").Lit(0).Block(
						jen.Id("wOptions").Op("=").Id("options").Index(jen.Lit(0)),
					)))

					g.Add(jen.If(jen.Id("wOptions").Dot("TaskQueue").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("wOptions").Dot("TaskQueue").Op("=").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)))
					}))

					g.Add(
						jen.If(jen.Id("wOptions").Dot("WorkflowID").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("wOptions").Dot("WorkflowID").Op("=").Id(getFmtObject(gf, "Sprintf")).CallFunc(func(g *jen.Group) {
								g.Add(jen.Lit("%s/%s"))
								g.Add(jen.Lit(methName))
								g.Add(jen.Id(getUUIDObject(gf, "NewString")).Parens(jen.Null()))
							}))
						},
						),
					)

					if workflowOptions != nil {
						if workflowOptions.WorkflowExecutionTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowExecutionTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowExecutionTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(workflowOptions.WorkflowExecutionTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowRunTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(workflowOptions.WorkflowRunTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowTaskTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(workflowOptions.WorkflowRunTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.RetryPolicy != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("RetryPolicy").Op("==").Nil())).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("RetryPolicy").Op("=").Op("&").Id(getTemporalObject(gf, "RetryPolicy")).BlockFunc(func(g *jen.Group) {
									if workflowOptions.RetryPolicy.InitialInterval != nil {
										g.Add(jen.Id("InitialInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(workflowOptions.RetryPolicy.InitialInterval.Value))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumInterval != nil {
										g.Add(jen.Id("MaximumInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(workflowOptions.RetryPolicy.MaximumInterval.Value))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.BackoffCoefficient != nil {
										g.Add(jen.Id("BackoffCoefficient").Op(":").Float64().Call(jen.Lit(workflowOptions.RetryPolicy.BackoffCoefficient.Value)).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumAttempts != nil {
										g.Add(jen.Id("MaximumAttempts").Op(":").Lit(workflowOptions.RetryPolicy.MaximumAttempts.Value).Op(","))
									}
									if workflowOptions.RetryPolicy.NonRetryableErrorTypes != nil {
										g.Add(jen.Id("NonRetryableErrorTypes").Op(":").Index(jen.Null()).String().Block(jen.ListFunc(func(g *jen.Group) {
											for i := 0; i < len(workflowOptions.RetryPolicy.NonRetryableErrorTypes); i++ {
												g.Add(jen.Lit(workflowOptions.RetryPolicy.NonRetryableErrorTypes[i].Value).Op(","))
											}
										})).Op(","))
									}
								}))
							})
						}
					}

					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id(getTemporalWorkflowObject(gf, "ExecuteChildWorkflow")).CallFunc(func(g *jen.Group) {
							g.Add(jen.Id(getTemporalWorkflowObject(gf, "WithChildOptions")).CallFunc(func(g *jen.Group) {
								g.Add(jen.Id("ctx"))
								g.Add(jen.Id("wOptions"))
							}))
							g.Add(jen.Lit(methName))
							g.Add(jen.Id("req"))
						}))
					}))
				}).Line().Line()

			// Executes the workflow as a child workflow synchronously
			client.Comment(fmt.Sprintf("ExecuteChild%sSync executes the workflow as a child workflow and returns the result when finished", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("ExecuteChild%sSync", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context")))
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
				g.Add(jen.Id("options").Op("...").Id(getTemporalWorkflowObject(gf, "ChildWorkflowOptions")))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("future").Op(":=").Id("c").Dot(fmt.Sprintf("ExecuteChild%s", method.GoName)).CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("req"))
						g.Add(jen.Id("options").Op("..."))
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

			/*
				Workflow result structs

				This creates a struct like so

				type SomeServiceSomeFuncWorkflow struct {
					WorkflowID string
					RunID string

					client temporal.Client
				}

				And a bunch of methods like
				func (s *SomeServiceSomeFuncWorkflow).Cancel(ctx context.Context) error {
					wf :=s.client.Cancel(ctx, s.WorkflowID, s.RunID)
				}

				and so on
			*/

			wfObjName := fmt.Sprintf("%s%s", service.GoName, method.GoName)
			client.Comment(fmt.Sprintf("%s is a struct that wraps a workflow", wfObjName)).Line().
				Type().Id(wfObjName).StructFunc(func(g *jen.Group) {
				g.Add(jen.Id("WorkflowID").String())
				g.Add(jen.Id("RunID").String())
				g.Add(jen.Id("client").Id(getTemporalClientObject(gf, "Client")))
			}).Line()

			// Gets an instance of a workflow
			client.Comment(fmt.Sprintf("Get%s gets an instance of a given workflow", method.GoName)).Line().
				Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("Get%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("workflowId").String())
				g.Add(jen.Id("runId").String())
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(wfObjName))
				g.Add(jen.Error())
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
						}))
						g.Add(jen.Nil())
					}))
				}).Line().Line()

			// Cancels the workflow
			client.Comment("Cancel cancels a given workflow").Line().
				Func().Parens(jen.Id("c").Op("*").Id(wfObjName)).Id("Cancel").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("c").Dot("client").Dot("CancelWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("c").Dot("WorkflowID"))
							g.Add(jen.Id("c").Dot("RunID"))
						}))
					}))
				}).Line()

			// Terminates the workflow
			client.Comment("Terminates terminates a given workflow").Line().
				Func().Parens(jen.Id("c").Op("*").Id(wfObjName)).Id("Terminate").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
				g.Add(jen.Id("reason").String())
				g.Add(jen.Id("details").Op("...").Id("interface{}"))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.ReturnFunc(func(g *jen.Group) {
						g.Add(jen.Id("c").Dot("client").Dot("TerminateWorkflow").CallFunc(func(g *jen.Group) {
							g.Add(jen.Id("ctx"))
							g.Add(jen.Id("c").Dot("WorkflowID"))
							g.Add(jen.Id("c").Dot("RunID"))
							g.Add(jen.Id("reason"))
							g.Add(jen.Id("details").Op("..."))
						}))
					}))
				}).Line()

			// Gets the result of a workflow
			client.Comment("Get gets the result of a given workflow").Line().
				Func().Parens(jen.Id("c").Op("*").Id(wfObjName)).Id("Get").ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx").Id(getContext(gf)))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}).
				BlockFunc(func(g *jen.Group) {
					g.Add(jen.Id("future").Op(":=").Id("c").Dot("client").Dot("GetWorkflow").CallFunc(func(g *jen.Group) {
						g.Add(jen.Id("ctx"))
						g.Add(jen.Id("c").Dot("WorkflowID"))
						g.Add(jen.Id("c").Dot("RunID"))
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

			client.Line()
		case MethodTypeActivity:
			// Executes activity asynchronously and sets a bunch of defaults
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

					g.Add(
						jen.If(jen.Id("aOptions").Dot("ActivityID").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("aOptions").Dot("ActivityID").Op("=").Id(getFmtObject(gf, "Sprintf")).CallFunc(func(g *jen.Group) {
								g.Add(jen.Lit("%s/%s"))
								g.Add(jen.Lit(methName))
								g.Add(jen.Id(getUUIDObject(gf, "NewString")).Parens(jen.Null()))
							}))
						},
						),
					)

					g.Add(
						jen.If(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("=").Id(fmt.Sprintf("Default%sScheduleToCloseTimeout", service.GoName)))
						}),
					)

					g.Add(jen.If(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("=").Id(fmt.Sprintf("Default%sStartToCloseTimeout", service.GoName)))
					}))

					if activityOptions != nil {
						if activityOptions.StartToCloseTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(activityOptions.StartToCloseTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.ScheduleToCloseTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(activityOptions.ScheduleToCloseTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.ScheduleToStartTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToStartTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToStartTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(activityOptions.ScheduleToStartTimeout.Value))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.RetryPolicy != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("RetryPolicy").Op("==").Nil())).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("RetryPolicy").Op("=").Op("&").Id(getTemporalObject(gf, "RetryPolicy")).BlockFunc(func(g *jen.Group) {
									if activityOptions.RetryPolicy.InitialInterval != nil {
										g.Add(jen.Id("InitialInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(activityOptions.RetryPolicy.InitialInterval.Value))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if activityOptions.RetryPolicy.MaximumInterval != nil {
										g.Add(jen.Id("MaximumInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(activityOptions.RetryPolicy.MaximumInterval.Value))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if activityOptions.RetryPolicy.BackoffCoefficient != nil {
										g.Add(jen.Id("BackoffCoefficient").Op(":").Float64().Call(jen.Lit(activityOptions.RetryPolicy.BackoffCoefficient.Value)).Op(","))
									}
									if activityOptions.RetryPolicy.MaximumAttempts != nil {
										g.Add(jen.Id("MaximumAttempts").Op(":").Lit(activityOptions.RetryPolicy.MaximumAttempts.Value).Op(","))
									}
									if activityOptions.RetryPolicy.NonRetryableErrorTypes != nil {
										g.Add(jen.Id("NonRetryableErrorTypes").Op(":").Index(jen.Null()).String().Block(jen.ListFunc(func(g *jen.Group) {
											for i := 0; i < len(activityOptions.RetryPolicy.NonRetryableErrorTypes); i++ {
												g.Add(jen.Lit(activityOptions.RetryPolicy.NonRetryableErrorTypes[i].Value).Op(","))
											}
										})).Op(","))
									}
								}))
							})
						}
					}

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

			// Excutes the activity synchronously
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
						g.Add(jen.Id("TaskQueue").Op(":").Id("c").Dot("taskQueue").Op(","))
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
