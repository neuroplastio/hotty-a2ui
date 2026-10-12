package cells

import "github.com/neuroplastio/hotty-a2ui/view"

// The pointer over the cells (profile §3.7): its shape, and the hover a
// frame draws where it is, as a browser shows both over a page.

// Pointer is the pointer's shape over a cell of the last frame, as CSS
// names it, for a program to set with OSC 22, as a browser shows it:
// "text", an I-beam, over a text control's value; over a longText's grip,
// the axes it resizes along ("nwse-resize"); "pointer", a hand, over
// what a click acts on (clickable): a Button, a CheckBox, a Switch, an
// option, a select, a tab, a link, in place or with a scheme, a tree's
// node, a list's item, a table's row, a hunk, a paginator's dot, a toast,
// a Slider's track; "text" over a Text's text, which a drag selects;
// else "", the terminal's own, a disabled control's too.
func (r *Rendition) Pointer(col, row int) string {
	if h := r.gripAt(col, row); h != nil {
		if e := r.c.V.Find(h.id); e != nil {
			return gripPointer(e)
		}
	}
	switch p, _ := r.point(spot{col: col, row: row, in: col >= 0 && row >= 0}); {
	case p.field:
		return "text"
	case p.id != "":
		return "pointer"
	}
	if f := r.last; f != nil && row >= 0 && row < f.Rows && col >= 0 && col < f.Cols && f.Cells[row][col].Link != "" {
		return "pointer"
	}
	if _, _, ok := r.textAt(col, row); ok {
		// Text a drag selects (select.go).
		return "text"
	}
	return ""
}

// clickable reports whether a click on a hit acts on what it is over: an
// item of something (an option of an open select, a row, a node, a dot, a
// hunk, a suggestion), or a control a click works; not a container's own
// hit (a table's header, a list's frame, a HottyScrollView), which a click
// gives the keyboard to and no more.
func (r *Rendition) clickable(h *hit) bool {
	e := r.c.V.Find(h.id)
	if e == nil || e.Disabled {
		return false
	}
	if h.opt >= 0 {
		return true
	}
	if e.Kind == view.FilePicker && e.FileDir != "" && h.opt == -1 {
		return true
	}
	switch e.Kind {
	case view.Button, view.CheckBox, view.Switch, view.Option, view.Tab, view.Link,
		view.Toast, view.ToastAction, view.Slider, view.RangeSlider, view.Knob:
		return true
	case view.Choice:
		return isSelect(e)
	}
	return e.Clickable
}

// pointed is what a hover draws: an element, its item (opt, -1 for the
// element), and whether it is a text control's value.
type pointed struct {
	id    string
	opt   int
	field bool
}

// spot is where the pointer is, in cells of the frame; in is false once
// it has left the frame (Hover(-1, -1)).
type spot struct {
	col, row int
	in       bool
}

// point is what the pointer at a cell of the last frame would have a hover
// draw, and the hit it is in: what Pointer gives a shape to; nothing
// elsewhere.
func (r *Rendition) point(at spot) (pointed, *hit) {
	if !at.in {
		return pointed{}, nil
	}
	if h := r.gripAt(at.col, at.row); h != nil {
		return pointed{id: h.id, opt: -1, field: true}, h
	}
	switch h := r.hitAt(at.col, at.row); {
	case h == nil || h.disabled:
	case h.field != nil:
		if a := h.field; at.row >= a.y && at.row < a.y+a.rows && at.col >= a.x && at.col < a.x+a.w {
			return pointed{id: h.id, opt: -1, field: true}, h
		}
	case r.clickable(h):
		return pointed{id: h.id, opt: h.opt}, h
	}
	return pointed{}, nil
}

// The hover's tints: a step past what a thing's own look has, so that it
// shows where the pointer is without drawing the eye.
const (
	hoverTint    = 20  // 8% toward the text, under what has no fill
	hoverButton  = 72  // a Button's fill, from buttonTint's 48
	hoverLineMix = 255 // a field's underline, from fieldLineMix's 50%
	hoverBarMix  = 90  // a longText's bar, 35% from border toward the text
)

// paintHover draws what the pointer is over hovered, past what painted
// the frame, as a browser's :hover is: a field's underline brighter (a
// long one's bar, where its blend shows), a Button's fill a step lighter, a link in the accent,
// anything else clickable on a faint tint of the text, where the
// terminal's or the theme's colours let a tint show. What is reversed (a
// selection, the keyboard's) keeps its look. The hover is worked out from
// the frame just drawn, so that it stays under the pointer when what is
// there moves or a Modal opens over it.
func (r *Rendition) paintHover(cv *canvas) {
	p, h := r.point(r.mouse)
	r.pointing = p
	if h == nil {
		return
	}
	e := r.c.V.Find(p.id)
	if e == nil {
		return
	}
	x, y, w, ht := h.x, h.y, h.w, h.h
	if a := h.field; a != nil {
		x, y, w, ht = a.x, a.y, a.w, a.rows
		if isLongText(e) {
			// Its bar, and its grip.
			x, w = h.x, 1
			if a.grip {
				cv.restyle(a.x+a.w, a.y+a.rows-1, 1, 1, func(c *Cell) {
					if c.Role == Border && c.Mix == 0 {
						c.To, c.Mix = Fg, hoverBarMix
					}
				})
			}
		}
	} else if isField(e) && w > gutter {
		// A field's gutter is the keyboard's bar, not its face.
		x, w = x+gutter, w-gutter
	}
	cv.restyle(x, y, w, ht, func(c *Cell) {
		switch {
		case c.Attr&Reverse != 0:
		case p.field && isLongText(e):
			if c.Role == Border && c.Mix == 0 {
				c.To, c.Mix = Fg, hoverBarMix
			}
		case p.field || isSelect(e) && p.opt < 0:
			if c.LineSet && c.Line == Border && c.LineMix == fieldLineMix {
				c.LineMix = hoverLineMix
			}
		case e.Kind == view.Button && c.Back == Fg && c.BackMix == buttonTint:
			c.BackMix = hoverButton
		case e.Kind == view.Link:
			c.Role = Accent
		case c.BackMix == 0:
			c.Back, c.BackMix = Fg, hoverTint
		}
	})
}
