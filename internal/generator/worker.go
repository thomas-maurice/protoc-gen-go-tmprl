package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

func Worker(gf *protogen.GeneratedFile, service *protogen.Service) error {
	workerName := fmt.Sprintf("%sWorker", service.GoName)

	worker := jen.Comment(fmt.Sprintf("%s: Worker for the %s service", workerName, service.GoName)).Line().
		Type().Id(workerName).
		StructFunc(func(g *jen.Group) {
			g.Add(jen.Id("client").Id(getTemporalClientObject(gf, "Client")))
			g.Add(jen.Id("worker").Id(getTemporalWorkerObject(gf, "Worker")))
			g.Add(jen.Id("svc").Id(getSvcName(service)))
		}).Line().Line().
		// New worker func
		Comment(fmt.Sprintf("New%s: Returns a new instance of the worker.", workerName)).Line().
		Comment("If `taskQueue` stays empty the default one will be used").Line().
		Func().Id(fmt.Sprintf("New%s", workerName)).
		ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("client").Id(getTemporalClientObject(gf, "Client")))
			g.Add(jen.Id("svc").Id(getSvcName(service)))
			g.Add(jen.Id("taskQueue").String())
			g.Add(jen.Id("workerOptions").Op("...").Id(getTemporalWorkerObject(gf, "Options")))
		}).
		Parens(jen.ListFunc(func(g *jen.Group) {
			g.Add(jen.Op("*").Id(workerName))
			g.Add(jen.Error())
		})).
		BlockFunc(func(g *jen.Group) {
			g.Add(jen.Id("wOpts").Op(":=").Id(getTemporalWorkerObject(gf, "Options")).Block())
			g.Add(jen.If(jen.Id("taskQueue").Op("==").Lit("").Block(
				jen.Id("taskQueue").Op("=").Id(fmt.Sprintf("Default%sTaskQueueName", service.GoName)),
			)))
			g.Add(jen.If(jen.Len(jen.Id("workerOptions")).Op(">").Lit(0).Block(
				jen.Id("wOpts").Op("=").Id("workerOptions").Index(jen.Lit(0)),
			)))
			g.Add(
				jen.Id("w").Op(":=").Id(gf.QualifiedGoIdent(
					protogen.GoIdent{
						GoImportPath: workerImport,
						GoName:       "New",
					},
				)).Parens(
					jen.List(
						jen.Id("client"),
						jen.Id("taskQueue"),
						jen.Id("wOpts"),
					),
				))

			g.Add(
				jen.ReturnFunc(func(g *jen.Group) {
					g.ListFunc(func(g *jen.Group) {
						g.Add(jen.Op("&").Id(workerName).BlockFunc(func(g *jen.Group) {
							g.Add(jen.Id("client").Op(":").Id("client")).Op(",")
							g.Add(jen.Id("svc").Op(":").Id("svc")).Op(",")
							g.Add(jen.Id("worker").Op(":").Id("w")).Op(",")
						}),
						)
						g.Add(jen.Nil())
					},
					)
				}),
			)
		}).Line().Line().
		// Register func, this will register activities and workflows in the client
		Comment("Register registers the worker and its activities/workflows in temporal").Line().
		Func().Parens(jen.Id("w").Op("*").Id(workerName)).Id("Register").ParamsFunc(func(g *jen.Group) {}).BlockFunc(func(g *jen.Group) {
		for _, m := range service.Methods {
			switch t, _ := getMethodType(m); t {
			case MethodTypeActivity:
				/*
					w.client.RegisterActivityWithOptions(w.svc.Activity, activity.RegisterOptions{
						Name: "example.v1.Activity",
					})
				*/
				name, err := getMethodRegisteredName(m)
				if err != nil {
					panic(err)
				}
				g.Add(
					jen.Comment(fmt.Sprintf("Registers activity %s", m.GoName)).Line().
						Id("w").Dot("worker").Dot("RegisterActivityWithOptions").Parens(
						jen.Id("w").Dot("svc").Dot(m.GoName).Op(",").Id(getTemporalActivityObject(gf, "RegisterOptions")).Block(
							jen.Id("Name").Op(":").Lit(name).Op(","),
						),
					),
				)

			case MethodTypeWorkflow:
				/*
					w.client.RegisterActivityWithOptions(w.svc.Workflow, workflow.RegisterOptions{
						Name: "example.v1.Workflow",
					})
				*/
				name, err := getMethodRegisteredName(m)
				if err != nil {
					panic(err)
				}
				g.Add(
					jen.Comment(fmt.Sprintf("Registers workflow %s", m.GoName)).Line().
						Id("w").Dot("worker").Dot("RegisterWorkflowWithOptions").Parens(
						jen.Id("w").Dot("svc").Dot(m.GoName).Op(",").Id(getTemporalWorkflowObject(gf, "RegisterOptions")).Block(
							jen.Id("Name").Op(":").Lit(name).Op(","),
						),
					),
				)
			}
		}

		jen.ReturnFunc(func(g *jen.Group) {
			g.Add(jen.Nil())
		})
	}).Line().
		/*
			// Run func like so
			func (w *Worker) Run() error {
				return w.Worker.Run(worker.InterruptCh())
			}
		*/
		Comment("Run will run the worker").Line().
		Func().Parens(jen.Id("w").Op("*").Id(workerName)).Id("Run").ParamsFunc(func(g *jen.Group) {}).Error().BlockFunc(func(g *jen.Group) {
		g.Add(
			jen.Return(
				jen.Id("w").Dot("worker").Dot("Run").Parens(
					jen.Id(getTemporalWorkerObject(gf, "InterruptCh")).Parens(jen.Null()),
				),
			),
		)
	}).Line().
		/*
			// Stop func like so
			func (w *Worker) Stop() error {
				return w.Worker.Stop()
			}
		*/
		Comment("Stop will stop the worker, may panic if called twice").Line().
		Func().Parens(jen.Id("w").Op("*").Id(workerName)).Id("Stop").ParamsFunc(func(g *jen.Group) {}).BlockFunc(func(g *jen.Group) {
		g.Add(
			jen.Id("w").Dot("worker").Dot("Stop").Parens(
				jen.Null(),
			),
		)
	})

	buf := bytes.NewBufferString("")

	err := worker.Render(buf)
	if err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
