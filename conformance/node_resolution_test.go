package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"go.yaml.in/yaml/v3"
)

// TestConformanceNodeResolution runs node_resolution.yaml
// (node_resolution.md has the contract). The fixtures are written in
// A2UI v0.9.1's vocabulary ({path}, {call}) against the v0.9.1 basic
// catalog; this core is v1.0, so they are translated to {"@path"} and
// {"@call"} and run against the v1.0 basic catalog, whose components the
// cases use the same way. The catalog's functions record their calls and
// return nil, as the contract has them.
func TestConformanceNodeResolution(t *testing.T) {
	for _, c := range suite(t, "node_resolution.yaml") {
		if c["action"] != "resolve_nodes" {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			fx := fixture(t, str(c["fixture"]))
			var calls []map[string]any
			cat := basic.Catalog()
			for name, f := range cat.Functions {
				name := name
				f.Impl = func(_ *a2ui.Context, args map[string]any) (any, error) {
					calls = append(calls, map[string]any{"name": name, "args": args})
					return nil, nil
				}
			}
			var events []map[string]any
			env := a2ui.Env{Action: func(_ *a2ui.Surface, a *a2ui.ActionMessage) {
				events = append(events, map[string]any{"name": a.Name, "source_component_id": a.SourceComponentID, "context": a.Context})
			}}
			s := a2ui.NewSurface("s", basic.ID, a2ui.CatalogSet{basic.ID: cat}, env)
			if data, ok := fx["data"].(map[string]any); ok {
				if err := s.Data.Set("/", data); err != nil {
					t.Fatal(err)
				}
			}
			s.Update(defs(fx["components"]))
			emitted := map[*a2ui.Node]int{}
			destroyed := map[*a2ui.Node]int{}
			s.Tree.OnEmit = func(n *a2ui.Node) { emitted[n]++ }
			s.Tree.OnDestroy = func(n *a2ui.Node) { destroyed[n]++ }
			s.Tree.Resolve()

			before := positions(s.Tree)
			if len(calls)+len(events) > 0 {
				t.Errorf("resolution ran actions %v or functions %v", events, calls)
			}
			checkNodes(t, "initial", before, c["expect"].(map[string]any))

			steps, _ := c["steps"].([]any)
			for i, st := range steps {
				step := st.(map[string]any)
				name := "step " + strconv.Itoa(i) + " " + str(step["op"])
				emitted, destroyed = map[*a2ui.Node]int{}, map[*a2ui.Node]int{}
				calls, events = nil, nil
				last := map[*a2ui.Node]string{}
				for pos, n := range before {
					last[n] = pos
				}
				switch step["op"] {
				case "set_data":
					if err := s.Write(str(step["path"]), step["value"]); err != nil {
						t.Fatalf("%s: %v", name, err)
					}
				case "update_components":
					s.Update(defs(step["components"]))
					s.Tree.Resolve()
				case "remove_component":
					s.Remove(str(step["component_id"]))
					s.Tree.Resolve()
				case "write":
					n := before[str(step["node"])]
					if err := s.Tree.Write(n, str(step["property"]), step["value"]); err != nil {
						t.Fatalf("%s: %v", name, err)
					}
				case "invoke":
					n := before[str(step["node"])]
					if err := s.Tree.Invoke(n, str(step["property"]), true); err != nil {
						t.Fatalf("%s: %v", name, err)
					}
				case "dispose":
					s.Tree.Dispose()
				default:
					t.Fatalf("%s: unknown op", name)
				}
				after := positions(s.Tree)
				exp := step["expect"].(map[string]any)
				checkNodes(t, name, after, exp)

				where := map[*a2ui.Node]string{}
				for pos, n := range after {
					where[n] = pos
				}
				gotEmit := map[string]any{}
				for n, k := range emitted {
					gotEmit[where[n]] = k
				}
				if want, _ := exp["emissions"].(map[string]any); !jsonEqual(t, gotEmit, orEmpty(want)) {
					t.Errorf("%s: emissions %v, want %v", name, gotEmit, want)
				}
				gotDestroyed := map[string]any{}
				for n, k := range destroyed {
					gotDestroyed[last[n]] = k
				}
				if want, _ := exp["destroyed"].(map[string]any); !jsonEqual(t, gotDestroyed, orEmpty(want)) {
					t.Errorf("%s: destroyed %v, want %v", name, gotDestroyed, want)
				}
				for _, p := range list(exp["same_nodes"]) {
					if before[p] == nil || before[p] != after[p] {
						t.Errorf("%s: %s is not the same node", name, p)
					}
				}
				for _, p := range list(exp["replaced_nodes"]) {
					if before[p] == nil || after[p] == nil || before[p] == after[p] {
						t.Errorf("%s: %s was not replaced", name, p)
					}
				}
				if want, ok := exp["data"].(map[string]any); ok {
					for p, v := range want {
						if got := s.Data.Value(p); !jsonEqual(t, got, v) {
							t.Errorf("%s: data %s = %v, want %v", name, p, got, v)
						}
					}
				}
				wantEvents, _ := exp["events"].([]any)
				if len(events)+len(wantEvents) > 0 && !jsonEqual(t, events, wantEvents) {
					t.Errorf("%s: events %v, want %v", name, events, wantEvents)
				}
				wantCalls, _ := exp["functions"].([]any)
				if len(calls)+len(wantCalls) > 0 && !jsonEqual(t, calls, wantCalls) {
					t.Errorf("%s: functions %v, want %v", name, calls, wantCalls)
				}
				before = after
			}
		})
	}
}

func orEmpty(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

func list(v any) []string {
	l, _ := v.([]any)
	out := make([]string, len(l))
	for i, x := range l {
		out[i] = str(x)
	}
	return out
}

// fixture reads a node fixture, translated to v1.0's vocabulary.
func fixture(t *testing.T, name string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(conformanceDir, name))
	if err != nil {
		t.Fatal(err)
	}
	var raw any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	return asJSON(t, raw).(map[string]any)
}

// defs reads component definitions, translated from v0.9.1's vocabulary.
func defs(v any) []map[string]any {
	l, _ := v.([]any)
	out := make([]map[string]any, 0, len(l))
	for _, d := range l {
		m, _ := v09to10(d).(map[string]any)
		out = append(out, m)
	}
	return out
}

// v09to10 rewrites v0.9's {path} bindings and {call, args} calls as v1.0's
// {"@path"} and {"@call"}. A template ({componentId, path}) stays.
func v09to10(v any) any {
	switch x := v.(type) {
	case map[string]any:
		if p, ok := x["path"].(string); ok && len(x) == 1 {
			return map[string]any{"@path": p}
		}
		out := map[string]any{}
		for k, e := range x {
			out[k] = v09to10(e)
		}
		if c, ok := x["call"].(string); ok {
			delete(out, "call")
			delete(out, "returnType")
			out["@call"] = c
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = v09to10(e)
		}
		return out
	}
	return v
}

// positions addresses every mounted node by its position: "root", then
// each child property's name and, in a list, the index.
func positions(r *a2ui.Resolver) map[string]*a2ui.Node {
	out := map[string]*a2ui.Node{}
	var walk func(pos string, n *a2ui.Node)
	walk = func(pos string, n *a2ui.Node) {
		out[pos] = n
		for _, k := range a2ui.SortedKeys(n.Props) {
			switch x := n.Props[k].(type) {
			case *a2ui.Node:
				walk(pos+"/"+k, x)
			case []*a2ui.Node:
				for i, c := range x {
					walk(pos+"/"+k+"/"+strconv.Itoa(i), c)
				}
			}
		}
	}
	if r.Root != nil {
		walk("root", r.Root)
	}
	return out
}

// checkNodes compares the mounted positions with expect.nodes, and each
// named field of each node.
func checkNodes(t *testing.T, step string, got map[string]*a2ui.Node, exp map[string]any) {
	t.Helper()
	want, _ := exp["nodes"].(map[string]any)
	var gotPos, wantPos []string
	for p := range got {
		gotPos = append(gotPos, p)
	}
	for p := range want {
		wantPos = append(wantPos, p)
	}
	sort.Strings(gotPos)
	sort.Strings(wantPos)
	if !jsonEqual(t, orNone(gotPos), orNone(wantPos)) {
		t.Errorf("%s: positions %v, want %v", step, gotPos, wantPos)
	}
	where := map[*a2ui.Node]string{}
	for p, n := range got {
		where[n] = p
	}
	for p, w := range want {
		n := got[p]
		if n == nil {
			continue
		}
		fields, _ := w.(map[string]any)
		for k, v := range fields {
			var g any
			switch k {
			case "component_id":
				g = n.ComponentID
			case "type":
				g = n.Type
			case "state":
				g = string(n.State)
			case "data_path":
				g = n.Scope.Path
			case "props":
				for prop, pv := range v.(map[string]any) {
					gp := propView(n.Props[prop], where)
					if !jsonEqual(t, gp, pv) {
						t.Errorf("%s: %s.%s = %v, want %v", step, p, prop, gp, pv)
					}
				}
				continue
			default:
				t.Fatalf("%s: unknown node field %s", step, k)
			}
			if !jsonEqual(t, g, v) {
				t.Errorf("%s: %s.%s = %v, want %v", step, p, k, g, v)
			}
		}
	}
}

func orNone(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// propView writes a resolved prop as node_resolution.md does.
func propView(v any, where map[*a2ui.Node]string) any {
	switch x := v.(type) {
	case a2ui.Bound:
		return map[string]any{"value": x.Value, "writable": x.Writable()}
	case *a2ui.Node:
		return map[string]any{"node": where[x]}
	case []*a2ui.Node:
		out := make([]any, len(x))
		for i, n := range x {
			out[i] = map[string]any{"node": where[n]}
		}
		return out
	case a2ui.ActionRef:
		return map[string]any{"action": true}
	}
	b, _ := json.Marshal(v)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}
