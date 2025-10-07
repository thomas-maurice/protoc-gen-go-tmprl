package renderer

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/model"
	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/tmpl"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// Renderer: Handles template rendering for code generation
type Renderer struct {
	templates *template.Template
	gf        *protogen.GeneratedFile
}

// NewRenderer: Creates a new renderer instance
func NewRenderer(gf *protogen.GeneratedFile) (*Renderer, error) {
	funcMap := tmpl.FuncMap(gf)

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Renderer{
		templates: tmpl,
		gf:        gf,
	}, nil
}

// RenderAll: Renders all components for a service using the comprehensive service template
//
// The service.tmpl template contains all sections:
// - Constants (task queue, timeouts, workflow/activity/signal/query names)
// - Service Interface (methods that users must implement)
// - Client Implementation (workflow and activity execution methods)
// - Worker Implementation (registration and lifecycle management)
// - Workflow Wrapper Objects (type-safe workflow management)
// - Signal Helper Functions (send and receive signals)
// - Query Helper Functions (query workflows)
func (r *Renderer) RenderAll(service *model.Service) (string, error) {
	var buf bytes.Buffer
	err := r.templates.ExecuteTemplate(&buf, "service.tmpl", service)
	if err != nil {
		return "", fmt.Errorf("failed to execute service template: %w", err)
	}
	return buf.String(), nil
}
