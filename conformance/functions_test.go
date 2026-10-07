package conformance

import (
	"errors"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
)

// TestConformanceFunctions runs functions.yaml's evaluate_function cases:
// a basic function called with the case's arguments as they are, in a
// context over the case's data model. A ValidationResult compares with a
// boolean by its validity, as Dart's harness compares them.
func TestConformanceFunctions(t *testing.T) {
	for _, c := range suite(t, "functions.yaml") {
		if c["action"] != "evaluate_function" || !v1(c) {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			data := a2ui.NewDataModel(nil)
			if dm, ok := c["dataModel"].(map[string]any); ok {
				for k, v := range dm {
					if err := data.Set("/"+k, v); err != nil {
						t.Fatal(err)
					}
				}
			}
			ctx := &a2ui.Context{Data: data, Scope: a2ui.RootScope, OpenURL: func(string) error { return nil }}
			impl := basic.Functions[str(c["function"])]
			if impl == nil {
				t.Fatalf("no function %v", c["function"])
			}
			args, _ := c["args"].(map[string]any)
			if args == nil {
				args = map[string]any{}
			}
			got, err := impl(ctx, args)
			if want := expectedError(c); want != nil {
				expectCategory(t, err, want)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := c["expect"]
			if _, ok := want.(bool); ok {
				if v, isResult := got.(a2ui.ValidationResult); isResult {
					got = v.Valid
				}
			}
			if !jsonEqual(t, got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}

// v1 reports whether a case applies to A2UI v1.0: it names that version,
// or none.
func v1(c map[string]any) bool {
	v := str(c["protocolVersion"])
	if cat, ok := c["catalog"].(map[string]any); ok && v == "" {
		v = str(cat["protocolVersion"])
	}
	return v == "" || strings.TrimPrefix(v, "v") == "1.0"
}

func expectedError(c map[string]any) map[string]any {
	for _, k := range []string{"expectError", "expect_error"} {
		if m, ok := c[k].(map[string]any); ok {
			return m
		}
	}
	return nil
}

// expectCategory checks err against an expected error's category and
// message.
func expectCategory(t *testing.T, err error, want map[string]any) {
	t.Helper()
	expectError(t, err, want)
	if err == nil {
		return
	}
	switch want["category"] {
	case "ExpressionError", "ParseError":
		var e *a2ui.ExpressionError
		if !errors.As(err, &e) {
			t.Errorf("error %T %v, want an ExpressionError", err, err)
		}
	case "DataError":
		if !a2ui.IsDataError(err) {
			t.Errorf("error %T %v, want a DataError", err, err)
		}
	}
}
