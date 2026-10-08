package a2ui

import "testing"

// TestCompositionErrorCodes checks that a component placed under an
// unallowed parent, and an unallowed child inside a container, reach the
// agent with A2UI's own wire codes ("UNALLOWED_PARENT", "UNALLOWED_CHILD")
// rather than the generic "VALIDATION_FAILED" a ValidationError defaults
// to (a2ui_protocol.md, Composition validation rules).
func TestCompositionErrorCodes(t *testing.T) {
	cases := []struct {
		name     string
		catalog  string
		root     map[string]any
		child    map[string]any
		wantCode string
	}{
		{
			name: "unallowed parent",
			catalog: `{
				"catalogId": "custom",
				"protocolVersion": "1.0",
				"components": {
					"ColumnHeader": {"type": "object", "allowedParents": ["Column"], "properties": {"component": {"const": "ColumnHeader"}}},
					"Row": {"type": "object", "properties": {"component": {"const": "Row"}, "children": {"$ref": "common_types.json#/$defs/ChildList"}}}
				}
			}`,
			root:     map[string]any{"id": "root", "component": "Row", "children": []any{"invalid_header"}},
			child:    map[string]any{"id": "invalid_header", "component": "ColumnHeader"},
			wantCode: "UNALLOWED_PARENT",
		},
		{
			name: "unallowed child",
			catalog: `{
				"catalogId": "custom",
				"protocolVersion": "1.0",
				"components": {
					"RestrictedBox": {"type": "object", "allowedChildren": ["Text", "Button"], "properties": {"component": {"const": "RestrictedBox"}, "children": {"$ref": "common_types.json#/$defs/ChildList"}}},
					"Video": {"type": "object", "properties": {"component": {"const": "Video"}}}
				}
			}`,
			root:     map[string]any{"id": "root", "component": "RestrictedBox", "children": []any{"v1"}},
			child:    map[string]any{"id": "v1", "component": "Video"},
			wantCode: "UNALLOWED_CHILD",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cat, err := ParseCatalog([]byte(c.catalog))
			if err != nil {
				t.Fatal(err)
			}
			p := NewProcessor(cat)
			var got []Outbound
			p.Send = func(o Outbound) { got = append(got, o) }
			err = p.Process([]any{map[string]any{"version": Version, "createSurface": map[string]any{
				"surfaceId": "s1", "catalogId": "custom", "components": []any{c.root, c.child},
			}}})
			if err == nil {
				t.Fatal("no error")
			}
			if code := Code(err); code != c.wantCode {
				t.Errorf("Code(err) = %q, want %q", code, c.wantCode)
			}
			if len(got) == 0 || got[len(got)-1].Error == nil {
				t.Fatalf("no error message sent: %v", got)
			}
			if code := got[len(got)-1].Error.Code; code != c.wantCode {
				t.Errorf("sent error code %q, want %q", code, c.wantCode)
			}
		})
	}
}
