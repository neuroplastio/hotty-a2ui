package html

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// TestPickerOnHost: a HottyPicker is a focusable box holding its query,
// with the program's caret, its rows (a heading where their section
// changes, the selected one marked) and its line of keys. A click on a
// row selects it, and a second runs its first action. The program edits
// the query by the keys, and the host follows by deltas.
func TestPickerOnHost(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"q":"","sel":"","act":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"HottyPicker","catalogId":"`+hottycat.ID+`","query":{"@path":"/q"},"selected":{"@path":"/sel"},"action":{"@path":"/act"},
  "chip":">","placeholder":"Type a command","panelKey":"mod+k",
  "rows":[{"value":"split","group":"Pane","title":"Split","section":"Used","match":[[0,2]],"actions":[{"id":"run","title":"Run"},{"id":"again","title":"Again","key":"mod+enter"}]},
   {"value":"p-1","group":"Pane","title":"fish","detail":"~/code","section":"Panes"}],
  "onAction":{"event":{"name":"act","context":{"row":{"@path":"/sel"},"action":{"@path":"/act"}}}}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	for _, want := range [][3]string{{"root", "tabindex", "0"}, {"root", "role", "combobox"}, {"root", "data-keys", listKeys}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	row := func(i int) string { return partID("root", partItem+strconv.Itoa(i)) }
	if got := s.TextOf(row(0)); got != "Pane: Splitctrl+enter" {
		t.Errorf("the first row reads %q", got)
	}
	if v, _ := s.Attr(row(0), "aria-selected"); v != "true" {
		t.Errorf("the first row is not selected")
	}
	if got := s.TextOf(partID("root", partTitle)); got != ">Type a command" {
		t.Errorf("the query reads %q", got)
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if got := r.C.S.Data.Value("/sel"); got != "p-1" || r.C.St.Focus != "root" {
		t.Fatalf("a click on fish: selected %v, focus %q", got, r.C.St.Focus)
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if len(x.actions) != 1 || x.actions[0].Context["row"] != "p-1" || x.actions[0].Context["action"] != "" {
		t.Fatalf("a second click: %+v", x.actions)
	}
	for _, k := range []string{"s", "p"} {
		if _, ok, err := r.Key(k); !ok || err != nil {
			t.Fatalf("key %s: %v %v", k, ok, err)
		}
		x.update(r)
		x.check(r)
	}
	if got := s.TextOf(partID("root", partTitle)); got != ">sp" || r.C.S.Data.Value("/q") != "sp" {
		t.Errorf("after s, p: the query reads %q, /q %v", got, r.C.S.Data.Value("/q"))
	}
}
