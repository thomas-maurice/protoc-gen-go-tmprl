package tmpl

import (
	"reflect"
	"testing"
	"time"

	"github.com/thomas-maurice/protoc-gen-go-tmprl/internal/model"
)

// TestFormatStringSlice Verifies markdown-friendly rendering of string slices
// used in the generated documentation (each entry wrapped in backticks so
// identifier-like values render as inline code).
func TestFormatStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"nil slice", nil, ""},
		{"empty slice", []string{}, ""},
		{"single entry", []string{"FATAL"}, "`FATAL`"},
		{"multiple entries", []string{"FATAL", "NOT_FOUND"}, "`FATAL`, `NOT_FOUND`"},
		{"preserves entries verbatim", []string{"a b", "c"}, "`a b`, `c`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatStringSlice(tt.input); got != tt.expected {
				t.Errorf("formatStringSlice(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestToSeconds Tests duration to seconds conversion
func TestToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"int32", int32(50), 50},
		{"time.Duration seconds", 30 * time.Second, 30},
		{"time.Duration minutes", 2 * time.Minute, 120},
		{"unknown type", "string", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("toSeconds(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWorkflowObjectName Tests workflow object name generation
func TestWorkflowObjectName(t *testing.T) {
	tests := []struct {
		serviceName string
		methodName  string
		expected    string
	}{
		{"Example", "Ping", "ExamplePing"},
		{"User", "Create", "UserCreate"},
		{"", "Test", "Test"},
		{"Service", "", "Service"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := workflowObjectName(tt.serviceName, tt.methodName)
			if result != tt.expected {
				t.Errorf("workflowObjectName(%q, %q) = %q, expected %q",
					tt.serviceName, tt.methodName, result, tt.expected)
			}
		})
	}
}

// TestChildWorkflowObjectName Tests child workflow object name generation
func TestChildWorkflowObjectName(t *testing.T) {
	tests := []struct {
		serviceName string
		methodName  string
		expected    string
	}{
		{"Example", "Ping", "ChildExamplePingExecution"},
		{"User", "Create", "ChildUserCreateExecution"},
		{"", "Test", "ChildTestExecution"},
		{"Service", "", "ChildServiceExecution"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := childWorkflowObjectName(tt.serviceName, tt.methodName)
			if result != tt.expected {
				t.Errorf("childWorkflowObjectName(%q, %q) = %q, expected %q",
					tt.serviceName, tt.methodName, result, tt.expected)
			}
		})
	}
}

// TestCommentOneLine Tests multiline comment collapse
func TestCommentOneLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"multiline with newlines",
			"This is a comment\nwith multiple lines\nand more text",
			"This is a comment with multiple lines and more text",
		},
		{
			"single line",
			"This is a single line comment",
			"This is a single line comment",
		},
		{
			"with carriage returns",
			"Line1\r\nLine2\r\nLine3",
			"Line1 Line2 Line3",
		},
		{
			"multiple spaces",
			"This   has    multiple     spaces",
			"This has multiple spaces",
		},
		{
			"mixed newlines and spaces",
			"First line\n  Second line  \n    Third line",
			"First line Second line Third line",
		},
		{
			"empty string",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commentOneLine(tt.input)
			if result != tt.expected {
				t.Errorf("commentOneLine(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCommentBlock Tests conversion of proto comments to multi-line Go doc blocks
func TestCommentBlock(t *testing.T) {
	tests := []struct {
		name     string
		indent   string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			indent:   "",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only input",
			indent:   "\t",
			input:    "   \n   \n",
			expected: "",
		},
		{
			name:     "single line",
			indent:   "",
			input:    " Throws a d6 and returns the result",
			expected: "// Throws a d6 and returns the result",
		},
		{
			name:     "single line with tab indent",
			indent:   "\t",
			input:    " Throws a d6",
			expected: "\t// Throws a d6",
		},
		{
			name:     "multi line preserves newlines",
			indent:   "",
			input:    " First line\n Second line\n Third line",
			expected: "// First line\n// Second line\n// Third line",
		},
		{
			name:     "blank proto lines become bare //",
			indent:   "",
			input:    " Para one\n\n Para two",
			expected: "// Para one\n//\n// Para two",
		},
		{
			name:   "code fence preserved",
			indent: "",
			input: " Description\n" +
				" ```golang\n" +
				" func main() {}\n" +
				" ```",
			expected: "// Description\n" +
				"// ```golang\n" +
				"// func main() {}\n" +
				"// ```",
		},
		{
			name:     "strips trailing whitespace per line",
			indent:   "",
			input:    " trailing spaces   \n another  ",
			expected: "// trailing spaces\n// another",
		},
		{
			name:     "CRLF is normalised",
			indent:   "",
			input:    " line1\r\n line2",
			expected: "// line1\n// line2",
		},
		{
			name:     "no double space when leading space already stripped",
			indent:   "",
			input:    "no leading space",
			expected: "// no leading space",
		},
		{
			name:     "surrounding blank lines trimmed",
			indent:   "",
			input:    "\n\n real\n\n",
			expected: "// real",
		},
		{
			name:     "indent applied to every line",
			indent:   "    ",
			input:    " one\n\n two",
			expected: "    // one\n    //\n    // two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commentBlock(tt.indent, tt.input)
			if result != tt.expected {
				t.Errorf("commentBlock(%q, %q) =\n%q\nexpected\n%q",
					tt.indent, tt.input, result, tt.expected)
			}
		})
	}
}

// TestDocComment Tests Go-convention doc comment rendering with a name prefix
func TestDocComment(t *testing.T) {
	tests := []struct {
		name     string
		indent   string
		ident    string
		body     string
		expected string
	}{
		{
			name:     "name only, no body",
			ident:    "ThrowDies",
			expected: "// ThrowDies",
		},
		{
			name:     "name only, whitespace body",
			ident:    "ThrowDies",
			body:     "   \n\n",
			expected: "// ThrowDies",
		},
		{
			name:     "name with single line body",
			ident:    "ThrowDies",
			body:     " Throws dies a few times",
			expected: "// ThrowDies Throws dies a few times",
		},
		{
			name:     "name with multi-line body",
			ident:    "Ping",
			body:     " Just a simple ping\n Takes no parameters\n returns nothing",
			expected: "// Ping Just a simple ping\n// Takes no parameters\n// returns nothing",
		},
		{
			name:   "name with body containing blank line and code fence",
			ident:  "DieRoll",
			indent: "",
			body: " Description\n" +
				"\n" +
				" ```golang\n" +
				" foo()\n" +
				" ```",
			expected: "// DieRoll Description\n" +
				"//\n" +
				"// ```golang\n" +
				"// foo()\n" +
				"// ```",
		},
		{
			name:     "tab indent applied to every line",
			indent:   "\t",
			ident:    "ThrowDies",
			body:     " first\n second",
			expected: "\t// ThrowDies first\n\t// second",
		},
		{
			name:     "empty name emits plain block",
			ident:    "",
			body:     " just text\n and more",
			expected: "// just text\n// and more",
		},
		{
			name:     "empty name and empty body returns empty",
			ident:    "",
			body:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := docComment(tt.indent, tt.ident, tt.body)
			if result != tt.expected {
				t.Errorf("docComment(%q, %q, %q) =\n%q\nexpected\n%q",
					tt.indent, tt.ident, tt.body, result, tt.expected)
			}
		})
	}
}

// TestKebab Verifies snake, camel, Pascal, acronym, digit, and edge-case
// inputs lower-kebab correctly.
func TestKebab(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"snake_case", "order_id", "order-id"},
		{"camelCase", "orderID", "order-id"},
		{"PascalCase", "OrderID", "order-id"},
		{"single word lower", "order", "order"},
		{"single word upper", "Order", "order"},
		{"acronym then word", "HTTPSServer", "https-server"},
		{"all caps acronym", "HTTP", "http"},
		{"digits cluster with letters", "order123Id", "order123-id"},
		{"mixed snake and camel", "shipping_addressLine", "shipping-address-line"},
		{"nested snake", "a_b_c", "a-b-c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kebab(tt.input); got != tt.expected {
				t.Errorf("kebab(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestPflagFuncName Covers every FlagKind constant in the model package
// and asserts the panic path for an out-of-range FlagKind.
func TestPflagFuncName(t *testing.T) {
	tests := []struct {
		kind     model.FlagKind
		expected string
	}{
		{model.FlagKindString, "StringVar"},
		{model.FlagKindStringSlice, "StringArrayVar"},
		{model.FlagKindInt32, "Int32Var"},
		{model.FlagKindInt32Slice, "Int32SliceVar"},
		{model.FlagKindInt64, "Int64Var"},
		{model.FlagKindInt64Slice, "Int64SliceVar"},
		{model.FlagKindUint32, "Uint32Var"},
		{model.FlagKindUint64, "Uint64Var"},
		{model.FlagKindFloat32, "Float32Var"},
		{model.FlagKindFloat64, "Float64Var"},
		{model.FlagKindBool, "BoolVar"},
		{model.FlagKindBytes, "BytesBase64Var"},
		{model.FlagKindDuration, "DurationVar"},
		{model.FlagKindTimestamp, "Var"},
		{model.FlagKindEnum, "Var"},
		{model.FlagKindFieldMask, "StringSliceVar"},
		{model.FlagKindStringToString, "StringToStringVar"},
		{model.FlagKindStringToInt, "StringToIntVar"},
		{model.FlagKindJSONEscape, "Var"},
	}
	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			if got := pflagFuncName(tt.kind); got != tt.expected {
				t.Errorf("pflagFuncName(%v) = %q, want %q", tt.kind, got, tt.expected)
			}
		})
	}

	t.Run("panics on unknown kind", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("pflagFuncName(unknown) did not panic")
			}
		}()
		_ = pflagFuncName(model.FlagKind(999))
	})
}

// TestGoFieldAssignPath Verifies the single-expression getter-chain form
// for nested proto field paths.
func TestGoFieldAssignPath(t *testing.T) {
	tests := []struct {
		name     string
		path     []string
		rootVar  string
		expected string
	}{
		{"empty path returns root", nil, "req", "req"},
		{"single segment", []string{"customer_id"}, "req", "req.CustomerId"},
		{"two segments", []string{"shipping", "street"}, "req", "req.GetShipping().Street"},
		{"three segments", []string{"shipping", "address", "street"}, "req", "req.GetShipping().GetAddress().Street"},
		{"underscore in leaf", []string{"order_id"}, "r", "r.OrderId"},
		{"multi-word segments", []string{"shipping_address", "postal_code"}, "req", "req.GetShippingAddress().PostalCode"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := goFieldAssignPath(tt.path, tt.rootVar); got != tt.expected {
				t.Errorf("goFieldAssignPath(%v, %q) = %q, want %q", tt.path, tt.rootVar, got, tt.expected)
			}
		})
	}
}

// TestGoFieldAssignSteps Verifies the step-wise form the template uses
// to allocate intermediate messages before assigning the leaf.
func TestGoFieldAssignSteps(t *testing.T) {
	tests := []struct {
		name     string
		path     []string
		rootVar  string
		expected []AssignStep
	}{
		{"nil path yields nil", nil, "req", nil},
		{
			"single segment leaf",
			[]string{"customer_id"},
			"req",
			[]AssignStep{{Expr: "req", FieldName: "CustomerId", IsLeaf: true}},
		},
		{
			"two segments",
			[]string{"shipping", "street"},
			"req",
			[]AssignStep{
				{Expr: "req", FieldName: "Shipping", IsLeaf: false},
				{Expr: "req.Shipping", FieldName: "Street", IsLeaf: true},
			},
		},
		{
			"three segments",
			[]string{"shipping", "address", "street"},
			"req",
			[]AssignStep{
				{Expr: "req", FieldName: "Shipping", IsLeaf: false},
				{Expr: "req.Shipping", FieldName: "Address", IsLeaf: false},
				{Expr: "req.Shipping.Address", FieldName: "Street", IsLeaf: true},
			},
		},
		{
			"underscore in intermediate",
			[]string{"shipping_address", "postal_code"},
			"r",
			[]AssignStep{
				{Expr: "r", FieldName: "ShippingAddress", IsLeaf: false},
				{Expr: "r.ShippingAddress", FieldName: "PostalCode", IsLeaf: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goFieldAssignSteps(tt.path, tt.rootVar)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("goFieldAssignSteps(%v, %q) = %+v, want %+v", tt.path, tt.rootVar, got, tt.expected)
			}
		})
	}
}

// TestEnumValueTypeName Verifies the generated pflag.Value wrapper name
// for enums is the expected `enum<GoName>Value` identifier.
func TestEnumValueTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"OrderStatus", "enumOrderStatusValue"},
		{"Color", "enumColorValue"},
		{"HTTPMethod", "enumHTTPMethodValue"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := enumValueTypeName(tt.input); got != tt.expected {
				t.Errorf("enumValueTypeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestJsonExampleFor Covers each JSON-escape sub-case from PLAN.md §6.1
// plus the no-op non-JSONEscape case.
func TestJsonExampleFor(t *testing.T) {
	tests := []struct {
		name     string
		spec     *model.FlagSpec
		expected string
	}{
		{
			name:     "non-JSONEscape kind returns empty",
			spec:     &model.FlagSpec{Kind: model.FlagKindString},
			expected: "",
		},
		{
			name:     "nil spec returns empty",
			spec:     nil,
			expected: "",
		},
		{
			name:     "repeated message",
			spec:     &model.FlagSpec{Kind: model.FlagKindJSONEscape, Repeated: true, JSONGoType: "example.v1.Item"},
			expected: `'[{"field":"value"}]'`,
		},
		{
			name:     "Any / Struct / Value WKT",
			spec:     &model.FlagSpec{Kind: model.FlagKindJSONEscape, JSONGoType: "google.protobuf.Struct"},
			expected: `'<JSON value>'`,
		},
		{
			name:     "cyclic subtree names the target type",
			spec:     &model.FlagSpec{Kind: model.FlagKindJSONEscape, JSONGoType: "example.v1.Tree"},
			expected: `'<JSON value of type Tree>'`,
		},
		{
			name:     "map of message (seam falls through to typed JSON value)",
			spec:     &model.FlagSpec{Kind: model.FlagKindJSONEscape, JSONGoType: "example.v1.Address"},
			expected: `'<JSON value of type Address>'`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonExampleFor(tt.spec); got != tt.expected {
				t.Errorf("jsonExampleFor(%+v) = %q, want %q", tt.spec, got, tt.expected)
			}
		})
	}
}
