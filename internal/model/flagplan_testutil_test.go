package model

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	// Side-effect imports: register WKT file descriptors with
	// protoregistry.GlobalFiles so synthetic files can depend on them.
	_ "google.golang.org/protobuf/types/known/anypb"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/emptypb"
	_ "google.golang.org/protobuf/types/known/fieldmaskpb"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
)

// fileBuilder is a tiny convenience wrapper to build a
// *descriptorpb.FileDescriptorProto from typed pieces. Tests use it to
// synthesize tailored proto fixtures for FlagPlan cases without writing
// full .proto files on disk.
type fileBuilder struct {
	name     string
	pkg      string
	imports  []string
	messages []*descriptorpb.DescriptorProto
	enums    []*descriptorpb.EnumDescriptorProto
}

func newFile(name, pkg string) *fileBuilder {
	return &fileBuilder{
		name:    name,
		pkg:     pkg,
		imports: nil,
	}
}

func (b *fileBuilder) withImport(path string) *fileBuilder {
	b.imports = append(b.imports, path)
	return b
}

func (b *fileBuilder) addMessage(m *descriptorpb.DescriptorProto) *fileBuilder {
	b.messages = append(b.messages, m)
	return b
}

func (b *fileBuilder) addEnum(e *descriptorpb.EnumDescriptorProto) *fileBuilder {
	b.enums = append(b.enums, e)
	return b
}

// build resolves all imports via protoregistry.GlobalFiles + any extra
// FileDescriptors provided by the caller, and returns the fully-linked
// FileDescriptor. Panics on failure — tests call it inline.
func (b *fileBuilder) build(t *testing.T, extra ...protoreflect.FileDescriptor) protoreflect.FileDescriptor {
	t.Helper()
	syntax := "proto3"
	fdp := &descriptorpb.FileDescriptorProto{
		Name:        proto.String(b.name),
		Package:     proto.String(b.pkg),
		Syntax:      proto.String(syntax),
		Dependency:  b.imports,
		MessageType: b.messages,
		EnumType:    b.enums,
	}

	// Start with the global registry so WKTs resolve, then layer in any
	// extras we were handed. We build a private registry so each test is
	// independent.
	reg := new(protoregistry.Files)
	// Seed only the imports we asked for to avoid duplicate-registration
	// errors when tests add the same file to multiple registries.
	for _, imp := range b.imports {
		if _, err := reg.FindFileByPath(imp); err == nil {
			continue
		}
		f, err := protoregistry.GlobalFiles.FindFileByPath(imp)
		if err != nil {
			t.Fatalf("resolving import %q: %v", imp, err)
		}
		if err := reg.RegisterFile(f); err != nil {
			t.Fatalf("registering %q: %v", imp, err)
		}
	}
	for _, ex := range extra {
		if err := reg.RegisterFile(ex); err != nil {
			t.Fatalf("registering extra %q: %v", ex.Path(), err)
		}
	}

	fd, err := protodesc.NewFile(fdp, reg)
	if err != nil {
		t.Fatalf("building %s: %v", b.name, err)
	}
	return fd
}

// planForMessage is a one-liner used by tests: build a FlagPlan from a
// message descriptor inside a synthesized file.
func planForMessage(t *testing.T, fd protoreflect.FileDescriptor, name string) (*FlagPlan, error) {
	t.Helper()
	msg := fd.Messages().ByName(protoreflect.Name(name))
	if msg == nil {
		t.Fatalf("message %q not found in %s", name, fd.Path())
	}
	plan := &FlagPlan{
		GoTypeName: name,
		GoTypePkg:  "example.com/fake/" + string(fd.Package()),
	}
	err := buildFlagPlan(plan, msg)
	return plan, err
}

// findFlag returns the FlagSpec with the given CLI name, or nil. Tests
// assert on fields of the result.
func findFlag(plan *FlagPlan, name string) *FlagSpec {
	for _, fs := range plan.Flags {
		if fs.CLIName == name {
			return fs
		}
	}
	return nil
}

// Field-builder helpers: shorter than writing out descriptorpb struct
// literals inline. They only cover what the tests need.

func scalarField(name string, num int32, kind descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	k := kind
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(num),
		Type:   &k,
		Label:  &label,
	}
}

func repeatedField(name string, num int32, kind descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	f := scalarField(name, num, kind)
	l := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	f.Label = &l
	return f
}

func messageField(name string, num int32, typeName string) *descriptorpb.FieldDescriptorProto {
	k := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(num),
		Type:     &k,
		Label:    &label,
		TypeName: proto.String(typeName),
	}
}

func repeatedMessageField(name string, num int32, typeName string) *descriptorpb.FieldDescriptorProto {
	f := messageField(name, num, typeName)
	l := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	f.Label = &l
	return f
}

func enumField(name string, num int32, typeName string) *descriptorpb.FieldDescriptorProto {
	k := descriptorpb.FieldDescriptorProto_TYPE_ENUM
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(num),
		Type:     &k,
		Label:    &label,
		TypeName: proto.String(typeName),
	}
}

// mapFieldDesc creates a field + the synthetic entry message required by
// protobuf for map fields. The caller appends the entry message to its
// container's nested type list.
func mapFieldDesc(name string, num int32, keyType, valueType descriptorpb.FieldDescriptorProto_Type, valueTypeName string) (*descriptorpb.FieldDescriptorProto, *descriptorpb.DescriptorProto) {
	entryName := mapEntryName(name)

	keyKind := keyType
	labelOpt := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	keyField := &descriptorpb.FieldDescriptorProto{
		Name:   proto.String("key"),
		Number: proto.Int32(1),
		Type:   &keyKind,
		Label:  &labelOpt,
	}
	valKind := valueType
	valField := &descriptorpb.FieldDescriptorProto{
		Name:   proto.String("value"),
		Number: proto.Int32(2),
		Type:   &valKind,
		Label:  &labelOpt,
	}
	if valueTypeName != "" {
		valField.TypeName = proto.String(valueTypeName)
	}

	isMapEntry := true
	entry := &descriptorpb.DescriptorProto{
		Name:  proto.String(entryName),
		Field: []*descriptorpb.FieldDescriptorProto{keyField, valField},
		Options: &descriptorpb.MessageOptions{
			MapEntry: &isMapEntry,
		},
	}

	mapKind := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	labelRep := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	field := &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(num),
		Type:     &mapKind,
		Label:    &labelRep,
		TypeName: proto.String(entryName),
	}
	return field, entry
}

// mapEntryName returns the synthetic map entry message name for a given
// field name (UpperCamel + "Entry").
func mapEntryName(fieldName string) string {
	var b []rune
	next := true
	for _, r := range fieldName {
		if r == '_' {
			next = true
			continue
		}
		if next && r >= 'a' && r <= 'z' {
			b = append(b, r-('a'-'A'))
		} else {
			b = append(b, r)
		}
		next = false
	}
	return string(b) + "Entry"
}

// optionalField marks a scalar field as proto3 `optional` by putting it in
// a synthetic oneof. Used by the optional-presence test.
func optionalField(name string, num int32, kind descriptorpb.FieldDescriptorProto_Type, oneofIndex int32) *descriptorpb.FieldDescriptorProto {
	f := scalarField(name, num, kind)
	p := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	f.Label = &p
	f.Proto3Optional = proto.Bool(true)
	f.OneofIndex = proto.Int32(oneofIndex)
	return f
}

// oneofField marks a field as part of a declared oneof.
func oneofField(name string, num int32, kind descriptorpb.FieldDescriptorProto_Type, oneofIndex int32) *descriptorpb.FieldDescriptorProto {
	f := scalarField(name, num, kind)
	f.OneofIndex = proto.Int32(oneofIndex)
	return f
}

// oneofMessageField marks a message field as a oneof branch.
func oneofMessageField(name string, num int32, typeName string, oneofIndex int32) *descriptorpb.FieldDescriptorProto {
	f := messageField(name, num, typeName)
	f.OneofIndex = proto.Int32(oneofIndex)
	return f
}
