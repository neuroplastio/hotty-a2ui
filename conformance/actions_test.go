package conformance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"go.yaml.in/yaml/v3"
)

// actionSkips are the actions.yaml cases this v1.0 core does not run.
var actionSkips = map[string]string{
	"test_action_direct_name_shorthand": "v0.9's {name, context} shorthand; a v1.0 Action is {event} or {functionCall}",
}

// TestConformanceActions runs actions.yaml (dispatch_action in the
// brief): an Action resolved in a scope of the case's data and dispatched
// from a component. The suite writes bindings in v0.9's vocabulary
// ({path}); each context value is translated to v1.0's ({"@path"}).
func TestConformanceActions(t *testing.T) {
	for _, c := range suite(t, "actions.yaml") {
		if c["action"] != "dispatch_action" {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			if why, ok := actionSkips[str(c["name"])]; ok {
				t.Skip(why)
			}
			var sent []*a2ui.ActionMessage
			env := a2ui.Env{Action: func(_ *a2ui.Surface, a *a2ui.ActionMessage) { sent = append(sent, a) }}
			id := str(c["surfaceId"])
			if id == "" {
				id = "main"
			}
			s := a2ui.NewSurface(id, "test_catalog", a2ui.CatalogSet{"test_catalog": a2ui.NewOpenCatalog("test_catalog", a2ui.Version)}, env)
			if dm, ok := c["dataModel"]; ok {
				if err := s.Data.Set("/", dm); err != nil {
					t.Fatal(err)
				}
			}
			scope := str(c["scope"])
			if scope == "" {
				scope = "/"
			}
			action, _ := asJSON(t, c["actionPayload"]).(map[string]any)
			if ev, ok := action["event"].(map[string]any); ok {
				if ctx, ok := ev["context"].(map[string]any); ok {
					for k, v := range ctx {
						ctx[k] = v09to10(v)
					}
				}
			}
			if err := s.Dispatch(action, a2ui.Scope{Path: scope}, "test_component", false); err != nil {
				t.Fatal(err)
			}
			want, ok := c["expectDispatched"].(map[string]any)
			if !ok {
				return
			}
			if len(sent) != 1 {
				t.Fatalf("dispatched %d actions, want 1", len(sent))
			}
			got := sent[0]
			if got.Name != want["name"] || got.SurfaceID != id || got.SourceComponentID != "test_component" {
				t.Errorf("dispatched %s", mustJSON(t, got))
			}
			if w, ok := want["context"]; ok && !jsonEqual(t, got.Context, w) {
				t.Errorf("context %v, want %v", got.Context, w)
			}
			if w, ok := want["userMessage"]; ok && got.UserMessage != w {
				t.Errorf("userMessage %q, want %v", got.UserMessage, w)
			}
		})
	}
}

// TestConformanceResolvePath runs data_context.yaml (resolve_path): a
// path resolved in a scope. The suite is tagged v0.9; path resolution is
// the same in v1.0.
func TestConformanceResolvePath(t *testing.T) {
	for _, c := range suite(t, "data_context.yaml") {
		if c["action"] != "resolve_path" {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			args, _ := c["args"].(map[string]any)
			scope, ok := args["contextPath"].(string)
			if !ok {
				scope, ok = args["context_path"].(string)
			}
			if !ok {
				scope = "/"
			}
			ctx := &a2ui.Context{Scope: a2ui.Scope{Path: scope}}
			if got := ctx.Path(str(args["path"])); got != c["expect"] {
				t.Errorf("%q in %q is %q, want %v", args["path"], scope, got, c["expect"])
			}
		})
	}
}

// selectSkips are the select_catalog cases this core does not run.
var selectSkips = map[string]string{
	"test_multi_catalog_version_mismatch_error": "a surface here never holds a catalog of another version, so there is no mix to refuse; " +
		"a component that names one is a CatalogError (message_processor_v1_0 test_v10_update_components_mismatched_catalog_protocol_version_error)",
}

// TestConformanceSelectCatalog runs multi_catalog.yaml's select_catalog
// cases through the processor: a surface over the case's catalogs (open,
// by id), then its components, the catalog the last one uses; or a
// function call, the catalog the agent is asked to run it in.
func TestConformanceSelectCatalog(t *testing.T) {
	order := componentOrder(t, "multi_catalog.yaml")
	for _, c := range suite(t, "multi_catalog.yaml") {
		if c["action"] != "select_catalog" {
			continue
		}
		name := str(c["name"])
		t.Run(name, func(t *testing.T) {
			if why, ok := selectSkips[name]; ok {
				t.Skip(why)
			}
			args, _ := c["args"].(map[string]any)
			surface, _ := args["surface"].(map[string]any)
			sid, def := str(surface["id"]), str(surface["defaultCatalogId"])
			if sid == "" {
				sid = "main_surface"
			}
			if def == "" {
				def = "basic"
			}
			p := a2ui.NewProcessor()
			if cats, ok := args["catalogs"].(map[string]any); ok {
				for id, m := range cats {
					v := str(m.(map[string]any)["protocolVersion"])
					if v == "" {
						v = a2ui.Version
					}
					p.Register(a2ui.NewOpenCatalog(id, v))
				}
			} else {
				ids := []any{def}
				if l, ok := surface["supportedCatalogIds"].([]any); ok {
					ids = l
				}
				for _, id := range ids {
					p.Register(a2ui.NewOpenCatalog(str(id), a2ui.Version))
				}
			}
			var called []a2ui.CallFunction
			p.Send = func(o a2ui.Outbound) {
				if o.CallAgentFunction != nil {
					called = append(called, o.CallAgentFunction.CallFunction)
				}
			}
			msgs := []any{map[string]any{"version": a2ui.Version, "createSurface": map[string]any{"surfaceId": sid, "catalogId": def}}}
			var selected string
			var err error
			if comps, ok := args["components"].(map[string]any); ok {
				var defs []any
				for _, id := range order[name] {
					d, _ := asJSON(t, comps[id]).(map[string]any)
					d["id"] = id
					defs = append(defs, d)
				}
				msgs = append(msgs, map[string]any{"version": a2ui.Version, "updateComponents": map[string]any{"surfaceId": sid, "components": defs}})
				if err = p.Process(msgs); err == nil {
					s := p.Surface(sid)
					selected = s.CatalogOf(s.Components[order[name][len(order[name])-1]])
				}
			} else if fc, ok := args["functionCall"].(map[string]any); ok {
				if err = p.Process(msgs); err == nil {
					call := map[string]any{"@call": fc["call"], "args": fc["args"]}
					if id, ok := fc["catalogId"]; ok {
						call["catalogId"] = id
					}
					_, err = p.Surface(sid).Context(a2ui.RootScope).Resolve(call)
					if err == nil && len(called) == 1 {
						selected = called[0].Catalog
					}
				}
			}
			if want, ok := c["expectError"].(map[string]any); ok {
				checkError(t, "select", err, want)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if selected != c["expectSelected"] {
				t.Errorf("selected %q, want %v", selected, c["expectSelected"])
			}
		})
	}
}

// componentOrder reads each case's args.components keys in the order the
// file writes them, which a map loses.
func componentOrder(t *testing.T, file string) map[string][]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(conformanceDir, "core", file))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	get := func(n *yaml.Node, key string) *yaml.Node {
		for i := 0; n != nil && i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == key {
				return n.Content[i+1]
			}
		}
		return nil
	}
	out := map[string][]string{}
	for _, c := range doc.Content[0].Content {
		comps := get(get(c, "args"), "components")
		if comps == nil {
			continue
		}
		var ids []string
		for i := 0; i+1 < len(comps.Content); i += 2 {
			ids = append(ids, comps.Content[i].Value)
		}
		out[get(c, "name").Value] = ids
	}
	return out
}
