package tmpl

import (
	"fmt"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/model"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// FuncMap Returns the template function map
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
		"kebab":                   kebab,
		"pflagFuncName":           pflagFuncName,
		"goFieldAssignPath":       goFieldAssignPath,
		"goFieldAssignSteps":      goFieldAssignSteps,
		"enumValueTypeName":       enumValueTypeName,
		"jsonExampleFor":          jsonExampleFor,
		"Cobra":                   func(name string) string { return qualifiedIdent(gf, "github.com/spf13/cobra", name) },
		"Pflag":                   func(name string) string { return qualifiedIdent(gf, "github.com/spf13/pflag", name) },
		"Protojson":               func(name string) string { return qualifiedIdent(gf, "google.golang.org/protobuf/encoding/protojson", name) },
		"Timestamppb":             func(name string) string { return qualifiedIdent(gf, "google.golang.org/protobuf/types/known/timestamppb", name) },
		"Durationpb":              func(name string) string { return qualifiedIdent(gf, "google.golang.org/protobuf/types/known/durationpb", name) },
		"Fieldmaskpb":             func(name string) string { return qualifiedIdent(gf, "google.golang.org/protobuf/types/known/fieldmaskpb", name) },
		"Reflect":                 func(name string) string { return qualifiedIdent(gf, "reflect", name) },
		"Strings":                 func(name string) string { return qualifiedIdent(gf, "strings", name) },
		"Strconv":                 func(name string) string { return qualifiedIdent(gf, "strconv", name) },
		"Os":                      func(name string) string { return qualifiedIdent(gf, "os", name) },
		"EncodingJSON":            func(name string) string { return qualifiedIdent(gf, "encoding/json", name) },
		"Proto":                   func(name string) string { return qualifiedIdent(gf, "google.golang.org/protobuf/proto", name) },
		"Errors":                  func(name string) string { return qualifiedIdent(gf, "errors", name) },
		"IO":                      func(name string) string { return qualifiedIdent(gf, "io", name) },
		"Enumspb":                 func(name string) string { return qualifiedIdent(gf, "go.temporal.io/api/enums/v1", name) },
		"dict":                    dict,
		"enumShortName":           enumShortName,
		"cliHelpString":           cliHelpString,
		"cliVarDecl":              cliVarDecl,
		"cliRegisterLine":         cliRegisterLine,
		"cliApplyBlock":           cliApplyBlock,
		"cliVarName":              cliVarName,
		"oneofGroupsGoLiteral":    oneofGroupsGoLiteral,
		"oneofExclusiveCalls":     oneofExclusiveCalls,
	}
}

// qualifiedTemporalIdent Returns a qualified identifier for temporal imports
func qualifiedTemporalIdent(gf *protogen.GeneratedFile, importPath, name string) string {
	return gf.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: protogen.GoImportPath(importPath),
		GoName:       name,
	})
}

// qualifiedIdent Returns a qualified identifier for standard imports
func qualifiedIdent(gf *protogen.GeneratedFile, importPath, name string) string {
	return gf.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: protogen.GoImportPath(importPath),
		GoName:       name,
	})
}

// toSeconds Converts a duration to seconds for template usage
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

// workflowObjectName Generates the workflow object name
func workflowObjectName(serviceName, methodName string) string {
	return fmt.Sprintf("%s%s", serviceName, methodName)
}

// childWorkflowObjectName Generates the child workflow object name
func childWorkflowObjectName(serviceName, methodName string) string {
	return fmt.Sprintf("Child%s%sExecution", serviceName, methodName)
}

// commentOneLine Converts a multiline comment to a single line
func commentOneLine(s string) string {
	// Replace newlines with spaces and collapse multiple spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	// Collapse multiple spaces
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

// commentBlock Converts a proto leading-comment string into a Go doc comment
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

// docComment Renders a Go-style doc comment for a named declaration, keeping
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

// normaliseCommentLines Splits a raw proto leading-comment string into
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

// joinCommentLines Emits normalised lines as `// `-prefixed output. The first
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

// makeAnchor Creates a markdown-safe anchor identifier
func makeAnchor(parts ...string) string {
	in := strings.Join(parts, ":")
	replaced := ":.-"
	for _, r := range replaced {
		in = strings.ReplaceAll(in, string(r), "_")
	}
	return in
}

// trimComment Trims comment formatting from protobuf comments
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

// formatDuration Formats an int64 seconds value as a duration string
func formatDuration(seconds interface{}) string {
	s := toSeconds(seconds)
	d := time.Duration(s) * time.Second
	return d.String()
}

// formatFloat Formats a float64 value
func formatFloat(f float64) string {
	return fmt.Sprintf("%f", f)
}

// formatStringSlice Formats a string slice for markdown table cells, wrapping
// each entry in backticks so identifier-like values (e.g. error type names)
// render as inline code. Returns an empty string for an empty/nil slice so
// callers can rely on the enclosing template's {{if}} guard.
func formatStringSlice(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	parts := make([]string, len(slice))
	for i, s := range slice {
		parts[i] = "`" + s + "`"
	}
	return strings.Join(parts, ", ")
}

// cardinalityToString Converts protoreflect.Cardinality to string
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

// deprecatedIcon Returns an icon indicating if a field is deprecated
func deprecatedIcon(opts interface{}) string {
	if fieldOpts, ok := opts.(*descriptorpb.FieldOptions); ok && fieldOpts != nil {
		if fieldOpts.GetDeprecated() {
			return "🗿"
		}
	}
	return "✅"
}

// fullNameToString Converts protoreflect.FullName to string
func fullNameToString(name protoreflect.FullName) string {
	return string(name)
}

// kebab converts a snake_case, camelCase, or PascalCase identifier to
// lower-kebab-case. Runs of consecutive capitals are treated as a single
// acronym token, with a split inserted before the last capital when it is
// followed by a lowercase letter (so `HTTPSServer` becomes `https-server`
// rather than `h-t-t-p-s-server`). Digits cluster with the preceding run:
// `order123Id` yields `order123-id`. An empty input returns an empty string.
func kebab(s string) string {
	if s == "" {
		return ""
	}
	// First normalise underscores: treat each underscore as a hard split.
	parts := strings.Split(s, "_")
	var tokens []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		tokens = append(tokens, splitCamel(p)...)
	}
	for i, t := range tokens {
		tokens[i] = strings.ToLower(t)
	}
	return strings.Join(tokens, "-")
}

// splitCamel splits a camelCase/PascalCase token into its component words,
// preserving acronym runs as single tokens and attaching trailing digits to
// the preceding alphabetic run.
func splitCamel(s string) []string {
	if s == "" {
		return nil
	}
	runes := []rune(s)
	var out []string
	start := 0
	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		cur := runes[i]
		// Lower-to-upper boundary: `orderId` → split before `I`.
		if !unicode.IsUpper(prev) && unicode.IsUpper(cur) {
			out = append(out, string(runes[start:i]))
			start = i
			continue
		}
		// Digit-to-letter boundary: `123Id` → split before `I` and before
		// `1` when coming from a letter.
		if unicode.IsDigit(prev) && unicode.IsLetter(cur) {
			out = append(out, string(runes[start:i]))
			start = i
			continue
		}
		// Acronym boundary: a run of uppercase letters followed by a
		// lowercase letter splits just before the final uppercase. `HTTPSServer`
		// → `HTTPS` + `Server`.
		if i+1 < len(runes) && unicode.IsUpper(prev) && unicode.IsUpper(cur) && unicode.IsLower(runes[i+1]) {
			out = append(out, string(runes[start:i]))
			start = i
			continue
		}
	}
	out = append(out, string(runes[start:]))
	return out
}

// pflagFuncName returns the name of the pflag binder function for a given
// FlagKind — the identifier used after `cmd.Flags().` at code-generation
// time. Slice kinds map to `ArrayVar` variants where pflag offers them so
// that values with commas survive unescaped. Kinds that need a custom
// pflag.Value (timestamp, enum, JSON-escape) map to `Var`, and the template
// is expected to wrap the backing value with the generated helper type.
// Panics on an unrecognised kind since the set is closed at build time.
func pflagFuncName(kind model.FlagKind) string {
	switch kind {
	case model.FlagKindString:
		return "StringVar"
	case model.FlagKindStringSlice:
		return "StringArrayVar"
	case model.FlagKindInt32:
		return "Int32Var"
	case model.FlagKindInt32Slice:
		return "Int32SliceVar"
	case model.FlagKindInt64:
		return "Int64Var"
	case model.FlagKindInt64Slice:
		return "Int64SliceVar"
	case model.FlagKindUint32:
		return "Uint32Var"
	case model.FlagKindUint64:
		return "Uint64Var"
	case model.FlagKindFloat32:
		return "Float32Var"
	case model.FlagKindFloat64:
		return "Float64Var"
	case model.FlagKindBool:
		return "BoolVar"
	case model.FlagKindBytes:
		return "BytesBase64Var"
	case model.FlagKindDuration:
		return "DurationVar"
	case model.FlagKindTimestamp:
		return "Var"
	case model.FlagKindEnum:
		return "Var"
	case model.FlagKindFieldMask:
		return "StringSliceVar"
	case model.FlagKindStringToString:
		return "StringToStringVar"
	case model.FlagKindStringToInt:
		return "StringToIntVar"
	case model.FlagKindJSONEscape:
		return "Var"
	}
	panic(fmt.Sprintf("pflagFuncName: unknown FlagKind %v", kind))
}

// AssignStep describes one level of a nested proto-field assignment. A
// slice of AssignSteps walks from the root receiver down to the leaf
// field, and the template emits intermediate-pointer allocations on each
// non-leaf step and the final field set on the leaf. Expr is the Go
// expression referring to the struct at this level (e.g. `req`, then
// `req.Shipping`, then `req.Shipping.Address`). FieldName is the Go field
// name on that struct (e.g. `Shipping`, `Address`, `Street`). GoTypeRef is
// a hint — empty for the root step and intentionally left empty for
// intermediates so the template computes the concrete type name from
// context (the plan layer does not carry nested Go type names today).
type AssignStep struct {
	// Expr is the Go expression that addresses the struct at this level,
	// rooted at the caller-provided rootVar. The root step has Expr equal
	// to rootVar itself.
	Expr string

	// FieldName is the Go field name to access on `Expr`. On the leaf
	// step this is the scalar field to assign; on intermediate steps it
	// is the nested message field whose pointer may need allocating.
	FieldName string

	// IsLeaf is true for the final step. Templates set the scalar value
	// on a leaf and allocate a nested message on non-leaves.
	IsLeaf bool
}

// goFieldAssignPath returns the Go expression that reaches the leaf field
// of a proto-field path using `Get*()` accessors for every intermediate
// message. rootVar is the variable name holding the root request proto.
// The returned expression is suitable for reading; for writing, the
// template should combine the step list from goFieldAssignSteps with a
// per-level nil check so intermediate messages get allocated before the
// final assignment. A nil/empty path returns rootVar unchanged so callers
// can feed the helper without guarding for edge cases.
func goFieldAssignPath(path []string, rootVar string) string {
	if len(path) == 0 {
		return rootVar
	}
	var b strings.Builder
	b.WriteString(rootVar)
	for i, seg := range path {
		field := goFieldName(seg)
		if i == len(path)-1 {
			b.WriteByte('.')
			b.WriteString(field)
		} else {
			b.WriteString(".Get")
			b.WriteString(field)
			b.WriteString("()")
		}
	}
	return b.String()
}

// goFieldAssignSteps returns an ordered slice of AssignSteps for a proto
// field path. The template ranges over the slice: intermediate steps
// guard a nil pointer with `if expr.Field == nil { expr.Field = &T{} }`,
// and the final step performs the scalar assignment. Expr on the Nth step
// is the dotted Go expression up to (not including) that step's
// FieldName. See the type doc for shape details.
func goFieldAssignSteps(path []string, rootVar string) []AssignStep {
	if len(path) == 0 {
		return nil
	}
	steps := make([]AssignStep, 0, len(path))
	expr := rootVar
	for i, seg := range path {
		field := goFieldName(seg)
		steps = append(steps, AssignStep{
			Expr:      expr,
			FieldName: field,
			IsLeaf:    i == len(path)-1,
		})
		expr = expr + "." + field
	}
	return steps
}

// goFieldName converts a proto snake_case field name to the Go field name
// protoc-gen-go produces — leading letter uppercased, each underscore
// consumes the next letter and uppercases it. `customer_id` → `CustomerId`,
// `shipping` → `Shipping`, `a_b_c` → `ABC` only if each segment is a
// single letter, otherwise `ABC_` rules do not apply (protoc-gen-go joins
// without underscores and uppercases each segment's first letter).
func goFieldName(snake string) string {
	if snake == "" {
		return ""
	}
	var b strings.Builder
	upperNext := true
	for _, r := range snake {
		if r == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// enumValueTypeName returns the Go identifier of the generated custom
// pflag.Value wrapper for a given enum type. `OrderStatus` becomes
// `enumOrderStatusValue`. The name is intentionally unexported so it
// lives alongside the consuming CLI constructors without polluting the
// generated package's public surface.
func enumValueTypeName(enumGoName string) string {
	return "enum" + enumGoName + "Value"
}

// wktJSONValueNames is the set of Well-Known Type full names whose CLI
// representation is a free-form JSON value rather than a message with a
// known shape. Used by jsonExampleFor to pick the right example snippet.
var wktJSONValueNames = map[string]bool{
	"google.protobuf.Any":       true,
	"google.protobuf.Struct":    true,
	"google.protobuf.Value":     true,
	"google.protobuf.ListValue": true,
}

// jsonExampleFor returns the example JSON snippet appended to the help
// text of a FlagKindJSONEscape flag, per PLAN.md §6.1. Non-JSONEscape
// flags return an empty string so the template can call this helper
// unconditionally. Sub-case selection is driven by FlagSpec.Repeated and
// FlagSpec.JSONGoType:
//
//   - Repeated message: `'[{"field":"value"}]'`
//   - Any / Struct / Value / ListValue: `'<JSON value>'`
//   - Otherwise (cyclic subtree or map-of-message seam):
//     `'<JSON value of type Foo>'` where Foo is the last path segment of
//     JSONGoType.
//
// The walker does not currently distinguish map-of-message from a cyclic
// seam; both surface the same "JSON value of type Foo" phrasing, which is
// still unambiguous to a human reader.
func jsonExampleFor(spec *model.FlagSpec) string {
	if spec == nil || spec.Kind != model.FlagKindJSONEscape {
		return ""
	}
	if spec.Repeated {
		return `'[{"field":"value"}]'`
	}
	if wktJSONValueNames[spec.JSONGoType] {
		return `'<JSON value>'`
	}
	return "'<JSON value of type " + lastDotSegment(spec.JSONGoType) + ">'"
}

// lastDotSegment returns the substring after the final '.' in s, or s
// unchanged when no dot is present. Used to trim a proto full name down
// to its terminal message name for help text.
func lastDotSegment(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		return s[i+1:]
	}
	return s
}

// enumShortName returns the short Go-level name of an enum given its
// proto full name. `example.v1.Color` yields `Color`. Used by the CLI
// template to construct the generated pflag.Value wrapper type name.
func enumShortName(fullName string) string {
	return lastDotSegment(fullName)
}

// dict builds a map[string]interface{} from alternating key/value pairs.
// Emulates Helm's well-known `dict` helper so CLI template subtemplates
// can be invoked with multiple named arguments without defining a custom
// Go struct per call site. Odd-length arg lists are truncated.
func dict(args ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			continue
		}
		m[k] = args[i+1]
	}
	return m
}

// cliHelpString returns the full help string for a FlagSpec, combining
// the proto-derived comment with the format hint from PLAN.md §6.1. The
// hint explicitly names the expected shape (RFC3339 timestamp, Go
// duration, repeatable flag, JSON array of <type>, etc.) so the user
// never has to guess. Called by the CLI template in flag-registration.
func cliHelpString(spec *model.FlagSpec) string {
	if spec == nil {
		return ""
	}
	help := strings.TrimSpace(spec.Help)
	if help == "" {
		help = spec.CLIName
	}
	hint := cliHelpHint(spec)
	if hint == "" {
		return help
	}
	return help + " " + hint
}

// cliHelpHint returns the parenthetical format hint appended to a
// FlagSpec's help string, matching the PLAN.md §6.1 kind table.
func cliHelpHint(spec *model.FlagSpec) string {
	switch spec.Kind {
	case model.FlagKindBytes:
		return "(base64-encoded)"
	case model.FlagKindDuration:
		return `(Go duration, e.g. "5m", "1h30m")`
	case model.FlagKindTimestamp:
		return `(RFC3339 timestamp, e.g. "2025-01-02T15:04:05Z")`
	case model.FlagKindEnum:
		return cliEnumHint(spec)
	case model.FlagKindFieldMask:
		return `(comma-separated field paths, e.g. "name,address.city")`
	case model.FlagKindStringSlice, model.FlagKindInt32Slice, model.FlagKindInt64Slice:
		return "(repeatable flag: pass --" + spec.CLIName + " once per element)"
	case model.FlagKindStringToString:
		return "(repeatable flag in key=value form, e.g. --" + spec.CLIName + " foo=bar --" + spec.CLIName + " baz=qux)"
	case model.FlagKindStringToInt:
		return "(repeatable flag in key=value form with integer values)"
	case model.FlagKindJSONEscape:
		return cliJSONHint(spec)
	}
	return ""
}

// cliEnumHint renders the hint for an enum flag, listing the full set of
// valid names so the user sees every option up-front.
func cliEnumHint(spec *model.FlagSpec) string {
	return "(one of: " + strings.Join(spec.EnumValues, ", ") + "; case-insensitive)"
}

// cliJSONHint renders the hint for a JSON-escape flag, calling out the
// target message or WKT shape so the user knows what JSON to provide.
func cliJSONHint(spec *model.FlagSpec) string {
	if spec.Repeated {
		return "(JSON array of " + enumShortName(spec.JSONGoType) + `, e.g. '[{"field":"value"}]')`
	}
	if wktJSONValueNames[spec.JSONGoType] {
		return "(JSON value)"
	}
	return "(JSON value of type " + enumShortName(spec.JSONGoType) + ")"
}

// cliVarDecl emits the local-variable declaration for a FlagSpec at the
// top of a RunE constructor. Scalars and maps use typed zero values;
// slice kinds use `nil`; custom-Value kinds (timestamp, enum, JSON)
// allocate pointer values upfront.
func cliVarDecl(spec *model.FlagSpec, prefix string) string {
	name := cliVarName(spec)
	switch spec.Kind {
	case model.FlagKindString:
		return "var " + name + " string"
	case model.FlagKindStringSlice:
		return "var " + name + " []string"
	case model.FlagKindInt32:
		return "var " + name + " int32"
	case model.FlagKindInt32Slice:
		return "var " + name + " []int32"
	case model.FlagKindInt64:
		return "var " + name + " int64"
	case model.FlagKindInt64Slice:
		return "var " + name + " []int64"
	case model.FlagKindUint32:
		return "var " + name + " uint32"
	case model.FlagKindUint64:
		return "var " + name + " uint64"
	case model.FlagKindFloat32:
		return "var " + name + " float32"
	case model.FlagKindFloat64:
		return "var " + name + " float64"
	case model.FlagKindBool:
		return "var " + name + " bool"
	case model.FlagKindBytes:
		return "var " + name + " []byte"
	case model.FlagKindDuration:
		return "var " + name + " time.Duration"
	case model.FlagKindTimestamp:
		return name + " := &" + prefix + "TimestampValue{}"
	case model.FlagKindEnum:
		return name + " := &" + prefix + "Enum" + enumShortName(spec.EnumGoType) + "Value{}"
	case model.FlagKindFieldMask:
		return "var " + name + " []string"
	case model.FlagKindStringToString:
		return "var " + name + " map[string]string"
	case model.FlagKindStringToInt:
		return "var " + name + " map[string]int"
	case model.FlagKindJSONEscape:
		return name + " := &" + prefix + "JSONValue{}"
	}
	return "// unknown flag kind " + spec.Kind.String()
}

// cliVarName returns the local Go variable name bound to a flag. The
// name is derived from the CLI flag name by camel-casing around hyphens
// and prefixing with `flag` to avoid collisions with Go keywords and
// with the workflowID / runID targeting vars declared above it.
func cliVarName(spec *model.FlagSpec) string {
	return "flag" + camelCaseFromKebab(spec.CLIName)
}

// camelCaseFromKebab upper-camels each delimited segment of a CLI flag
// name. Splits on both '-' (word boundary inside a single proto field) and
// '.' (nesting boundary between proto path segments) so `shipping.address
// .street` and `input-customer-id` both become valid Go identifiers
// (`ShippingAddressStreet` / `InputCustomerId`).
func camelCaseFromKebab(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '-' || r == '.' })
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}

// cliRegisterLine emits the `cmd.Flags().XxxVar(...)` registration for a
// FlagSpec. Custom-Value kinds (timestamp, enum, JSON) use `.Var(...)`
// with the pre-allocated pointer; everything else calls the pflag binder
// returned by pflagFuncName with the appropriate default value.
func cliRegisterLine(spec *model.FlagSpec, prefix string) string {
	name := cliVarName(spec)
	flagName := fmt.Sprintf("%q", spec.CLIName)
	help := fmt.Sprintf("%q", cliHelpString(spec))
	fn := pflagFuncName(spec.Kind)
	switch spec.Kind {
	case model.FlagKindTimestamp, model.FlagKindEnum, model.FlagKindJSONEscape:
		return fmt.Sprintf("cmd.Flags().%s(%s, %s, %s)", fn, name, flagName, help)
	case model.FlagKindStringToString:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, nil, %s)", fn, name, flagName, help)
	case model.FlagKindStringToInt:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, nil, %s)", fn, name, flagName, help)
	case model.FlagKindStringSlice, model.FlagKindInt32Slice, model.FlagKindInt64Slice, model.FlagKindFieldMask:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, nil, %s)", fn, name, flagName, help)
	case model.FlagKindBytes:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, nil, %s)", fn, name, flagName, help)
	case model.FlagKindBool:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, false, %s)", fn, name, flagName, help)
	case model.FlagKindDuration:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, 0, %s)", fn, name, flagName, help)
	case model.FlagKindString:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, \"\", %s)", fn, name, flagName, help)
	case model.FlagKindFloat32, model.FlagKindFloat64:
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, 0, %s)", fn, name, flagName, help)
	default:
		// Numeric scalars.
		return fmt.Sprintf("cmd.Flags().%s(&%s, %s, 0, %s)", fn, name, flagName, help)
	}
}

// cliApplyBlock emits the RunE body that assigns a flag's value onto
// `req` when the flag was provided on the command line. Each block calls
// the generated reflection-based setter with the ProtoPath and the
// kind-appropriate Go value. Presence-gating (`if cmd.Flags().Changed`)
// is done around this block by the caller template.
func cliApplyBlock(spec *model.FlagSpec, prefix string) string {
	name := cliVarName(spec)
	path := protoPathGoLiteral(spec.ProtoPath)
	setCall := fmt.Sprintf("%sSetField(req, %s, ", prefix, path)
	switch spec.Kind {
	case model.FlagKindTimestamp:
		return "if " + name + ".ts != nil { if err := " + setCall + name + ".ts); err != nil { return err } }"
	case model.FlagKindEnum:
		return "if " + name + ".set { if err := " + setCall + name + ".v); err != nil { return err } }"
	case model.FlagKindJSONEscape:
		return "if " + name + ".set { if err := " + prefix + "UnmarshalJSONInto(req, " + path + ", " + name + ".raw); err != nil { return err } }"
	case model.FlagKindDuration:
		return "if err := " + setCall + prefix + "DurationProto(" + name + ")); err != nil { return err }"
	case model.FlagKindFieldMask:
		return "if err := " + setCall + prefix + "FieldMaskProto(" + name + ")); err != nil { return err }"
	default:
		return "if err := " + setCall + name + "); err != nil { return err }"
	}
}

// protoPathGoLiteral renders a ProtoPath as a Go []string literal.
func protoPathGoLiteral(path []string) string {
	if len(path) == 0 {
		return "nil"
	}
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = fmt.Sprintf("%q", p)
	}
	return "[]string{" + strings.Join(parts, ", ") + "}"
}

// oneofGroupsGoLiteral renders a [][]string as a Go literal inline.
// Used by the CLI template when generating the oneof-exclusive call
// list for a command's Long description.
func oneofGroupsGoLiteral(groups [][]string) string {
	if len(groups) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("[][]string{")
	for i, g := range groups {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("{")
		for j, n := range g {
			if j > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%q", n)
		}
		b.WriteString("}")
	}
	b.WriteString("}")
	return b.String()
}

// oneofExclusiveCalls emits one `cmd.MarkFlagsMutuallyExclusive(...)`
// line per oneof group, separated by newlines. Each line lists the
// CLI-level flag names that form the group.
func oneofExclusiveCalls(groups [][]string) string {
	if len(groups) == 0 {
		return ""
	}
	var b strings.Builder
	for i, g := range groups {
		if i > 0 {
			b.WriteString("\n\t")
		}
		b.WriteString("cmd.MarkFlagsMutuallyExclusive(")
		for j, n := range g {
			if j > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%q", n)
		}
		b.WriteString(")")
	}
	return b.String()
}
