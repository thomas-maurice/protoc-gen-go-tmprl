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
			g.Add(jen.Id("client").Op("*").Id(getTemporalClient(gf)))
			g.Add(jen.Id("svc").Id(getSvcName(service)))
		}).Line().Line().
		// New worker func
		Comment(fmt.Sprintf("New%s: Returns a new instance of the worker", workerName)).Line().
		Func().Id(fmt.Sprintf("New%s", workerName)).
		ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("client").Op("*").Id(getTemporalClient(gf)))
			g.Add(jen.Id("svc").Id(getSvcName(service)))
		}).
		Parens(jen.ListFunc(func(g *jen.Group) {
			g.Add(jen.Op("*").Id(workerName))
			g.Add(jen.Error())
		})).
		Block(
			jen.ReturnFunc(func(g *jen.Group) {
				g.ListFunc(func(g *jen.Group) {
					g.Add(jen.Op("&").Id(workerName).BlockFunc(func(g *jen.Group) {
						g.Add(jen.Id("client").Op(":").Id("client")).Op(",")
						g.Add(jen.Id("svc").Op(":").Id("svc")).Op(",")
					}),
					)
					g.Add(jen.Nil())
				},
				)
			},
			),
		)

	buf := bytes.NewBufferString("")

	err := worker.Render(buf)
	if err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
