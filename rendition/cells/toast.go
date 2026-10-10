package cells

import (
	"github.com/neuroplastio/hotty-a2ui/icons"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// Toasts in cells (profile §3.4, §6.23): each a rounded box in its kind's
// colour, stacked down from the frame's top right corner, the newest at
// the top, over what is under them, as an open Modal's panel is drawn
// over the surface. A toast is its kind's mark, its message, and its
// action at the end; the mark says the kind without colour.

const (
	// toastWidth is the most columns a toast takes.
	toastWidth = 40
	// toastGutter is the columns kept between the toasts and the frame's
	// right edge, as opencode keeps two: a box glued to the edge reads as
	// the terminal's, and a terminal's last column is where its writes
	// wrap.
	toastGutter = 1
)

// toastRoles are each kind's colour: its border's and its mark's.
var toastRoles = map[string]Role{
	view.ToastInfo:    Info,
	view.ToastSuccess: Success,
	view.ToastWarning: Warning,
	view.ToastError:   Error,
}

// toastBox is one toast as the stack lays it out: its message's rows,
// past the borders and the padding, and its action's face, at the end of
// the last row when it fits there, else on a row of its own.
type toastBox struct {
	e      *view.Element
	rows   [][]glyph
	action []glyph
	own    bool // the action has a row of its own
}

// height is a toast's rows, its borders included.
func (t toastBox) height() int {
	h := len(t.rows) + 2
	if t.own {
		h++
	}
	return h
}

// toastStack lays the toasts out cols wide: each as wide as the widest
// needs, at most toastWidth and the frame past its gutter. It returns the boxes, the
// newest first, and their width.
func (r *Rendition) toastStack(cols int) (boxes []toastBox, w int) {
	ts := r.c.V.Toasts
	if len(ts) == 0 {
		return nil, 0
	}
	for _, e := range ts {
		n := 2 + Width(e.Label)
		if a := toastAction(e); a != nil {
			n += 2 + Width(a.Label)
		}
		w = max(w, n+4)
	}
	w = max(min(w, toastWidth, cols-toastGutter), 1)
	iw := max(w-4, 1)
	for _, e := range ts {
		role := toastRoles[e.Variant]
		// The kind's mark, the glyph an Icon of it draws (view.ToastIcons:
		// ⓘ ✓ ! ✗), so that under NO_COLOR the kind still reads.
		mark := line(icons.Glyph(view.ToastIcons[e.Variant])+" ", style{role: role, attr: Bold})
		t := toastBox{e: e, rows: wrap(concat(mark, line(e.Label, style{})), iw, nil, repeat(" ", 2, style{}))}
		if a := toastAction(e); a != nil {
			st := style{attr: Bold | Underline}
			if r.focused(a.ID) {
				st = style{role: Accent, attr: Bold | Reverse}
			}
			t.action = fit(line(a.Label, st), iw)
			last := t.rows[len(t.rows)-1]
			t.own = width(last)+2+width(t.action) > iw
		}
		boxes = append(boxes, t)
	}
	return boxes, w
}

// toastAction is a toast's action, nil when it has none.
func toastAction(e *view.Element) *view.Element {
	for _, c := range e.Children {
		if c.Kind == view.ToastAction {
			return c
		}
	}
	return nil
}

// toastsHeight is the rows the stack takes.
func toastsHeight(boxes []toastBox) int {
	h := 0
	for _, t := range boxes {
		h += t.height()
	}
	return h
}

// paintToasts paints the stack at the frame's top right corner, over what
// is there, its gutter blank. A toast takes a click anywhere on it, which
// dismisses it, and its action one of its own, on top. A field's cursor
// under one is not shown.
func (r *Rendition) paintToasts(cv *canvas, boxes []toastBox, w int) {
	x, y := max(cv.f.Cols-w-toastGutter, 0), 0
	for _, t := range boxes {
		h := t.height()
		if f := cv.f; f.cursor && (box{x, y, w + toastGutter, h}).holds(f.cursorCol, f.cursorRow) {
			f.cursor = false
		}
		cv.fill(x, y, min(w+toastGutter, cv.f.Cols-x), h)
		cv.boxIn(x, y, w, h, style{role: toastRoles[t.e.Variant]})
		r.boxes[t.e.ID] = box{x, y, w, h}
		r.hits = append(r.hits, hit{x: x, y: y, w: w, h: h, id: t.e.ID, opt: -1})
		for i, row := range t.rows {
			cv.write(x+2, y+1+i, w-4, row)
		}
		if a := toastAction(t.e); a != nil {
			ay := y + len(t.rows)
			if t.own {
				ay++
			}
			ax := x + w - 2 - width(t.action)
			cv.write(ax, ay, width(t.action), t.action)
			r.boxes[a.ID] = box{ax, ay, width(t.action), 1}
			r.hits = append(r.hits, hit{x: ax, y: ay, w: width(t.action), h: 1, id: a.ID, opt: -1})
		}
		y += h
	}
}

// onToast reports whether a hit is a toast's or its action's.
func (r *Rendition) onToast(h *hit) bool {
	if h == nil {
		return false
	}
	e := r.c.V.Find(h.id)
	return e != nil && (e.Kind == view.Toast || e.Kind == view.ToastAction)
}

// heldToast reports whether the pointer is on a toast (Hover), whose time
// then waits (view.Controller.TickToasts).
func (r *Rendition) heldToast(id string) bool { return id != "" && id == r.held }

// Hover is the pointer at a cell of the last frame with no button down,
// as a terminal reports it when asked for every move (mode 1003). Over a
// toast, the toast waits; over an element with an accessibility
// description, the nearest one (view.Controller.Description), that
// description shows in the HottyKeyHints in place of the keyboard's, until
// the pointer leaves it, or a key or a click hides it, as a tooltip does
// (profile §6.23). changed reports whether the frame is to be drawn again.
func (r *Rendition) Hover(col, row int) (changed bool) {
	was, held := r.c.Tooltip(r.hover), r.held
	r.hover, r.held = "", ""
	if h := r.hitAt(col, row); r.onToast(h) {
		r.held = r.c.V.Find(h.id).Name
	}
	if r.held == "" {
		r.hover = r.describedAt(col, row)
	}
	if r.hover != r.quiet {
		r.quiet = ""
	}
	if r.quiet != "" {
		r.hover = ""
	}
	return r.c.Tooltip(r.hover) != was || r.held != held
}

// hush hides the hovered description, as a key or a click hides a
// tooltip, until the pointer leaves the element it is on.
func (r *Rendition) hush() {
	if r.hover != "" {
		r.quiet, r.hover = r.hover, ""
	}
}

// describedAt is the innermost element at a cell of the last frame that
// has a description, "" for none: under an open Modal's panel, only its
// content's; under a toast, none.
func (r *Rendition) describedAt(col, row int) string {
	if col < 0 || row < 0 {
		return ""
	}
	v := r.c.V
	for _, t := range v.Toasts {
		if b, ok := r.boxes[t.ID]; ok && b.holds(col, row) {
			return ""
		}
	}
	root := v.Root
	if r.panel != nil {
		if !r.panel.contains(col, row) {
			return ""
		}
		root = v.Overlay
	}
	// The last element in tree order under the cell with a description of
	// its own or one it is in has it: the innermost.
	found := ""
	var walk func(e *view.Element, desc string)
	walk = func(e *view.Element, desc string) {
		if e == nil || e.A11y.Hidden {
			return
		}
		if e.A11y.Description != "" {
			desc = e.A11y.Description
		}
		if b, ok := r.boxes[e.ID]; ok && desc != "" && b.holds(col, row) {
			found = e.ID
		}
		for _, k := range e.Children {
			walk(k, desc)
		}
	}
	walk(root, "")
	return found
}
