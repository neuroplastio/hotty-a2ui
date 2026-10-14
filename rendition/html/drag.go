package html

import (
	"slices"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Drag and drop on a host (profile §6.21, SPEC §9.1). What a press may
// drag has drag in its data-on, which also keeps the press from selecting
// text: a HottyList's item, a HottyTable's row or a HottyTree's node that
// can be dragged (view.Element.Draggable), a moving List's child, a drag
// source. Its dragstart arms a drag; once the pointer leaves it with the
// button down, as the host's drag names another element or none, the
// item is lifted (view.Controller.Lift) and shows faint, and a line shows
// where it would land: in what takes it around the element the host
// names (view.Controller.DropTarget), against the item it names. What
// takes a drag has drop in its data-on, so that the host names it though
// a press does not drag it (SPEC §9.1): a list's box, which an empty
// list takes an item by, and a drop target. An item that moves also has
// data-steps, its halves ("0 1"), or a tree's node that takes children
// its before, into and after ("0 2"), and the host says which the
// pointer is in (tx, ty), which the view's halves and thirds take
// (view.Controller.Over); a host that does not, or a step of no item,
// leaves the item to take the place of the one it is over, as in cells
// without pixels. The release drops it there; Escape cancels it (Key). A
// press that never left its item is a click, which the host sends after
// dragend; one that lifted and came back to it is not: that click is the
// drag's.

// press is a drag armed or under way: the DOM id its dragstart named,
// what it drags (an element's item, as Lift has it), and whether the item
// is lifted; done once Escape cancelled it, or nothing could be lifted,
// and its events until dragend are nothing.
type press struct {
	t            string
	id           string
	item         int
	lifted, done bool
}

// moveDrag takes the events of a drag of an item (above). done reports
// that ev was one of them.
func (r *Rendition) moveDrag(ev hotty.Event) (done bool, err error) {
	c := r.C
	switch ev.Kind {
	case hotty.EventDragStart:
		r.press = nil
		id, item, ok := r.grab(ev.Target)
		if !ok {
			return false, nil
		}
		r.press = &press{t: ev.Target, id: id, item: item}
		return true, nil
	case hotty.EventDrag, hotty.EventDragEnd:
		p := r.press
		if p == nil {
			return false, nil
		}
		end := ev.Kind == hotty.EventDragEnd
		if end {
			r.press = nil
		}
		switch {
		case p.done:
			return true, nil
		case !p.lifted && ev.Target == p.t:
			return true, nil
		case !p.lifted:
			if !c.Lift(p.id, p.item) {
				p.done = true
				return true, nil
			}
			p.lifted = true
		}
		r.over(ev)
		if !end {
			return true, nil
		}
		r.dropped = ev.Target == p.t
		return true, c.Drop()
	}
	return false, nil
}

// grab is what a press on the element with DOM id t drags: a list's,
// table's or tree's item, the element itself as a drag source, or a
// moving List's child; ok is false for nothing.
func (r *Rendition) grab(t string) (id string, item int, ok bool) {
	c := r.C
	vid, part, ok := viewID(t)
	e := c.V.Find(vid)
	if !ok || e == nil {
		return "", 0, false
	}
	if i := itemOf(e, part); i >= 0 {
		return e.ID, i, e.Draggable(i)
	}
	if part != "" && part != partWrap {
		return "", 0, false
	}
	if e.Draggable(-1) {
		return e.ID, -1, true
	}
	if anc := c.Ancestors(vid); len(anc) > 0 {
		s := anc[len(anc)-1]
		if i := slices.Index(s.Children, e); s.Kind == view.Stack && i >= 0 && s.Draggable(i) {
			return s.ID, i, true
		}
	}
	return "", 0, false
}

// over sets where the drag would land with the pointer over the element
// ev's t names ("": over nothing that takes a drag), at the step of it
// the host says (stepFrac). Over a list's box between its items, the line
// stays where it was.
func (r *Rendition) over(ev hotty.Event) {
	c := r.C
	vid, part, ok := viewID(ev.Target)
	if !ok {
		c.Over("", -1, -1)
		return
	}
	target, child, item := c.DropTarget(vid)
	e := c.V.Find(target)
	frac := -1.0
	switch {
	case e != nil && rows(e):
		// Over the list's own box, its item the host names; over its
		// container, past its last (DropTarget's item; -1 elsewhere).
		if e.ID == vid {
			item = itemOf(e, part)
		}
		frac = stepFrac(ev, dropSteps(e, item), false)
	case e != nil && e.Kind == view.Stack && child == vid:
		frac = stepFrac(ev, dropSteps(e, item), e.Dir == view.Horizontal)
	}
	if d := c.St.Drag; item < 0 && d != nil && d.Target == target && e != nil && (rows(e) || e.Kind == view.Stack) {
		return
	}
	c.Over(target, item, frac)
}

// dropSteps is how many steps element e's item item is cut into for a
// drop (data-steps): 2 for a tree's node that takes children (before,
// into, after), 1 for any other item of a list whose items move (before
// and after), 0 for none.
func dropSteps(e *view.Element, item int) int {
	switch {
	case !e.Movable || item < 0:
		return 0
	case e.Kind == view.Tree && item < len(e.Nodes) && e.Nodes[item].Takes:
		return 2
	}
	return 1
}

// stepFrac is how far down t (across it, in a Row) the pointer is, from
// the step of k the host says it is at (SPEC §9.1, ty or tx): the middle
// of what that step spans, a step being the nearest whole number of k
// times the fraction. -1 where the host says none.
func stepFrac(ev hotty.Event, k int, across bool) float64 {
	d := ev.Drag
	if d == nil || k <= 0 {
		return -1
	}
	step, has := d.TY, d.HasTY
	if across {
		step, has = d.TX, d.HasTX
	}
	if !has {
		return -1
	}
	lo := max(0, (float64(step)-0.5)/float64(k))
	hi := min(1, (float64(step)+0.5)/float64(k))
	return (lo + hi) / 2
}

// rows reports whether an element's items are rows the markup draws with
// ids of their own: a HottyList's, a HottyTable's, a HottyTree's.
func rows(e *view.Element) bool {
	return e.Kind == view.RichList || e.Kind == view.Table || e.Kind == view.Tree
}

// itemOf is the item of element e a part names ("i3": a HottyList's item
// 3, or a HottyFilePicker's entry), or -1.
func itemOf(e *view.Element, part string) int {
	prefix := ""
	switch e.Kind {
	case view.RichList, view.FilePicker:
		prefix = partItem
	case view.Table:
		prefix = partRow
	case view.Tree:
		prefix = partNode
	case view.DiffView:
		prefix = partHunk
	}
	rest, ok := strings.CutPrefix(part, prefix)
	if prefix == "" || !ok {
		return -1
	}
	i, err := strconv.Atoi(rest)
	if err != nil || i < 0 {
		return -1
	}
	return i
}

// drags marks node n, element e's item (-1: e itself; a List's child is
// the List's item), for drag and drop: drag in its data-on where a press
// may drag it; drop where the host names it to say where a drag is (an
// item that moves, with its steps, dropSteps; a list's box whose items
// move; a drop target); and while a drag is under way, k-lifted on what
// it lifted (a tree's node with the nodes under it), and where it would
// land: k-drop-before or k-drop-after on the item the line is along,
// k-drop-into on a tree's node it would go into, k-drop-on on a target as
// a whole. A web page has no program to drag for.
func (m *markup) drags(n *node, e *view.Element, item int) {
	if m.page {
		return
	}
	if e.Draggable(item) {
		opt(n, "drag")
	}
	if k := dropSteps(e, item); k > 0 {
		opt(n, "drop")
		if _, has := n.attr("data-steps"); !has {
			steps := "0 " + strconv.Itoa(k)
			if e.Kind == view.Stack && e.Dir == view.Horizontal {
				steps = strconv.Itoa(k)
			}
			n.set("data-steps", steps)
		}
	}
	if item < 0 && (len(e.Accepts) > 0 || e.Movable && (rows(e) || e.Kind == view.Stack) || view.ListOf(e) != nil) {
		// A container of a list whose items move takes its drags too
		// (view.Controller.DropTarget): the host names it where the
		// pointer is over its heading or the room under the list.
		opt(n, "drop")
	}
	d := m.drag
	if d == nil {
		return
	}
	class := ""
	switch {
	case d.Source == e.ID && (d.Item == item || item >= 0 && e.Kind == view.Tree && under(e, item, d.Item)):
		class = "k-lifted"
	case d.Target != e.ID || d.At != item:
	case item < 0:
		class = "k-drop-on"
	case d.Where == view.Into:
		class = "k-drop-into"
	case d.Where == view.After:
		class = "k-drop-after"
	default:
		class = "k-drop-before"
	}
	if class != "" {
		have, _ := n.attr("class")
		n.set("class", strings.TrimSpace(have+" "+class))
	}
}

// opt adds an event to what node n opts in to (data-on).
func opt(n *node, event string) {
	on, _ := n.attr("data-on")
	if !slices.Contains(strings.Fields(on), event) {
		n.set("data-on", strings.TrimSpace(on+" "+event))
	}
}

// under reports whether a tree's node i is inside node top.
func under(e *view.Element, i, top int) bool {
	if top < 0 || i >= len(e.Nodes) {
		return false
	}
	for j := e.Nodes[i].Parent; j >= 0; j = e.Nodes[j].Parent {
		if j == top {
			return true
		}
	}
	return false
}
