package model

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TestFlagPlanAllScalarKinds exercises every singular scalar/bytes/bool
// kind and asserts the walker picks the expected FlagKind for each.
func TestFlagPlanAllScalarKinds(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("Scalars"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("str", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			scalarField("i32", 2, descriptorpb.FieldDescriptorProto_TYPE_INT32),
			scalarField("i64", 3, descriptorpb.FieldDescriptorProto_TYPE_INT64),
			scalarField("u32", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT32),
			scalarField("u64", 5, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			scalarField("f32", 6, descriptorpb.FieldDescriptorProto_TYPE_FLOAT),
			scalarField("f64", 7, descriptorpb.FieldDescriptorProto_TYPE_DOUBLE),
			scalarField("b", 8, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
			scalarField("raw", 9, descriptorpb.FieldDescriptorProto_TYPE_BYTES),
		},
	}
	fd := newFile("scalars.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Scalars")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	expect := []struct {
		name string
		kind FlagKind
	}{
		{"str", FlagKindString},
		{"i32", FlagKindInt32},
		{"i64", FlagKindInt64},
		{"u32", FlagKindUint32},
		{"u64", FlagKindUint64},
		{"f32", FlagKindFloat32},
		{"f64", FlagKindFloat64},
		{"b", FlagKindBool},
		{"raw", FlagKindBytes},
	}
	if len(plan.Flags) != len(expect) {
		t.Fatalf("flags len = %d, want %d: %+v", len(plan.Flags), len(expect), plan.Flags)
	}
	for i, want := range expect {
		got := plan.Flags[i]
		if got.CLIName != want.name {
			t.Errorf("flag %d: CLIName = %q, want %q", i, got.CLIName, want.name)
		}
		if got.Kind != want.kind {
			t.Errorf("flag %q: Kind = %s, want %s", got.CLIName, got.Kind, want.kind)
		}
	}
}

// TestFlagPlanRepeatedScalarKinds covers the slice FlagKinds.
func TestFlagPlanRepeatedScalarKinds(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("Slices"),
		Field: []*descriptorpb.FieldDescriptorProto{
			repeatedField("names", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			repeatedField("scores", 2, descriptorpb.FieldDescriptorProto_TYPE_INT32),
			repeatedField("counts", 3, descriptorpb.FieldDescriptorProto_TYPE_INT64),
			repeatedField("bytes_list", 4, descriptorpb.FieldDescriptorProto_TYPE_BYTES),
		},
	}
	fd := newFile("slices.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Slices")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	cases := []struct {
		name string
		kind FlagKind
	}{
		{"names", FlagKindStringSlice},
		{"scores", FlagKindInt32Slice},
		{"counts", FlagKindInt64Slice},
		{"bytes-list", FlagKindStringSlice},
	}
	for _, c := range cases {
		f := findFlag(plan, c.name)
		if f == nil {
			t.Fatalf("missing flag %q: got %+v", c.name, plan.Flags)
		}
		if f.Kind != c.kind {
			t.Errorf("flag %q: Kind = %s, want %s", c.name, f.Kind, c.kind)
		}
		if !f.Repeated {
			t.Errorf("flag %q: expected Repeated=true", c.name)
		}
	}
}

// TestFlagPlanEnumKind verifies FlagKindEnum populates EnumGoType /
// EnumValues and carries the declared value list in declaration order.
func TestFlagPlanEnumKind(t *testing.T) {
	enum := &descriptorpb.EnumDescriptorProto{
		Name: proto.String("Status"),
		Value: []*descriptorpb.EnumValueDescriptorProto{
			{Name: proto.String("STATUS_UNSPECIFIED"), Number: proto.Int32(0)},
			{Name: proto.String("STATUS_ACTIVE"), Number: proto.Int32(1)},
			{Name: proto.String("STATUS_CANCELLED"), Number: proto.Int32(2)},
		},
	}
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("EnumMsg"),
		Field: []*descriptorpb.FieldDescriptorProto{
			enumField("status", 1, ".test.Status"),
		},
	}
	fd := newFile("enum.proto", "test").addEnum(enum).addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "EnumMsg")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "status")
	if f == nil {
		t.Fatalf("missing status flag: %+v", plan.Flags)
	}
	if f.Kind != FlagKindEnum {
		t.Fatalf("Kind = %s, want Enum", f.Kind)
	}
	if f.EnumGoType != "test.Status" {
		t.Errorf("EnumGoType = %q, want test.Status", f.EnumGoType)
	}
	want := []string{"STATUS_UNSPECIFIED", "STATUS_ACTIVE", "STATUS_CANCELLED"}
	if len(f.EnumValues) != len(want) {
		t.Fatalf("EnumValues = %v, want %v", f.EnumValues, want)
	}
	for i, v := range want {
		if f.EnumValues[i] != v {
			t.Errorf("EnumValues[%d] = %q, want %q", i, f.EnumValues[i], v)
		}
	}
}

// TestFlagPlanWKTLeafs covers Timestamp, Duration, FieldMask → dedicated
// FlagKinds.
func TestFlagPlanWKTLeafs(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("WKTLeafs"),
		Field: []*descriptorpb.FieldDescriptorProto{
			messageField("when", 1, ".google.protobuf.Timestamp"),
			messageField("dur", 2, ".google.protobuf.Duration"),
			messageField("mask", 3, ".google.protobuf.FieldMask"),
		},
	}
	fd := newFile("wkt_leafs.proto", "test").
		withImport("google/protobuf/timestamp.proto").
		withImport("google/protobuf/duration.proto").
		withImport("google/protobuf/field_mask.proto").
		addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "WKTLeafs")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	cases := []struct {
		name string
		kind FlagKind
	}{
		{"when", FlagKindTimestamp},
		{"dur", FlagKindDuration},
		{"mask", FlagKindFieldMask},
	}
	for _, c := range cases {
		f := findFlag(plan, c.name)
		if f == nil {
			t.Fatalf("missing %q: %+v", c.name, plan.Flags)
		}
		if f.Kind != c.kind {
			t.Errorf("%q: Kind = %s, want %s", c.name, f.Kind, c.kind)
		}
	}
}

// TestFlagPlanWKTWrappers verifies google.protobuf.*Value wrappers produce
// the underlying scalar kind with Optional=true so the CLI layer can
// presence-gate assignment.
func TestFlagPlanWKTWrappers(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("Wrapped"),
		Field: []*descriptorpb.FieldDescriptorProto{
			messageField("maybe_s", 1, ".google.protobuf.StringValue"),
			messageField("maybe_i32", 2, ".google.protobuf.Int32Value"),
			messageField("maybe_i64", 3, ".google.protobuf.Int64Value"),
			messageField("maybe_b", 4, ".google.protobuf.BoolValue"),
			messageField("maybe_d", 5, ".google.protobuf.DoubleValue"),
			messageField("maybe_bytes", 6, ".google.protobuf.BytesValue"),
		},
	}
	fd := newFile("wrap.proto", "test").
		withImport("google/protobuf/wrappers.proto").
		addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Wrapped")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	cases := []struct {
		name string
		kind FlagKind
	}{
		{"maybe-s", FlagKindString},
		{"maybe-i32", FlagKindInt32},
		{"maybe-i64", FlagKindInt64},
		{"maybe-b", FlagKindBool},
		{"maybe-d", FlagKindFloat64},
		{"maybe-bytes", FlagKindBytes},
	}
	for _, c := range cases {
		f := findFlag(plan, c.name)
		if f == nil {
			t.Fatalf("missing %q: %+v", c.name, plan.Flags)
		}
		if f.Kind != c.kind {
			t.Errorf("%q: Kind = %s, want %s", c.name, f.Kind, c.kind)
		}
		if !f.Optional {
			t.Errorf("%q: expected Optional=true", c.name)
		}
	}
}

// TestFlagPlanWKTSkippedAndJSON covers Empty (skipped) and
// Any/Struct/Value (FlagKindJSONEscape).
func TestFlagPlanWKTSkippedAndJSON(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("SpecialWKT"),
		Field: []*descriptorpb.FieldDescriptorProto{
			messageField("nada", 1, ".google.protobuf.Empty"),
			messageField("blob", 2, ".google.protobuf.Any"),
			messageField("cfg", 3, ".google.protobuf.Struct"),
			messageField("val", 4, ".google.protobuf.Value"),
		},
	}
	fd := newFile("special.proto", "test").
		withImport("google/protobuf/empty.proto").
		withImport("google/protobuf/any.proto").
		withImport("google/protobuf/struct.proto").
		addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "SpecialWKT")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if findFlag(plan, "nada") != nil {
		t.Errorf("expected Empty field to emit no flag, got %+v", plan.Flags)
	}
	for _, n := range []string{"blob", "cfg", "val"} {
		f := findFlag(plan, n)
		if f == nil {
			t.Fatalf("missing %q flag: %+v", n, plan.Flags)
		}
		if f.Kind != FlagKindJSONEscape {
			t.Errorf("%q: Kind = %s, want JSONEscape", n, f.Kind)
		}
	}
}

// TestFlagPlanNestedMessagePrefix verifies deep nesting joins segments
// with '-' as the prefix path walks down.
func TestFlagPlanNestedMessagePrefix(t *testing.T) {
	street := &descriptorpb.DescriptorProto{
		Name: proto.String("Address"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("street", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
	shipping := &descriptorpb.DescriptorProto{
		Name: proto.String("Shipping"),
		Field: []*descriptorpb.FieldDescriptorProto{
			messageField("address", 1, ".test.Address"),
		},
	}
	order := &descriptorpb.DescriptorProto{
		Name: proto.String("Order"),
		Field: []*descriptorpb.FieldDescriptorProto{
			messageField("shipping", 1, ".test.Shipping"),
		},
	}
	fd := newFile("order.proto", "test").
		addMessage(street).addMessage(shipping).addMessage(order).build(t)

	plan, err := planForMessage(t, fd, "Order")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "shipping.address.street")
	if f == nil {
		t.Fatalf("expected flag shipping.address.street, got %+v", plan.Flags)
	}
	if got, want := strings.Join(f.ProtoPath, "."), "shipping.address.street"; got != want {
		t.Errorf("ProtoPath = %q, want %q", got, want)
	}
}

// TestFlagPlanRepeatedMessageJSONEscape verifies that a repeated message
// field becomes a single FlagKindJSONEscape flag naming the element type.
func TestFlagPlanRepeatedMessageJSONEscape(t *testing.T) {
	item := &descriptorpb.DescriptorProto{
		Name: proto.String("Item"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("sku", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
	order := &descriptorpb.DescriptorProto{
		Name: proto.String("OrderWithItems"),
		Field: []*descriptorpb.FieldDescriptorProto{
			repeatedMessageField("items", 1, ".test.Item"),
		},
	}
	fd := newFile("items.proto", "test").addMessage(item).addMessage(order).build(t)

	plan, err := planForMessage(t, fd, "OrderWithItems")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "items")
	if f == nil {
		t.Fatalf("missing items flag: %+v", plan.Flags)
	}
	if f.Kind != FlagKindJSONEscape {
		t.Errorf("Kind = %s, want JSONEscape", f.Kind)
	}
	if f.JSONGoType != "test.Item" {
		t.Errorf("JSONGoType = %q, want test.Item", f.JSONGoType)
	}
	if !f.Repeated {
		t.Errorf("expected Repeated=true")
	}
	// Item's inner fields must NOT be present — we don't recurse into the
	// repeated element.
	if findFlag(plan, "items-sku") != nil {
		t.Errorf("unexpected items-sku recurse: %+v", plan.Flags)
	}
}

// TestFlagPlanMapStringToString covers the simplest map form.
func TestFlagPlanMapStringToString(t *testing.T) {
	field, entry := mapFieldDesc("labels", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, descriptorpb.FieldDescriptorProto_TYPE_STRING, "")
	msg := &descriptorpb.DescriptorProto{
		Name:        proto.String("MapMsg"),
		Field:       []*descriptorpb.FieldDescriptorProto{field},
		NestedType:  []*descriptorpb.DescriptorProto{entry},
	}
	fd := newFile("map_ss.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "MapMsg")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "labels")
	if f == nil {
		t.Fatalf("missing labels flag: %+v", plan.Flags)
	}
	if f.Kind != FlagKindStringToString {
		t.Errorf("Kind = %s, want StringToString", f.Kind)
	}
}

// TestFlagPlanMapStringToInt covers the numeric-value map form.
func TestFlagPlanMapStringToInt(t *testing.T) {
	field, entry := mapFieldDesc("counts", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, descriptorpb.FieldDescriptorProto_TYPE_INT64, "")
	msg := &descriptorpb.DescriptorProto{
		Name:       proto.String("MapMsg"),
		Field:      []*descriptorpb.FieldDescriptorProto{field},
		NestedType: []*descriptorpb.DescriptorProto{entry},
	}
	fd := newFile("map_si.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "MapMsg")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "counts")
	if f == nil {
		t.Fatalf("missing counts flag: %+v", plan.Flags)
	}
	if f.Kind != FlagKindStringToInt {
		t.Errorf("Kind = %s, want StringToInt", f.Kind)
	}
}

// TestFlagPlanMapNonStringKeyJSON covers maps with non-string keys.
func TestFlagPlanMapNonStringKeyJSON(t *testing.T) {
	field, entry := mapFieldDesc("by_code", 1, descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_STRING, "")
	msg := &descriptorpb.DescriptorProto{
		Name:       proto.String("MapMsg"),
		Field:      []*descriptorpb.FieldDescriptorProto{field},
		NestedType: []*descriptorpb.DescriptorProto{entry},
	}
	fd := newFile("map_is.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "MapMsg")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "by-code")
	if f == nil {
		t.Fatalf("missing by-code flag: %+v", plan.Flags)
	}
	if f.Kind != FlagKindJSONEscape {
		t.Errorf("Kind = %s, want JSONEscape", f.Kind)
	}
}

// TestFlagPlanMapStringToMessageJSON covers map<string, Message>.
func TestFlagPlanMapStringToMessageJSON(t *testing.T) {
	inner := &descriptorpb.DescriptorProto{
		Name: proto.String("Detail"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("v", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
	field, entry := mapFieldDesc("details", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, ".test.Detail")
	msg := &descriptorpb.DescriptorProto{
		Name:       proto.String("MapMsg"),
		Field:      []*descriptorpb.FieldDescriptorProto{field},
		NestedType: []*descriptorpb.DescriptorProto{entry},
	}
	fd := newFile("map_sm.proto", "test").addMessage(inner).addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "MapMsg")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "details")
	if f == nil {
		t.Fatalf("missing details flag: %+v", plan.Flags)
	}
	if f.Kind != FlagKindJSONEscape {
		t.Errorf("Kind = %s, want JSONEscape", f.Kind)
	}
	if f.JSONGoType != "test.Detail" {
		t.Errorf("JSONGoType = %q, want test.Detail", f.JSONGoType)
	}
}

// TestFlagPlanRecursiveType verifies a self-referential message degrades
// to JSONEscape at the seam but still produces a plan for the outer
// fields.
func TestFlagPlanRecursiveType(t *testing.T) {
	tree := &descriptorpb.DescriptorProto{
		Name: proto.String("Tree"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("label", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			messageField("left", 2, ".test.Tree"),
			messageField("right", 3, ".test.Tree"),
		},
	}
	fd := newFile("tree.proto", "test").addMessage(tree).build(t)

	plan, err := planForMessage(t, fd, "Tree")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if findFlag(plan, "label") == nil {
		t.Errorf("expected label flag present")
	}
	for _, n := range []string{"left", "right"} {
		f := findFlag(plan, n)
		if f == nil {
			t.Fatalf("missing %q flag: %+v", n, plan.Flags)
		}
		if f.Kind != FlagKindJSONEscape {
			t.Errorf("%q: Kind = %s, want JSONEscape (cycle seam)", n, f.Kind)
		}
	}
}

// TestFlagPlanOneofGroup verifies oneof branches get OneofName set and
// the group is recorded in OneofGroups with the branches' CLI names.
func TestFlagPlanOneofGroup(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("Payment"),
		OneofDecl: []*descriptorpb.OneofDescriptorProto{
			{Name: proto.String("method")},
		},
		Field: []*descriptorpb.FieldDescriptorProto{
			oneofField("card_number", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, 0),
			oneofField("bank_account", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING, 0),
			scalarField("amount", 3, descriptorpb.FieldDescriptorProto_TYPE_INT64),
		},
	}
	fd := newFile("payment.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Payment")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	for _, n := range []string{"card-number", "bank-account"} {
		f := findFlag(plan, n)
		if f == nil {
			t.Fatalf("missing %q flag: %+v", n, plan.Flags)
		}
		if f.OneofName != "method" {
			t.Errorf("%q: OneofName = %q, want method", n, f.OneofName)
		}
	}
	if len(plan.OneofGroups) != 1 {
		t.Fatalf("OneofGroups len = %d, want 1", len(plan.OneofGroups))
	}
	group := plan.OneofGroups[0]
	if len(group) != 2 {
		t.Fatalf("group = %v, want 2 entries", group)
	}
	if group[0] != "card-number" || group[1] != "bank-account" {
		t.Errorf("group members = %v, want [card-number bank-account]", group)
	}
	// Non-oneof field should not carry OneofName.
	if f := findFlag(plan, "amount"); f == nil || f.OneofName != "" {
		t.Errorf("amount flag OneofName mismatch: %+v", f)
	}
}

// TestFlagPlanNameCollision verifies two fields at the same level whose
// rendered CLI names collide produce an explicit error rather than silently
// dropping one branch. Nesting-vs-underscore no longer collides (segments
// are joined with '.' while underscores become '-' inside a segment), so we
// engineer the clash via two identical `json_name` overrides on sibling
// fields — proto lets you do this and it is the user-reachable failure
// mode the collision check still needs to cover.
func TestFlagPlanNameCollision(t *testing.T) {
	f1 := scalarField("customer_name", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)
	f1.JsonName = proto.String("handle")
	f2 := scalarField("merchant_tag", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING)
	f2.JsonName = proto.String("handle")
	msg := &descriptorpb.DescriptorProto{
		Name:  proto.String("Outer"),
		Field: []*descriptorpb.FieldDescriptorProto{f1, f2},
	}
	fd := newFile("collide.proto", "test").addMessage(msg).build(t)

	_, err := planForMessage(t, fd, "Outer")
	if err == nil {
		t.Fatalf("expected collision error, got nil")
	}
	if !strings.Contains(err.Error(), "collision") && !strings.Contains(err.Error(), "handle") {
		t.Errorf("error should mention collision / handle, got: %v", err)
	}
}

// TestFlagPlanJSONNameOverride verifies a custom [json_name="..."] is
// honored and kebab-cased in the CLI flag name.
func TestFlagPlanJSONNameOverride(t *testing.T) {
	f1 := scalarField("customer_name", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)
	f1.JsonName = proto.String("buyerHandle")
	msg := &descriptorpb.DescriptorProto{
		Name:  proto.String("Overridden"),
		Field: []*descriptorpb.FieldDescriptorProto{f1},
	}
	fd := newFile("override.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Overridden")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if findFlag(plan, "buyer-handle") == nil {
		t.Fatalf("expected buyer-handle flag, got %+v", plan.Flags)
	}
	if findFlag(plan, "customer-name") != nil {
		t.Errorf("did not expect default customer-name flag when json_name override set")
	}
}

// TestFlagPlanReservedRename verifies that an input-derived flag named
// workflow-id or run-id is rewritten to input-workflow-id / input-run-id
// so the CLI layer's targeting flags can coexist with input fields.
func TestFlagPlanReservedRename(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("Targeted"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("workflow_id", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			scalarField("run_id", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			scalarField("note", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
	fd := newFile("targeted.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Targeted")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if findFlag(plan, "workflow-id") != nil {
		t.Errorf("expected workflow-id to be renamed away")
	}
	if findFlag(plan, "run-id") != nil {
		t.Errorf("expected run-id to be renamed away")
	}
	if findFlag(plan, "input-workflow-id") == nil {
		t.Errorf("expected input-workflow-id flag: %+v", plan.Flags)
	}
	if findFlag(plan, "input-run-id") == nil {
		t.Errorf("expected input-run-id flag: %+v", plan.Flags)
	}
	if findFlag(plan, "note") == nil {
		t.Errorf("unrelated note flag should remain")
	}
}

// TestFlagPlanProto3Optional verifies proto3 `optional` marks the flag
// Optional=true so the CLI layer can presence-gate assignment.
func TestFlagPlanProto3Optional(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("OptMsg"),
		OneofDecl: []*descriptorpb.OneofDescriptorProto{
			{Name: proto.String("_maybe")},
		},
		Field: []*descriptorpb.FieldDescriptorProto{
			optionalField("maybe", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING, 0),
		},
	}
	fd := newFile("opt.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "OptMsg")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	f := findFlag(plan, "maybe")
	if f == nil {
		t.Fatalf("missing maybe flag: %+v", plan.Flags)
	}
	if !f.Optional {
		t.Errorf("expected Optional=true for proto3 optional field")
	}
	if f.OneofName != "" {
		t.Errorf("proto3 optional synthetic oneof should not propagate OneofName, got %q", f.OneofName)
	}
}

// TestFlagPlanGoTypeInfo verifies the root Go type name and package path
// are carried through to the plan (populated via NewFlagPlan path; here we
// synthesize them directly to cover the struct contract).
func TestFlagPlanGoTypeInfo(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("Foo"),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("x", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
	fd := newFile("foo.proto", "test").addMessage(msg).build(t)

	plan, err := planForMessage(t, fd, "Foo")
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if plan.GoTypeName != "Foo" {
		t.Errorf("GoTypeName = %q, want Foo", plan.GoTypeName)
	}
	if plan.GoTypePkg == "" {
		t.Errorf("GoTypePkg empty")
	}
}
