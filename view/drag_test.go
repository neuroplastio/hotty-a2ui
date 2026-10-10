package view_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// surface makes a surface of a data model and components, and the actions
// it sent.
func surface(t *testing.T, data, components string) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
		if o.Error != nil {
			t.Errorf("error to the agent: %+v", o.Error)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":` + data + `}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":` + components + `}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s")), actions
}

// jsonOf is a value as JSON, to compare.
func jsonOf(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// todos is a reorderable HottyList, bound, with its moves written to /move.
func todos(t *testing.T) (*view.Controller, *[]*a2ui.ActionMessage) {
	return surface(t, `{"sel":"a","move":null,"todos":[{"label":"A","value":"a"},{"label":"B","value":"b"},{"label":"C","value":"c"},{"label":"D","value":"d"}]}`,
		`[{"id":"root","component":"HottyList","catalogId":"`+hotty.ID+`","items":{"@path":"/todos"},"selected":{"@path":"/sel"},
		 "reorderable":true,"moved":{"@path":"/move"},"onMove":{"event":{"name":"moved","context":{"move":{"@path":"/move"}}}}}]`)
}

func order(c *view.Controller, path string) string {
	var out string
	l, _ := c.S.Data.Value(path).([]any)
	for _, x := range l {
		m, _ := x.(map[string]any)
		out += a2ui.ToString(m["value"])
	}
	return out
}

// A drag lifts an item, a line shows where it would land, and the drop
// moves it there: the items are written, the item stays selected, the move
// is written to moved, and onMove runs with it.
func TestDragMovesAnItem(t *testing.T) {
	c, actions := todos(t)
	if !c.Lift("root", 0) {
		t.Fatal("Lift refused a reorderable list's item")
	}
	// Without the pointer's pixels, the item takes the place of the one it
	// is over: after it, coming from above.
	c.Over("root", 2, -1)
	if d := c.St.Drag; d.Target != "root" || d.At != 2 || d.Where != view.After {
		t.Fatalf("over C from above: %+v, want after 2", d)
	}
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if got := order(c, "/todos"); got != "bcad" {
		t.Errorf("order %s, want bcad", got)
	}
	if got := c.S.Data.Value("/sel"); got != "a" {
		t.Errorf("selected %v, want a, which moved", got)
	}
	want := `{"from":{"index":0,"path":"/todos"},"item":"a","to":{"index":2,"path":"/todos"}}`
	if got := jsonOf(c.S.Data.Value("/move")); got != want {
		t.Errorf("moved %s, want %s", got, want)
	}
	if len(*actions) != 1 || (*actions)[0].Name != "moved" || jsonOf((*actions)[0].Context["move"]) != want {
		t.Errorf("actions %+v, want moved with the move", *actions)
	}
	if c.St.Drag != nil {
		t.Error("the drag outlived its drop")
	}
}

// With the pointer's pixels the half of the row decides; where the item
// would not move, no line shows and a drop lands nothing.
func TestDragHalvesAndStays(t *testing.T) {
	c, actions := todos(t)
	c.Lift("root", 2)
	for _, tc := range []struct {
		item   int
		frac   float64
		target bool
		where  view.Where
	}{
		{0, 0.2, true, view.Before},
		{0, 0.8, true, view.After},
		{1, 0.9, false, 0}, // after B is where C is
		{3, 0.1, false, 0}, // before D too
		{3, 0.6, true, view.After},
		{2, -1, false, 0}, // over itself
	} {
		c.Over("root", tc.item, tc.frac)
		d := c.St.Drag
		if (d.Target != "") != tc.target || tc.target && (d.At != tc.item || d.Where != tc.where) {
			t.Errorf("over %d at %v: %+v, want target %v %v", tc.item, tc.frac, d, tc.target, tc.where)
		}
	}
	c.Over("root", 1, 0.9)
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if got := order(c, "/todos"); got != "abcd" || len(*actions) != 0 {
		t.Errorf("a drop where it stays: %s, %d actions; want abcd and none", got, len(*actions))
	}
}

// Escape cancels: nothing moves.
func TestCancelDrag(t *testing.T) {
	c, actions := todos(t)
	c.Lift("root", 0)
	c.Over("root", 3, -1)
	c.CancelDrag()
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if got := order(c, "/todos"); got != "abcd" || len(*actions) != 0 {
		t.Errorf("after a cancel: %s, %d actions", got, len(*actions))
	}
}

// Alt+ArrowUp and Alt+ArrowDown move the selected item a place, and stay
// at the ends; other lists leave the keys.
func TestMoveKeys(t *testing.T) {
	c, actions := todos(t)
	for _, step := range []struct{ key, want string }{
		{"Alt+ArrowUp", "abcd"},
		{"Alt+ArrowDown", "bacd"},
		{"Alt+ArrowDown", "bcad"},
		{"Alt+ArrowDown", "bcda"},
		{"Alt+ArrowDown", "bcda"},
		{"Alt+ArrowUp", "bcad"},
	} {
		if ok, err := c.MoveKey("root", step.key); !ok || err != nil {
			t.Fatalf("%s: %v %v", step.key, ok, err)
		}
		if got := order(c, "/todos"); got != step.want {
			t.Errorf("%s: %s, want %s", step.key, got, step.want)
		}
	}
	if c.S.Data.Value("/sel") != "a" || len(*actions) != 4 {
		t.Errorf("selected %v, %d actions; want a, 4", c.S.Data.Value("/sel"), len(*actions))
	}
	if ok, _ := c.MoveKey("root", "Alt+ArrowLeft"); ok {
		t.Error("a list took Alt+ArrowLeft")
	}
	if ok, _ := c.MoveKey("root", "ArrowDown"); ok {
		t.Error("MoveKey took ArrowDown")
	}
}

// A literal list's moves are the renderer's, until the literal changes.
func TestMoveLiteral(t *testing.T) {
	c, _ := surface(t, `{}`, `[{"id":"root","component":"HottyTable","catalogId":"`+hotty.ID+`","reorderable":true,"rowKey":"id",
	 "columns":[{"key":"id","header":"Id"}],"rows":[{"id":"x"},{"id":"y"},{"id":"z"}]}]`)
	c.Focus("root")
	if err := c.SelectRow("root", 0); err != nil {
		t.Fatal(err)
	}
	if ok, err := c.MoveKey("root", "Alt+ArrowDown"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if e := c.V.Find("root"); jsonOf(e.RowIDs) != `["y","x","z"]` || e.SelectedRow() != 1 {
		t.Errorf("rows %v selected %d, want y x z, the second", e.RowIDs, e.SelectedRow())
	}
	c.S.Update([]map[string]any{{"id": "root", "component": "HottyTable", "catalogId": hotty.ID, "reorderable": true, "rowKey": "id",
		"columns": []any{map[string]any{"key": "id"}}, "rows": []any{map[string]any{"id": "x"}, map[string]any{"id": "w"}}}})
	c.S.Tree.Resolve()
	c.Rebuild()
	if e := c.V.Find("root"); jsonOf(e.RowIDs) != `["x","w"]` {
		t.Errorf("rows %v after the agent's, want x w", e.RowIDs)
	}
}

// fileTree is a reorderable tree: docs/ (a.md, b.md), src/ (main.go), an
// empty folder, and README.
func fileTree(t *testing.T) (*view.Controller, *[]*a2ui.ActionMessage) {
	return surface(t, `{"sel":"","move":null,"files":[
	 {"label":"docs","value":"docs","children":[{"label":"a.md","value":"docs/a.md"},{"label":"b.md","value":"docs/b.md"}]},
	 {"label":"src","value":"src","children":[{"label":"main.go","value":"src/main.go"}]},
	 {"label":"empty","value":"empty","children":[]},
	 {"label":"README","value":"README"}],"open":["docs","src"]}`,
		`[{"id":"root","component":"HottyTree","catalogId":"`+hotty.ID+`","items":{"@path":"/files"},"selected":{"@path":"/sel"},
		 "expanded":{"@path":"/open"},"reorderable":true,"moved":{"@path":"/move"},"onMove":{"event":{"name":"moved","context":{"move":{"@path":"/move"}}}}}]`)
}

// treeOrder is a tree's values in pre-order, a level's children in brackets.
func treeOrder(v any) string {
	var out string
	l, _ := v.([]any)
	for _, x := range l {
		m, _ := x.(map[string]any)
		out += a2ui.ToString(m["value"]) + " "
		if k, ok := m["children"].([]any); ok && len(k) > 0 {
			out += "[ " + treeOrder(k) + "] "
		}
	}
	return out
}

// node is a tree's node index by its value.
func node(c *view.Controller, id string) int {
	e := c.V.Find("root")
	for i, r := range e.RowIDs {
		if r == id {
			return i
		}
	}
	return -1
}

// In a tree, a node goes into a node that takes children (a folder, empty
// or not) over its middle third, or without pixels; a folder never goes
// into itself or what is inside it.
func TestDragTree(t *testing.T) {
	c, actions := fileTree(t)
	c.Lift("root", node(c, "docs"))
	for _, tc := range []struct {
		over   string
		frac   float64
		target bool
		where  view.Where
		at     string
	}{
		{"docs/a.md", -1, false, 0, ""},                // inside itself
		{"docs", 0.5, false, 0, ""},                    // itself
		{"src", -1, true, view.Into, "src"},            // a folder, without pixels
		{"src", 0.1, false, 0, ""},                     // its top third: before it is where docs is
		{"src", 0.5, true, view.Into, "src"},           // its middle
		{"src", 0.9, true, view.Before, "src/main.go"}, // under an open folder: its first child's place
		{"README", 0.7, true, view.After, "README"},
		{"empty", -1, true, view.Into, "empty"},
	} {
		c.Over("root", node(c, tc.over), tc.frac)
		d := c.St.Drag
		got := d.Target != ""
		if got != tc.target || got && (d.Where != tc.where || d.At != node(c, tc.at)) {
			t.Errorf("over %s at %v: %+v, want %v %v %s", tc.over, tc.frac, d, tc.target, tc.where, tc.at)
		}
	}
	c.Over("root", node(c, "empty"), -1)
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if got, want := treeOrder(c.S.Data.Value("/files")), "src [ src/main.go ] empty [ docs [ docs/a.md docs/b.md ] ] README "; got != want {
		t.Errorf("tree %q, want %q", got, want)
	}
	want := `{"from":{"index":0,"parent":null,"path":"/files"},"item":"docs","to":{"index":0,"parent":"empty","path":"/files"}}`
	if got := jsonOf(c.S.Data.Value("/move")); got != want {
		t.Errorf("moved %s, want %s", got, want)
	}
	if c.S.Data.Value("/sel") != "docs" || len(*actions) != 1 {
		t.Errorf("selected %v, %d actions", c.S.Data.Value("/sel"), len(*actions))
	}
	if e := c.V.Find("root"); !e.Nodes[node(c, "empty")].Open {
		t.Error("the folder it went into is closed")
	}
}

// A tree's keys: up and down among its siblings, left out of its branch,
// right into the node above it that takes children.
func TestTreeMoveKeys(t *testing.T) {
	c, _ := fileTree(t)
	if err := c.SelectNode("root", node(c, "docs/b.md")); err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct{ key, want string }{
		{"Alt+ArrowUp", "docs [ docs/b.md docs/a.md ] src [ src/main.go ] empty README "},
		{"Alt+ArrowUp", "docs [ docs/b.md docs/a.md ] src [ src/main.go ] empty README "},
		{"Alt+ArrowLeft", "docs [ docs/a.md ] docs/b.md src [ src/main.go ] empty README "},
		{"Alt+ArrowDown", "docs [ docs/a.md ] src [ src/main.go ] docs/b.md empty README "},
		{"Alt+ArrowRight", "docs [ docs/a.md ] src [ src/main.go docs/b.md ] empty README "},
		{"Alt+ArrowLeft", "docs [ docs/a.md ] src [ src/main.go ] docs/b.md empty README "},
		{"Alt+ArrowDown", "docs [ docs/a.md ] src [ src/main.go ] empty docs/b.md README "},
		{"Alt+ArrowRight", "docs [ docs/a.md ] src [ src/main.go ] empty [ docs/b.md ] README "},
		{"Alt+ArrowLeft", "docs [ docs/a.md ] src [ src/main.go ] empty docs/b.md README "},
		{"Alt+ArrowDown", "docs [ docs/a.md ] src [ src/main.go ] empty README docs/b.md "},
		{"Alt+ArrowRight", "docs [ docs/a.md ] src [ src/main.go ] empty README docs/b.md "},
	} {
		if ok, err := c.MoveKey("root", step.key); !ok || err != nil {
			t.Fatalf("%s: %v %v", step.key, ok, err)
		}
		if got := treeOrder(c.S.Data.Value("/files")); got != step.want {
			t.Errorf("%s: %q, want %q", step.key, got, step.want)
		}
		if c.S.Data.Value("/sel") != "docs/b.md" {
			t.Errorf("%s: selected %v", step.key, c.S.Data.Value("/sel"))
		}
	}
}

// board is a kanban: two Lists of a template, cards of one type, and an
// assignee that takes a card dropped on it.
func board(t *testing.T) (*view.Controller, *[]*a2ui.ActionMessage) {
	return surface(t, `{"todo":[{"id":"c1","title":"One"},{"id":"c2","title":"Two"}],"doing":[{"id":"c3","title":"Three"}],"move":null,"dropped":null}`,
		`[{"id":"root","component":"Row","children":["todo","doing","ada"]},
		 {"id":"todo","component":"List","children":{"componentId":"card","path":"/todo"},
		  "metadata":{"extensions":{"io_neuroplast_hotty":{"reorder":{"type":"card","moved":{"@path":"/move"},"onMove":{"event":{"name":"moved","context":{"move":{"@path":"/move"}}}}}}}}},
		 {"id":"doing","component":"List","children":{"componentId":"card","path":"/doing"},
		  "metadata":{"extensions":{"io_neuroplast_hotty":{"reorder":{"type":"card","moved":{"@path":"/move"}}}}}},
		 {"id":"card","component":"Button","child":"title","action":{"event":{"name":"open","context":{"id":{"@path":"id"}}}}},
		 {"id":"title","component":"Text","text":{"@path":"title"}},
		 {"id":"ada","component":"Card","child":"ada_t",
		  "metadata":{"extensions":{"io_neuroplast_hotty":{"drop":{"accepts":["card"],"value":{"@path":"/dropped"},
		   "action":{"event":{"name":"assign","context":{"card":{"@path":"/dropped/id"},"to":"ada"}}}}}}}},
		 {"id":"ada_t","component":"Text","text":"Ada"}]`)
}

// Lists of a template with reorder take each other's items: the template's
// lists are written, the keyboard stays on the card it was on, and the
// List the card went to runs its onMove, else the one it left.
func TestDragBetweenLists(t *testing.T) {
	c, actions := board(t)
	if e := c.V.Find("todo"); !e.Movable || e.ItemType != "card" {
		t.Fatalf("todo: %+v, want a movable List of cards", e)
	}
	c.Focus("card[/todo/1]")
	if !c.Lift("todo", 1) {
		t.Fatal("Lift refused a List's item")
	}
	target, child, item := c.DropTarget("title[/doing/0]")
	if target != "doing" || child != "card[/doing/0]" || item != 0 {
		t.Fatalf("DropTarget: %s %s %d, want doing's first card", target, child, item)
	}
	c.Over(target, item, -1)
	if d := c.St.Drag; d.Target != "doing" || d.Where != view.Before {
		t.Fatalf("over doing's card: %+v, want before it", d)
	}
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if got := jsonOf(c.S.Data.Value("/doing")); got != `[{"id":"c2","title":"Two"},{"id":"c3","title":"Three"}]` {
		t.Errorf("doing %s", got)
	}
	if got := jsonOf(c.S.Data.Value("/todo")); got != `[{"id":"c1","title":"One"}]` {
		t.Errorf("todo %s", got)
	}
	if c.St.Focus != "card[/doing/0]" {
		t.Errorf("focus %s, want the card where it went", c.St.Focus)
	}
	// doing has no onMove: todo's runs.
	want := `{"from":{"index":1,"path":"/todo"},"item":{"id":"c2","title":"Two"},"to":{"index":0,"path":"/doing"}}`
	if len(*actions) != 1 || (*actions)[0].Name != "moved" || jsonOf((*actions)[0].Context["move"]) != want {
		t.Errorf("actions %v, want moved: %s", jsonOf(*actions), want)
	}
	// The keys move the card the keyboard is on, in its List.
	c.Focus("card[/doing/0]")
	if ok, err := c.MoveKey("card[/doing/0]", "Alt+ArrowDown"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if got := jsonOf(c.S.Data.Value("/doing")); got != `[{"id":"c3","title":"Three"},{"id":"c2","title":"Two"}]` || c.St.Focus != "card[/doing/1]" {
		t.Errorf("doing %s, focus %s", got, c.St.Focus)
	}
	// A List takes an item into its end when empty.
	c.Lift("todo", 0)
	c.Over("todo", -1, -1)
	if c.St.Drag.Target != "" {
		t.Error("a List with items took one at no item")
	}
}

// A drop target takes what is dropped on it: written where its value is
// bound, and its action runs, which the agent gives a meaning.
func TestDropOnTarget(t *testing.T) {
	c, actions := board(t)
	c.Lift("doing", 0)
	target, child, _ := c.DropTarget("ada_t")
	if target != "ada" || child != "" {
		t.Fatalf("DropTarget(ada_t): %s %s, want ada", target, child)
	}
	c.Over(target, -1, -1)
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if got := jsonOf(c.S.Data.Value("/dropped")); got != `{"id":"c3","title":"Three"}` {
		t.Errorf("dropped %s", got)
	}
	if len(*actions) != 1 || (*actions)[0].Name != "assign" || !reflect.DeepEqual((*actions)[0].Context, map[string]any{"card": "c3", "to": "ada"}) {
		t.Errorf("actions %s, want assign c3 to ada", jsonOf(*actions))
	}
	if got := jsonOf(c.S.Data.Value("/doing")); got != `[{"id":"c3","title":"Three"}]` {
		t.Errorf("doing %s: a drop on a target moves nothing", got)
	}
}

// A drag source carries its value, resolved in its scope; what takes no
// such type is no target.
func TestDragSource(t *testing.T) {
	c, actions := surface(t, `{"people":[{"name":"Ada"}],"tickets":[{"id":"T-1"}],"got":null}`,
		`[{"id":"root","component":"Column","children":["ts","ps"]},
		 {"id":"ts","component":"Column","children":{"componentId":"ticket","path":"/tickets"}},
		 {"id":"ticket","component":"Text","text":{"@path":"id"},"metadata":{"extensions":{"io_neuroplast_hotty":{"drag":{"type":"ticket","value":{"@path":"id"}}}}}},
		 {"id":"ps","component":"Column","children":{"componentId":"person","path":"/people"}},
		 {"id":"person","component":"Text","text":{"@path":"name"},"metadata":{"extensions":{"io_neuroplast_hotty":{"drop":{"accepts":"ticket","value":{"@path":"/got"},
		  "action":{"event":{"name":"assign","context":{"ticket":{"@path":"/got"},"person":{"@path":"name"}}}}}}}}}]`)
	if !c.Lift("ticket[/tickets/0]", -1) || c.St.Drag.Value != "T-1" {
		t.Fatalf("drag %+v, want T-1", c.St.Drag)
	}
	if target, _, _ := c.DropTarget("ts"); target != "" {
		t.Errorf("a Column took a ticket: %s", target)
	}
	target, _, _ := c.DropTarget("person[/people/0]")
	c.Over(target, -1, -1)
	if err := c.Drop(); err != nil {
		t.Fatal(err)
	}
	if len(*actions) != 1 || jsonOf((*actions)[0].Context) != `{"person":"Ada","ticket":"T-1"}` {
		t.Errorf("actions %s", jsonOf(*actions))
	}
}
