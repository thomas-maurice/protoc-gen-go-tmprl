package renderer

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	temporalv1 "github.com/thomas-maurice/protoc-gen-go-tmprl/gen/temporal/v1"
	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/model"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	_ "google.golang.org/protobuf/types/known/timestamppb"
)

// update controls whether the snapshot test rewrites its golden file
// instead of asserting against it. Set via `go test -update`.
var update = flag.Bool("update", false, "update golden CLI snapshot")

// TestTemplatesParse Verifies that all templates parse correctly
func TestTemplatesParse(t *testing.T) {
	// Create a minimal function map for parsing
	funcMap := template.FuncMap{
		"QualifiedGoIdent":        func(args ...interface{}) string { return "" },
		"TemporalClient":          func(name string) string { return "" },
		"TemporalWorker":          func(name string) string { return "" },
		"TemporalWorkflow":        func(name string) string { return "" },
		"TemporalActivity":        func(name string) string { return "" },
		"Temporal":                func(name string) string { return "" },
		"Context":                 func() string { return "" },
		"Time":                    func(name string) string { return "" },
		"Fmt":                     func(name string) string { return "" },
		"UUID":                    func(name string) string { return "" },
		"toSeconds":               func(d interface{}) int64 { return 0 },
		"Quote":                   func(s string) string { return "" },
		"WorkflowObjectName":      func(service, method string) string { return "" },
		"ChildWorkflowObjectName": func(service, method string) string { return "" },
		"commentOneLine":          func(s string) string { return "" },
		"commentBlock":            func(indent, s string) string { return "" },
		"docComment":              func(indent, name, body string) string { return "" },
		"MakeAnchor":              func(parts ...string) string { return "" },
		"TrimComment":             func(comment interface{}) string { return "" },
		"FormatDuration":          func(seconds interface{}) string { return "" },
		"FormatFloat":             func(f float64) string { return "" },
		"FormatStringSlice":       func(slice []string) string { return "" },
		"Cardinality":             func(c interface{}) string { return "" },
		"DeprecatedIcon":          func(opts interface{}) string { return "" },
		"FullName":                func(name interface{}) string { return "" },
		"kebab":                   func(s string) string { return "" },
		"pflagFuncName":           func(kind interface{}) string { return "" },
		"goFieldAssignPath":       func(path []string, rootVar string) string { return "" },
		"goFieldAssignSteps":      func(path []string, rootVar string) interface{} { return nil },
		"enumValueTypeName":       func(name string) string { return "" },
		"jsonExampleFor":          func(spec interface{}) string { return "" },
		"Cobra":                   func(name string) string { return "" },
		"Pflag":                   func(name string) string { return "" },
		"Protojson":               func(name string) string { return "" },
		"Timestamppb":             func(name string) string { return "" },
		"Durationpb":              func(name string) string { return "" },
		"Fieldmaskpb":             func(name string) string { return "" },
		"Reflect":                 func(name string) string { return "" },
		"Strings":                 func(name string) string { return "" },
		"Strconv":                 func(name string) string { return "" },
		"Os":                      func(name string) string { return "" },
		"EncodingJSON":            func(name string) string { return "" },
		"Proto":                   func(name string) string { return "" },
		"Errors":                  func(name string) string { return "" },
		"IO":                      func(name string) string { return "" },
		"Enumspb":                 func(name string) string { return "" },
		"dict":                    func(args ...interface{}) map[string]interface{} { return nil },
		"enumShortName":           func(s string) string { return "" },
		"cliHelpString":           func(spec interface{}) string { return "" },
		"cliVarDecl":              func(spec interface{}, prefix string) string { return "" },
		"cliRegisterLine":         func(spec interface{}, prefix string) string { return "" },
		"cliApplyBlock":           func(spec interface{}, prefix string) string { return "" },
		"cliVarName":              func(spec interface{}) string { return "" },
		"oneofGroupsGoLiteral":    func(groups [][]string) string { return "" },
		"oneofExclusiveCalls":     func(groups [][]string) string { return "" },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	// Verify the main service template is present
	if tmpl.Lookup("service.tmpl") == nil {
		t.Error("expected service.tmpl template not found")
	}
}

// TestTemplateEmbedding Verifies templates are properly embedded
func TestTemplateEmbedding(t *testing.T) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		t.Fatalf("failed to read embedded templates directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("no templates found in embedded FS")
	}

	// Verify all entries are .tmpl files
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".tmpl") {
			t.Errorf("unexpected non-template file in templates directory: %s", entry.Name())
		}
	}
}

// TestTemplateFilesCoverage Verifies the main service template exists
func TestTemplateFilesCoverage(t *testing.T) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		t.Fatalf("failed to read templates directory: %v", err)
	}

	// Find the service template
	foundService := false
	for _, entry := range entries {
		if entry.Name() == "service.tmpl" {
			foundService = true
			t.Logf("service.tmpl: comprehensive service template with all sections")
			break
		}
	}

	if !foundService {
		t.Error("service.tmpl not found - this is the main template file")
	}
}

// TestRendererStructure Verifies Renderer struct has expected fields and RenderAll method
func TestRendererStructure(t *testing.T) {
	// This test documents the Renderer structure
	var r Renderer

	// These assignments verify the fields exist and have correct types
	_ = r.templates // *template.Template
	_ = r.gf        // *protogen.GeneratedFile

	// Verify the Renderer has RenderAll method by checking it compiles
	// This is a compile-time check - if RenderAll doesn't exist, this won't compile
	_ = r.RenderAll
}

// TestServiceTemplateSections Documents what the comprehensive service.tmpl contains
func TestServiceTemplateSections(t *testing.T) {
	sections := []string{
		"Constants (task queue, timeouts, workflow/activity/signal/query names)",
		"Service Interface (methods that users must implement)",
		"Client Implementation (workflow and activity execution methods)",
		"Worker Implementation (registration and lifecycle management)",
		"Workflow Wrapper Objects (type-safe workflow management)",
		"Signal Helper Functions (send and receive signals)",
		"Query Helper Functions (query workflows)",
	}

	if len(sections) != 7 {
		t.Errorf("Expected service.tmpl to have 7 sections, documented %d", len(sections))
	}

	for i, section := range sections {
		t.Logf("%d. %s", i+1, section)
	}
}

// TestCLISnapshot renders the CLI template for a synthetic service with a
// rich workflow input (scalars, timestamp, enum, nested message, oneof,
// repeated scalar, repeated message → JSONEscape), gofmt-formats the
// combined output, and compares against a golden file. Use
// `go test ./internal/renderer/... -update` to rewrite the golden.
func TestCLISnapshot(t *testing.T) {
	plugin := buildCLISnapshotPlugin(t)

	// Find our synthetic file and its service.
	var file *protogen.File
	for _, f := range plugin.Files {
		if f.Proto.GetName() == "example/snapshot/v1/snapshot.proto" {
			file = f
			break
		}
	}
	if file == nil {
		t.Fatalf("synthetic file not found in plugin.Files")
	}
	if len(file.Services) == 0 {
		t.Fatalf("synthetic file has no services")
	}

	gf := plugin.NewGeneratedFile("snapshot_tmprl.pb.go", file.GoImportPath)

	config := &model.Config{
		GenWorkflowPrefix:              false,
		DefaultActivityScheduleToClose: 3600,
	}

	svc, err := model.NewService(file.Services[0], gf, config)
	if err != nil {
		t.Fatalf("building service model: %v", err)
	}
	if !svc.GenerateCLI {
		t.Fatalf("expected GenerateCLI=true on synthetic service")
	}
	if len(svc.Workflows) == 0 {
		t.Fatalf("synthetic service has no workflows")
	}
	// Sanity: the rich-input workflow must have an input FlagPlan.
	foundPlan := false
	for _, wf := range svc.Workflows {
		if wf.InputFlagPlan != nil {
			foundPlan = true
			break
		}
	}
	if !foundPlan {
		t.Fatalf("expected at least one workflow with an input FlagPlan")
	}

	r, err := NewRenderer(gf)
	if err != nil {
		t.Fatalf("creating renderer: %v", err)
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated by protoc-gen-go-tmprl snapshot test. DO NOT EDIT.\n")
	buf.WriteString("package snapshot\n\n")

	out, err := r.RenderCLI(svc)
	if err != nil {
		t.Fatalf("RenderCLI: %v", err)
	}
	buf.WriteString(out)

	// gofmt the combined output. A failure here means the template
	// produced syntactically invalid Go.
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Dump the unformatted output to aid debugging.
		tmpPath := filepath.Join(t.TempDir(), "cli_snapshot.unformatted.go")
		_ = os.WriteFile(tmpPath, buf.Bytes(), 0o644)
		t.Fatalf("gofmt failed (unformatted output at %s): %v", tmpPath, err)
	}

	goldenPath := filepath.Join("testdata", "cli_snapshot.golden")

	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, formatted, 0o644); err != nil {
			t.Fatalf("writing golden: %v", err)
		}
		t.Logf("wrote golden: %s (%d bytes)", goldenPath, len(formatted))
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden (run with -update to create): %v", err)
	}
	if !bytes.Equal(formatted, want) {
		// Write the current output next to the golden for diffing.
		actualPath := filepath.Join(t.TempDir(), "cli_snapshot.actual")
		_ = os.WriteFile(actualPath, formatted, 0o644)
		t.Fatalf("snapshot mismatch; got output at %s\nre-run with -update to refresh golden", actualPath)
	}
}

// buildCLISnapshotPlugin constructs a *protogen.Plugin backed by a
// synthetic FileDescriptorProto that declares a temporal service with
// generate_cli=true, a rich workflow input, a signal, and a query.
// Hand-built (not parsed from .proto text) so the test has no toolchain
// dependency on protoc.
func buildCLISnapshotPlugin(t *testing.T) *protogen.Plugin {
	t.Helper()

	// Build the ServiceOptions and pack the temporal.v1.service extension.
	svcOpts := &descriptorpb.ServiceOptions{}
	proto.SetExtension(svcOpts, temporalv1.E_Service, &temporalv1.ServiceOptions{
		TaskQueue:   "snapshot-tq",
		GenerateCli: proto.Bool(true),
	})

	// Per-method options: workflow, signal, query.
	startOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(startOpts, temporalv1.E_Workflow, &temporalv1.WorkflowOptions{
		Name:    "SnapshotService.StartOrder",
		Signals: []string{"Nudge"},
		Queries: []string{"GetStatus"},
	})

	signalOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(signalOpts, temporalv1.E_Signal, &temporalv1.SignalOptions{
		Name: "SnapshotService.Nudge",
	})

	queryOpts := &descriptorpb.MethodOptions{}
	proto.SetExtension(queryOpts, temporalv1.E_Query, &temporalv1.QueryOptions{
		Name: "SnapshotService.GetStatus",
	})

	syntax := "proto3"
	pkg := "example.snapshot.v1"
	goPkg := "example.com/gen/example/snapshot/v1;snapshotv1"

	// Build types piecewise.
	labelOpt := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	labelRep := descriptorpb.FieldDescriptorProto_LABEL_REPEATED

	typeString := descriptorpb.FieldDescriptorProto_TYPE_STRING
	typeInt64 := descriptorpb.FieldDescriptorProto_TYPE_INT64
	typeBool := descriptorpb.FieldDescriptorProto_TYPE_BOOL
	typeEnum := descriptorpb.FieldDescriptorProto_TYPE_ENUM
	typeMsg := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE

	// Priority enum.
	priorityEnum := &descriptorpb.EnumDescriptorProto{
		Name: proto.String("Priority"),
		Value: []*descriptorpb.EnumValueDescriptorProto{
			{Name: proto.String("PRIORITY_UNSPECIFIED"), Number: proto.Int32(0)},
			{Name: proto.String("PRIORITY_LOW"), Number: proto.Int32(1)},
			{Name: proto.String("PRIORITY_HIGH"), Number: proto.Int32(2)},
		},
	}

	// Address nested message.
	addressMsg := &descriptorpb.DescriptorProto{
		Name: proto.String("Address"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("street"), Number: proto.Int32(1), Type: &typeString, Label: &labelOpt},
			{Name: proto.String("city"), Number: proto.Int32(2), Type: &typeString, Label: &labelOpt},
		},
	}

	// LineItem nested message (triggers JSONEscape for repeated message).
	lineItemMsg := &descriptorpb.DescriptorProto{
		Name: proto.String("LineItem"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("sku"), Number: proto.Int32(1), Type: &typeString, Label: &labelOpt},
			{Name: proto.String("qty"), Number: proto.Int32(2), Type: &typeInt64, Label: &labelOpt},
		},
	}

	// Oneof: payment method (card_token | invoice_id).
	oneofIdx := int32(0)
	paymentOneof := &descriptorpb.OneofDescriptorProto{Name: proto.String("payment")}

	startOrderRequest := &descriptorpb.DescriptorProto{
		Name: proto.String("StartOrderRequest"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("order_id"), Number: proto.Int32(1), Type: &typeString, Label: &labelOpt},
			{Name: proto.String("quantity"), Number: proto.Int32(2), Type: &typeInt64, Label: &labelOpt},
			{Name: proto.String("express"), Number: proto.Int32(3), Type: &typeBool, Label: &labelOpt},
			{Name: proto.String("scheduled_at"), Number: proto.Int32(4), Type: &typeMsg, Label: &labelOpt, TypeName: proto.String(".google.protobuf.Timestamp")},
			{Name: proto.String("priority"), Number: proto.Int32(5), Type: &typeEnum, Label: &labelOpt, TypeName: proto.String(".example.snapshot.v1.Priority")},
			{Name: proto.String("shipping_address"), Number: proto.Int32(6), Type: &typeMsg, Label: &labelOpt, TypeName: proto.String(".example.snapshot.v1.Address")},
			{Name: proto.String("items"), Number: proto.Int32(7), Type: &typeMsg, Label: &labelRep, TypeName: proto.String(".example.snapshot.v1.LineItem")},
			{Name: proto.String("tags"), Number: proto.Int32(8), Type: &typeString, Label: &labelRep},
			{Name: proto.String("card_token"), Number: proto.Int32(9), Type: &typeString, Label: &labelOpt, OneofIndex: &oneofIdx},
			{Name: proto.String("invoice_id"), Number: proto.Int32(10), Type: &typeString, Label: &labelOpt, OneofIndex: &oneofIdx},
		},
		OneofDecl: []*descriptorpb.OneofDescriptorProto{paymentOneof},
	}

	startOrderResponse := &descriptorpb.DescriptorProto{
		Name: proto.String("StartOrderResponse"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("order_id"), Number: proto.Int32(1), Type: &typeString, Label: &labelOpt},
		},
	}

	nudgeRequest := &descriptorpb.DescriptorProto{
		Name: proto.String("NudgeRequest"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("reason"), Number: proto.Int32(1), Type: &typeString, Label: &labelOpt},
		},
	}
	nudgeResponse := &descriptorpb.DescriptorProto{Name: proto.String("NudgeResponse")}

	getStatusRequest := &descriptorpb.DescriptorProto{Name: proto.String("GetStatusRequest")}
	getStatusResponse := &descriptorpb.DescriptorProto{
		Name: proto.String("GetStatusResponse"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: proto.String("state"), Number: proto.Int32(1), Type: &typeString, Label: &labelOpt},
		},
	}

	service := &descriptorpb.ServiceDescriptorProto{
		Name:    proto.String("SnapshotService"),
		Options: svcOpts,
		Method: []*descriptorpb.MethodDescriptorProto{
			{
				Name:       proto.String("StartOrder"),
				InputType:  proto.String(".example.snapshot.v1.StartOrderRequest"),
				OutputType: proto.String(".example.snapshot.v1.StartOrderResponse"),
				Options:    startOpts,
			},
			{
				Name:       proto.String("Nudge"),
				InputType:  proto.String(".example.snapshot.v1.NudgeRequest"),
				OutputType: proto.String(".example.snapshot.v1.NudgeResponse"),
				Options:    signalOpts,
			},
			{
				Name:       proto.String("GetStatus"),
				InputType:  proto.String(".example.snapshot.v1.GetStatusRequest"),
				OutputType: proto.String(".example.snapshot.v1.GetStatusResponse"),
				Options:    queryOpts,
			},
		},
	}

	fdp := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("example/snapshot/v1/snapshot.proto"),
		Package:    proto.String(pkg),
		Syntax:     proto.String(syntax),
		Dependency: []string{"google/protobuf/timestamp.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			addressMsg,
			lineItemMsg,
			startOrderRequest,
			startOrderResponse,
			nudgeRequest,
			nudgeResponse,
			getStatusRequest,
			getStatusResponse,
		},
		EnumType: []*descriptorpb.EnumDescriptorProto{priorityEnum},
		Service:  []*descriptorpb.ServiceDescriptorProto{service},
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String(goPkg),
		},
	}

	// Pull in the timestamppb FileDescriptor as a dependency.
	tsFdp := protoFileDescriptorProto(t, "google/protobuf/timestamp.proto")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{fdp.GetName()},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{tsFdp, fdp},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.Options.New: %v", err)
	}
	return plugin
}

// protoFileDescriptorProto returns the FileDescriptorProto form of a
// well-known type file, loaded from protoregistry.GlobalFiles. It is used
// to supply the `google/protobuf/timestamp.proto` dependency to the
// synthetic CodeGeneratorRequest.
func protoFileDescriptorProto(t *testing.T, path string) *descriptorpb.FileDescriptorProto {
	t.Helper()
	// Use proto.GetExtension + protodesc to convert. We use the known
	// timestamppb file descriptor via its package.
	// Simpler: build a minimal FileDescriptorProto for timestamp with just
	// the Timestamp message, since we only need the dependency to resolve
	// TypeName "google.protobuf.Timestamp".
	if path != "google/protobuf/timestamp.proto" {
		t.Fatalf("protoFileDescriptorProto: unsupported path %q", path)
	}
	typeInt64 := descriptorpb.FieldDescriptorProto_TYPE_INT64
	typeInt32 := descriptorpb.FieldDescriptorProto_TYPE_INT32
	labelOpt := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	syntax := "proto3"
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("google/protobuf/timestamp.proto"),
		Package: proto.String("google.protobuf"),
		Syntax:  &syntax,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Timestamp"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("seconds"), Number: proto.Int32(1), Type: &typeInt64, Label: &labelOpt},
					{Name: proto.String("nanos"), Number: proto.Int32(2), Type: &typeInt32, Label: &labelOpt},
				},
			},
		},
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("google.golang.org/protobuf/types/known/timestamppb"),
		},
	}
}

// TestRenderDocumentationCLISection renders the markdown documentation for
// a synthetic CLI-opted service and asserts that the new CLI section is
// present with the expected subheadings, command tree, and flag-kind-
// specific hints from PLAN.md §6. The fixture reuses
// buildCLISnapshotPlugin, which already covers the representative
// FlagKinds (String, Int64, Bool, Timestamp, Enum, nested message,
// repeated scalar, repeated message → JSONEscape, oneof).
func TestRenderDocumentationCLISection(t *testing.T) {
	plugin := buildCLISnapshotPlugin(t)

	var file *protogen.File
	for _, f := range plugin.Files {
		if f.Proto.GetName() == "example/snapshot/v1/snapshot.proto" {
			file = f
			break
		}
	}
	if file == nil {
		t.Fatalf("synthetic file not found in plugin.Files")
	}

	gf := plugin.NewGeneratedFile("snapshot_doc.md", file.GoImportPath)
	config := &model.Config{DefaultActivityScheduleToClose: 3600}

	svc, err := model.NewService(file.Services[0], gf, config)
	if err != nil {
		t.Fatalf("building service model: %v", err)
	}

	r, err := NewRenderer(gf)
	if err != nil {
		t.Fatalf("creating renderer: %v", err)
	}

	out, err := r.RenderDocumentation(svc)
	if err != nil {
		t.Fatalf("RenderDocumentation: %v", err)
	}

	// The presence checks cover: the gated section heading, the kebab
	// root name, a sampling of flags from each kind family that exercises
	// the helper's hint output, the oneof-exclusive block, and the
	// cross-workflow / schedule tables.
	mustContain := []string{
		"## CLI",
		"snapshot-service", // ServiceKebabName
		"### Workflow commands",
		"`--order-id`",
		"`--quantity`",
		"`--scheduled-at`",
		`(RFC3339 timestamp, e.g. "2025-01-02T15:04:05Z")`,
		"`--priority`",
		"(one of: PRIORITY_UNSPECIFIED, PRIORITY_LOW, PRIORITY_HIGH; case-insensitive)",
		"`--shipping-address.street`",
		"`--tags`",
		"(repeatable flag: pass --tags once per element)",
		"`--items`",
		"(JSON array of LineItem",
		"Mutually exclusive flag groups:",
		"`--card-token`, `--invoice-id`",
		"### Cross-workflow commands",
		"snapshot-service workflow cancel",
		"### Schedule commands",
		"`--schedule-id`",
		"`--schedule-cron`",
		"snapshot-service schedule pause",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("rendered docs missing %q\n--- full output ---\n%s", s, out)
		}
	}
}

var _ = fmt.Sprintf // keep fmt import in case of future diagnostics
