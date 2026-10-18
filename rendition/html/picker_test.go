package html

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// TestPickerOnHost: a HottyPicker is a box holding its query as an input
// of its own, which the host edits and gives the program the picker's keys
// from, its rows (the selected one marked) and its line of keys. A click on
// a row selects it, and a second runs its first action. A query the program
// writes comes in a new input, which has the keyboard (SPEC §6.2).
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
	input := func(n int) string { return partID("root", partTitle+strconv.Itoa(n)) }
	for _, want := range [][3]string{{input(0), "role", "combobox"}, {input(0), "placeholder", "Type a command"},
		{input(0), "data-keys", pickKeys(r.C.V.Find("root"))}} {
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
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if got := r.C.S.Data.Value("/sel"); got != "p-1" || r.C.St.Focus != "root" || s.Focused() != input(0) {
		t.Fatalf("a click on fish: selected %v, focus %q, the host's on %q", got, r.C.St.Focus, s.Focused())
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if len(x.actions) != 1 || x.actions[0].Context["row"] != "p-1" || x.actions[0].Context["action"] != "" {
		t.Fatalf("a second click: %+v", x.actions)
	}
	for _, k := range []string{"s", "p"} {
		if !x.h.Key(k) {
			t.Fatalf("the host did not type %s into the query", k)
		}
		x.pump()
		x.update(r)
	}
	if got := r.C.S.Data.Value("/q"); got != "sp" {
		t.Fatalf("after s, p: /q %v", got)
	}
	// The program writes another query: a new input, with the keyboard.
	must(t, r.C.S.Write("/q", "split"))
	r.C.Rebuild()
	x.update(r)
	x.check(r)
	if v, _ := s.Attr(input(1), "value"); v != "split" || s.Focused() != input(1) {
		t.Fatalf("the program's query: value %q, the host's keyboard on %q", v, s.Focused())
	}
}
