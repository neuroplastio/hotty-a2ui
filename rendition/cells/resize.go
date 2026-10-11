package cells

import "github.com/neuroplastio/hotty-a2ui/view"

// A longText's corner (profile §3.5): a grip the pointer drags to resize
// the field along the axes its resize allows (io_neuroplast_hotty.resize),
// as a web textarea's.

// gripGlyph is the grip, in the value's last column on its last row.
const gripGlyph = "◢"

// minGripCols is the fewest columns a grip leaves a field's value: two
// of text and the grip's.
const minGripCols = 3

// fieldSize is the size a longText's grip gave it: the rows and the
// columns of its value, the grip's included; 0 on an axis the grip has
// not set (its lines, 3 to 8, and its box's width).
type fieldSize struct{ rows, cols int }

// resizing is a drag of a longText's grip: the field, where the press
// was, and the field's size then, as drawn.
type resizing struct {
	id       string
	col, row int
	from     fieldSize
}

// gripped reports whether a longText shows its grip: one its resize
// allows any axis of, not disabled, at least minGripCols wide.
func gripped(e *view.Element, vw int) bool {
	return isLongText(e) && e.Resize != "" && e.Resize != "none" && !e.Disabled && vw >= minGripCols
}

// gripAt is the hit of the longText whose grip is at a cell of the last
// frame, on top there; nil for none.
func (r *Rendition) gripAt(col, row int) *hit {
	h := r.hitAt(col, row)
	if h == nil || h.field == nil || !h.field.grip {
		return nil
	}
	if a := h.field; col == a.x+a.w && row == a.y+a.rows-1 {
		return h
	}
	return nil
}

// gripPointer is the pointer's shape over a grip, as CSS names it: the
// axes it resizes along.
func gripPointer(e *view.Element) string {
	switch e.Resize {
	case "vertical":
		return "ns-resize"
	case "horizontal":
		return "ew-resize"
	}
	return "nwse-resize"
}

// grip starts a drag of the grip at a cell, if there is one: it takes
// the press, and nothing else does (the keyboard stays where it is).
func (r *Rendition) grip(col, row int) bool {
	h := r.gripAt(col, row)
	if h == nil {
		return false
	}
	a := h.field
	r.resize = &resizing{id: h.id, col: col, row: row, from: fieldSize{rows: a.rows, cols: a.w + 1}}
	return true
}

// resizeTo is a grip's drag at a cell: the field's value is its size at
// the press, moved as far as the pointer has, on the axes its resize
// allows: at least a row, and minGripCols. Wider than its box, it shows
// as wide as the box (paintField).
func (r *Rendition) resizeTo(col, row int) {
	z := r.resize
	e := r.c.V.Find(z.id)
	if e == nil || !isLongText(e) {
		r.resize = nil
		return
	}
	s := r.sized[z.id]
	if e.Resize != "horizontal" {
		s.rows = max(z.from.rows+row-z.row, 1)
	}
	if e.Resize != "vertical" {
		s.cols = max(z.from.cols+col-z.col, minGripCols)
	}
	r.sized[z.id] = s
}

// fieldRows is how many rows a field's value takes: one; a longText's,
// what its grip gave it, else its lines, at least 3 and at most 8.
func (r *Rendition) fieldRows(e *view.Element) int {
	if !isLongText(e) {
		return 1
	}
	if s := r.sized[e.ID]; s.rows > 0 {
		return s.rows
	}
	v, _ := e.Value.(string)
	return min(max(len(splitClusters(clusters(v))), 3), 8)
}
