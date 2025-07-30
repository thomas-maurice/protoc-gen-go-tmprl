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

func Client(gf *protogen.GeneratedFile, service *protogen.Service, config *Config) error {
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
		defaultActivityOptions := getDefaultActivityOptions(service)
		if activityOptions == nil {
			activityOptions = defaultActivityOptions
		} else if defaultActivityOptions != nil {
			if activityOptions.ScheduleToStartTimeout == nil {
				activityOptions.ScheduleToStartTimeout = defaultActivityOptions.ScheduleToStartTimeout
			}
			if activityOptions.ScheduleToCloseTimeout == nil {
				activityOptions.ScheduleToCloseTimeout = defaultActivityOptions.ScheduleToCloseTimeout
			}
			if activityOptions.StartToCloseTimeout == nil {
				activityOptions.StartToCloseTimeout = defaultActivityOptions.StartToCloseTimeout
			}
			if activityOptions.RetryPolicy == nil {
				activityOptions.RetryPolicy = defaultActivityOptions.RetryPolicy
			}
		}

		workflowOptions := getWorkflowOptions(method)
		defaultWorkflowOptions := getDefaultWorkflowOptions(service)
		if workflowOptions == nil {
			workflowOptions = defaultWorkflowOptions
		} else if defaultActivityOptions != nil {
			if workflowOptions.WorkflowExecutionTimeout == nil {
				workflowOptions.WorkflowExecutionTimeout = defaultWorkflowOptions.WorkflowExecutionTimeout
			}
			if workflowOptions.WorkflowRunTimeout == nil {
				workflowOptions.WorkflowRunTimeout = defaultWorkflowOptions.WorkflowRunTimeout
			}
			if workflowOptions.WorkflowTaskTimeout == nil {
				workflowOptions.WorkflowTaskTimeout = defaultWorkflowOptions.WorkflowTaskTimeout
			}
			if workflowOptions.RetryPolicy == nil {
				workflowOptions.RetryPolicy = defaultWorkflowOptions.RetryPolicy
			}
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
									g.Add(jen.Lit(*workflowOptions.WorkflowExecutionTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowRunTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*workflowOptions.WorkflowRunTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowTaskTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*workflowOptions.WorkflowRunTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.RetryPolicy != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("RetryPolicy").Op("==").Nil())).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("RetryPolicy").Op("=").Op("&").Id(getTemporalObject(gf, "RetryPolicy")).BlockFunc(func(g *jen.Group) {
									if workflowOptions.RetryPolicy.InitialInterval != nil {
										g.Add(jen.Id("InitialInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(*workflowOptions.RetryPolicy.InitialInterval))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumInterval != nil {
										g.Add(jen.Id("MaximumInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(*workflowOptions.RetryPolicy.MaximumInterval))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.BackoffCoefficient != nil {
										g.Add(jen.Id("BackoffCoefficient").Op(":").Float64().Call(jen.Lit(*workflowOptions.RetryPolicy.BackoffCoefficient)).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumAttempts != nil {
										g.Add(jen.Id("MaximumAttempts").Op(":").Lit(*workflowOptions.RetryPolicy.MaximumAttempts).Op(","))
									}
									if workflowOptions.RetryPolicy.NonRetryableErrorTypes != nil {
										g.Add(jen.Id("NonRetryableErrorTypes").Op(":").Index(jen.Null()).String().Block(jen.ListFunc(func(g *jen.Group) {
											for _, errType := range activityOptions.RetryPolicy.NonRetryableErrorTypes {
												g.Lit(errType)
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
				g.Add(jen.Error())
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

					if config.GenWorkflowPrefix {
						// we use side effects here to ensure that the workflow history won't be altered in case of replay
						g.Add(
							jen.If(jen.Id("wOptions").Dot("WorkflowID").Op("==").Lit("")).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Var().Id("id").String())
								g.Add(jen.Id("genId").Op(":=").Id(getTemporalWorkflowObject(gf, "SideEffect")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Id("ctx"))
									g.Add(jen.Func().Params(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context"))).Interface(jen.Null()).BlockFunc(func(g *jen.Group) {
										g.Add(jen.Return(jen.Id(getFmtObject(gf, "Sprintf")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit("%s/%s"))
											g.Add(jen.Lit(methName))
											g.Add(jen.Id(getUUIDObject(gf, "NewString")).Parens(jen.Null()))
										})))
									}))
								}))

								g.Add(jen.Id("err").Op(":=").Id("genId").Dot("Get").Params(jen.Op("&").Id("id")))

								g.Add(IfErrNilDouble)

								g.Add(jen.Id("wOptions").Dot("WorkflowID").Op("=").Id("id"))
							},
							),
						)
					}

					if workflowOptions != nil {
						if workflowOptions.WorkflowExecutionTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowExecutionTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowExecutionTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*workflowOptions.WorkflowExecutionTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowRunTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowRunTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*workflowOptions.WorkflowRunTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.WorkflowTaskTimeout != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("WorkflowTaskTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*workflowOptions.WorkflowRunTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if workflowOptions.RetryPolicy != nil {
							g.Add(jen.If(jen.Id("wOptions").Dot("RetryPolicy").Op("==").Nil())).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("wOptions").Dot("RetryPolicy").Op("=").Op("&").Id(getTemporalObject(gf, "RetryPolicy")).BlockFunc(func(g *jen.Group) {
									if workflowOptions.RetryPolicy.InitialInterval != nil {
										g.Add(jen.Id("InitialInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(*workflowOptions.RetryPolicy.InitialInterval))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumInterval != nil {
										g.Add(jen.Id("MaximumInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(*workflowOptions.RetryPolicy.MaximumInterval))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if workflowOptions.RetryPolicy.BackoffCoefficient != nil {
										g.Add(jen.Id("BackoffCoefficient").Op(":").Float64().Call(jen.Lit(*workflowOptions.RetryPolicy.BackoffCoefficient)).Op(","))
									}
									if workflowOptions.RetryPolicy.MaximumAttempts != nil {
										g.Add(jen.Id("MaximumAttempts").Op(":").Lit(*workflowOptions.RetryPolicy.MaximumAttempts).Op(","))
									}
									if workflowOptions.RetryPolicy.NonRetryableErrorTypes != nil {
										g.Add(jen.Id("NonRetryableErrorTypes").Op(":").Index(jen.Null()).String().Block(jen.ListFunc(func(g *jen.Group) {
											for _, errType := range activityOptions.RetryPolicy.NonRetryableErrorTypes {
												g.Lit(errType)
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
						g.Add(jen.Nil())
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
					g.Add(jen.Id("future").Op(",").Id("err").Op(":=").Id("c").Dot(fmt.Sprintf("ExecuteChild%s", method.GoName)).CallFunc(func(g *jen.Group) {
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

					if activityOptions != nil {
						if activityOptions.StartToCloseTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*activityOptions.StartToCloseTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.ScheduleToCloseTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*activityOptions.ScheduleToCloseTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						} else {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Id(fmt.Sprintf("Default%sActivityScheduleToCloseTimeout", service.GoName)))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.ScheduleToStartTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToStartTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToStartTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*activityOptions.ScheduleToStartTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.RetryPolicy != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("RetryPolicy").Op("==").Nil())).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("RetryPolicy").Op("=").Op("&").Id(getTemporalObject(gf, "RetryPolicy")).BlockFunc(func(g *jen.Group) {
									if activityOptions.RetryPolicy.InitialInterval != nil {
										g.Add(jen.Id("InitialInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(*activityOptions.RetryPolicy.InitialInterval))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if activityOptions.RetryPolicy.MaximumInterval != nil {
										g.Add(jen.Id("MaximumInterval").Op(":").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
											g.Add(jen.Lit(*activityOptions.RetryPolicy.MaximumInterval))
										}).Op("*").Id(getTimeObject(gf, "Second")).Op(","))
									}
									if activityOptions.RetryPolicy.BackoffCoefficient != nil {
										g.Add(jen.Id("BackoffCoefficient").Op(":").Float64().Call(jen.Lit(*activityOptions.RetryPolicy.BackoffCoefficient)).Op(","))
									}
									if activityOptions.RetryPolicy.MaximumAttempts != nil {
										g.Add(jen.Id("MaximumAttempts").Op(":").Lit(*activityOptions.RetryPolicy.MaximumAttempts).Op(","))
									}
									if activityOptions.RetryPolicy.NonRetryableErrorTypes != nil {
										g.Add(jen.Id("NonRetryableErrorTypes").Op(":").Index(jen.Null()).String().Values(jen.ListFunc(func(g *jen.Group) {
											for _, errType := range activityOptions.RetryPolicy.NonRetryableErrorTypes {
												g.Lit(errType)
											}
										})).Op(","))
									}
								}))
							})
						}

						if activityOptions.ScheduleToCloseTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*activityOptions.ScheduleToCloseTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.StartToCloseTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("StartToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*activityOptions.StartToCloseTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}

						if activityOptions.ScheduleToStartTimeout != nil {
							g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToStartTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
								g.Add(jen.Id("aOptions").Dot("ScheduleToStartTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
									g.Add(jen.Lit(*activityOptions.ScheduleToStartTimeout))
								}).Op("*").Id(getTimeObject(gf, "Second")))
							}))
						}
					} else {
						// At least specify a default start to close activity timeout otherwise temporal won't run them
						g.Add(jen.If(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("aOptions").Dot("ScheduleToCloseTimeout").Op("=").Id(getTimeObject(gf, "Duration")).CallFunc(func(g *jen.Group) {
								g.Add(jen.Id(fmt.Sprintf("Default%sActivityScheduleToCloseTimeout", service.GoName)))
							}).Op("*").Id(getTimeObject(gf, "Second")))
						}))
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
