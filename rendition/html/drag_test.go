package html

import (
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// onHost is a drag and drop story on a host: the host, the rendition,
// the story's run, and pump, which sends the host the rendition's deltas
// and plays the events the host has sent since.
type onHost struct {
	t    *testing.T
	h    *hottytest.Host
	r    *Rendition
	run  *story.Run
	seen int
}

func dragStory(t *testing.T, name string) *onHost {
	t.Helper()
	run, err := story.Start(story.Find(name))
	if err != nil {
		t.Fatal(err)
	}
	x := &onHost{t: t, h: hottytest.New(t), r: New(run.Surfaces()[0].C, "drag"), run: run}
	x.send(hotty.Doc("drag", x.r.Doc()), hotty.Place("drag", hotty.Placement{Cols: 72, Rows: 24}))
	x.pump()
	return x
}

func (x *onHost) send(cmds ...string) {
	x.t.Helper()
	for _, c := range cmds {
		_, _ = io.WriteString(x.h, c)
	}
	if errs := x.h.Errors(); len(errs) > 0 {
		x.t.Fatalf("the host refused: %v", errs)
	}
}

func (x *onHost) pump() {
	x.t.Helper()
	x.send(x.r.Update()...)
	for evs := x.h.Events(); x.seen < len(evs); evs = x.h.Events() {
		for _, ev := range evs[x.seen:] {
			x.seen++
			if err := x.r.Event(ev); err != nil {
				x.t.Fatalf("%s %s: %v", ev.Kind, ev.Target, err)
			}
			x.send(x.r.Update()...)
		}
	}
}

// do runs a step of a drag on the host, then pumps.
func (x *onHost) do(err error) {
	x.t.Helper()
	if err != nil {
		x.t.Fatal(err)
	}
	x.pump()
}

// classes are the classes of the element with a DOM id, as the host has
// them.
func (x *onHost) classes(id string) []string {
	c, _ := x.h.Surface("drag").Attr(id, "class")
	return strings.Fields(c)
}

// actions are the names of the actions the run sent.
func (x *onHost) actions() []string {
	var out []string
	for _, e := range x.run.Actions() {
		var m struct {
			Action struct {
				Name string `json:"name"`
			} `json:"action"`
		}
		_ = json.Unmarshal(e.JSON, &m)
		out = append(out, m.Action.Name)
	}
	return out
}

// ids are the ids of a list's items in the data model, at a path: each
// item's value, key or id.
func ids(c *view.Controller, path string) string {
	var out []string
	items, _ := c.S.Data.Value(path).([]any)
	for _, it := range items {
		m, _ := it.(map[string]any)
		for _, k := range []string{"value", "key", "id"} {
			if s, ok := m[k].(string); ok {
				out = append(out, s)
				break
			}
		}
	}
	return strings.Join(out, " ")
}

func item(id string, i int) string { return partID(id, partItem+strconv.Itoa(i)) }

// A reorderable HottyList's items opt in to drags (SPEC §9.1), so a press
// on one selects no text. A press and a move onto another item lifts it,
// faint, and a line shows where it would land, after the item it is over
// when it comes from above; the release drops it there, writes the items
// and sends moved. No click follows a release elsewhere.
func TestDragOnHost(t *testing.T) {
	x := dragStory(t, "hotty/drag-todo")
	c := x.r.C
	s := x.h.Surface("drag")
	if on, _ := s.Attr(item("list", 0), "data-on"); on != "click drag" {
		t.Fatalf("an item's data-on: %q", on)
	}
	x.do(x.h.DragStart("drag", item("list", 0), 4, 3))
	if c.St.Drag != nil {
		t.Fatalf("lifted on the press: %+v", *c.St.Drag)
	}
	x.do(x.h.DragMove(item("list", 2), 4, 5))
	if d := c.St.Drag; d == nil || d.Target != "list" || d.At != 2 || d.Where != view.After {
		t.Fatalf("over Water the ferns: %+v, want after it", d)
	}
	if got := x.classes(item("list", 0)); !slices.Contains(got, "k-lifted") {
		t.Errorf("the lifted item's classes: %v", got)
	}
	if got := x.classes(item("list", 2)); !slices.Contains(got, "k-drop-after") {
		t.Errorf("the classes of the item it would land after: %v", got)
	}
	if got := x.classes(surfaceID); !slices.Contains(got, "k-dragging") {
		t.Errorf("the surface's classes while dragging, which turn hover off: %v", got)
	}
	x.do(x.h.DragEnd(item("list", 2), 4, 5))
	if got := x.classes(surfaceID); slices.Contains(got, "k-dragging") {
		t.Errorf("the surface's classes after the drop: %v", got)
	}
	if got := ids(c, "/todos"); got != "plumber ferns milk passport backup dentist" {
		t.Errorf("after the drop: %s", got)
	}
	if got := x.actions(); !slices.Equal(got, []string{"moved"}) {
		t.Errorf("sent %v", got)
	}
	if got := x.classes(item("list", 2)); slices.Contains(got, "k-lifted") || slices.Contains(got, "k-drop-after") {
		t.Errorf("after the drop, the item's classes: %v", got)
	}
	x.check()
}

// A press and a release on an item with no move between is a click,
// which selects it; a drag that comes back where it began lands nothing,
// and the click after it is the drag's: a branch it lifted stays open.
// Escape cancels a drag, and the rest of it is nothing.
func TestDragOrClickOnHost(t *testing.T) {
	x := dragStory(t, "hotty/drag-todo")
	c := x.r.C
	x.do(x.h.DragStart("drag", item("list", 1), 4, 4))
	x.do(x.h.DragEnd(item("list", 1), 4, 4))
	if c.St.Drag != nil || ids(c, "/todos") != "milk plumber ferns passport backup dentist" {
		t.Errorf("a press and a release moved: %s", ids(c, "/todos"))
	}
	if sel := c.S.Data.Value("/selected"); sel != "plumber" {
		t.Errorf("a click selected %v", sel)
	}

	x.do(x.h.DragStart("drag", item("list", 1), 4, 4))
	x.do(x.h.DragMove(item("list", 3), 4, 6))
	if _, ok, _ := x.r.Key("Escape"); !ok || c.St.Drag != nil {
		t.Fatalf("Escape: took it %v, drag %+v", ok, c.St.Drag)
	}
	x.do(x.h.DragMove(item("list", 4), 4, 7))
	x.do(x.h.DragEnd(item("list", 4), 4, 7))
	if c.St.Drag != nil || ids(c, "/todos") != "milk plumber ferns passport backup dentist" {
		t.Errorf("after Escape: %s", ids(c, "/todos"))
	}

	x = dragStory(t, "hotty/drag-files")
	c = x.r.C
	tree := c.V.Find("tree")
	src := slices.IndexFunc(tree.Nodes, func(n view.TreeNode) bool { return n.Label == "src" })
	readme := slices.IndexFunc(tree.Nodes, func(n view.TreeNode) bool { return n.Label == "README.md" })
	node := func(i int) string { return partID("tree", partNode+strconv.Itoa(i)) }
	x.do(x.h.DragStart("drag", node(src), 4, 3))
	x.do(x.h.DragMove(node(readme), 4, 9))
	x.do(x.h.DragMove(node(src), 4, 3))
	x.do(x.h.DragEnd(node(src), 4, 3))
	if !c.V.Find("tree").Nodes[src].Open {
		t.Errorf("the click after a drag that came back closed src")
	}
	if got := x.actions(); len(got) != 0 {
		t.Errorf("sent %v", got)
	}
}

// A HottyTree's node goes into a folder it is dropped on, which shows in
// the accent, reversed, while it would; a HottyTable's row takes the
// place of the one it is dropped on.
func TestDragTreeAndTableOnHost(t *testing.T) {
	x := dragStory(t, "hotty/drag-files")
	c := x.r.C
	tree := c.V.Find("tree")
	at := func(label string) string {
		i := slices.IndexFunc(tree.Nodes, func(n view.TreeNode) bool { return n.Label == label })
		return partID("tree", partNode+strconv.Itoa(i))
	}
	x.do(x.h.DragStart("drag", at("README.md"), 4, 9))
	x.do(x.h.DragMove(at("archive"), 4, 8))
	if d := c.St.Drag; d == nil || d.Target != "tree" || d.Where != view.Into {
		t.Fatalf("over archive: %+v, want into it", d)
	}
	if got := x.classes(at("archive")); !slices.Contains(got, "k-drop-into") {
		t.Errorf("archive's classes: %v", got)
	}
	x.do(x.h.DragEnd(at("archive"), 4, 8))
	if files, _ := json.Marshal(c.S.Data.Value("/files/2")); !strings.Contains(string(files), `"label":"README.md"`) {
		t.Errorf("archive: %s", files)
	}
	x.check()

	x = dragStory(t, "hotty/drag-backlog")
	c = x.r.C
	row := func(i int) string { return partID("table", partRow+strconv.Itoa(i)) }
	if on, _ := x.h.Surface("drag").Attr(row(3), "data-on"); on != "click drag" {
		t.Fatalf("a row's data-on: %q", on)
	}
	x.do(x.h.DragStart("drag", row(3), 4, 5))
	x.do(x.h.DragMove(row(1), 4, 3))
	if got := x.classes(row(1)); !slices.Contains(got, "k-drop-before") {
		t.Errorf("the classes of the row it would land before: %v", got)
	}
	x.do(x.h.DragEnd(row(1), 4, 3))
	if got := ids(c, "/backlog"); got != "KIT-31 KIT-22 KIT-27 KIT-35 KIT-40" {
		t.Errorf("after the drop: %s", got)
	}
	if got := x.actions(); !slices.Equal(got, []string{"reprioritized"}) {
		t.Errorf("sent %v", got)
	}
}

// A List's children opt in to drags: a card goes from one column to
// another, before the card it is dropped on, with a line before it.
func TestDragKanbanOnHost(t *testing.T) {
	x := dragStory(t, "hotty/drag-kanban")
	c := x.r.C
	card, onto := DOMID("card[/board/todo/1]"), DOMID("card[/board/doing/0]")
	if on, _ := x.h.Surface("drag").Attr(card, "data-on"); on != "drag" {
		t.Fatalf("a card's data-on: %q", on)
	}
	x.do(x.h.DragStart("drag", card, 4, 5))
	x.do(x.h.DragMove(onto, 30, 3))
	if got := x.classes(onto); !slices.Contains(got, "k-drop-before") {
		t.Errorf("the classes of the card it would land before: %v", got)
	}
	if got := x.classes(card); !slices.Contains(got, "k-lifted") {
		t.Errorf("the lifted card's classes: %v", got)
	}
	x.do(x.h.DragEnd(onto, 30, 3))
	if got := ids(c, "/board/doing"); got != "c2 c4" {
		t.Errorf("doing: %s", got)
	}
	if got := ids(c, "/board/todo"); got != "c1 c3" {
		t.Errorf("todo: %s", got)
	}
	x.check()
}

// check is that the deltas the host had made the document a new one
// would be.
func (x *onHost) check() {
	x.t.Helper()
	got := x.h.Surface("drag").HTML()
	h2 := hottytest.New(x.t)
	_, _ = io.WriteString(h2, hotty.Doc("drag", New(x.r.C, "drag").Doc()))
	if want := h2.Surface("drag").HTML(); got != want {
		x.t.Errorf("the deltas made\n%s\nthe document is\n%s", got, want)
	}
}

// A double click on a host, whose second click has count 2 (SPEC §9):
// on a tree's branch it closes it, as its first click does, and leaves it
// so; on a list's item not selected, it selects it and acts on it once.
func TestDoubleClickOnHost(t *testing.T) {
	x := dragStory(t, "hotty/drag-files")
	c := x.r.C
	tree := c.V.Find("tree")
	src := slices.IndexFunc(tree.Nodes, func(n view.TreeNode) bool { return n.Label == "src" })
	x.do(x.h.DoubleClick("drag", partID("tree", partNode+strconv.Itoa(src))))
	if c.V.Find("tree").Nodes[src].Open {
		t.Errorf("a double click on src left it open")
	}
	x = dragStory(t, "hotty/drag-backlog")
	c = x.r.C
	x.do(x.h.DoubleClick("drag", partID("table", partRow+"2")))
	if sel := c.S.Data.Value("/selected"); sel != "KIT-35" {
		t.Errorf("a double click on KIT-35 selected %v", sel)
	}
}
