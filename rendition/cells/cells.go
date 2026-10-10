// Package cells is the rendition on a terminal that is not a HOTTY host:
// a surface laid out and painted in cells, identically in every
// implementation (docs/profile.md §3). The renderer does what a host would
// do with the surface's document: it takes keys and clicks, and turns them
// into the view.Controller's calls.
package cells

import (
	"time"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottyedit"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Rendition is one surface in cells. Besides the controller's state, which
// belongs to the surface, it keeps what only cells have: where each text
// control's cursor is and how it is scrolled, which select's list is open,
// and where the last Draw put each thing a click can land on.
type Rendition struct {
	c *view.Controller

	cursor  map[string]int // a text control's cursor: an index into its value's clusters
	hscroll map[string]int // the first column it shows
	vscroll map[string]int // a longText's first line shown
	rows    map[string]int // a longText's rows, a Table's body rows, as last drawn
	list    string         // the select whose list is open
	hi      int            // the option highlighted in it
	keys    string         // the text controls' keymap (SetKeys)
	// fields are the text controls as last edited, which keep a run of
	// row moves' place along the row (editKey).
	fields map[string]*hottyedit.Field

	hits  []hit
	panel *hit           // the open Modal's panel, as last drawn
	boxes map[string]box // the cells each element covers, as last drawn
	// reveal is the part of an element a scroll view keeps in sight while
	// it has the keyboard, where that is not all of it: a HottyDiff's
	// selected hunk.
	reveal map[string]box
	// scrolls are the HottyScrollViews as last drawn, scrollOrder their
	// ids in the order painted, the innermost last (Wheel).
	scrolls     map[string]scrolled
	scrollOrder []string
	drag        string        // the Slider a click on its track started dragging
	anim        time.Duration // how soon the last Draw changes again (Animating)

	// Clock is the time a Draw paints at: a Spinner's frame, an
	// indeterminate Progress's place. New sets time.Now.
	Clock func() time.Time
}

// box is a rectangle of cells.
type box struct{ x, y, w, h int }

// New is a surface's rendition in cells.
func New(c *view.Controller) *Rendition {
	return &Rendition{c: c, cursor: map[string]int{}, hscroll: map[string]int{}, vscroll: map[string]int{}, rows: map[string]int{},
		keys: hotty.TerminalKeys, fields: map[string]*hottyedit.Field{}, Clock: time.Now}
}

// Animating is how soon the last Draw's frame changes by itself: the
// shortest interval of the Spinners that spin and the indeterminate
// Progress bars it painted; 0 when nothing moves. A program that shows the
// frame draws it again after that long (profile §3.4, the clock).
func (r *Rendition) Animating() time.Duration { return r.anim }

// animate notes that the frame changes again after d.
func (r *Rendition) animate(d time.Duration) {
	if d > 0 && (r.anim == 0 || d < r.anim) {
		r.anim = d
	}
}

// SetKeys sets the keymap the text controls edit by, as rendition/html's
// SetKeys sets a surface's: the default keymap, then keys, then the
// components' own (io_neuroplast_hotty.keys). New starts with
// hotty.TerminalKeys; "" is SPEC §10.2's default keymap alone.
func (r *Rendition) SetKeys(keys string) { r.keys = keys }

// focused reports whether id has the keyboard.
func (r *Rendition) focused(id string) bool {
	return r.c.St.Keyboard && r.c.St.Focus == id
}

// cursorOf is a text control's cursor, at most n; at the end when it has
// none.
func (r *Rendition) cursorOf(id string, n int) int {
	if p, ok := r.cursor[id]; ok && p <= n {
		return p
	}
	return n
}

// listOpen reports whether a select shows its options: after Space,
// Enter or a click opened them, until it loses the keyboard.
func (r *Rendition) listOpen(e *view.Element) bool {
	return r.list == e.ID && r.focused(e.ID)
}

// ListOpen reports whether a select shows its options. They push what is
// under the select down, so the frame is taller until the list closes.
func (r *Rendition) ListOpen() bool { return r.list != "" && r.focused(r.list) }

// Draw lays the surface out cols wide, as many rows as it takes, and
// paints it (profile §3). The open Modal's content is a rounded panel over
// it, centered; the frame grows when the panel is taller.
func (r *Rendition) Draw(cols int) *Frame {
	cols = max(cols, 1)
	if !r.focused(r.list) {
		r.list = ""
	}
	v := r.c.V
	l := newLayout(r)
	root := v.Root
	if root != nil && root.A11y.Hidden {
		root = nil
	}
	rootH, rows := 0, 0
	if root != nil {
		rootH = l.height(root, cols)
		rows = rootH
	}
	var pw, ph int
	if v.Overlay != nil {
		most := cols
		if cols >= 20 {
			most = cols - 4
		}
		pw = min(l.natural(v.Overlay)+4, most)
		ph = l.height(v.Overlay, pw-4) + 2
		rows = max(rows, ph)
	}
	f := newFrame(cols, rows)
	cv := &canvas{f: f}
	r.hits, r.panel, r.boxes, r.reveal, r.anim = nil, nil, map[string]box{}, map[string]box{}, 0
	r.scrolls, r.scrollOrder = map[string]scrolled{}, nil
	if root != nil {
		l.paint(cv, root, 0, 0, cols, rootH)
	}
	if v.Overlay != nil {
		cv.restyle(0, 0, cols, rows, func(c *Cell) { c.Attr |= Faint })
		r.hits = nil
		f.cursor = false
		px, py := (cols-pw)/2, (rows-ph)/2
		cv.fill(px, py, pw, ph)
		cv.box(px, py, pw, ph)
		r.panel = &hit{x: px, y: py, w: pw, h: ph}
		l.paint(cv, v.Overlay, px+2, py+1, pw-4, ph-2)
	}
	return f
}

// Box is the cells an element covers in the last frame Draw returned, by
// its view id: its column and row, its width and height. ok is false when
// it was not drawn. A control that does not fill its box (a Button, a
// CheckBox, a Tabs' title, an option) covers what it painted.
func (r *Rendition) Box(id string) (col, row, w, h int, ok bool) {
	b, ok := r.boxes[id]
	return b.x, b.y, b.w, b.h, ok
}

// Sight is the cells of an element to keep in sight in the last frame,
// as Box has them: the part of it it reveals (a HottyDiff's selected
// hunk), or else all of it.
func (r *Rendition) Sight(id string) (col, row, w, h int, ok bool) {
	b, ok := r.reveal[id]
	if !ok {
		b, ok = r.boxes[id]
	}
	return b.x, b.y, b.w, b.h, ok
}

// Click is a click on a cell of the last frame drawn (SPEC §10.1): what
// takes focus there gets the keyboard and is activated, as the host would
// (a Button runs, a CheckBox toggles, a select opens, a field puts its
// cursor there); a click on nothing that takes focus gives the keyboard
// back. Outside the open Modal's panel, a click closes the Modal.
func (r *Rendition) Click(col, row int) error {
	c := r.c
	if r.panel != nil && !r.panel.contains(col, row) {
		r.list = ""
		c.CloseModal()
		c.Focus("")
		return nil
	}
	var h *hit
	for i := len(r.hits) - 1; i >= 0; i-- {
		if r.hits[i].contains(col, row) {
			h = &r.hits[i]
			break
		}
	}
	e := (*view.Element)(nil)
	if h != nil {
		e = c.V.Find(h.id)
	}
	if e == nil {
		r.list = ""
		c.Focus("")
		return nil
	}
	if h.disabled {
		r.list = ""
		c.Focus("")
		return c.Activate(e.ID)
	}
	if c.St.Focus != e.ID {
		r.list = ""
	}
	c.Focus(e.ID)
	switch {
	case isTextControl(e):
		if a := h.field; a != nil && row >= a.y {
			v, _ := e.Value.(string)
			lines := [][]string{clusters(v)}
			if isLongText(e) {
				lines = splitClusters(lines[0])
			}
			li := min(row-a.y+a.voff, len(lines)-1)
			r.cursor[e.ID] = offset(lines, li, indexAt(lines[li], col-a.x+a.hoff, e.Variant == "obscured"))
		}
		return nil
	case isSelect(e):
		if h.opt >= 0 {
			r.list = ""
			return c.SetValue(e.ID, e.Options[h.opt].Value)
		}
		r.toggleList(e)
		return nil
	case e.Kind == view.Table:
		return r.clickTable(e, h.opt)
	case e.Kind == view.DiffView:
		return r.clickDiff(e, h.opt)
	case e.Kind == view.RichList:
		return r.clickList(e, h.opt)
	case e.Kind == view.Tree:
		return r.c.ClickNode(e.ID, h.opt)
	case e.Kind == view.Slider:
		if t := h.track; t != nil && col >= t.x && col < t.x+t.n {
			r.drag = e.ID
			return r.slideTo(e, t, col)
		}
		return nil
	}
	return c.Activate(e.ID)
}

// Drag is the pointer at a cell with the primary button still down,
// after a Click: a Slider whose track the click landed on follows it,
// clamped to the track's ends, as last drawn. After any other click it
// does nothing.
func (r *Rendition) Drag(col, row int) error {
	if r.drag == "" {
		return nil
	}
	e := r.c.V.Find(r.drag)
	if e == nil || e.Kind != view.Slider {
		r.drag = ""
		return nil
	}
	for _, h := range r.hits {
		if h.id == e.ID && h.track != nil {
			return r.slideTo(e, h.track, col)
		}
	}
	return nil
}

// Release is the primary button let go: a drag ends.
func (r *Rendition) Release() { r.drag = "" }

// slideTo sets a Slider to the value at a column of its track:
// min + (max − min) × column / (track − 1), clamped and stepped.
func (r *Rendition) slideTo(e *view.Element, t *trackArea, col int) error {
	f := 0.0
	if t.n > 1 {
		f = float64(min(max(col-t.x, 0), t.n-1)) / float64(t.n-1)
	}
	return r.c.SetValue(e.ID, e.Min+f*(e.Max-e.Min))
}

// toggleList opens a select's list, its value highlighted, or closes it.
func (r *Rendition) toggleList(e *view.Element) {
	if r.list == e.ID {
		r.list = ""
		return
	}
	r.list, r.hi = e.ID, max(picked(e), 0)
}
