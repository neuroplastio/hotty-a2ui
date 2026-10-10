package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
)

// A text field's suggestions on a host are a datalist of the agent's
// options, which its input names (profile §2, §6.22): the field edits as
// ever, onInput tells the agent each change, and the options the agent
// writes reach the datalist as deltas. A longText takes none.
func TestSuggestionsOnHost(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"q":"","opts":["Berlin","Bern"]}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["city","notes"]},
 {"id":"city","component":"TextField","label":"City","value":{"@path":"/q"},
  "metadata":{"extensions":{"io_neuroplast_hotty":{"suggestions":{"options":{"@path":"/opts"},
   "onInput":{"event":{"name":"suggest","context":{"query":{"@path":"/q"}}}}}}}}},
 {"id":"notes","component":"TextField","label":"Notes","variant":"longText","value":"",
  "metadata":{"extensions":{"io_neuroplast_hotty":{"suggestions":{"options":["x"]}}}}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	if got, _ := s.Attr("city", "list"); got != partID("city", partList) {
		t.Fatalf("the input's list is %q", got)
	}
	doc := s.HTML()
	if !strings.Contains(doc, `<datalist id="city~x"><option value="Berlin"></option><option value="Bern"></option></datalist>`) {
		t.Errorf("no datalist of the options:\n%s", doc)
	}
	if strings.Contains(doc, "notes~x") {
		t.Error("a longText has a datalist")
	}
	r.C.Focus("city")
	x.update(r)
	for _, k := range []string{"B", "e"} {
		if !x.h.Key(k) {
			t.Fatalf("the host did not type %s", k)
		}
		x.pump()
	}
	if got := r.C.S.Data.Value("/q"); got != "Be" {
		t.Errorf("/q is %v", got)
	}
	if len(x.actions) != 2 || x.actions[1].Name != "suggest" || x.actions[1].Context["query"] != "Be" {
		t.Errorf("onInput sent %+v", x.actions)
	}
	x.process(map[string]any{"version": "v1.0", "updateDataModel": map[string]any{"surfaceId": "s", "path": "/opts", "value": []any{"Bergen"}}})
	if doc := s.HTML(); !strings.Contains(doc, `<datalist id="city~x"><option value="Bergen"></option></datalist>`) {
		t.Errorf("the rewritten options did not reach the datalist:\n%s", doc)
	}
	// Once the field is left, its value goes out too (no echo while edited).
	r.C.Focus("")
	x.update(r)
	x.pump()
	x.check(r)
}
