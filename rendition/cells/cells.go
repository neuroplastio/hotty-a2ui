// Package cells is the rendition on a terminal that is not a HOTTY host:
// a surface laid out and painted in cells, identically in every
// implementation (docs/profile.md §3). The renderer does what a host would
// do with the surface's document: it takes keys and clicks, and turns them
// into the view.Controller's calls.
package cells

import (
	"time"

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

	hits    []hit
	panel   *hit           // the open Modal's panel, as last drawn
	suggest *suggestPlace  // where the focused field's suggestions go over the frame, in the Draw under way
	boxes   map[string]box // the cells each element covers, as last drawn
	// reveal is the part of an element a scroll view keeps in sight while
	// it has the keyboard, where that is not all of it: a HottyDiff's
	// selected hunk.
	reveal map[string]box
	// scrolls are the HottyScrollViews as last drawn, scrollOrder their
	// ids in the order painted, the innermost last (Wheel).
	scrolls     map[string]scrolled
	scrollOrder []string
	drag        string        // the Slider or HottyRangeSlider a click on its track started dragging, or the text control a press in its value did
	knob        int           // which of a HottyRangeSlider's knobs the drag moves; -1 until its first move picks one
	anchor      int           // where in the text control the press was, which a drag selects from
	blinkID     string        // the text control whose caret blinks (caretOn)
	blinkFrom   time.Time     // when its caret last showed anew: it took the keyboard, a key, a press
	press       *press        // a press that may start a drag and drop (drag.go)
	anim        time.Duration // how soon the last Draw changes again (Animating)
	// hover is the element whose description the pointer shows (Hover),
	// quiet the one a key or a click hid it on, until the pointer leaves
	// it; held is the toast the pointer is on, whose time waits.
	hover, quiet, held string

	// Clock is the time a Draw paints at: a Spinner's frame, an
	// indeterminate Progress's place. New sets time.Now.
	Clock func() time.Time
}

// box is a rectangle of cells.
type box struct{ x, y, w, h int }

// New is a surface's rendition in cells.
func New(c *view.Controller) *Rendition {
	return &Rendition{c: c, cursor: map[string]int{}, hscroll: map[string]int{}, vscroll: map[string]int{}, rows: map[string]int{},
		fields: map[string]*hottyedit.Field{}, Clock: time.Now}
}

// Animating is how soon the last Draw's frame changes by itself: the
// shortest interval of the Spinners that spin and the indeterminate
// Progress bars it painted, and of the toasts and the timers that count
// (view.ToastStep, view.TimerStep); 0 when nothing moves. A program that shows the
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
// components' own (io_neuroplast_hotty.keys). New starts with "", SPEC
// §10.2's default keymap alone, a GUI field's; hotty.TerminalKeys gives
// Bubble Tea's.
func (r *Rendition) SetKeys(keys string) { r.keys = keys }

// blockCursor reports whether a text control's cursor is a block: under a
// keymap set over the default one, a terminal's (hotty.TerminalKeys),
// whose keys a block says, as a line says a GUI field's (profile §3.5).
func (r *Rendition) blockCursor() bool { return r.keys != "" }

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
// it, centered; the toasts stack over both at the top right corner, after
// their time is counted on the rendition's clock (view.Controller
// .TickToasts), as the timers' is first (TickTimers), whose onTimeout may
// show one. The frame grows when the panel or the stack is taller.
func (r *Rendition) Draw(cols int) *Frame {
	cols = max(cols, 1)
	timers := r.c.TickTimers(r.Clock())
	toasts := r.c.TickToasts(r.Clock(), r.heldToast)
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
	stack, tw := r.toastStack(cols)
	rows = max(rows, toastsHeight(stack))
	f := newFrame(cols, rows)
	cv := &canvas{f: f}
	r.hits, r.panel, r.suggest, r.boxes, r.reveal, r.anim = nil, nil, nil, map[string]box{}, map[string]box{}, 0
	r.animate(toasts)
	r.animate(timers)
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
	if r.suggest != nil {
		l.paintSuggestions(cv, r.suggest)
	}
	l.paintDrag(cv)
	r.paintToasts(cv, stack, tw)
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
//
// A press on what can be dragged (profile §6.21) may be a drag's start: it
// selects a list's item there now, and what else the click does (a second
// click's action, a branch opening, a Button in a card running) waits for
// the Release, which a drag does not do.
func (r *Rendition) Click(col, row int) error {
	r.press = nil
	r.hush()
	h := r.hitAt(col, row)
	if (r.panel == nil || r.panel.contains(col, row)) && !r.holds(h) && !r.onToast(h) {
		if id, item, at, ok := r.grab(col, row, h); ok {
			p := &press{id: id, item: item, at: at}
			r.press = p
			if e := r.c.V.Find(id); item >= 0 && e.Kind != view.Stack && item != e.SelectedRow() && !(e.Kind == view.Tree && e.Nodes[item].Branch()) {
				return r.click(col, row)
			}
			if h != nil && !h.disabled {
				r.c.Focus(h.id)
			}
			p.click = func() error { return r.click(col, row) }
			return nil
		}
	}
	return r.click(col, row)
}

// ShiftClick is a primary press with Shift held. In the text control that
// has the keyboard, it extends the selection to where the press is, from
// the selection's anchor, or else from the caret, and a drag from there
// goes on extending it, as in a GUI's field. Anywhere else it is a Click.
func (r *Rendition) ShiftClick(col, row int) error {
	if h := r.hitAt(col, row); h != nil && h.field != nil && row >= h.field.y && r.focused(h.id) {
		if e := r.c.V.Find(h.id); e != nil && isTextControl(e) {
			f := r.field(e)
			anchor, _ := f.Anchor()
			p := posAt(e, h.field, col, row)
			f.Select(anchor, p)
			r.cursor[e.ID] = p
			r.press, r.drag, r.anchor = nil, e.ID, anchor
			r.wake()
			return nil
		}
	}
	return r.Click(col, row)
}

// Pointer is the pointer's shape over a cell of the last frame, as CSS
// names it, for a program to set with OSC 22: "text", an I-beam, over a
// text control's value, as a browser shows over a field; else "", the
// terminal's own.
func (r *Rendition) Pointer(col, row int) string {
	if h := r.hitAt(col, row); h != nil && !h.disabled && h.field != nil {
		if a := h.field; row >= a.y && row < a.y+a.rows && col >= a.x && col < a.x+a.w {
			return "text"
		}
	}
	return ""
}

// hitAt is the topmost hit at a cell of the last frame, or nil.
func (r *Rendition) hitAt(col, row int) *hit {
	for i := len(r.hits) - 1; i >= 0; i-- {
		if r.hits[i].contains(col, row) {
			return &r.hits[i]
		}
	}
	return nil
}

// holds reports whether a press on a hit is the pointer's own and starts
// no drag: in a text control, which a drag selects in, or on a Slider or
// a HottyRangeSlider, which a drag moves.
func (r *Rendition) holds(h *hit) bool {
	if h == nil {
		return false
	}
	e := r.c.V.Find(h.id)
	return e != nil && (isTextControl(e) || e.Kind == view.Slider || e.Kind == view.RangeSlider || e.Kind == view.Knob)
}

func (r *Rendition) click(col, row int) error {
	c := r.c
	if h := r.hitAt(col, row); r.onToast(h) {
		// A toast dismisses, and its action is picked: a click on the
		// toast is on nothing that takes focus, and gives the keyboard
		// back; a click on its action gives it the action, which then
		// goes with the toast (view.Controller.DismissToast).
		e := c.V.Find(h.id)
		r.list = ""
		if e.Kind == view.ToastAction {
			c.Focus(e.ID)
		} else {
			c.Focus("")
		}
		return c.Activate(e.ID)
	}
	if r.panel != nil && !r.panel.contains(col, row) {
		r.list = ""
		c.CloseModal()
		c.Focus("")
		return nil
	}
	h := r.hitAt(col, row)
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
		if h.opt >= 0 {
			// A row of its suggestions: picked, the caret after it.
			delete(r.cursor, e.ID)
			return c.PickSuggestion(e.ID, h.opt)
		}
		if a := h.field; a != nil && row >= a.y {
			// The caret goes where the press is, and a drag from there
			// selects (Drag), as in a GUI's field.
			p := posAt(e, a, col, row)
			r.field(e).Select(p, p)
			r.cursor[e.ID] = p
			r.drag, r.anchor = e.ID, p
			r.wake()
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
	case e.Kind == view.RangeSlider:
		// The track off the knobs: the nearer knob takes the keyboard and
		// goes there, and a drag moves it on (view.Controller.PressKnob).
		var err error
		r.drag = e.ID
		r.knob, err = c.PressKnob(e.ID, trackValue(e, h.track, col))
		return err
	case e.Kind == view.Knob:
		// A knob takes the keyboard, where it is, and a drag moves it; where
		// the knobs meet, the drag's first move picks the one it goes
		// towards, as from the track.
		if anc := c.Ancestors(e.ID); len(anc) > 0 {
			rg := anc[len(anc)-1]
			r.drag, r.knob = rg.ID, e.Selected
			if lo, hi := rg.Range(); lo == hi {
				r.knob = -1
			}
		}
		return nil
	}
	return c.Activate(e.ID)
}

// Drag is the pointer at a cell with the primary button still down,
// after a Click: a Slider whose track the click landed on follows it, and
// so does the knob of a HottyRangeSlider the click picked (stopped where
// it meets the other), clamped to the track's ends, as last drawn; a text
// control the click landed in selects from there to the pointer, a row
// above or below its lines selecting to its first or last; after a press
// on what can be dragged, once the pointer leaves it, it is lifted and a
// line shows where it would land (drag.go). After any other click it does
// nothing.
func (r *Rendition) Drag(col, row int) error { return r.DragAt(col, row, -1) }

// slide is Drag for a Slider, a HottyRangeSlider or a text control.
func (r *Rendition) slide(col, row int) error {
	e := r.c.V.Find(r.drag)
	if e == nil || e.Kind != view.Slider && e.Kind != view.RangeSlider && !isTextControl(e) {
		r.drag = ""
		return nil
	}
	for _, h := range r.hits {
		if h.id != e.ID {
			continue
		}
		if h.field != nil {
			p := posAt(e, h.field, col, row)
			r.field(e).Select(r.anchor, p)
			r.cursor[e.ID] = p
			r.wake()
			return nil
		}
		if h.track == nil {
			continue
		}
		if e.Kind == view.RangeSlider {
			k, err := r.c.DragKnob(e.ID, r.knob, trackValue(e, h.track, col))
			r.knob = k
			return err
		}
		return r.slideTo(e, h.track, col)
	}
	return nil
}

// posAt is the place in a text control's value at a cell of its area, as
// last drawn: before its first character left of it, after its last right
// of it, and in its first or last line above or below it.
func posAt(e *view.Element, a *fieldArea, col, row int) int {
	v, _ := e.Value.(string)
	lines := [][]string{clusters(v)}
	if isLongText(e) {
		lines = splitClusters(lines[0])
	}
	li := min(max(row-a.y+a.voff, 0), len(lines)-1)
	return offset(lines, li, indexAt(lines[li], col-a.x+a.hoff, e.Variant == "obscured"))
}

// Release is the primary button let go: a drag ends; a drag and drop
// drops, and a press that did not drag does what its click waited for.
func (r *Rendition) Release() error {
	r.drag = ""
	return r.release()
}

// slideTo sets a Slider to the value at a column of its track (trackValue).
func (r *Rendition) slideTo(e *view.Element, t *trackArea, col int) error {
	return r.c.SetValue(e.ID, trackValue(e, t, col))
}

// trackValue is the value at a column of a Slider's track:
// min + (max − min) × column / (track − 1), the column clamped to the
// track; the controller steps it.
func trackValue(e *view.Element, t *trackArea, col int) float64 {
	f := 0.0
	if t != nil && t.n > 1 {
		f = float64(min(max(col-t.x, 0), t.n-1)) / float64(t.n-1)
	}
	return e.Min + f*(e.Max-e.Min)
}

// TrackCell is the cell of a Slider's or a HottyRangeSlider's track that
// stands for v, in the last frame drawn (profile §3.4): where a click sets
// v, or the nearest to it. ok is false when the track was not drawn.
func (r *Rendition) TrackCell(id string, v float64) (col, row int, ok bool) {
	e := r.c.V.Find(id)
	if e == nil {
		return 0, 0, false
	}
	for _, h := range r.hits {
		if h.id == id && h.track != nil {
			return h.track.x + trackCol(e, v, h.track.n), h.y, true
		}
	}
	return 0, 0, false
}

// toggleList opens a select's list, its value highlighted, or closes it.
func (r *Rendition) toggleList(e *view.Element) {
	if r.list == e.ID {
		r.list = ""
		return
	}
	r.list, r.hi = e.ID, max(picked(e), 0)
}
