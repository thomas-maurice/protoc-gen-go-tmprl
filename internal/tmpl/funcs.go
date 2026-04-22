package tmpl

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
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
		"Quote":                   func(s string) string { return fmt.Sprintf("%q", s) },
		"WorkflowObjectName":      workflowObjectName,
		"ChildWorkflowObjectName": childWorkflowObjectName,
		"commentOneLine":          commentOneLine,
		"commentBlock":            commentBlock,
		"docComment":              docComment,
		"MakeAnchor":              makeAnchor,
		"TrimComment":             trimComment,
		"FormatDuration":          formatDuration,
		"FormatFloat":             formatFloat,
		"FormatStringSlice":       formatStringSlice,
		"Cardinality":             cardinalityToString,
		"DeprecatedIcon":          deprecatedIcon,
		"FullName":                fullNameToString,
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

// workflowObjectName: Generates the workflow object name
func workflowObjectName(serviceName, methodName string) string {
	return fmt.Sprintf("%s%s", serviceName, methodName)
}

// childWorkflowObjectName: Generates the child workflow object name
func childWorkflowObjectName(serviceName, methodName string) string {
	return fmt.Sprintf("Child%s%sExecution", serviceName, methodName)
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

// commentBlock: Converts a proto leading-comment string into a Go doc comment
// block. Each input line becomes a `// `-prefixed output line; empty proto
// lines become bare `//`, which keeps paragraph breaks and code fences intact.
// `indent` is prepended to every emitted line so the caller can place the
// block inside an interface body, struct decl, etc. Returns an empty string
// when the comment is empty, so templates can elide the block cleanly.
//
// Example:
//
//	commentBlock("\t", " Does a thing\n\n ```go\n foo()\n ```")
//	// =>
//	// \t// Does a thing
//	// \t//
//	// \t// ```go
//	// \t// foo()
//	// \t// ```
func commentBlock(indent, s string) string {
	lines := normaliseCommentLines(s)
	if len(lines) == 0 {
		return ""
	}
	return joinCommentLines(indent, "", lines)
}

// docComment: Renders a Go-style doc comment for a named declaration, keeping
// the convention that the first line starts with the identifier. If `body` is
// empty, returns `<indent>// <name>`. Otherwise emits
// `<indent>// <name> <firstLine>` followed by the remaining proto lines each
// prefixed with `// `, with blank lines preserved as bare `//`. Used where we
// want idiomatic Go godoc on interface methods and generated funcs.
func docComment(indent, name, body string) string {
	lines := normaliseCommentLines(body)
	if len(lines) == 0 {
		if name == "" {
			return ""
		}
		return indent + "// " + name
	}
	if name == "" {
		return joinCommentLines(indent, "", lines)
	}
	return joinCommentLines(indent, name+" ", lines)
}

// normaliseCommentLines: Splits a raw proto leading-comment string into
// trimmed lines, normalising CRLF, stripping a single leading space (the one
// left over from "// foo" → " foo"), and trimming blank lines from both ends
// while preserving interior blank lines.
func normaliseCommentLines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.Trim(s, "\n")
	raw := strings.Split(s, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimPrefix(line, " ")
		line = strings.TrimRight(line, " \t")
		lines = append(lines, line)
	}
	// Drop leading/trailing blanks so the block is compact without losing
	// interior paragraph breaks.
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// joinCommentLines: Emits normalised lines as `// `-prefixed output. The first
// line receives `firstPrefix` (e.g. the identifier) between `// ` and the
// line body, so callers can produce `// Name <body>` while still preserving
// a multi-line tail.
func joinCommentLines(indent, firstPrefix string, lines []string) string {
	out := make([]string, 0, len(lines))
	for i, line := range lines {
		prefix := ""
		if i == 0 {
			prefix = firstPrefix
		}
		switch {
		case line == "" && prefix == "":
			out = append(out, indent+"//")
		case line == "":
			// A blank first line with a prefix collapses to `// <prefix>`
			// without a trailing space.
			out = append(out, indent+"// "+strings.TrimRight(prefix, " "))
		default:
			out = append(out, indent+"// "+prefix+line)
		}
	}
	return strings.Join(out, "\n")
}

// makeAnchor: Creates a markdown-safe anchor identifier
func makeAnchor(parts ...string) string {
	in := strings.Join(parts, ":")
	replaced := ":.-"
	for _, r := range replaced {
		in = strings.ReplaceAll(in, string(r), "_")
	}
	return in
}

// trimComment: Trims comment formatting from protobuf comments
func trimComment(comment interface{}) string {
	var s string
	switch v := comment.(type) {
	case string:
		s = v
	case protogen.Comments:
		s = string(v)
	default:
		return ""
	}
	s = strings.TrimPrefix(s, "//")
	return strings.TrimSpace(s)
}

// formatDuration: Formats an int64 seconds value as a duration string
func formatDuration(seconds interface{}) string {
	s := toSeconds(seconds)
	d := time.Duration(s) * time.Second
	return d.String()
}

// formatFloat: Formats a float64 value
func formatFloat(f float64) string {
	return fmt.Sprintf("%f", f)
}

// formatStringSlice: Formats a string slice as a bracketed list
func formatStringSlice(slice []string) string {
	return fmt.Sprintf("%v", slice)
}

// cardinalityToString: Converts protoreflect.Cardinality to string
func cardinalityToString(c protoreflect.Cardinality) string {
	switch c {
	case protoreflect.Repeated:
		return "Repeated"
	case protoreflect.Optional:
		return "Optional"
	case protoreflect.Required:
		return "Required"
	}
	return "Invalid cardinality"
}

// deprecatedIcon: Returns an icon indicating if a field is deprecated
func deprecatedIcon(opts interface{}) string {
	if fieldOpts, ok := opts.(*descriptorpb.FieldOptions); ok && fieldOpts != nil {
		if fieldOpts.GetDeprecated() {
			return "🗿"
		}
	}
	return "✅"
}

// fullNameToString: Converts protoreflect.FullName to string
func fullNameToString(name protoreflect.FullName) string {
	return string(name)
}
