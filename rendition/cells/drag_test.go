package cells

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// dragging is a drag and drop story's surface in cells, drawn at cols,
// and the names of the actions it sent, with their contexts.
func dragging(t *testing.T, name string, cols int) (*Rendition, *view.Controller, *story.Run) {
	t.Helper()
	st := story.Find(name)
	if st == nil {
		t.Fatalf("no story %s", name)
	}
	run, err := story.Start(st)
	if err != nil {
		t.Fatal(err)
	}
	c := run.Surfaces()[0].C
	r := New(c)
	r.Draw(cols)
	return r, c, run
}

// sent are a run's actions, each its name and its context as JSON.
func sent(run *story.Run) []string {
	var out []string
	for _, e := range run.Actions() {
		var m struct {
			Action struct {
				Name    string         `json:"name"`
				Context map[string]any `json:"context"`
			} `json:"action"`
		}
		_ = json.Unmarshal(e.JSON, &m)
		b, _ := json.Marshal(m.Action.Context)
		out = append(out, m.Action.Name+" "+string(b))
	}
	return out
}

// rowOf is the first row of a frame whose text has s.
func rowOf(t *testing.T, f *Frame, s string) int {
	t.Helper()
	for y, line := range strings.Split(f.Plain(), "\n") {
		if strings.Contains(line, s) {
			return y
		}
	}
	t.Fatalf("no row has %q in\n%s", s, f.Plain())
	return -1
}

// marks are a frame's rows as Plain has them, with what a drag paints
// shown beside each: "_" under a column underlined in a colour of its
// own (the line), "‾" over one overlined, "~" faint, "#" reversed in the
// accent. Columns that have none of these are blank in the row of marks.
func marks(f *Frame) string {
	var b strings.Builder
	for y, row := range f.Cells {
		var line, mark strings.Builder
		for _, c := range row {
			line.WriteString(c.Text)
			m := " "
			switch {
			case c.LineSet && c.Attr&Underline != 0:
				m = "_"
			case c.LineSet && c.Attr&Overline != 0:
				m = "‾"
			case c.Attr&Reverse != 0 && c.Role == Accent && y > 0:
				m = "#"
			case c.Attr&Faint != 0:
				m = "~"
			}
			if c.Width != 0 || c.Text != "" {
				mark.WriteString(m)
			}
		}
		b.WriteString(strings.TrimRight(line.String(), " ") + "\n")
		if m := strings.TrimRight(mark.String(), " "); m != "" {
			b.WriteString(m + "\n")
		}
	}
	return b.String()
}

// A press on a reorderable HottyList's item and a move off it lifts the
// item, which shows faint, and a line shows where it would land; the
// release drops it there, writes the items and sends moved.
func TestDragTodo(t *testing.T) {
	r, c, run := dragging(t, "hotty/drag-todo", 40)
	f := r.Draw(40)
	milk, ferns := rowOf(t, f, "Buy oat milk"), rowOf(t, f, "Water the ferns")
	must(t, r.Click(4, milk))
	must(t, r.Drag(4, milk+1))
	if c.St.Drag == nil {
		t.Fatal("no drag after moving off the item")
	}
	must(t, r.Drag(4, ferns))
	f = r.Draw(40)
	t.Logf("over ferns:\n%s", marks(f))
	if d := c.St.Drag; d.Target != "list" || d.At != 2 || d.Where != view.After {
		t.Errorf("over Water the ferns: %+v, want after it", *d)
	}
	if c := f.Cells[milk][4]; c.Attr&Faint == 0 {
		t.Errorf("the item lifted is not faint:\n%s", marks(f))
	}
	for _, x := range []int{0, 20, 39} {
		if c := f.Cells[ferns][x]; c.Attr&Underline == 0 || !c.LineSet || c.Line != Accent {
			t.Errorf("no line under Water the ferns at %d:\n%s", x, marks(f))
		}
	}
	must(t, r.Release())
	f = r.Draw(40)
	t.Logf("dropped:\n%s", f.Plain())
	if got := order(c, "/todos"); got != "plumber ferns milk passport backup dentist" {
		t.Errorf("after the drop: %s", got)
	}
	if got := sent(run); len(got) != 1 || !strings.HasPrefix(got[0], `moved {"move":{"from":{"index":0,`) {
		t.Errorf("sent %v", got)
	}
}

// Escape during a drag leaves the item where it was, and so does a
// release where no line shows; a press and a release on an item with no
// move between is a click.
func TestDragCancels(t *testing.T) {
	r, c, run := dragging(t, "hotty/drag-todo", 40)
	f := r.Draw(40)
	milk, ferns := rowOf(t, f, "Buy oat milk"), rowOf(t, f, "Water the ferns")
	must(t, r.Click(4, milk))
	must(t, r.Drag(4, ferns))
	if ok, err := r.Key("Escape"); !ok || err != nil || c.St.Drag != nil {
		t.Fatalf("Escape: %v %v %+v", ok, err, c.St.Drag)
	}
	must(t, r.Release())
	must(t, r.Click(4, milk))
	must(t, r.Drag(4, ferns))
	must(t, r.Drag(4, f.Rows+5))
	if d := c.St.Drag; d == nil || d.Target != "" {
		t.Fatalf("below the surface: %+v, want a drag over nothing", d)
	}
	must(t, r.Release())
	if got := order(c, "/todos"); got != "milk plumber ferns passport backup dentist" || len(sent(run)) != 0 {
		t.Errorf("after the cancels: %s, sent %v", got, sent(run))
	}
	r.Draw(40)
	must(t, r.Click(4, ferns))
	must(t, r.Release())
	if c.S.Data.Value("/selected") != "ferns" || c.St.Drag != nil {
		t.Errorf("a click: selected %v", c.S.Data.Value("/selected"))
	}
}

// order is the values of the items at path, in order.
func order(c *view.Controller, path string) string {
	var out []string
	items, _ := c.S.Data.Value(path).([]any)
	for _, it := range items {
		m, _ := it.(map[string]any)
		v := m["value"]
		if v == nil {
			v = m["key"]
		}
		if v == nil {
			v = m["id"]
		}
		s, _ := v.(string)
		out = append(out, s)
	}
	return strings.Join(out, " ")
}

// Alt with the arrows moves the selected item a place, in cells as in the
// view.
func TestMoveKeysInCells(t *testing.T) {
	r, c, run := dragging(t, "hotty/drag-backlog", 60)
	c.Focus("table")
	if ok, err := r.Key("Alt+ArrowDown"); !ok || err != nil {
		t.Fatalf("Alt+ArrowDown: %v %v", ok, err)
	}
	if got := order(c, "/backlog"); got != "KIT-27 KIT-31 KIT-35 KIT-22 KIT-40" {
		t.Errorf("after Alt+ArrowDown: %s", got)
	}
	if c.S.Data.Value("/selected") != "KIT-31" || len(sent(run)) != 1 {
		t.Errorf("selected %v, sent %v", c.S.Data.Value("/selected"), sent(run))
	}
	t.Logf("\n%s", r.Draw(60).Plain())
}

// Over a folder's row a tree's node goes into it, which the row in the
// accent shows; a folder over its own child goes nowhere.
func TestDragFiles(t *testing.T) {
	r, c, _ := dragging(t, "hotty/drag-files", 40)
	f := r.Draw(40)
	t.Logf("\n%s", f.Plain())
	readme, archive := rowOf(t, f, "README.md"), rowOf(t, f, "archive")
	must(t, r.Click(6, readme))
	must(t, r.Drag(6, archive))
	f = r.Draw(40)
	t.Logf("over archive:\n%s", marks(f))
	if d := c.St.Drag; d == nil || d.Target != "tree" || d.Where != view.Into {
		t.Fatalf("over archive: %+v, want into it", d)
	}
	must(t, r.Release())
	if files, _ := json.Marshal(c.S.Data.Value("/files/2")); !strings.Contains(string(files), `"children":[{"icon":"description","label":"README.md"`) {
		t.Errorf("archive: %s", files)
	}
	f = r.Draw(40)
	src, util := rowOf(t, f, "src"), rowOf(t, f, "util")
	must(t, r.Click(6, src))
	must(t, r.Drag(6, util))
	if d := c.St.Drag; d == nil || d.Target != "" {
		t.Errorf("src over util, inside it: %+v, want nowhere", d)
	}
	f = r.Draw(40)
	t.Logf("src over util:\n%s", marks(f))
	must(t, r.Release())
}

// A card goes from one column's List to another's; dropped on a person,
// it is written to /assigning and assign is sent, and it stays where it
// was.
func TestDragKanban(t *testing.T) {
	r, c, run := dragging(t, "hotty/drag-kanban", 72)
	f := r.Draw(72)
	t.Logf("\n%s", f.Plain())
	x, y, _, _, _ := r.Box("card[/board/todo/1]")
	_, ty, _, _, _ := r.Box("card[/board/doing/0]")
	dx, _, _, _, _ := r.Box("doing")
	must(t, r.Click(x+1, y))
	must(t, r.Drag(dx+1, ty))
	f = r.Draw(72)
	t.Logf("over Move the CI:\n%s", marks(f))
	must(t, r.Release())
	if got := order(c, "/board/doing"); got != "c2 c4" {
		t.Errorf("doing: %s", got)
	}
	r.Draw(72)
	x, y, _, _, _ = r.Box("card[/board/todo/0]")
	gx, gy, _, _, _ := r.Box("grace")
	must(t, r.Click(x+1, y))
	must(t, r.Drag(gx+1, gy+1))
	f = r.Draw(72)
	t.Logf("over Grace:\n%s", marks(f))
	must(t, r.Release())
	got := sent(run)
	if len(got) != 2 || got[1] != `assign {"card":"c1","to":"grace"}` {
		t.Errorf("sent %v", got)
	}
	if got := order(c, "/board/todo"); got != "c1 c3" {
		t.Errorf("todo: %s", got)
	}
}

// Where the terminal says how far down its cell the pointer is (DragAt),
// a row's top half is before it and its bottom half after it; a node that
// takes children has thirds, its middle one into it. After an open
// branch is before its first child.
func TestDragHalves(t *testing.T) {
	r, c, _ := dragging(t, "hotty/drag-todo", 40)
	f := r.Draw(40)
	milk, ferns := rowOf(t, f, "Buy oat milk"), rowOf(t, f, "Water the ferns")
	must(t, r.Click(4, milk))
	for _, tc := range []struct {
		sub   float64
		where view.Where
	}{{0.2, view.Before}, {0.8, view.After}} {
		must(t, r.DragAt(4, ferns, tc.sub))
		if d := c.St.Drag; d.Target != "list" || d.At != 2 || d.Where != tc.where {
			t.Errorf("at %.1f down Water the ferns: %+v", tc.sub, *d)
		}
	}
	_, _ = r.Key("Escape")
	must(t, r.Release())

	r, c, _ = dragging(t, "hotty/drag-files", 40)
	f = r.Draw(40)
	readme, src := rowOf(t, f, "README.md"), rowOf(t, f, "src")
	must(t, r.Click(6, readme))
	for _, tc := range []struct {
		sub   float64
		at    int
		where view.Where
	}{{0.1, 3, view.Before}, {0.5, 3, view.Into}, {0.9, 4, view.Before}} {
		must(t, r.DragAt(6, src, tc.sub))
		if d := c.St.Drag; d.Target != "tree" || d.At != tc.at || d.Where != tc.where {
			t.Errorf("at %.1f down src: %+v, want %d %v", tc.sub, *d, tc.at, tc.where)
		}
	}
}

// A press and a release on a card with no move between is its Button's
// click; a press that drags is not.
func TestDragOrClick(t *testing.T) {
	r, _, run := dragging(t, "hotty/drag-kanban", 72)
	x, y, _, _, _ := r.Box("card[/board/todo/2]")
	must(t, r.Click(x+1, y))
	if got := sent(run); len(got) != 0 {
		t.Fatalf("on the press: %v", got)
	}
	must(t, r.Release())
	if got := sent(run); len(got) != 1 || got[0] != `open {"card":"c3"}` {
		t.Errorf("after the release: %v", got)
	}
	r.Draw(72)
	must(t, r.Click(x+1, y))
	must(t, r.Drag(x+1, y-1))
	must(t, r.Drag(x+1, y-2))
	must(t, r.Release())
	if got := sent(run); len(got) != 2 || !strings.HasPrefix(got[1], "moved ") {
		t.Errorf("after a drag: %v", got)
	}
}
