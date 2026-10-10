package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// Drag and drop in cells (profile §3.7, §6.21). A press on an item that
// can be dragged (a reorderable HottyList's, HottyTable's or HottyTree's,
// a List's with reorder, or a drag source) arms a drag; once the pointer
// leaves the item with the button down, the item is lifted
// (view.Controller.Lift) and shows faint, and a line shows where it would
// land: an underline in the accent along the row above the place, or an
// overline along the row under it where there is no row above; into a
// tree's node, the node's row in the accent, reversed. The release drops
// it there; Escape, or a release where no line shows, cancels. A row is a
// cell high: where the terminal says where in the cell the pointer is
// (SGR-Pixels, DragAt's sub), its halves and thirds say before, into and
// after; elsewhere the view's rule does (view.Controller.Over).

// press is a primary press that may start a drag: on element id's item
// (-1: the element itself), whose cells are at; and what a click does
// there that waits for the release, so that a press that drags does not
// do it.
type press struct {
	id    string
	item  int
	at    box
	click func() error
}

// grab is what a press at a cell would drag: a list's, table's or tree's
// item under it (hit, the topmost hit there), a List's item that holds
// it, or a drag source; ok is false for nothing.
func (r *Rendition) grab(col, row int, h *hit) (id string, item int, at box, ok bool) {
	if h != nil && h.opt >= 0 {
		if e := r.c.V.Find(h.id); e != nil && (e.Kind == view.Table || e.Kind == view.RichList || e.Kind == view.Tree) {
			if e.Draggable(h.opt) {
				return e.ID, h.opt, box{h.x, h.y, h.w, h.h}, true
			}
			return "", 0, box{}, false
		}
	}
	path := r.under(col, row)
	for k := len(path) - 1; k >= 0; k-- {
		e := path[k]
		switch {
		case e.Kind == view.Stack && e.Movable && k+1 < len(path):
			i := indexOf(e.Children, path[k+1])
			if b, ok := r.boxes[path[k+1].ID]; ok && e.Draggable(i) {
				return e.ID, i, b, true
			}
		case e.DragType != "":
			if b, ok := r.boxes[e.ID]; ok {
				return e.ID, -1, b, true
			}
		}
	}
	return "", 0, box{}, false
}

// under is the elements drawn at a cell in the last frame, from the
// surface's root (the open Modal's content, while one is open) to the
// innermost: each one's box holds the cell.
func (r *Rendition) under(col, row int) []*view.Element {
	root := r.c.V.Root
	if r.c.V.Overlay != nil {
		root = r.c.V.Overlay
	}
	var path []*view.Element
	for e := root; e != nil; {
		b, ok := r.boxes[e.ID]
		if !ok || !(box{b.x, b.y, b.w, b.h}).holds(col, row) {
			break
		}
		path = append(path, e)
		var next *view.Element
		for _, c := range e.Children {
			if b, ok := r.boxes[c.ID]; ok && b.holds(col, row) {
				next = c
			}
		}
		e = next
	}
	return path
}

func (b box) holds(col, row int) bool {
	return col >= b.x && col < b.x+b.w && row >= b.y && row < b.y+b.h
}

func indexOf(es []*view.Element, e *view.Element) int {
	for i, x := range es {
		if x == e {
			return i
		}
	}
	return -1
}

// dragging reports whether a drag is under way: lifted, not only armed.
func (r *Rendition) dragging() bool { return r.c.St.Drag != nil }

// DragAt is Drag where the terminal reports the pointer in pixels: sub is
// how far down the cell the pointer is, from 0 to 1; -1 where it is not
// known.
func (r *Rendition) DragAt(col, row int, sub float64) error {
	if r.drag != "" {
		return r.slide(col, row)
	}
	p := r.press
	if p == nil {
		return nil
	}
	if !r.dragging() {
		if p.at.holds(col, row) {
			return nil
		}
		p.click = nil
		if !r.c.Lift(p.id, p.item) {
			r.press = nil
			return nil
		}
	}
	r.over(col, row, sub)
	return nil
}

// over finds where the drag would land with the pointer at a cell: the
// innermost element there that takes it (view.Controller.DropTarget), the
// item of it the pointer is over or nearest to, and how far down that
// item (frac).
func (r *Rendition) over(col, row int, sub float64) {
	c := r.c
	path := r.under(col, row)
	if len(path) == 0 {
		c.Over("", -1, -1)
		return
	}
	target, child, item := c.DropTarget(path[len(path)-1].ID)
	t := c.V.Find(target)
	if t == nil {
		c.Over("", -1, -1)
		return
	}
	var items []box
	switch {
	case t.Kind == view.Table || t.Kind == view.RichList || t.Kind == view.Tree:
		at := map[int]box{}
		for _, h := range r.hits {
			if h.id == t.ID && h.opt >= 0 {
				at[h.opt] = box{h.x, h.y, h.w, h.h}
			}
		}
		item, items = -1, nil
		var order []int
		for i := range len(t.RowIDs) {
			if b, ok := at[i]; ok {
				order = append(order, i)
				items = append(items, b)
			}
		}
		k, frac := nearest(items, row, sub)
		if k >= 0 {
			item = order[k]
		}
		c.Over(t.ID, item, frac)
		return
	case t.Kind == view.Stack:
		for _, ch := range t.Children {
			items = append(items, r.boxes[ch.ID])
		}
		if child != "" {
			b := r.boxes[child]
			c.Over(t.ID, item, rowFrac(b, row, sub))
			return
		}
		k, frac := nearest(items, row, sub)
		c.Over(t.ID, k, frac)
		return
	}
	c.Over(t.ID, -1, -1)
}

// nearest is the item, among boxes in order down, the pointer at row (sub
// down it) is over, and how far down it; between two, the one above at
// its bottom; above the first, the first at its top; below the last, the
// last at its bottom. -1 when there are none.
func nearest(items []box, row int, sub float64) (int, float64) {
	if len(items) == 0 {
		return -1, -1
	}
	for k, b := range items {
		switch {
		case row < b.y && k == 0:
			return 0, 0
		case row < b.y:
			return k - 1, 1
		case row < b.y+b.h:
			return k, rowFrac(b, row, sub)
		}
	}
	return len(items) - 1, 1
}

// rowFrac is how far down an item of cells b the pointer is at row: by
// its pixels where known (sub); else by its rows, where it has more than
// one and the pointer is not on the middle one; else -1.
func rowFrac(b box, row int, sub float64) float64 {
	if b.h <= 0 {
		return -1
	}
	if sub >= 0 {
		return (float64(row-b.y) + sub) / float64(b.h)
	}
	if b.h > 1 && !(b.h%2 == 1 && row-b.y == b.h/2) {
		return (float64(row-b.y) + 0.5) / float64(b.h)
	}
	return -1
}

// release ends a press: a drag drops where its line shows, and a press
// that did not drag does what its click waited for.
func (r *Rendition) release() error {
	p := r.press
	r.press = nil
	switch {
	case r.dragging():
		return r.c.Drop()
	case p != nil && p.click != nil:
		return p.click()
	}
	return nil
}

// cancelDrag drops a drag and the press under way: Escape.
func (r *Rendition) cancelDrag() bool {
	if r.press == nil && !r.dragging() {
		return false
	}
	r.press = nil
	r.c.CancelDrag()
	return true
}

// paintDrag shows a drag on the frame: what it lifted faint, and where it
// would land.
func (l *layout) paintDrag(cv *canvas) {
	r := l.r
	d := r.c.St.Drag
	if d == nil {
		return
	}
	faint := func(c *Cell) { c.Attr |= Faint }
	for _, b := range r.lifted(d) {
		cv.restyle(b.x, b.y, b.w, b.h, faint)
	}
	t := r.c.V.Find(d.Target)
	if t == nil {
		return
	}
	if d.At < 0 {
		if b, ok := r.boxes[t.ID]; ok && t.Kind != view.Stack && t.Kind != view.Table && t.Kind != view.RichList && t.Kind != view.Tree {
			// A drop target: the accent, reversed, over it.
			cv.restyle(b.x, b.y, b.w, b.h, func(c *Cell) { c.Role, c.Attr = Accent, c.Attr|Reverse })
			return
		}
		// An empty list: a line along its first row.
		if b, ok := r.boxes[t.ID]; ok {
			dragLine(cv, box{b.x, b.y, b.w, 1}, view.Before)
		}
		return
	}
	b, ok := r.itemBox(t.ID, d.At)
	if !ok {
		return
	}
	if d.Where == view.Into {
		x := b.x + treeIndent + treeGuide*t.Nodes[d.At].Level
		cv.restyle(x, b.y, b.x+b.w-x, 1, func(c *Cell) { c.Role, c.Attr = Accent, c.Attr|Reverse })
		return
	}
	if t.Kind == view.Tree {
		// The line starts where the item's level does.
		x := b.x + treeIndent + treeGuide*t.Nodes[d.At].Level
		b = box{x, b.y, b.x + b.w - x, b.h}
	}
	dragLine(cv, b, d.Where)
}

// lifted is the cells of what a drag lifted, as last drawn: its item, and
// a tree's node's rows under it that show.
func (r *Rendition) lifted(d *view.Drag) []box {
	b, ok := r.itemBox(d.Source, d.Item)
	if !ok {
		return nil
	}
	out := []box{b}
	if e := r.c.V.Find(d.Source); e.Kind == view.Tree && d.Item >= 0 {
		for i := d.Item + 1; i < len(e.Nodes) && e.Nodes[i].Level > e.Nodes[d.Item].Level; i++ {
			if b, ok := r.itemBox(e.ID, i); ok {
				out = append(out, b)
			}
		}
	}
	return out
}

// itemBox is the cells of element id's item (-1: of the element itself),
// as last drawn.
func (r *Rendition) itemBox(id string, item int) (box, bool) {
	e := r.c.V.Find(id)
	if e == nil {
		return box{}, false
	}
	switch {
	case item < 0:
		b, ok := r.boxes[id]
		return b, ok
	case e.Kind == view.Stack:
		if item < len(e.Children) {
			b, ok := r.boxes[e.Children[item].ID]
			return b, ok
		}
		return box{}, false
	}
	for _, h := range r.hits {
		if h.id == id && h.opt == item {
			return box{h.x, h.y, h.w, h.h}, true
		}
	}
	return box{}, false
}

// dragLine draws a drag's line before or after an item of cells b, across
// its columns: an underline in the accent along the row above the place
// (the item above, a blank row between items, a list's blank row under
// its status line, a table's rule, a List's heading), which leaves what is
// in it as it is. At the frame's top, where there is no row above, it is
// an overline along the row under the place, the row's text in the
// accent: an overline takes the text's colour.
func dragLine(cv *canvas, b box, where view.Where) {
	above := b.y - 1
	if where == view.After {
		above = b.y + b.h - 1
	}
	if above >= 0 {
		cv.restyle(b.x, above, b.w, 1, func(c *Cell) { c.Attr |= Underline; c.Line, c.LineSet = Accent, true })
		return
	}
	cv.restyle(b.x, above+1, b.w, 1, func(c *Cell) { c.Attr |= Overline; c.Role, c.Mix = Accent, 0 })
}
