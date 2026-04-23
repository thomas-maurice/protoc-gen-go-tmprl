package model

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// FlagKind identifies which pflag binder and assembly path a FlagSpec uses.
type FlagKind int

// Flag kinds. The ordering is stable; callers (templates and code-gen
// helpers) may switch on the specific values.
const (
	FlagKindString FlagKind = iota
	FlagKindStringSlice
	FlagKindInt32
	FlagKindInt32Slice
	FlagKindInt64
	FlagKindInt64Slice
	FlagKindUint32
	FlagKindUint64
	FlagKindFloat32
	FlagKindFloat64
	FlagKindBool
	FlagKindBytes
	FlagKindDuration
	FlagKindTimestamp
	FlagKindEnum
	FlagKindFieldMask
	FlagKindStringToString
	FlagKindStringToInt
	FlagKindJSONEscape
)

// String returns a stable debug name for the FlagKind. Intended for
// test failure messages and log output, not for code generation.
func (k FlagKind) String() string {
	switch k {
	case FlagKindString:
		return "String"
	case FlagKindStringSlice:
		return "StringSlice"
	case FlagKindInt32:
		return "Int32"
	case FlagKindInt32Slice:
		return "Int32Slice"
	case FlagKindInt64:
		return "Int64"
	case FlagKindInt64Slice:
		return "Int64Slice"
	case FlagKindUint32:
		return "Uint32"
	case FlagKindUint64:
		return "Uint64"
	case FlagKindFloat32:
		return "Float32"
	case FlagKindFloat64:
		return "Float64"
	case FlagKindBool:
		return "Bool"
	case FlagKindBytes:
		return "Bytes"
	case FlagKindDuration:
		return "Duration"
	case FlagKindTimestamp:
		return "Timestamp"
	case FlagKindEnum:
		return "Enum"
	case FlagKindFieldMask:
		return "FieldMask"
	case FlagKindStringToString:
		return "StringToString"
	case FlagKindStringToInt:
		return "StringToInt"
	case FlagKindJSONEscape:
		return "JSONEscape"
	default:
		return fmt.Sprintf("FlagKind(%d)", int(k))
	}
}

// FlagSpec describes a single flag derived from a proto field (or a nested
// subfield). Every flag that will be registered on a generated cobra
// command has exactly one FlagSpec in a FlagPlan.
type FlagSpec struct {
	// CLIName is the fully-qualified flag name as it appears on the
	// command line, kebab-case, path-joined with '-'.
	CLIName string

	// Help is the human-readable help text. The kind-specific format hint
	// is appended by the renderer, not here.
	Help string

	// ProtoPath is the proto field path from the root message
	// (snake_case segments). Used at assembly time to set the field on
	// the proto struct.
	ProtoPath []string

	// Kind drives which pflag function to call and which post-processing
	// is needed at assembly time.
	Kind FlagKind

	// EnumGoType is the fully-qualified Go identifier for the enum type.
	// Only populated when Kind == FlagKindEnum.
	EnumGoType string

	// EnumValues lists the valid UPPER_SNAKE enum names. Only populated
	// when Kind == FlagKindEnum.
	EnumValues []string

	// EnumNumbers lists the integer numbers corresponding to EnumValues,
	// in the same order. The CLI template uses these to emit the
	// string-to-int mapping inside the generated pflag.Value Set method
	// without needing to reach for the protoc-gen-go-generated `<Enum>_value`
	// map (which would require the Go ident of the enum, which the plan
	// layer does not carry).
	EnumNumbers []int32

	// JSONGoType is the Go type the JSON will be unmarshaled into. Only
	// populated when Kind == FlagKindJSONEscape.
	JSONGoType string

	// OneofName holds the proto name of the oneof this flag belongs to.
	// Empty string when the flag is not inside a oneof.
	OneofName string

	// Optional reports whether the field is proto3 `optional` or a WKT
	// wrapper. The CLI layer uses this to presence-gate assignment.
	Optional bool

	// Repeated reports whether the underlying proto field is `repeated`.
	// Renderers use it to pick between slice append and scalar assignment.
	Repeated bool
}

// FlagPlan describes how to turn one proto message into a set of
// cobra/pflag flags. One FlagPlan is built per workflow input, per signal
// input, per query input.
type FlagPlan struct {
	// GoTypeName is the root Go type name for the message this plan
	// represents (e.g. "OrderRequest").
	GoTypeName string

	// GoTypePkg is the Go import path for the root message type.
	GoTypePkg string

	// Flags is the ordered, deterministic list of flags to register.
	Flags []*FlagSpec

	// OneofGroups lists oneof groupings. Each inner slice holds the
	// CLINames of the branches that must be marked mutually exclusive.
	OneofGroups [][]string
}

// reservedTargetingFlags are the flag names claimed by the CLI layer for
// identifying a running workflow on signal / query / cancel / terminate
// commands. Any input-derived flag colliding with these is renamed with an
// "input-" prefix so both can coexist.
var reservedTargetingFlags = map[string]bool{
	"workflow-id": true,
	"run-id":      true,
}

// NewFlagPlan walks a message descriptor and produces a FlagPlan describing
// the CLI flags that should be generated for it. Unrepresentable subtrees
// (cyclic types, `Any`/`Struct`/`Value`, repeated messages, maps of
// messages, maps with non-string keys) fall back to a single
// FlagKindJSONEscape flag at the seam rather than failing the build.
//
// Returns an error only for hard failures such as flag-name collisions that
// cannot be resolved automatically.
func NewFlagPlan(msg *protogen.Message) (*FlagPlan, error) {
	if msg == nil {
		return nil, fmt.Errorf("NewFlagPlan: nil message")
	}
	plan := &FlagPlan{
		GoTypeName: msg.GoIdent.GoName,
		GoTypePkg:  string(msg.GoIdent.GoImportPath),
	}
	if err := buildFlagPlan(plan, msg.Desc); err != nil {
		return nil, err
	}
	return plan, nil
}

// buildFlagPlan walks msg.Desc and appends flag specs to plan. It is the
// descriptor-only entry point used by NewFlagPlan and by tests that
// synthesize descriptors without building a full *protogen.Plugin.
func buildFlagPlan(plan *FlagPlan, msg protoreflect.MessageDescriptor) error {
	visited := map[protoreflect.FullName]bool{msg.FullName(): true}
	walkMessage(plan, msg, nil, nil, "", visited)
	if err := applyReservedFlagRenames(plan); err != nil {
		return err
	}
	return checkUniqueFlagNames(plan)
}

// walkMessage walks msg's fields in proto field-number order, appending
// FlagSpecs to plan. cliPrefix is the kebab-prefix ("" at root), protoPath
// is the ordered proto field names leading here, oneofName is propagated
// when a caller is inside a oneof (empty otherwise). visited is the set of
// proto message full names currently on the recursion stack; it is
// copy-on-descend so sibling branches do not confuse the cycle check.
func walkMessage(
	plan *FlagPlan,
	msg protoreflect.MessageDescriptor,
	cliPrefix []string,
	protoPath []string,
	oneofName string,
	visited map[protoreflect.FullName]bool,
) {
	// Collect ordered fields: plain fields first in declaration order,
	// oneofs grouped by oneof, each oneof emitted after its first branch
	// appears. The protoreflect.FieldDescriptors iterator already returns
	// fields in proto field-number order, which matches §2.2 rule 11.
	seenOneofs := map[protoreflect.Name]bool{}

	fds := msg.Fields()
	for i := 0; i < fds.Len(); i++ {
		field := fds.Get(i)

		if od := field.ContainingOneof(); od != nil && !od.IsSynthetic() {
			name := od.Name()
			if seenOneofs[name] {
				continue
			}
			seenOneofs[name] = true
			group := walkOneof(plan, od, cliPrefix, protoPath, visited)
			if len(group) > 0 {
				plan.OneofGroups = append(plan.OneofGroups, group)
			}
			continue
		}

		walkField(plan, field, cliPrefix, protoPath, oneofName, visited)
	}
}

// walkOneof emits every branch of a oneof as its own FlagSpec with
// OneofName set. Returns the list of CLINames appended so the caller can
// record the exclusivity group.
func walkOneof(
	plan *FlagPlan,
	od protoreflect.OneofDescriptor,
	cliPrefix []string,
	protoPath []string,
	visited map[protoreflect.FullName]bool,
) []string {
	oneofName := string(od.Name())
	var group []string
	branches := od.Fields()
	before := len(plan.Flags)
	for i := 0; i < branches.Len(); i++ {
		branch := branches.Get(i)
		walkField(plan, branch, cliPrefix, protoPath, oneofName, visited)
	}
	for _, fs := range plan.Flags[before:] {
		if fs.OneofName == oneofName {
			group = append(group, fs.CLIName)
		}
	}
	return group
}

// walkField emits zero or more FlagSpecs for a single field. Plain scalars
// and enums produce one spec each. Nested messages recurse. `repeated` and
// `map` are handled here via kind selection. WKT special cases short-
// circuit before recursing.
func walkField(
	plan *FlagPlan,
	field protoreflect.FieldDescriptor,
	cliPrefix []string,
	protoPath []string,
	oneofName string,
	visited map[protoreflect.FullName]bool,
) {
	segment := fieldFlagSegment(field)
	cliName := joinKebab(append(append([]string(nil), cliPrefix...), segment))
	childPath := append(append([]string(nil), protoPath...), string(field.Name()))

	// Maps: distinct from repeated message for handling purposes.
	if field.IsMap() {
		walkMapField(plan, field, cliName, childPath, oneofName)
		return
	}

	// Repeated fields (but not maps): decide by element kind.
	if field.Cardinality() == protoreflect.Repeated {
		walkRepeatedField(plan, field, cliName, childPath, oneofName)
		return
	}

	// Singular message field (including WKT and nested).
	if field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind {
		walkSingularMessageField(plan, field, cliName, childPath, oneofName, visited)
		return
	}

	// Singular scalar or enum.
	walkSingularScalarOrEnum(plan, field, cliName, childPath, oneofName)
}

func walkSingularScalarOrEnum(
	plan *FlagPlan,
	field protoreflect.FieldDescriptor,
	cliName string,
	childPath []string,
	oneofName string,
) {
	if field.Kind() == protoreflect.EnumKind {
		enum := field.Enum()
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:     cliName,
			Help:        fieldHelp(field),
			ProtoPath:   childPath,
			Kind:        FlagKindEnum,
			EnumGoType:  string(enum.FullName()),
			EnumValues:  enumValueNames(enum),
			EnumNumbers: enumValueNumbers(enum),
			OneofName:   oneofName,
			Optional:    field.HasOptionalKeyword(),
		})
		return
	}

	kind, ok := scalarKind(field.Kind())
	if !ok {
		// Unknown scalar — fall back to JSONEscape at the seam.
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:    cliName,
			Help:       fieldHelp(field),
			ProtoPath:  childPath,
			Kind:       FlagKindJSONEscape,
			JSONGoType: field.Kind().String(),
			OneofName:  oneofName,
		})
		return
	}
	plan.Flags = append(plan.Flags, &FlagSpec{
		CLIName:   cliName,
		Help:      fieldHelp(field),
		ProtoPath: childPath,
		Kind:      kind,
		OneofName: oneofName,
		Optional:  field.HasOptionalKeyword(),
	})
}

func walkSingularMessageField(
	plan *FlagPlan,
	field protoreflect.FieldDescriptor,
	cliName string,
	childPath []string,
	oneofName string,
	visited map[protoreflect.FullName]bool,
) {
	msgDesc := field.Message()
	full := msgDesc.FullName()

	// WKT special cases first.
	if kind, wkt, ok := wktKind(full); ok {
		switch wkt {
		case wktEmpty:
			// Emit no flag.
			return
		case wktJSONValue:
			plan.Flags = append(plan.Flags, &FlagSpec{
				CLIName:    cliName,
				Help:       fieldHelp(field),
				ProtoPath:  childPath,
				Kind:       FlagKindJSONEscape,
				JSONGoType: string(full),
				OneofName:  oneofName,
			})
			return
		case wktWrapper:
			plan.Flags = append(plan.Flags, &FlagSpec{
				CLIName:   cliName,
				Help:      fieldHelp(field),
				ProtoPath: childPath,
				Kind:      kind,
				OneofName: oneofName,
				Optional:  true,
			})
			return
		case wktLeaf:
			plan.Flags = append(plan.Flags, &FlagSpec{
				CLIName:   cliName,
				Help:      fieldHelp(field),
				ProtoPath: childPath,
				Kind:      kind,
				OneofName: oneofName,
				Optional:  field.HasOptionalKeyword(),
			})
			return
		}
	}

	// Recursion detection via the visited-on-path set.
	if visited[full] {
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:    cliName,
			Help:       fieldHelp(field),
			ProtoPath:  childPath,
			Kind:       FlagKindJSONEscape,
			JSONGoType: string(full),
			OneofName:  oneofName,
		})
		return
	}

	// Recurse into the nested message. Copy visited so siblings don't
	// share cycle state with this descent.
	childVisited := copyVisited(visited)
	childVisited[full] = true
	walkMessage(plan, msgDesc, splitKebab(cliName), childPath, oneofName, childVisited)
}

func walkRepeatedField(
	plan *FlagPlan,
	field protoreflect.FieldDescriptor,
	cliName string,
	childPath []string,
	oneofName string,
) {
	// Repeated messages and groups: JSONEscape at the seam unless the WKT
	// has a repeatable scalar mapping (Timestamp, Duration).
	if field.Kind() == protoreflect.MessageKind || field.Kind() == protoreflect.GroupKind {
		full := field.Message().FullName()
		switch full {
		case "google.protobuf.Timestamp":
			// No slice form for timestamps in pflag standard; fall back to JSON.
			plan.Flags = append(plan.Flags, &FlagSpec{
				CLIName:    cliName,
				Help:       fieldHelp(field),
				ProtoPath:  childPath,
				Kind:       FlagKindJSONEscape,
				JSONGoType: string(full),
				OneofName:  oneofName,
				Repeated:   true,
			})
			return
		case "google.protobuf.Duration":
			// No slice form for durations in pflag standard; fall back to JSON.
			plan.Flags = append(plan.Flags, &FlagSpec{
				CLIName:    cliName,
				Help:       fieldHelp(field),
				ProtoPath:  childPath,
				Kind:       FlagKindJSONEscape,
				JSONGoType: string(full),
				OneofName:  oneofName,
				Repeated:   true,
			})
			return
		default:
			plan.Flags = append(plan.Flags, &FlagSpec{
				CLIName:    cliName,
				Help:       fieldHelp(field),
				ProtoPath:  childPath,
				Kind:       FlagKindJSONEscape,
				JSONGoType: string(full),
				OneofName:  oneofName,
				Repeated:   true,
			})
			return
		}
	}

	// Repeated enum.
	if field.Kind() == protoreflect.EnumKind {
		// pflag has no EnumSliceVar; model it as a repeatable string
		// slice and let the CLI layer validate/convert each element.
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:     cliName,
			Help:        fieldHelp(field),
			ProtoPath:   childPath,
			Kind:        FlagKindStringSlice,
			EnumGoType:  string(field.Enum().FullName()),
			EnumValues:  enumValueNames(field.Enum()),
			EnumNumbers: enumValueNumbers(field.Enum()),
			OneofName:   oneofName,
			Repeated:    true,
		})
		return
	}

	// Repeated scalar.
	kind, ok := repeatedScalarKind(field.Kind())
	if !ok {
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:    cliName,
			Help:       fieldHelp(field),
			ProtoPath:  childPath,
			Kind:       FlagKindJSONEscape,
			JSONGoType: field.Kind().String(),
			OneofName:  oneofName,
			Repeated:   true,
		})
		return
	}
	plan.Flags = append(plan.Flags, &FlagSpec{
		CLIName:   cliName,
		Help:      fieldHelp(field),
		ProtoPath: childPath,
		Kind:      kind,
		OneofName: oneofName,
		Repeated:  true,
	})
}

func walkMapField(
	plan *FlagPlan,
	field protoreflect.FieldDescriptor,
	cliName string,
	childPath []string,
	oneofName string,
) {
	mapMsg := field.Message()
	keyField := mapMsg.Fields().ByNumber(1)
	valField := mapMsg.Fields().ByNumber(2)

	// Map with non-string key → JSONEscape.
	if keyField == nil || keyField.Kind() != protoreflect.StringKind {
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:   cliName,
			Help:      fieldHelp(field),
			ProtoPath: childPath,
			Kind:      FlagKindJSONEscape,
			JSONGoType: string(func() protoreflect.FullName {
				if valField != nil && (valField.Kind() == protoreflect.MessageKind || valField.Kind() == protoreflect.GroupKind) {
					return valField.Message().FullName()
				}
				return mapMsg.FullName()
			}()),
			OneofName: oneofName,
		})
		return
	}

	// map<string, message> → JSONEscape.
	if valField.Kind() == protoreflect.MessageKind || valField.Kind() == protoreflect.GroupKind {
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:    cliName,
			Help:       fieldHelp(field),
			ProtoPath:  childPath,
			Kind:       FlagKindJSONEscape,
			JSONGoType: string(valField.Message().FullName()),
			OneofName:  oneofName,
		})
		return
	}

	// map<string, scalar>: pick the right StringTo* kind.
	switch valField.Kind() {
	case protoreflect.StringKind:
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:   cliName,
			Help:      fieldHelp(field),
			ProtoPath: childPath,
			Kind:      FlagKindStringToString,
			OneofName: oneofName,
		})
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:   cliName,
			Help:      fieldHelp(field),
			ProtoPath: childPath,
			Kind:      FlagKindStringToInt,
			OneofName: oneofName,
		})
	default:
		// Anything else (bool, float, bytes, enum) → JSONEscape. pflag
		// doesn't ship a StringTo<X> for them.
		plan.Flags = append(plan.Flags, &FlagSpec{
			CLIName:    cliName,
			Help:       fieldHelp(field),
			ProtoPath:  childPath,
			Kind:       FlagKindJSONEscape,
			JSONGoType: string(mapMsg.FullName()),
			OneofName:  oneofName,
		})
	}
}

// scalarKind maps a protoreflect.Kind to the corresponding singular
// FlagKind, returning false for kinds that must be JSON-escaped.
func scalarKind(k protoreflect.Kind) (FlagKind, bool) {
	switch k {
	case protoreflect.StringKind:
		return FlagKindString, true
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return FlagKindInt32, true
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return FlagKindInt64, true
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return FlagKindUint32, true
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return FlagKindUint64, true
	case protoreflect.FloatKind:
		return FlagKindFloat32, true
	case protoreflect.DoubleKind:
		return FlagKindFloat64, true
	case protoreflect.BoolKind:
		return FlagKindBool, true
	case protoreflect.BytesKind:
		return FlagKindBytes, true
	}
	return 0, false
}

// repeatedScalarKind maps a protoreflect.Kind to the corresponding slice
// FlagKind. Kinds without a pflag slice binder collapse to
// FlagKindStringSlice or are rejected by the caller.
func repeatedScalarKind(k protoreflect.Kind) (FlagKind, bool) {
	switch k {
	case protoreflect.StringKind, protoreflect.BytesKind:
		return FlagKindStringSlice, true
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return FlagKindInt32Slice, true
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return FlagKindInt64Slice, true
	case protoreflect.BoolKind, protoreflect.FloatKind, protoreflect.DoubleKind:
		// No widely-supported pflag binder; stringify.
		return FlagKindStringSlice, true
	}
	return 0, false
}

// wktKindClass classifies a WKT by how it should be rendered.
type wktKindClass int

const (
	wktNone wktKindClass = iota
	wktLeaf
	wktWrapper
	wktEmpty
	wktJSONValue
)

// wktKind returns the appropriate FlagKind for a well-known type message,
// along with a class telling the caller how to emit the flag (as a leaf
// with the returned kind, as an optional-wrapped leaf, as a JSON value, or
// skipped entirely for Empty).
func wktKind(full protoreflect.FullName) (FlagKind, wktKindClass, bool) {
	switch full {
	case "google.protobuf.Timestamp":
		return FlagKindTimestamp, wktLeaf, true
	case "google.protobuf.Duration":
		return FlagKindDuration, wktLeaf, true
	case "google.protobuf.FieldMask":
		return FlagKindFieldMask, wktLeaf, true
	case "google.protobuf.Empty":
		return 0, wktEmpty, true
	case "google.protobuf.Any", "google.protobuf.Struct", "google.protobuf.Value", "google.protobuf.ListValue":
		return FlagKindJSONEscape, wktJSONValue, true
	case "google.protobuf.StringValue":
		return FlagKindString, wktWrapper, true
	case "google.protobuf.Int32Value":
		return FlagKindInt32, wktWrapper, true
	case "google.protobuf.Int64Value":
		return FlagKindInt64, wktWrapper, true
	case "google.protobuf.UInt32Value":
		return FlagKindUint32, wktWrapper, true
	case "google.protobuf.UInt64Value":
		return FlagKindUint64, wktWrapper, true
	case "google.protobuf.FloatValue":
		return FlagKindFloat32, wktWrapper, true
	case "google.protobuf.DoubleValue":
		return FlagKindFloat64, wktWrapper, true
	case "google.protobuf.BoolValue":
		return FlagKindBool, wktWrapper, true
	case "google.protobuf.BytesValue":
		return FlagKindBytes, wktWrapper, true
	}
	return 0, wktNone, false
}

// fieldFlagSegment returns the single kebab-case segment derived from a
// proto field: `json_name` override if set and non-default, otherwise the
// proto field's snake_case name converted to kebab-case.
func fieldFlagSegment(field protoreflect.FieldDescriptor) string {
	protoName := string(field.Name())
	jsonName := field.JSONName()
	// The default jsonName is lowerCamel of the proto name. If the author
	// explicitly set a [json_name=...] override, it typically differs from
	// both the snake name and the default camel form; prefer it.
	if jsonName != "" && jsonName != defaultJSONName(protoName) {
		return kebabFromJSONName(jsonName)
	}
	return strings.ReplaceAll(protoName, "_", "-")
}

// defaultJSONName returns the lowerCamel form of a snake_case proto field
// name — the same algorithm protoc uses to populate the default
// `json_name` attribute on a FieldDescriptor.
func defaultJSONName(snake string) string {
	var b strings.Builder
	upperNext := false
	for i, r := range snake {
		if r == '_' {
			upperNext = true
			continue
		}
		if i == 0 {
			b.WriteRune(toLower(r))
			continue
		}
		if upperNext {
			b.WriteRune(toUpper(r))
			upperNext = false
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// kebabFromJSONName converts an identifier (typically lowerCamel or a
// user-supplied override) to kebab-case by inserting '-' before each
// interior uppercase letter and lowercasing everything.
func kebabFromJSONName(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && isUpper(r) {
			b.WriteRune('-')
		}
		b.WriteRune(toLower(r))
	}
	// Replace any remaining underscores (defensive — user override may
	// include them).
	return strings.ReplaceAll(b.String(), "_", "-")
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func toLower(r rune) rune {
	if isUpper(r) {
		return r + ('a' - 'A')
	}
	return r
}
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - ('a' - 'A')
	}
	return r
}

// joinKebab joins a path of kebab-case segments with '.'. Empty segments
// are filtered to stay idempotent at the root level.
func joinKebab(parts []string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, ".")
}

// splitKebab is the inverse of joinKebab at the path-segment level. Used
// by the walker to rebuild a prefix slice from a composed CLI name when
// recursing into a nested message. Segments are separated by '.'; the
// dashes inside each kebab-cased segment are preserved.
func splitKebab(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ".")
}

// fieldHelp extracts the help text for a field. The plan layer returns the
// raw proto comment (or a fallback kebab-name); kind-specific format hints
// are appended later by the template.
func fieldHelp(field protoreflect.FieldDescriptor) string {
	// Descriptor-only access gives us the field name — SourceLocation
	// comments live on the containing file descriptor.
	if parent := field.ParentFile(); parent != nil {
		loc := parent.SourceLocations().ByDescriptor(field)
		if c := strings.TrimSpace(loc.LeadingComments); c != "" {
			return oneLine(c)
		}
		if c := strings.TrimSpace(loc.TrailingComments); c != "" {
			return oneLine(c)
		}
	}
	return strings.ReplaceAll(string(field.Name()), "_", "-")
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	// Collapse runs of whitespace.
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if lastSpace {
				continue
			}
			lastSpace = true
			b.WriteRune(' ')
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// enumValueNames returns the UPPER_SNAKE value names of an enum in
// declaration order.
func enumValueNames(enum protoreflect.EnumDescriptor) []string {
	values := enum.Values()
	names := make([]string, 0, values.Len())
	for i := 0; i < values.Len(); i++ {
		names = append(names, string(values.Get(i).Name()))
	}
	return names
}

// enumValueNumbers returns the integer numbers of an enum's values in
// declaration order. The slice is aligned with enumValueNames.
func enumValueNumbers(enum protoreflect.EnumDescriptor) []int32 {
	values := enum.Values()
	nums := make([]int32, 0, values.Len())
	for i := 0; i < values.Len(); i++ {
		nums = append(nums, int32(values.Get(i).Number()))
	}
	return nums
}

func copyVisited(in map[protoreflect.FullName]bool) map[protoreflect.FullName]bool {
	out := make(map[protoreflect.FullName]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// applyReservedFlagRenames rewrites any input-derived flag that would
// collide with `--workflow-id` or `--run-id` (reserved for CLI-layer
// targeting) to `input-<name>`. See PLAN.md §4.
func applyReservedFlagRenames(plan *FlagPlan) error {
	for _, fs := range plan.Flags {
		if reservedTargetingFlags[fs.CLIName] {
			fs.CLIName = "input-" + fs.CLIName
		}
	}
	// Also fix up any oneof groups that referenced the renamed flag.
	for gi, group := range plan.OneofGroups {
		for i, name := range group {
			if reservedTargetingFlags[name] {
				plan.OneofGroups[gi][i] = "input-" + name
			}
		}
	}
	return nil
}

// checkUniqueFlagNames asserts every FlagSpec.CLIName is unique. A
// collision is a generator error — the caller (or the proto author) must
// disambiguate, typically with a `json_name` override on one of the
// colliding fields.
func checkUniqueFlagNames(plan *FlagPlan) error {
	seen := make(map[string]int, len(plan.Flags))
	for i, fs := range plan.Flags {
		if prev, dup := seen[fs.CLIName]; dup {
			return fmt.Errorf("flag name collision on %q between fields %v and %v",
				fs.CLIName,
				strings.Join(plan.Flags[prev].ProtoPath, "."),
				strings.Join(fs.ProtoPath, "."))
		}
		seen[fs.CLIName] = i
	}
	return nil
}
