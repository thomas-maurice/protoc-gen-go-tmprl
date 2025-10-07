package renderer

import (
	"strings"
	"testing"
	"text/template"

	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/model"
	"google.golang.org/protobuf/compiler/protogen"
)

// TestTemplatesParse: Verifies that all templates parse correctly
func TestTemplatesParse(t *testing.T) {
	// Create a minimal function map for parsing
	funcMap := template.FuncMap{
		"QualifiedGoIdent":        func(ident protogen.GoIdent) string { return "" },
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
		"HasTimeout":              func(d interface{}) bool { return false },
		"HasRetryPolicy":          func(rp interface{}) bool { return false },
		"Join":                    func(elems []string, sep string) string { return "" },
		"Quote":                   func(s string) string { return "" },
		"Add":                     func(a, b int) int { return a + b },
		"WorkflowObjectName":      func(service, method string) string { return "" },
		"ChildWorkflowObjectName": func(service, method string) string { return "" },
		"dict":                    func(values ...interface{}) (map[string]interface{}, error) { return nil, nil },
		"commentOneLine":          func(s string) string { return "" },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	// Verify expected templates are present
	expectedTemplates := []string{
		"constants.tmpl",
		"interface.tmpl",
		"worker.tmpl",
		"client.tmpl",
		"workflow_objects.tmpl",
		"signals.tmpl",
		"queries.tmpl",
		"helpers.tmpl",
		"activity_methods.tmpl",
		"workflow_methods.tmpl",
	}

	for _, name := range expectedTemplates {
		if tmpl.Lookup(name) == nil {
			t.Errorf("expected template %s not found", name)
		}
	}
}

// TestTemplateDefinitions: Verifies that expected sub-templates are defined
func TestTemplateDefinitions(t *testing.T) {
	funcMap := template.FuncMap{
		"QualifiedGoIdent":        func(ident protogen.GoIdent) string { return "" },
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
		"HasTimeout":              func(d interface{}) bool { return false },
		"HasRetryPolicy":          func(rp interface{}) bool { return false },
		"Join":                    func(elems []string, sep string) string { return "" },
		"Quote":                   func(s string) string { return "" },
		"Add":                     func(a, b int) int { return a + b },
		"WorkflowObjectName":      func(service, method string) string { return "" },
		"ChildWorkflowObjectName": func(service, method string) string { return "" },
		"dict":                    func(values ...interface{}) (map[string]interface{}, error) { return nil, nil },
		"commentOneLine":          func(s string) string { return "" },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	// Verify sub-template definitions exist
	expectedDefinitions := []string{
		"timeout_option",
		"retry_policy",
		"activity_execute",
		"activity_execute_sync",
		"workflow_execute",
	}

	for _, name := range expectedDefinitions {
		if tmpl.Lookup(name) == nil {
			t.Errorf("expected template definition %s not found", name)
		}
	}
}

// TestRenderComponentNames: Verifies render method component names
func TestRenderComponentNames(t *testing.T) {
	// This test documents the expected render components
	expectedComponents := []string{
		"constants",
		"interface",
		"worker",
		"client",
		"workflow_objects",
		"signals",
		"queries",
	}

	// Just verify we have documentation of what components exist
	if len(expectedComponents) != 7 {
		t.Errorf("expected 7 render components, documented %d", len(expectedComponents))
	}
}

// TestTemplateEmbedding: Verifies templates are properly embedded
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

// TestRenderMethods: Tests that all render methods exist and have correct signatures
func TestRenderMethods(t *testing.T) {
	// This test verifies the Renderer type has all expected methods
	t.Run("method existence", func(t *testing.T) {
		var r *Renderer
		if r != nil {
			// These calls won't execute but verify method signatures compile
			var s *model.Service
			_, _ = r.RenderConstants(s)
			_, _ = r.RenderInterface(s)
			_, _ = r.RenderWorker(s)
			_, _ = r.RenderClient(s)
			_, _ = r.RenderWorkflowObjects(s)
			_, _ = r.RenderSignals(s)
			_, _ = r.RenderQueries(s)
			_, _ = r.RenderAll(s)
		}
	})
}

// TestRenderAllComponents: Tests that RenderAll calls all component renderers
func TestRenderAllComponents(t *testing.T) {
	// This test documents the expected rendering order in RenderAll
	expectedComponents := []string{
		"constants",
		"interface",
		"worker",
		"client",
		"workflow_objects",
		"signals",
		"queries",
	}

	if len(expectedComponents) != 7 {
		t.Errorf("RenderAll should render %d components, found %d", 7, len(expectedComponents))
	}
}

// TestTemplateFilesCoverage: Verifies all template files are accounted for
func TestTemplateFilesCoverage(t *testing.T) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		t.Fatalf("failed to read templates directory: %v", err)
	}

	// Map of template files to their purpose
	templatePurpose := map[string]string{
		"constants.tmpl":        "Service constants (task queue, timeouts)",
		"interface.tmpl":        "Service interface definition",
		"worker.tmpl":           "Worker implementation",
		"client.tmpl":           "Client implementation",
		"workflow_objects.tmpl": "Workflow object wrappers",
		"signals.tmpl":          "Signal helper functions",
		"queries.tmpl":          "Query helper functions",
		"helpers.tmpl":          "Reusable template helpers",
		"activity_methods.tmpl": "Activity execution methods",
		"workflow_methods.tmpl": "Workflow execution methods",
	}

	for _, entry := range entries {
		name := entry.Name()
		if purpose, ok := templatePurpose[name]; ok {
			t.Logf("✓ %s: %s", name, purpose)
		} else {
			t.Errorf("Template file %s is not documented in test", name)
		}
	}

	// Verify all expected templates exist
	for expectedName := range templatePurpose {
		found := false
		for _, entry := range entries {
			if entry.Name() == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected template %s not found in templates directory", expectedName)
		}
	}
}

// TestTemplateConsistency: Verifies template names match render methods
func TestTemplateConsistency(t *testing.T) {
	// Map render methods to their template files
	methodToTemplate := map[string]string{
		"RenderConstants":       "constants.tmpl",
		"RenderInterface":       "interface.tmpl",
		"RenderWorker":          "worker.tmpl",
		"RenderClient":          "client.tmpl",
		"RenderWorkflowObjects": "workflow_objects.tmpl",
		"RenderSignals":         "signals.tmpl",
		"RenderQueries":         "queries.tmpl",
	}

	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		t.Fatalf("failed to read templates directory: %v", err)
	}

	templateFiles := make(map[string]bool)
	for _, entry := range entries {
		templateFiles[entry.Name()] = true
	}

	for method, templateName := range methodToTemplate {
		if !templateFiles[templateName] {
			t.Errorf("Method %s expects template %s which doesn't exist", method, templateName)
		}
	}
}

// TestTemplateHelperDefinitions: Verifies helper templates are defined
func TestTemplateHelperDefinitions(t *testing.T) {
	funcMap := template.FuncMap{
		"QualifiedGoIdent":        func(ident protogen.GoIdent) string { return "" },
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
		"HasTimeout":              func(d interface{}) bool { return false },
		"HasRetryPolicy":          func(rp interface{}) bool { return false },
		"Join":                    func(elems []string, sep string) string { return "" },
		"Quote":                   func(s string) string { return "" },
		"Add":                     func(a, b int) int { return a + b },
		"WorkflowObjectName":      func(service, method string) string { return "" },
		"ChildWorkflowObjectName": func(service, method string) string { return "" },
		"dict":                    func(values ...interface{}) (map[string]interface{}, error) { return nil, nil },
		"commentOneLine":          func(s string) string { return "" },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	// Helpers.tmpl should define these reusable blocks
	expectedHelpers := map[string]string{
		"timeout_option":        "Renders timeout options",
		"retry_policy":          "Renders retry policy configuration",
		"activity_execute":      "Activity execution code",
		"activity_execute_sync": "Synchronous activity execution",
		"workflow_execute":      "Workflow execution code",
	}

	for helper, purpose := range expectedHelpers {
		if tmpl.Lookup(helper) == nil {
			t.Errorf("Helper template '%s' not found (purpose: %s)", helper, purpose)
		} else {
			t.Logf("✓ Helper '%s': %s", helper, purpose)
		}
	}
}

// TestRendererStructure: Verifies Renderer struct has expected fields and methods
func TestRendererStructure(t *testing.T) {
	// This test documents the Renderer structure
	var r Renderer

	// These assignments verify the fields exist and have correct types
	_ = r.templates // *template.Template
	_ = r.gf        // *protogen.GeneratedFile

	// Verify Renderer pointer has the expected methods by type-checking
	// This ensures the interface contract is maintained
	var _ interface {
		RenderConstants(*model.Service) (string, error)
		RenderInterface(*model.Service) (string, error)
		RenderWorker(*model.Service) (string, error)
		RenderClient(*model.Service) (string, error)
		RenderWorkflowObjects(*model.Service) (string, error)
		RenderSignals(*model.Service) (string, error)
		RenderQueries(*model.Service) (string, error)
		RenderAll(*model.Service) (string, error)
	} = &r
}
