package conformance

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// TestConformanceExpressionTemplates runs expressions.yaml's
// parse_expression_template cases. Parts are compared with adjacent
// literals joined, as the suite asks; bindings and calls in the suite's
// vocabulary ({path}, {call, args, returnType}).
func TestConformanceExpressionTemplates(t *testing.T) {
	for _, c := range suite(t, "expressions.yaml") {
		if c["action"] != "parse_expression_template" {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			parts, err := a2ui.ParseTemplate(str(c["input"]))
			if want, ok := c["expect_error"].(map[string]any); ok {
				expectError(t, err, want)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := joinLiterals(suiteParts(parts))
			want := joinLiterals(c["expect"].([]any))
			if !jsonEqual(t, got, want) {
				t.Errorf("got %v, want %v", asJSON(t, got), want)
			}
		})
	}
}

func suiteParts(parts []any) []any {
	out := make([]any, len(parts))
	for i, p := range parts {
		out[i] = suitePart(p)
	}
	return out
}

func suitePart(p any) any {
	switch v := p.(type) {
	case a2ui.Binding:
		return map[string]any{"path": v.Path}
	case a2ui.Call:
		args := map[string]any{}
		for k, a := range v.Args {
			args[k] = suitePart(a)
		}
		return map[string]any{"call": v.Name, "args": args, "returnType": "any"}
	}
	return p
}

func joinLiterals(parts []any) []any {
	var out []any
	for _, p := range parts {
		s, ok := p.(string)
		if !ok {
			out = append(out, p)
			continue
		}
		if s == "" {
			continue
		}
		if n := len(out); n > 0 {
			if prev, ok := out[n-1].(string); ok {
				out[n-1] = prev + s
				continue
			}
		}
		out = append(out, s)
	}
	return out
}
