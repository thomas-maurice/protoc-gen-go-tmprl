package generator

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

const (
	workflowContextImport = "go.temporal.io/sdk/workflow"
	clientContextImport   = "go.temporal.io/sdk/client"
)

func getContext(gf *protogen.GeneratedFile) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: "context",
			GoName:       "Context",
		},
	)
}

func getWorkflowContext(gf *protogen.GeneratedFile) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: workflowContextImport,
			GoName:       "Context",
		},
	)
}
func getTemporalClient(gf *protogen.GeneratedFile) string {
	return gf.QualifiedGoIdent(
		protogen.GoIdent{
			GoImportPath: clientContextImport,
			GoName:       "Client",
		},
	)
}

func getSvcName(svc *protogen.Service) string {
	return fmt.Sprintf("%sService", svc.GoName)
}
