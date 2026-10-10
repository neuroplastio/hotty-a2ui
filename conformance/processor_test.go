package conformance

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/a2ui/schema"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
)

// The suites whose cases feed messages to a message processor, and the
// cases this v1.0 core does not run: v0.9.1's version string.
var (
	processorSuites = []string{
		"message_processor_v1_0.yaml", "reserved_keys.yaml", "data_model_pointers.yaml",
		"data_deletion.yaml", "expressions.yaml", "functions.yaml", "index_function.yaml",
		"validation_result.yaml", "multi_catalog.yaml", "validator_v1_0.yaml", "composition_constraints.yaml",
	}
	processorSkips = map[string]string{
		"test_v091_version_string_compatibility": "v0.9.1 messages; this core speaks v1.0",
		"test_validator_theme_schema":            "v0.9 messages: a v1.0 createSurface has no theme",
	}
)

// TestConformanceProcessor runs the process_messages, validate and
// get_renderer_capabilities cases (process_messages.md and validate.md
// in the brief have the contract). process_messages applies the messages
// as they come; validate does so on a strict processor, then resolves
// every surface, and an error the surfaces report fails the step as a
// ValidationError.
func TestConformanceProcessor(t *testing.T) {
	for _, name := range processorSuites {
		for _, c := range suite(t, name) {
			action := str(c["action"])
			if action != "process_messages" && action != "validate" && action != "get_renderer_capabilities" {
				continue
			}
			t.Run(strings.TrimSuffix(name, ".yaml")+"/"+str(c["name"]), func(t *testing.T) {
				if why, ok := processorSkips[str(c["name"])]; ok {
					t.Skip(why)
				}
				if v := caseVersion(c); v != "" && v != a2ui.Version {
					t.Skipf("protocol version %s", v)
				}
				p := schema.NewProcessor(catalogsFor(t, c)...)
				if action == "get_renderer_capabilities" {
					args, _ := c["args"].(map[string]any)
					if got := p.Capabilities(str(args["version"])); !jsonEqual(t, got, c["expect"]) {
						t.Errorf("capabilities %v, want %v", asJSON(t, got), c["expect"])
					}
					return
				}
				p.Strict = action == "validate" || c["strictMode"] == true
				var reported []*a2ui.ErrorMessage
				p.Send = func(o a2ui.Outbound) {
					if o.Error != nil {
						reported = append(reported, o.Error)
					}
				}
				steps := stepsOf(c)
				for i, st := range steps {
					last := i == len(steps)-1
					expErr, exp := errorOf(st["expectError"]), asMap(st["expect"])
					if last && expErr == nil && exp == nil {
						expErr, exp = errorOf(c["expectError"]), asMap(c["expect"])
					}
					payload := st["messages"]
					if payload == nil {
						payload = st["payload"]
					}
					reported = nil
					err := p.Process(payload)
					if action == "validate" && err == nil {
						for _, s := range p.Surfaces() {
							s.Tree.Resolve()
						}
						if len(reported) > 0 {
							err = &a2ui.ValidationError{Msg: reported[0].Message}
						}
					}
					if expErr != nil {
						checkError(t, fmt.Sprintf("step %d", i), err, expErr)
						continue
					}
					if err != nil {
						t.Fatalf("step %d: %v", i, err)
					}
					if exp != nil {
						checkSurfaces(t, fmt.Sprintf("step %d", i), p, exp)
					}
				}
			})
		}
	}
}

func asMap(v any) map[string]any { m, _ := v.(map[string]any); return m }

// errorOf reads an expectError: {category, message}, or the message
// alone.
func errorOf(v any) map[string]any {
	if s, ok := v.(string); ok {
		return map[string]any{"message": s}
	}
	return asMap(v)
}

// expectsOtherError reports whether any step of a case expects an error
// that is not a CatalogError.
func expectsOtherError(c map[string]any) bool {
	if e := errorOf(c["expectError"]); e != nil && e["category"] != "CatalogError" {
		return true
	}
	steps, _ := c["steps"].([]any)
	for _, s := range steps {
		if e := errorOf(asMap(s)["expectError"]); e != nil && e["category"] != "CatalogError" {
			return true
		}
	}
	return false
}

// caseVersion is the protocol version a case declares, as the Python
// runner reads it: its own, its catalog's, its first catalog's.
func caseVersion(c map[string]any) string {
	if v := str(c["protocolVersion"]); v != "" {
		return v
	}
	if cat, ok := c["catalog"].(map[string]any); ok && str(cat["protocolVersion"]) != "" {
		return str(cat["protocolVersion"])
	}
	if cats, ok := c["catalogs"].([]any); ok && len(cats) > 0 {
		if m, ok := cats[0].(map[string]any); ok {
			return str(m["protocolVersion"])
		}
	}
	if args, ok := c["args"].(map[string]any); ok {
		return str(args["version"])
	}
	return ""
}

func stepsOf(c map[string]any) []map[string]any {
	if l, ok := c["steps"].([]any); ok && c["messages"] == nil {
		out := make([]map[string]any, len(l))
		for i, s := range l {
			out[i], _ = s.(map[string]any)
		}
		return out
	}
	return []map[string]any{{"messages": c["messages"], "payload": c["payload"]}}
}

// catalogsFor registers a case's catalogs in the reference runners'
// order: its inline catalog, its catalogs (open stand-ins when they have
// no components), its catalog files, the built-in "basic", and for any
// other catalog a createSurface names, an open v1.0 stand-in, or, when
// the case expects an error other than a CatalogError, a copy of the
// first of the case's own catalogs, so that its schemas apply.
func catalogsFor(t *testing.T, c map[string]any) []*a2ui.Catalog {
	t.Helper()
	var out []*a2ui.Catalog
	var first func(id string) *a2ui.Catalog
	have := map[string]bool{}
	add := func(make func(id string) *a2ui.Catalog, id string) {
		if have[id] {
			return
		}
		have[id] = true
		out = append(out, make(id))
		if first == nil {
			first = make
		}
	}
	parse := func(doc map[string]any, version string) func(string) *a2ui.Catalog {
		return func(id string) *a2ui.Catalog {
			d := asJSON(t, doc).(map[string]any)
			d["catalogId"] = id
			if str(d["protocolVersion"]) == "" {
				d["protocolVersion"] = version
			}
			cat, err := a2ui.ParseCatalog(mustJSON(t, d))
			if err != nil {
				t.Fatal(err)
			}
			return cat
		}
	}
	version := caseVersion(c)
	if version == "" {
		version = a2ui.Version
	}
	if cat, ok := c["catalog"].(map[string]any); ok {
		doc := cat
		if cs, ok := cat["catalogSchema"].(map[string]any); ok {
			doc = cs
		}
		if doc["components"] != nil {
			id := str(doc["catalogId"])
			if id == "" {
				id = "custom"
			}
			add(parse(doc, fallback(str(cat["protocolVersion"]), version)), id)
		}
	}
	cats, _ := c["catalogs"].([]any)
	for _, e := range cats {
		m, _ := e.(map[string]any)
		v := str(m["protocolVersion"])
		if m["components"] != nil || m["functions"] != nil {
			add(parse(m, v), str(m["catalogId"]))
		} else {
			add(func(id string) *a2ui.Catalog { return a2ui.NewOpenCatalog(id, v) }, str(m["catalogId"]))
		}
	}
	if c["action"] == "get_renderer_capabilities" {
		return out
	}
	basicAs := func(id string) *a2ui.Catalog {
		b := basic.Catalog()
		b.ID = id
		return b
	}
	paths, _ := c["catalogPaths"].([]any)
	for _, p := range paths {
		if regexp.MustCompile(`catalogs/basic/v1/catalog\.json$`).MatchString(str(p)) {
			add(basicAs, basic.ID)
		}
	}
	specified := first
	add(basicAs, "basic")
	standIn := func(id string) *a2ui.Catalog { return a2ui.NewOpenCatalog(id, a2ui.Version) }
	if specified != nil && expectsOtherError(c) {
		standIn = specified
	}
	var walk func(v any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if cs, ok := x["createSurface"].(map[string]any); ok {
				if id := str(cs["catalogId"]); id != "" {
					add(standIn, id)
				}
			}
			for _, e := range x {
				walk(e)
			}
		case []any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	walk(c["messages"])
	walk(c["payload"])
	walk(c["steps"])
	return out
}

func fallback(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

// checkError checks an error against expectError: its category, A2UI's
// error classes (an IntegrityError or a RecursionError is a
// ValidationError), its wire code, and its message, a substring or a
// regular expression.
func checkError(t *testing.T, step string, err error, want map[string]any) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: no error; want %v", step, want)
		return
	}
	if cat := str(want["category"]); cat != "" && !isCategory(err, cat) {
		t.Errorf("%s: %T %v is not a %s", step, err, err, cat)
	}
	if code := str(want["code"]); code != "" && a2ui.Code(err) != code {
		t.Errorf("%s: error code %q, want %q", step, a2ui.Code(err), code)
	}
	if m := str(want["message"]); m != "" && !messageMatches(err.Error(), m) {
		t.Errorf("%s: error %q does not say %q", step, err, m)
	}
}

func messageMatches(got, want string) bool {
	if strings.Contains(got, want) {
		return true
	}
	re, err := regexp.Compile(want)
	return err == nil && re.MatchString(got)
}

func isCategory(err error, cat string) bool {
	var (
		v *a2ui.ValidationError
		i *a2ui.IntegrityError
		r *a2ui.RecursionError
		c *a2ui.CatalogError
		d *a2ui.DataError
		e *a2ui.ExpressionError
	)
	switch cat {
	case "ValidationError":
		return errors.As(err, &v) || errors.As(err, &i) || errors.As(err, &r)
	case "IntegrityError":
		return errors.As(err, &i) || errors.As(err, &r)
	case "RecursionError":
		return errors.As(err, &r)
	case "CatalogError":
		return errors.As(err, &c)
	case "DataError":
		return errors.As(err, &d)
	case "ExpressionError", "ParseError":
		return errors.As(err, &e)
	}
	return true
}

// checkSurfaces checks expect.surfaces: each listed surface exists (or
// not, with exists: false), and its listed state.
func checkSurfaces(t *testing.T, step string, p *a2ui.Processor, exp map[string]any) {
	t.Helper()
	surfaces, _ := exp["surfaces"].(map[string]any)
	for id, e := range surfaces {
		want, _ := e.(map[string]any)
		s := p.Surface(id)
		if want["exists"] == false {
			if s != nil {
				t.Errorf("%s: surface %s exists", step, id)
			}
			continue
		}
		if s == nil {
			t.Errorf("%s: no surface %s", step, id)
			continue
		}
		if v, ok := want["sendDataModel"]; ok && v != s.SendDataModel {
			t.Errorf("%s: %s sendDataModel %v, want %v", step, id, s.SendDataModel, v)
		}
		if v, ok := want["dataModel"]; ok && !jsonEqual(t, s.Data.Root(), v) {
			t.Errorf("%s: %s dataModel %v, want %v", step, id, asJSON(t, s.Data.Root()), v)
		}
		nodes := nodesByName(s)
		switch comps := want["components"].(type) {
		case []any:
			var ids []string
			for _, e := range comps {
				m, _ := e.(map[string]any)
				ids = append(ids, str(m["id"]))
				checkComponent(t, step, s, nodes, str(m["id"]), m)
			}
			if got := s.ComponentIDs(); !jsonEqual(t, sorted(got), sorted(ids)) {
				t.Errorf("%s: %s components %v, want %v", step, id, got, ids)
			}
		case map[string]any:
			for cid, e := range comps {
				m, _ := e.(map[string]any)
				checkComponent(t, step, s, nodes, cid, m)
			}
		}
		if vr, ok := want["validationResult"].(map[string]any); ok {
			for cid, w := range vr {
				n := nodes[cid]
				if n == nil {
					t.Errorf("%s: no node %s", step, cid)
					continue
				}
				got := map[string]any{"valid": true}
				checks, _ := n.Props["checks"].([]a2ui.ValidationResult)
				for _, r := range checks {
					if !r.Valid {
						got = asJSON(t, r).(map[string]any)
						break
					}
				}
				if !jsonEqual(t, got, w) {
					t.Errorf("%s: %s validationResult %v, want %v", step, cid, got, w)
				}
			}
		}
	}
}

func sorted(s []string) []string {
	out := append([]string(nil), s...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// nodesByName indexes the resolved tree as the TypeScript runner does:
// by component id (the last instance wins), by key, and, for a node in a
// template, by id_index ("item_text_1").
func nodesByName(s *a2ui.Surface) map[string]*a2ui.Node {
	out := map[string]*a2ui.Node{}
	for _, n := range s.Tree.Nodes() {
		out[n.ComponentID] = n
		out[n.Key] = n
		if p, err := a2ui.ParsePointer(n.Scope.Path); err == nil && len(p) > 0 {
			if seg := p[len(p)-1]; regexp.MustCompile(`^[0-9]+$`).MatchString(seg) {
				out[n.ComponentID+"_"+seg] = n
			}
		}
	}
	return out
}

// checkComponent compares a component's listed properties, resolved, with
// what the case expects: structured values as JSON, scalars as
// JavaScript's String() writes them.
func checkComponent(t *testing.T, step string, s *a2ui.Surface, nodes map[string]*a2ui.Node, id string, want map[string]any) {
	t.Helper()
	var typ string
	var props map[string]any
	if n := nodes[id]; n != nil {
		typ, props = n.Type, n.Props
	} else if c := s.Components[id]; c != nil {
		typ, props = c.Type, c.Props
	} else {
		t.Errorf("%s: no component %s", step, id)
		return
	}
	for k, w := range want {
		switch k {
		case "id":
			continue
		case "component":
			if typ != w {
				t.Errorf("%s: %s is a %s, want %v", step, id, typ, w)
			}
			continue
		}
		got := plain(props[k])
		_, wObj := w.(map[string]any)
		_, wList := w.([]any)
		_, gObj := got.(map[string]any)
		_, gList := got.([]any)
		if wObj || wList || gObj || gList {
			if !jsonEqual(t, got, w) {
				t.Errorf("%s: %s.%s = %v, want %v", step, id, k, asJSON(t, got), w)
			}
		} else if jsString(got) != jsString(w) {
			t.Errorf("%s: %s.%s = %s, want %s", step, id, k, jsString(got), jsString(w))
		}
	}
}

// plain is a resolved property as a JSON value: a binding's value, a
// child's id, an action's definition.
func plain(v any) any {
	switch x := v.(type) {
	case a2ui.Bound:
		return plain(x.Value)
	case *a2ui.Node:
		return x.ComponentID
	case []*a2ui.Node:
		out := make([]any, len(x))
		for i, n := range x {
			out[i] = n.ComponentID
		}
		return out
	case a2ui.ActionRef:
		return x.Raw
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = plain(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = plain(e)
		}
		return out
	}
	return v
}

func jsString(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case float64:
		return a2ui.NumberString(x)
	case int:
		return a2ui.NumberString(float64(x))
	}
	return fmt.Sprint(v)
}
