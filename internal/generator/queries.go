package generator

import (
	"bytes"
	"fmt"

	"github.com/dave/jennifer/jen"
	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func ServiceQueries(gf *protogen.GeneratedFile, service *protogen.Service) error {
	clientName := getClientName(service)

	queries := jen.Null()

	for _, method := range service.Methods {
		t, err := getMethodType(method)
		if err != nil {
			return err
		}

		if t != MethodTypeQuery {
			continue
		}

		methName, err := getMethodRegisteredName(method)
		if err != nil {
			return err
		}

		queryOpts, _ := proto.GetExtension(method.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions)

		queries.Comment(fmt.Sprintf("Query%s sends the %s query to a workflow", method.GoName, method.GoName)).Line().
			Func().Parens(jen.Id("c").Op("*").Id(clientName)).Id(fmt.Sprintf("Query%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("ctx").Id(getContext(gf)))
			g.Add(jen.Id("workflowID").String())
			g.Add(jen.Id("runID").String())
			g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
		}).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
			g.Add(jen.Error())
		}).BlockFunc(func(g *jen.Group) {
			queryName := methName
			if queryOpts != nil && queryOpts.Name != "" {
				queryName = queryOpts.Name
			}

			g.Add(jen.Id("future").Op(",").Err().Op(":=").Id("c").Dot("client").Dot("QueryWorkflow").CallFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx"))
				g.Add(jen.Id("workflowID"))
				g.Add(jen.Id("runID"))
				g.Add(jen.Lit(queryName))
				g.Add(jen.Id("req"))
			}))

			g.Add(IfErrNilDouble)

			g.Add(jen.Var().Id("resp").Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))

			g.Add(jen.Id("err").Op("=").Id("future").Dot("Get").CallFunc(func(g *jen.Group) {
				g.Add(jen.Op("&").Id("resp"))
			}))

			g.Add(IfErrNilDouble)

			g.Add(jen.ReturnFunc(func(g *jen.Group) {
				g.Add(jen.Id("resp"))
				g.Add(jen.Nil())
			}))

		}).Line()

		queries.Comment(fmt.Sprintf("HandleQuery%s sets up the %s query and responds accordingly, returns an error if it failed", method.GoName, method.GoName)).Line().
			Func().Id(fmt.Sprintf("HandleQuery%s", method.GoName)).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("ctx").Id(getTemporalWorkflowObject(gf, "Context")))
			g.Add(jen.Id("queryFunc").Func().ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Id("req").Op("*").Id(gf.QualifiedGoIdent(method.Input.GoIdent)))
			}).ParamsFunc(func(g *jen.Group) {
				g.Add(jen.Op("*").Id(gf.QualifiedGoIdent(method.Output.GoIdent)))
				g.Add(jen.Error())
			}))
		}).ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Error())
		}).BlockFunc(func(g *jen.Group) {
			queryName := methName
			if queryOpts != nil && queryOpts.Name != "" {
				queryName = queryOpts.Name
			}

			g.Add(jen.Return().Id(getTemporalWorkflowObject(gf, "SetQueryHandler")).CallFunc(func(g *jen.Group) {
				g.Add(jen.Id("ctx"))
				g.Add(jen.Lit(queryName))
				g.Add(jen.Id("queryFunc"))
			}))
		}).Line()
	}

	buf := bytes.NewBufferString("")
	if err := queries.Render(buf); err != nil {
		return err
	}

	gf.P(buf.String())

	return nil
}
