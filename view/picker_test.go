package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// paneRows are a palette's rows: two used commands under a heading, then
// a pane with three actions, then a command with none.
const paneRows = `[
 {"value":"pane.split","title":"Split","group":"Pane","section":"Recently used","match":[[0,2]],
  "actions":[{"id":"run","title":"Run","key":"enter"},{"id":"again","title":"Run and stay","key":"mod+enter"}]},
 {"value":"tab.new","title":"New tab","group":"Tab","section":"Recently used","actions":[{"id":"run","title":"Run"}]},
 {"value":"p-1","title":"fish","group":"Pane","detail":"~/code","section":"Panes",
  "actions":[{"id":"focus","title":"Focus"},{"id":"close","title":"Close pane","key":"shift+delete"},{"id":"copy","title":"Copy id","key":"alt+enter"}]},
 {"value":"config.get","title":"Get a setting","group":"Config","section":"Panes"}]`

// palette is a surface with a HottyPicker on paneRows, its query, rows,
// selection and action bound, and the actions it sent.
func palette(t *testing.T, height string) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	return paletteOf(t, paneRows, height)
}

// paletteOf is palette on rows.
func paletteOf(t *testing.T, rows, height string) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"q":"","sel":"","act":"","rows":` + rows + `}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyPicker","catalogId":"` + hotty.ID + `","query":{"@path":"/q"},"rows":{"@path":"/rows"},
	  "selected":{"@path":"/sel"},"action":{"@path":"/act"},"panelKey":"mod+k","placeholder":"Run a command…",` + height + `
	  "onQuery":{"event":{"name":"query","context":{"q":{"@path":"/q"}}}},
	  "onAction":{"event":{"name":"act","context":{"row":{"@path":"/sel"},"action":{"@path":"/act"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s")), actions
}

// data is the surface's data model at path.
func data(c *view.Controller, path string) any {
	v, _, _ := c.S.Data.Get(path)
	return v
}

// A HottyPicker's rows are the agent's, in its order, with a heading where
// the section changes; with selected naming none, the first is selected.
func TestPickerBuilds(t *testing.T) {
	c, _ := palette(t, "")
	e := c.V.Find("root")
	if e == nil || e.Kind != view.Picker || !e.Focusable() || e.Placeholder != "Run a command…" || e.NoRows != "No matches." {
		t.Fatalf("root: %+v, want a focusable HottyPicker", e)
	}
	if len(e.Rows) != 4 || e.Rows[0].Group != "Pane" || e.Rows[0].Match[0] != [2]int{0, 2} || e.Rows[2].Actions[1].Key != "shift+delete" {
		t.Errorf("rows %+v", e.Rows)
	}
	if e.Selected != 0 || e.Height != 10 || e.Top != 0 {
		t.Errorf("selected %d, height %d, top %d; want 0, 10, 0", e.Selected, e.Height, e.Top)
	}
	lines := e.PickLines()
	want := []view.PickLine{{Row: -1, Heading: "Recently used"}, {Row: 0}, {Row: 1}, {Row: -1, Heading: "Panes"}, {Row: 2}, {Row: 3}}
	if len(lines) != len(want) {
		t.Fatalf("lines %+v, want %+v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: %+v, want %+v", i, lines[i], want[i])
		}
	}
	if enter, panel := e.PickFooter(); enter != "Run" || !panel {
		t.Errorf("footer %q, %v", enter, panel)
	}
}

// The arrows and Control+p and Control+n move the selection round, the
// selection is written where it is bound, and the lines follow it: a row
// under a heading brings the heading with it.
func TestPickerMoves(t *testing.T) {
	c, _ := palette(t, `"height":2,`)
	for _, step := range []struct {
		key      string
		sel, top int
	}{
		{"arrow_down", 1, 1}, {"ctrl+n", 2, 3}, {"arrow_down", 3, 4}, {"arrow_down", 0, 0},
		{"ctrl+p", 3, 4}, {"page_up", 1, 2}, {"page_down", 3, 4},
	} {
		if ok, err := c.PickerKey("root", step.key); !ok || err != nil {
			t.Fatalf("%s: ok %v, err %v", step.key, ok, err)
		}
		e := c.V.Find("root")
		if e.Selected != step.sel || e.Top != step.top {
			t.Errorf("%s: selected %d, top %d; want %d, %d", step.key, e.Selected, e.Top, step.sel, step.top)
		}
		if got := data(c, "/sel"); got != e.Rows[e.Selected].Value {
			t.Errorf("%s: /sel is %v, want %q", step.key, got, e.Rows[e.Selected].Value)
		}
	}
	if ok, _ := c.PickerKey("root", "x"); ok {
		t.Error("the picker took x: a typed key is the query's")
	}
}

// describedRows are commands, all but the second with a description.
const describedRows = `[
 {"value":"pane.new","title":"pane.new","description":"New pane (your shell)","detail":"ctrl+b n"},
 {"value":"pane.split","title":"pane.split"},
 {"value":"pane.close","title":"pane.close","description":"Close the pane"},
 {"value":"tab.new","title":"tab.new","description":"New tab"}]`

// A row's description is a line of its own under it: the lines show the
// selected row whole, never start on a description, and a page counts a
// description's line with its row.
func TestPickerDescriptions(t *testing.T) {
	c, _ := paletteOf(t, describedRows, `"height":3,`)
	e := c.V.Find("root")
	if e.Rows[0].Description != "New pane (your shell)" || e.Rows[1].Description != "" {
		t.Errorf("rows %+v", e.Rows)
	}
	lines := e.PickLines()
	want := []view.PickLine{{Row: 0}, {Row: 0, Desc: true}, {Row: 1}, {Row: 2}, {Row: 2, Desc: true}, {Row: 3}, {Row: 3, Desc: true}}
	if len(lines) != len(want) {
		t.Fatalf("lines %+v, want %+v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: %+v, want %+v", i, lines[i], want[i])
		}
	}
	for _, step := range []struct {
		key      string
		sel, top int
	}{
		{"arrow_down", 1, 0}, {"arrow_down", 2, 2}, {"arrow_down", 3, 5}, {"arrow_up", 2, 3},
		{"page_up", 0, 0}, {"page_down", 2, 2},
	} {
		if ok, err := c.PickerKey("root", step.key); !ok || err != nil {
			t.Fatalf("%s: ok %v, err %v", step.key, ok, err)
		}
		if e := c.V.Find("root"); e.Selected != step.sel || e.Top != step.top {
			t.Errorf("%s: selected %d, top %d; want %d, %d", step.key, e.Selected, e.Top, step.sel, step.top)
		}
	}
	// A line high, the selected row shows its title.
	c, _ = paletteOf(t, describedRows, `"height":1,`)
	if e := c.V.Find("root"); e.Top != 0 {
		t.Errorf("height 1: top %d, want the first row's title", e.Top)
	}
}

// Enter runs the selected row's first action; a key one of its actions
// names runs that one; the action's id and the row's value are written
// first, for onAction's context. Enter on a row with none runs "".
func TestPickerActs(t *testing.T) {
	c, actions := palette(t, "")
	run := func(key string, want ...string) {
		t.Helper()
		n := len(*actions)
		if ok, err := c.PickerKey("root", key); !ok || err != nil {
			t.Fatalf("%s: ok %v, err %v", key, ok, err)
		}
		if len(want) == 0 {
			if len(*actions) != n {
				t.Errorf("%s ran %+v", key, (*actions)[n:])
			}
			return
		}
		if len(*actions) != n+1 {
			t.Fatalf("%s: %d actions, want one", key, len(*actions)-n)
		}
		a := (*actions)[n]
		if a.Name != "act" || a.Context["row"] != want[0] || a.Context["action"] != want[1] {
			t.Errorf("%s: %s %v, want row %q, action %q", key, a.Name, a.Context, want[0], want[1])
		}
	}
	run("enter", "pane.split", "run")
	run("mod+enter", "pane.split", "again")
	if ok, _ := c.PickerKey("root", "shift+delete"); ok {
		t.Error("shift+delete was taken on a row with no action on it")
	}
	c.PickerKey("root", "arrow_down")
	c.PickerKey("root", "arrow_down")
	run("shift+delete", "p-1", "close")
	run("alt+enter", "p-1", "copy")
	run("enter", "p-1", "focus")
	c.PickerKey("root", "arrow_down")
	run("enter", "config.get", "")
}

// panelKey opens the selected row's actions: the arrows move in them,
// Enter runs the one highlighted, Escape closes them; a row with no
// actions has no panel.
func TestPickerPanel(t *testing.T) {
	c, actions := palette(t, "")
	c.PickerKey("root", "arrow_down")
	c.PickerKey("root", "arrow_down")
	if ok, _ := c.PickerKey("root", "ctrl+k"); !ok || !c.V.Find("root").Panel {
		t.Fatal("ctrl+k (mod+k) opened no panel")
	}
	c.PickerKey("root", "arrow_up")
	if at := c.V.Find("root").PanelAt; at != 2 {
		t.Errorf("arrow_up from the first action: at %d, want 2 (round)", at)
	}
	c.PickerKey("root", "escape")
	if c.V.Find("root").Panel {
		t.Error("escape left the panel open")
	}
	c.PickerKey("root", "ctrl+k")
	c.PickerKey("root", "arrow_down")
	c.PickerKey("root", "enter")
	if e := c.V.Find("root"); e.Panel || len(*actions) != 1 || (*actions)[0].Context["action"] != "close" {
		t.Errorf("enter in the panel: open %v, actions %+v", e.Panel, *actions)
	}
	c.PickerKey("root", "arrow_down")
	if ok, _ := c.PickerKey("root", "ctrl+k"); !ok || c.V.Find("root").Panel {
		t.Error("a row with no actions opened a panel")
	}
}

// A change to the query is written where it is bound, the selection goes
// back to the first row, and onQuery runs.
func TestPickerQuery(t *testing.T) {
	c, actions := palette(t, "")
	c.PickerKey("root", "arrow_down")
	if err := c.SetValue("root", "sp"); err != nil {
		t.Fatal(err)
	}
	e := c.V.Find("root")
	if e.Value != "sp" || data(c, "/q") != "sp" || e.Selected != 0 || data(c, "/sel") != "" {
		t.Errorf("query %v (/q %v), selected %d (/sel %v)", e.Value, data(c, "/q"), e.Selected, data(c, "/sel"))
	}
	if len(*actions) != 1 || (*actions)[0].Name != "query" || (*actions)[0].Context["q"] != "sp" {
		t.Errorf("actions %+v, want query with q sp", *actions)
	}
	if err := c.SetValue("root", "sp"); err != nil || len(*actions) != 1 {
		t.Errorf("the same query again ran onQuery: %v", err)
	}
}

// Tab completes the query to the selected row's title, as a query the
// agent wrote (Written), and onQuery runs.
func TestPickerTab(t *testing.T) {
	c, actions := palette(t, "")
	c.PickerKey("root", "arrow_down")
	if ok, err := c.PickerKey("root", "tab"); !ok || err != nil {
		t.Fatalf("tab: ok %v, err %v", ok, err)
	}
	e := c.V.Find("root")
	if e.Value != "New tab" || data(c, "/q") != "New tab" || e.Written != 1 || e.Selected != 0 {
		t.Errorf("query %v (/q %v), written %d, selected %d", e.Value, data(c, "/q"), e.Written, e.Selected)
	}
	if len(*actions) != 1 || (*actions)[0].Name != "query" {
		t.Errorf("actions %+v, want query", *actions)
	}
}
