package tmpl

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"google.golang.org/protobuf/compiler/protogen"
)

// FuncMap: Returns the template function map
func FuncMap(gf *protogen.GeneratedFile) template.FuncMap {
	return template.FuncMap{
		"QualifiedGoIdent":        func(ident protogen.GoIdent) string { return gf.QualifiedGoIdent(ident) },
		"TemporalClient":          func(name string) string { return qualifiedTemporalIdent(gf, "go.temporal.io/sdk/client", name) },
		"TemporalWorker":          func(name string) string { return qualifiedTemporalIdent(gf, "go.temporal.io/sdk/worker", name) },
		"TemporalWorkflow":        func(name string) string { return qualifiedTemporalIdent(gf, "go.temporal.io/sdk/workflow", name) },
		"TemporalActivity":        func(name string) string { return qualifiedTemporalIdent(gf, "go.temporal.io/sdk/activity", name) },
		"Temporal":                func(name string) string { return qualifiedTemporalIdent(gf, "go.temporal.io/sdk/temporal", name) },
		"Context":                 func() string { return qualifiedIdent(gf, "context", "Context") },
		"Time":                    func(name string) string { return qualifiedIdent(gf, "time", name) },
		"Fmt":                     func(name string) string { return qualifiedIdent(gf, "fmt", name) },
		"UUID":                    func(name string) string { return qualifiedIdent(gf, "github.com/google/uuid", name) },
		"toSeconds":               toSeconds,
		"HasTimeout":              hasTimeout,
		"HasRetryPolicy":          hasRetryPolicy,
		"Join":                    strings.Join,
		"Quote":                   func(s string) string { return fmt.Sprintf("%q", s) },
		"Add":                     func(a, b int) int { return a + b },
		"WorkflowObjectName":      workflowObjectName,
		"ChildWorkflowObjectName": childWorkflowObjectName,
		"dict":                    dict,
		"commentOneLine":          commentOneLine,
	}
}

// qualifiedTemporalIdent: Returns a qualified identifier for temporal imports
func qualifiedTemporalIdent(gf *protogen.GeneratedFile, importPath, name string) string {
	return gf.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: protogen.GoImportPath(importPath),
		GoName:       name,
	})
}

// qualifiedIdent: Returns a qualified identifier for standard imports
func qualifiedIdent(gf *protogen.GeneratedFile, importPath, name string) string {
	return gf.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: protogen.GoImportPath(importPath),
		GoName:       name,
	})
}

// toSeconds: Converts a duration to seconds for template usage
func toSeconds(d interface{}) int64 {
	switch v := d.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case int32:
		return int64(v)
	case time.Duration:
		return int64(v / time.Second)
	default:
		return 0
	}
}

// hasTimeout: Checks if a timeout value is set
func hasTimeout(d interface{}) bool {
	return toSeconds(d) > 0
}

// hasRetryPolicy: Checks if retry policy is not nil
func hasRetryPolicy(rp interface{}) bool {
	return rp != nil
}

// workflowObjectName: Generates the workflow object name
func workflowObjectName(serviceName, methodName string) string {
	return fmt.Sprintf("%s%s", serviceName, methodName)
}

// childWorkflowObjectName: Generates the child workflow object name
func childWorkflowObjectName(serviceName, methodName string) string {
	return fmt.Sprintf("Child%s%sExecution", serviceName, methodName)
}

// dict: Creates a map from key-value pairs for template usage
func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// commentOneLine: Converts a multiline comment to a single line
func commentOneLine(s string) string {
	// Replace newlines with spaces and collapse multiple spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	// Collapse multiple spaces
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}
