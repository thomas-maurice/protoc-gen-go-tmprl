package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
)

const (
	workflowImport = "go.temporal.io/sdk/workflow"
	activityImport = "go.temporal.io/sdk/activity"
	workerImport   = "go.temporal.io/sdk/worker"
	clientImport   = "go.temporal.io/sdk/client"
	temporalImport = "go.temporal.io/sdk/temporal"
	uuidImport     = "github.com/google/uuid"
	fmtImport      = "fmt"
)

var (
	IfErrNilDouble = jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.ReturnFunc(func(g *jen.Group) {
			g.Add(jen.Nil())
			g.Add(jen.Id("err"))
		}),
	)

	IfErrNilSingle = jen.If(jen.Id("err").Op("!=").Nil()).Block(
		jen.ReturnFunc(func(g *jen.Group) {
			g.Add(jen.Id("err"))
		}),
	)
)

func getContext(gf *protogen.GeneratedFile) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: "context",
			GoName:       "Context",
		},
	)
}

func getTemporalWorkerObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: workerImport,
			GoName:       o,
		},
	)
}

func getTimeObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: "time",
			GoName:       o,
		},
	)
}

func getTemporalClientObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: clientImport,
			GoName:       o,
		},
	)
}

func getUUIDObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: uuidImport,
			GoName:       o,
		},
	)
}

func getTemporalObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: temporalImport,
			GoName:       o,
		},
	)
}

func getFmtObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: fmtImport,
			GoName:       o,
		},
	)
}

func getTemporalWorkflowObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: workflowImport,
			GoName:       o,
		},
	)
}

func getTemporalActivityObject(gf *protogen.GeneratedFile, o string) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: activityImport,
			GoName:       o,
		},
	)
}

func getSvcName(svc *protogen.Service) string {
	return fmt.Sprintf("%sService", svc.GoName)
}
