package view

import (
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// Drag and drop (profile §6.21; vault knowledge/drag-and-drop.md): model C,
// both, with P2, a line. A HottyList, a HottyTable, a HottyTree and a List
// of a template move their items themselves, within themselves and between
// two of a kind whose items have one type, as a Slider writes its value,
// and then tell the agent (onMove). Besides, any component can be a drag
// source or a drop target (io_neuroplast_hotty.drag and .drop): a drop
// there is the target's action, whose meaning is the agent's. A rendition
// finds what the pointer is over (DropTarget) and how far down its item;
// the view decides where the item would land (Over), shows nothing where
// it would not move, and makes the move (Drop). The keys move the selected
// item a place (MoveKey).

// Where is where a dragged item lands against the item the pointer is
// over.
type Where int

// The places an item lands.
const (
	// Before the item: the line above it.
	Before Where = iota
	// After the item: the line under it.
	After
	// Into the item, a HottyTree's node that takes children: as its last
	// child.
	Into
)

// Drag is a drag the user is making (State.Drag): what it lifted, what it
// carries, and where it would land now.
type Drag struct {
	// Source is the element dragged from, and Item its item (an index into
	// its RowIDs, or a List's children); Item is -1 when the element itself
	// is dragged, a drag source.
	Source string
	Item   int
	// Type and Value are what it carries: the items' type and the item's
	// data, or a drag source's type and value.
	Type  string
	Value any
	// Target is the element it would land in now: "" over nothing that
	// takes it, or where it would not move, where a release lands nothing.
	// At and Where are the line: before or after the target's item At
	// (Before, After), or into it (Into); At is -1 for the target as a
	// whole (an empty list's end, a drop target).
	Target string
	At     int
	Where  Where
}

// Moved is a list whose items the user moved while it is not bound
// (State.Moved): the literal it was, and what the moves made it. A
// literal that changes drops it.
type Moved struct {
	From any
	To   []any
}

// isRows reports whether an element's items are its RowIDs: a HottyList's,
// a HottyTable's, a HottyTree's.
func isRows(e *Element) bool { return e.Kind == RichList || e.Kind == Table || e.Kind == Tree }

// moves reports whether an element moves its items itself: a list, a table
// or a tree that is reorderable, or a List of a template with reorder.
func moves(e *Element) bool { return e.Movable && (isRows(e) || e.Kind == Stack) }

// filtering reports whether a filter applies to a HottyList or a
// HottyTree: then its rows are not in their order (a list's are ranked),
// and nothing moves.
func (e *Element) filtering() bool {
	return (e.Kind == RichList || e.Kind == Tree) && e.Query.Text != ""
}

// Draggable reports whether a press on an element's item (item ≥ 0) or on
// the element itself (item -1) may start a drag (Lift).
func (e *Element) Draggable(item int) bool {
	switch {
	case item < 0:
		return e.DragType != ""
	case isRows(e):
		return (e.Movable || e.ItemType != "") && item < len(e.RowIDs) && !e.filtering()
	case e.Kind == Stack:
		return e.Movable && item < len(e.Children)
	}
	return false
}

// Lift starts a drag of element id's item (an index into its RowIDs, or a
// List's children), or with item -1 of element id itself, a drag source. A
// HottyList's, HottyTable's or HottyTree's item is selected as it is
// lifted. ok is false when it cannot be dragged; there is no drag then.
func (c *Controller) Lift(id string, item int) (ok bool) {
	c.St.Drag = nil
	e := c.V.Find(id)
	if e == nil || !e.Draggable(item) {
		return false
	}
	d := &Drag{Source: id, Item: item, At: -1}
	switch {
	case item < 0:
		d.Type, d.Value = e.DragType, c.dragValue(e)
	default:
		d.Type, d.Value = e.ItemType, a2ui.Clone(c.itemData(e, item))
		if isRows(e) && e.SelectedRow() != item {
			if e.Kind == Tree {
				_ = c.SelectNode(id, item)
			} else {
				_ = c.SelectRow(id, item)
			}
		}
	}
	c.St.Drag = d
	return true
}

// CancelDrag ends the drag without a drop: Escape, or a release over
// nothing.
func (c *Controller) CancelDrag() {
	if c.St.Drag != nil {
		c.St.Drag = nil
		c.Rebuild()
	}
}

// DropTarget is what would take the drag over element id, the innermost
// element the pointer is over: id itself, or the nearest element around it
// that takes it. For a List, also its child the pointer is in and the
// child's index; else child is "" and item -1, and a rendition finds a
// HottyList's, HottyTable's or HottyTree's item itself. target is "" when
// nothing takes the drag there.
func (c *Controller) DropTarget(id string) (target, child string, item int) {
	d := c.St.Drag
	e := c.V.Find(id)
	if d == nil || e == nil {
		return "", "", -1
	}
	path := append(c.Ancestors(id), e)
	for k := len(path) - 1; k >= 0; k-- {
		t := path[k]
		if move, drop := c.takes(t, d); !move && !drop {
			continue
		}
		if t.Kind == Stack && k+1 < len(path) {
			return t.ID, path[k+1].ID, slices.Index(t.Children, path[k+1])
		}
		return t.ID, "", -1
	}
	return "", "", -1
}

// takes reports whether an element would take a drag: as a move of its
// items (move: the drag's own list, or another of its kind whose items
// have the drag's type, both moving their items), or as a drop target that
// accepts the drag's type (drop).
func (c *Controller) takes(t *Element, d *Drag) (move, drop bool) {
	src := c.V.Find(d.Source)
	if src != nil && d.Item >= 0 && moves(src) && moves(t) && src.Kind == t.Kind && !t.filtering() &&
		(t.ID == src.ID || d.Type != "" && t.ItemType == d.Type) {
		move = true
	}
	drop = d.Type != "" && slices.Contains(t.Accepts, d.Type) && !(t.ID == d.Source && d.Item < 0)
	return move, drop
}

// Over is the pointer over element id, a target DropTarget found, at its
// item (an index as Lift's; -1 for none: an empty list, or a drop target),
// frac of the way down that item (0 at its top, 1 at its bottom; -1 where
// the rendition cannot tell, a row of cells without the pointer's pixels).
// It sets where the drag would land (Drag.Target, At, Where): before the
// item over its top half and after it over its bottom half; for a
// HottyTree's node that takes children, into it over its middle third.
// Without frac, the item takes the place of the one it is over (after it
// when it comes from above, before it when from below), and goes into a
// node that takes children; in another list, before it. Over a drop target
// it would land there. Nothing is set over what does not take it, where
// the item would not move, and into the item itself or anything inside
// it.
func (c *Controller) Over(id string, item int, frac float64) {
	d := c.St.Drag
	if d == nil {
		return
	}
	d.Target, d.At, d.Where = "", -1, Before
	t := c.V.Find(id)
	if t == nil {
		return
	}
	move, drop := c.takes(t, d)
	switch {
	case move:
		c.overMove(d, t, item, frac)
	case drop:
		d.Target = t.ID
	}
}

// overMove sets where a move would land in t, over its item.
func (c *Controller) overMove(d *Drag, t *Element, item int, frac float64) {
	same := t.ID == d.Source
	n := c.count(t)
	if item < 0 || item >= n {
		if n > 0 {
			return
		}
		// An empty list takes it as its first.
		d.Target, d.At = t.ID, -1
		return
	}
	where := Before
	if t.Kind == Tree {
		node := t.Nodes[item]
		switch {
		case frac < 0 && node.Takes && !(same && item == d.Item):
			where = Into
		case frac < 0 && same && item > d.Item:
			where = After
		case frac >= 0 && node.Takes && frac >= 1.0/3 && frac <= 2.0/3:
			where = Into
		case frac >= 0 && (node.Takes && frac > 2.0/3 || !node.Takes && frac >= 0.5):
			where = After
		}
		// Under an open branch is its first child's place.
		if where == After && node.Kids > 0 && node.Open && item+1 < len(t.Nodes) && t.Nodes[item+1].Parent == item {
			item, where = item+1, Before
		}
	} else {
		switch {
		case frac >= 0 && frac >= 0.5, frac < 0 && same && item > d.Item:
			where = After
		}
	}
	parent, ins := t.landing(item, where)
	if same && !t.lands(d.Item, parent, ins) {
		return
	}
	d.Target, d.At, d.Where = t.ID, item, where
}

// landing is where an item lands in e, before its removal, by the line
// (item, where): its parent node (-1 for a list, or a tree's roots) and its
// index among the parent's children, or in the list.
func (e *Element) landing(item int, where Where) (parent, ins int) {
	if item < 0 {
		return -1, 0
	}
	if e.Kind != Tree {
		if where == After {
			return -1, item + 1
		}
		return -1, item
	}
	switch where {
	case Into:
		return item, e.Nodes[item].Kids
	case After:
		return e.Nodes[item].Parent, e.sibling(item) + 1
	}
	return e.Nodes[item].Parent, e.sibling(item)
}

// lands reports whether e's own item moves when it lands at (parent, ins):
// it goes somewhere else, and not into itself or anything inside it.
func (e *Element) lands(item, parent, ins int) bool {
	if e.Kind != Tree {
		return ins != item && ins != item+1
	}
	if parent == item || parent >= 0 && e.inside(parent, item) {
		return false
	}
	if parent == e.Nodes[item].Parent {
		s := e.sibling(item)
		return ins != s && ins != s+1
	}
	return true
}

// count is how many items an element has: its rows, or a List's children.
func (c *Controller) count(e *Element) int {
	if isRows(e) {
		return len(e.RowIDs)
	}
	return len(e.Children)
}

// Drop ends the drag where it would land: a move of the item, or the drop
// target's action. Over nothing, it lands nothing.
func (c *Controller) Drop() error {
	d := c.St.Drag
	c.St.Drag = nil
	if d == nil || d.Target == "" {
		c.Rebuild()
		return nil
	}
	t := c.V.Find(d.Target)
	src := c.V.Find(d.Source)
	if t == nil {
		c.Rebuild()
		return nil
	}
	if move, _ := c.takes(t, d); move && src != nil {
		parent, ins := t.landing(d.At, d.Where)
		return c.move(src, d.Item, t, parent, ins)
	}
	return c.dropOn(t, d.Value)
}

// dropOn runs a drop target's drop: what the drag carries is written where
// its value is bound, and then its action runs.
func (c *Controller) dropOn(t *Element, v any) error {
	n := c.V.Node(t.ID)
	spec, _ := hottyExt(n)["drop"].(map[string]any)
	var err error
	if p, ok := a2ui.IsBinding(spec["value"]); ok && v != nil {
		err = c.S.Write(c.S.Context(n.Scope).Path(p), v)
	}
	if a, ok := spec["action"].(map[string]any); ok {
		if err2 := c.S.Dispatch(a, n.Scope, n.ComponentID, true); err == nil {
			err = err2
		}
	}
	c.Rebuild()
	return err
}

// MoveKey moves the item the keyboard is on a place by a key, as SPEC
// §10.4 names it: Alt+ArrowUp and Alt+ArrowDown move a HottyList's,
// HottyTable's or HottyTree's selected item, or a List's item that holds
// the focused element, before the one above it or after the one below;
// in a HottyTree, Alt+ArrowLeft moves the node out of its branch, after
// it, and Alt+ArrowRight into the node above it among its siblings, as its
// last child, where that takes children. At an end it stays. ok reports
// whether one of these took the key: a key of these on an element that
// moves its items.
func (c *Controller) MoveKey(id, key string) (ok bool, err error) {
	dir := ""
	for _, k := range []string{"ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight"} {
		if SameKey(key, "Alt+"+k) {
			dir = k
		}
	}
	e := c.V.Find(id)
	if dir == "" || e == nil {
		return false, nil
	}
	if !moves(e) || !isRows(e) {
		// A List's item, where the keyboard is inside one.
		anc := append(c.Ancestors(id), e)
		for k := len(anc) - 2; k >= 0; k-- {
			if anc[k].Kind == Stack && moves(anc[k]) {
				return c.moveChild(anc[k], slices.Index(anc[k].Children, anc[k+1]), dir)
			}
		}
		return false, nil
	}
	if e.filtering() || e.Kind != Tree && (dir == "ArrowLeft" || dir == "ArrowRight") {
		return false, nil
	}
	item := e.SelectedRow()
	if item < 0 {
		return true, nil
	}
	if e.Kind != Tree {
		return c.moveChild(e, item, dir)
	}
	n := e.Nodes[item]
	s := e.sibling(item)
	switch dir {
	case "ArrowUp":
		if s > 0 {
			return true, c.move(e, item, e, n.Parent, s-1)
		}
	case "ArrowDown":
		if s < e.siblings(n.Parent)-1 {
			return true, c.move(e, item, e, n.Parent, s+2)
		}
	case "ArrowLeft":
		if p := n.Parent; p >= 0 {
			return true, c.move(e, item, e, e.Nodes[p].Parent, e.sibling(p)+1)
		}
	case "ArrowRight":
		if prev := e.previousSibling(item); prev >= 0 && e.Nodes[prev].Takes {
			return true, c.move(e, item, e, prev, e.Nodes[prev].Kids)
		}
	}
	return true, nil
}

// moveChild moves a flat list's item i a place up or down (dir), if it can.
func (c *Controller) moveChild(e *Element, i int, dir string) (bool, error) {
	n := c.count(e)
	switch {
	case i < 0, dir == "ArrowLeft", dir == "ArrowRight":
		return dir == "ArrowUp" || dir == "ArrowDown", nil
	case dir == "ArrowUp" && i > 0:
		return true, c.move(e, i, e, -1, i-1)
	case dir == "ArrowDown" && i < n-1:
		return true, c.move(e, i, e, -1, i+2)
	}
	return true, nil
}

// sibling is node i's index among its siblings.
func (e *Element) sibling(i int) int {
	n := 0
	for j := range i {
		if e.Nodes[j].Parent == e.Nodes[i].Parent {
			n++
		}
	}
	return n
}

// siblings is how many children node parent has (-1: the roots).
func (e *Element) siblings(parent int) int {
	n := 0
	for _, x := range e.Nodes {
		if x.Parent == parent {
			n++
		}
	}
	return n
}

// previousSibling is the node before node i among its siblings, or -1.
func (e *Element) previousSibling(i int) int {
	for j := i - 1; j >= 0; j-- {
		if e.Nodes[j].Parent == e.Nodes[i].Parent {
			return j
		}
		if e.Nodes[j].Level < e.Nodes[i].Level {
			return -1
		}
	}
	return -1
}

// place is node i's index at each level, from the roots down.
func (e *Element) place(i int) []int {
	var p []int
	for j := i; j >= 0; j = e.Nodes[j].Parent {
		p = append(p, e.sibling(j))
	}
	slices.Reverse(p)
	return p
}

// nodeAt is the node at a place, or -1.
func (e *Element) nodeAt(p []int) int {
	for i := range e.Nodes {
		if slices.Equal(e.place(i), p) {
			return i
		}
	}
	return -1
}

// move moves src's item to dst (src itself, or another of its kind), at
// ins among the children of dst's node parent (-1: a flat list, or a
// tree's roots), ins counted before the item leaves: the items are written
// where they are bound (else kept as the renderer's, State.Moved), the
// item stays selected (and the keyboard goes with it to dst), the move is
// written where moved is bound, and onMove runs: dst's, else src's.
func (c *Controller) move(src *Element, item int, dst *Element, parent, ins int) error {
	from, fromParent := []int{item}, -1
	if src.Kind == Tree {
		from, fromParent = src.place(item), src.Nodes[item].Parent
	}
	var to []int
	if parent >= 0 {
		to = dst.place(parent)
	}
	srcList, dstList := c.items(src), c.items(dst)
	same := src.ID == dst.ID
	if same {
		dstList = srcList
	}
	node, srcList := cut(srcList, from)
	if same {
		dstList = srcList
		if len(to) >= len(from) && slices.Equal(to[:len(from)-1], from[:len(from)-1]) && from[len(from)-1] < to[len(from)-1] {
			to[len(from)-1]--
		}
		if slices.Equal(to, from[:len(from)-1]) && from[len(from)-1] < ins {
			ins--
		}
	}
	dstList = paste(dstList, to, ins, node)
	record := map[string]any{
		"item": c.itemID(src, item, node),
		"from": c.spot(src, fromParent, from[len(from)-1]),
	}
	var err error
	keep := func(e error) {
		if err == nil {
			err = e
		}
	}
	focused := c.St.Keyboard && c.St.Focus == src.ID
	focus := c.St.Focus
	if same {
		keep(c.setItems(src, dstList))
	} else {
		keep(c.setItems(src, srcList))
		keep(c.setItems(dst, dstList))
	}
	c.Rebuild()
	if d := c.V.Find(dst.ID); d != nil {
		dst = d
	}
	record["to"] = c.spot(dst, -1, ins)
	switch {
	case dst.Kind == Tree:
		if i := dst.nodeAt(append(slices.Clone(to), ins)); i >= 0 {
			record["to"] = c.spot(dst, dst.Nodes[i].Parent, ins)
			keep(c.SelectNode(dst.ID, i))
		}
	case isRows(dst):
		keep(c.SelectRow(dst.ID, ins))
	}
	switch {
	case src.Kind == Stack:
		// A template's nodes are keyed by their items' places, which
		// changed: the keyboard stays on what it was on.
		c.St.Focus = refocus(focus, c.listPath(src), c.listPath(dst), from[0], ins, same)
	case focused:
		c.St.Focus = dst.ID
	}
	for _, e := range []*Element{src, dst} {
		if p := c.movedPath(e); p != "" {
			keep(c.S.Write(p, record))
		}
		if same {
			break
		}
	}
	c.Rebuild()
	by := dst
	if c.onMove(dst) == nil {
		by = src
	}
	if a, n := c.onMove(by), c.V.Node(by.ID); a != nil && n != nil {
		if by.Kind == Stack {
			keep(c.S.Dispatch(a, n.Scope, n.ComponentID, true))
		} else {
			keep(c.S.Tree.Invoke(n, "onMove", true))
		}
	}
	c.Rebuild()
	return err
}

// cut takes the item at a place out of a list (a tree's roots, its
// children, theirs): the item, and a copy of the list without it.
func cut(list []any, at []int) (item any, out []any) {
	i := at[len(at)-1]
	out = edit(list, at[:len(at)-1], func(kids []any) []any {
		if i >= len(kids) {
			return kids
		}
		item = kids[i]
		return slices.Delete(kids, i, i+1)
	})
	return item, out
}

// paste is a copy of a list with an item put in at ins among the children
// of the node at a place (nil: the list itself).
func paste(list []any, at []int, ins int, item any) []any {
	return edit(list, at, func(kids []any) []any {
		return slices.Insert(kids, min(max(ins, 0), len(kids)), item)
	})
}

// edit is a copy of a list whose children of the node at a place (nil:
// the list itself) are what fn makes of a copy of them. The nodes on the
// way are copied, so that the list given, which may be the data model's
// or a component's, does not change.
func edit(list []any, at []int, fn func([]any) []any) []any {
	out := slices.Clone(list)
	if len(at) == 0 {
		if out == nil {
			out = []any{}
		}
		return fn(out)
	}
	if at[0] >= len(out) {
		return out
	}
	m, _ := out[at[0]].(map[string]any)
	m = maps.Clone(m)
	if m == nil {
		m = map[string]any{}
	}
	kids, _ := m["children"].([]any)
	m["children"] = edit(kids, at[1:], fn)
	out[at[0]] = m
	return out
}

// items is an element's items now: a list's, a table's rows, a tree's
// roots, a List's template's list. Not to be changed.
func (c *Controller) items(e *Element) []any {
	if e.Kind == Stack {
		l, _ := c.S.Data.Value(c.listPath(e)).([]any)
		return l
	}
	n := c.V.Node(e.ID)
	return (&Builder{S: c.S, St: c.St}).List(n, itemsProp(e))
}

// itemsProp is the prop that holds an element's items.
func itemsProp(e *Element) string {
	if e.Kind == Table {
		return "rows"
	}
	return "items"
}

// setItems writes an element's items: to where they are bound, or a
// List's template's list; else the renderer keeps them (State.Moved).
func (c *Controller) setItems(e *Element, l []any) error {
	if e.Kind == Stack {
		return c.S.Write(c.listPath(e), l)
	}
	n := c.V.Node(e.ID)
	prop := itemsProp(e)
	if bd, ok := n.Props[prop].(a2ui.Bound); ok && bd.Writable() {
		return c.S.Write(bd.Path, l)
	}
	c.St.Moved[e.ID] = Moved{From: (&Builder{S: c.S, St: c.St}).Raw(n, prop), To: l}
	return nil
}

// itemData is an element's item i as data: the object in its list (a
// tree's node with its children).
func (c *Controller) itemData(e *Element, i int) any {
	l := c.items(e)
	if e.Kind != Tree {
		if i < len(l) {
			return l[i]
		}
		return nil
	}
	at := e.place(i)
	var v any = l
	for k, j := range at {
		list, _ := v.([]any)
		if k > 0 {
			m, _ := v.(map[string]any)
			list, _ = m["children"].([]any)
		}
		if j >= len(list) {
			return nil
		}
		v = list[j]
	}
	return v
}

// itemID is how the move's record names the item: its id (a list's item,
// a table's row, a tree's node: its value, else its place before the
// move), or a List's item's data.
func (c *Controller) itemID(e *Element, i int, data any) any {
	if isRows(e) && i < len(e.RowIDs) {
		return e.RowIDs[i]
	}
	return data
}

// spot is where in an element an item is, for the move's record: its
// index; in a tree its parent's id (nil for the roots); where the items
// are bound, their path.
func (c *Controller) spot(e *Element, parent, index int) map[string]any {
	s := map[string]any{"index": float64(index)}
	if e.Kind == Tree {
		if parent >= 0 && parent < len(e.RowIDs) {
			s["parent"] = e.RowIDs[parent]
		} else {
			s["parent"] = nil
		}
	}
	if p := c.boundPath(e); p != "" {
		s["path"] = p
	}
	return s
}

// boundPath is the data path an element's items are bound to: a List's
// template's list, or a bound items or rows; "" for a literal.
func (c *Controller) boundPath(e *Element) string {
	if e.Kind == Stack {
		return c.listPath(e)
	}
	if bd, ok := c.V.Node(e.ID).Props[itemsProp(e)].(a2ui.Bound); ok {
		return bd.Path
	}
	return ""
}

// listPath is the data path of a List's template's list.
func (c *Controller) listPath(e *Element) string {
	n := c.V.Node(e.ID)
	if n == nil {
		return ""
	}
	return templatePath(c.S, n)
}

// templatePath is the absolute data path of a node's children's template,
// or "" when its children are no template.
func templatePath(s *a2ui.Surface, n *a2ui.Node) string {
	comp := s.Components[n.ComponentID]
	if comp == nil {
		return ""
	}
	t, _ := comp.Props["children"].(map[string]any)
	p, ok := t["path"].(string)
	if !ok {
		return ""
	}
	return s.Context(n.Scope).Path(p)
}

// movedPath is where an element's moves are written: its moved prop's
// path, or for a List its reorder's moved; "" when not bound.
func (c *Controller) movedPath(e *Element) string {
	n := c.V.Node(e.ID)
	if n == nil {
		return ""
	}
	if e.Kind == Stack {
		r, _ := hottyExt(n)["reorder"].(map[string]any)
		if p, ok := a2ui.IsBinding(r["moved"]); ok {
			return c.S.Context(n.Scope).Path(p)
		}
		return ""
	}
	if bd, ok := n.Props["moved"].(a2ui.Bound); ok && bd.Writable() {
		return bd.Path
	}
	return ""
}

// onMove is an element's onMove: its prop, or for a List its reorder's;
// nil when it has none.
func (c *Controller) onMove(e *Element) map[string]any {
	n := c.V.Node(e.ID)
	if n == nil {
		return nil
	}
	if e.Kind == Stack {
		r, _ := hottyExt(n)["reorder"].(map[string]any)
		a, _ := r["onMove"].(map[string]any)
		return a
	}
	if n.Props["onMove"] == nil {
		return nil
	}
	return map[string]any{}
}

// dragValue is what a drag source carries: its drag's value, resolved in
// its scope.
func (c *Controller) dragValue(e *Element) any {
	n := c.V.Node(e.ID)
	if n == nil {
		return nil
	}
	spec, _ := hottyExt(n)["drag"].(map[string]any)
	v, err := c.S.Context(n.Scope).Resolve(spec["value"])
	if err != nil {
		return nil
	}
	return v
}

// refocus is the element with the keyboard after a List's item moved, by
// its node key: the item from at index from of the list at path src went
// to index to of the list at path dst (indices after it left), and the
// items around them shifted. A key of a template's node holds its item's
// path, in brackets.
func refocus(key, src, dst string, from, to int, same bool) string {
	open := strings.IndexByte(key, '[')
	end := strings.LastIndexByte(key, ']')
	if open < 0 || end < open {
		return key
	}
	scope := key[open+1 : end]
	index := func(list string) (int, string, bool) {
		rest, ok := strings.CutPrefix(scope, list+"/")
		if !ok {
			return 0, "", false
		}
		tok, tail, _ := strings.Cut(rest, "/")
		i, err := strconv.Atoi(tok)
		if tail != "" {
			tail = "/" + tail
		}
		return i, tail, err == nil
	}
	moved := func(list string, i int, tail string) string {
		return key[:open+1] + list + "/" + strconv.Itoa(i) + tail + key[end:]
	}
	if i, tail, ok := index(src); ok {
		switch {
		case i == from:
			return moved(dst, to, tail)
		case same && i > from && i <= to:
			return moved(src, i-1, tail)
		case same && i < from && i >= to:
			return moved(src, i+1, tail)
		case !same && i > from:
			return moved(src, i-1, tail)
		}
		return key
	}
	if i, tail, ok := index(dst); ok && !same && i >= to {
		return moved(dst, i+1, tail)
	}
	return key
}

// List is a prop's list: a binding's value, a literal as it is, or what
// the user's moves made a literal (State.Moved), while the literal stays
// as it was.
func (b *Builder) List(n *a2ui.Node, prop string) []any {
	raw := b.Raw(n, prop)
	if bd, ok := n.Props[prop].(a2ui.Bound); !ok || !bd.Writable() {
		if m, ok := b.St.Moved[n.Key]; ok {
			if reflect.DeepEqual(m.From, raw) {
				return m.To
			}
			delete(b.St.Moved, n.Key)
		}
	}
	l, _ := raw.([]any)
	return l
}

// dragDrop reads a node's drag and drop extensions: what it is as a drag
// source (io_neuroplast_hotty.drag's type) and the types it takes as a
// drop target (io_neuroplast_hotty.drop's accepts).
func dragDrop(n *a2ui.Node, e *Element) {
	ext := hottyExt(n)
	if d, ok := ext["drag"].(map[string]any); ok {
		e.DragType, _ = d["type"].(string)
	}
	if d, ok := ext["drop"].(map[string]any); ok {
		switch a := d["accepts"].(type) {
		case string:
			e.Accepts = []string{a}
		case []any:
			e.Accepts = strs(a)
		}
	}
}
