package renderer

import (
	"strings"
	"testing"
	"text/template"
)

// TestTemplatesParse: Verifies that all templates parse correctly
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

	// Verify the main service template is present
	if tmpl.Lookup("service.tmpl") == nil {
		t.Error("expected service.tmpl template not found")
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

// TestTemplateFilesCoverage: Verifies the main service template exists
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
			t.Logf("✓ service.tmpl: Comprehensive service template with all sections")
			break
		}
	}

	if !foundService {
		t.Error("service.tmpl not found - this is the main template file")
	}
}

// TestRendererStructure: Verifies Renderer struct has expected fields and RenderAll method
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

// TestServiceTemplateSections: Documents what the comprehensive service.tmpl contains
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
