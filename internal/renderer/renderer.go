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

// RenderConstants: Renders the constants for a service
func (r *Renderer) RenderConstants(service *model.Service) (string, error) {
	return r.render("constants.tmpl", service)
}

// RenderInterface: Renders the service interface
func (r *Renderer) RenderInterface(service *model.Service) (string, error) {
	return r.render("interface.tmpl", service)
}

// RenderWorker: Renders the worker code
func (r *Renderer) RenderWorker(service *model.Service) (string, error) {
	return r.render("worker.tmpl", service)
}

// RenderClient: Renders the client code
func (r *Renderer) RenderClient(service *model.Service) (string, error) {
	return r.render("client.tmpl", service)
}

// RenderWorkflowObjects: Renders workflow object wrappers
func (r *Renderer) RenderWorkflowObjects(service *model.Service) (string, error) {
	return r.render("workflow_objects.tmpl", service)
}

// RenderSignals: Renders signal helper functions
func (r *Renderer) RenderSignals(service *model.Service) (string, error) {
	return r.render("signals.tmpl", service)
}

// RenderQueries: Renders query helper functions
func (r *Renderer) RenderQueries(service *model.Service) (string, error) {
	return r.render("queries.tmpl", service)
}

// render: Common rendering logic
func (r *Renderer) render(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := r.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// RenderAll: Renders all components for a service
func (r *Renderer) RenderAll(service *model.Service) (string, error) {
	var result bytes.Buffer

	components := []struct {
		name string
		fn   func(*model.Service) (string, error)
	}{
		{"constants", r.RenderConstants},
		{"interface", r.RenderInterface},
		{"worker", r.RenderWorker},
		{"client", r.RenderClient},
		{"workflow_objects", r.RenderWorkflowObjects},
		{"signals", r.RenderSignals},
		{"queries", r.RenderQueries},
	}

	for _, component := range components {
		output, err := component.fn(service)
		if err != nil {
			return "", fmt.Errorf("failed to render %s: %w", component.name, err)
		}
		result.WriteString(output)
		result.WriteString("\n\n")
	}

	return result.String(), nil
}
