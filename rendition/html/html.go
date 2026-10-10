// Package html is the rendition on a HOTTY host: a surface's view as an
// HTML document, which the host lays out, paints, and runs the keyboard
// in (SPEC §10), and then as deltas (SPEC §6) as the surface changes. The
// host's events come back as the user's acts on the view's Controller.
//
// What the host does alone stays there: typing in a field, moving focus
// with Tab, toggling a checkbox. The rendition hears the outcome (input,
// change, click, submit), which keeps the data model current as the user
// types, as A2UI's own renderers do.
package html

import (
	_ "embed"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

//go:embed kit.css
var kitCSS string

// head goes first in every document. Images are the only thing a view
// fetches: an Image's URL, or one in a Text's Markdown, over HTTPS. The
// host's own policy decides the rest (SPEC §7.2).
var head = `<meta name="hotty-network" content="img-src https:"><style>` + kitCSS + `</style>`

// Rendition is one surface on a host.
type Rendition struct {
	C    *view.Controller
	name string
	// sent is what the host has: the surface's two top elements as last
	// sent; nil before the document is.
	sent []*node
	// pending is the Shortcut key waiting for the host to commit the
	// focused field (Key).
	pending string
	// host is the keyboard as the host has it, as far as the rendition
	// knows: Update gives the host the controller's when they differ.
	host keyboard
	// theme is the colours the kit paints with; the zero theme is the
	// host's own.
	theme theme.Theme
	// keys is the surface's keymap (SetKeys).
	keys string
	// fit: the program sizes the surface to its document (SetFit).
	fit bool
	// steps: the host says where in a dragged element the pointer is
	// (SetSteps). drag is the Slider whose track such a drag is on, or the
	// HottyRangeSlider a drag is on (with steps or not), knob the knob it
	// moves (-1 until its first move picks one), and dragged the one whose
	// drag just ended: the click after it is the drag's.
	steps         bool
	drag, dragged string
	knob          int
	// pressed: a drag just started on a HottyRangeSlider's track, which
	// takes no focus, so that the host's blur follows (rangeDrag).
	pressed bool
	// Away is set while the keyboard is in another rendition of the same
	// surface (cells beside it): the host's surface gives it up, and its
	// blur leaves the controller's keyboard where it is.
	Away bool
	// back is set when the keyboard goes back after a Shortcut's blur:
	// to where the host had it, which Tab may have moved unseen.
	back bool
	// list is the select whose list is open, nil when none is.
	list *list
	// pop is the open list's surface as the host has it (Popover); nil
	// before its document is sent, and once the list closes.
	pop *node
	// Clock is the time a Doc or Update shows: a Spinner's frame, an
	// indeterminate Progress's sweep. New sets it to time.Now; a test
	// stops it.
	Clock func() time.Time
	// anim is how soon the last Doc or Update changes again (Animating).
	anim time.Duration
}

type keyboard struct {
	on    bool
	focus string
}

// New is the rendition of a controller's surface, as the host surface
// name.
func New(c *view.Controller, name string) *Rendition {
	return &Rendition{C: c, name: name, Clock: time.Now}
}

// Animating is how soon the document the last Doc or Update made changes
// by itself: the interval of its fastest Spinner or indeterminate
// Progress, or 0 when nothing in it moves. The program calls Update again
// then; the cells rendition keeps the same clock, so the two show the
// same frame.
func (r *Rendition) Animating() time.Duration { return r.anim }

// markup is a new markup of the view as it is at the rendition's clock.
func (r *Rendition) markup() *markup {
	now := time.Now
	if r.Clock != nil {
		now = r.Clock
	}
	short, full := r.C.KeyHints(r.keys, true)
	m := &markup{list: r.openList(), now: now(), short: short, full: full, steps: r.steps}
	if r.C.St.Keyboard && !r.Away {
		m.keyboard = r.C.St.Focus
	}
	return m
}

// SetTheme paints the surface in a theme's colours from the next Doc or
// Update on; the zero theme is the host's own.
func (r *Rendition) SetTheme(th theme.Theme) { r.theme = th }

// SetKeys sets the keymap of the surface's text fields from the next Doc
// or Update on: a data-keys value on the surface's top elements (SPEC
// §10.2), which a component's own (io_neuroplast_hotty.keys) overrides
// key by key. New starts with "", the host's default keymap, a GUI
// field's, as the cells rendition starts; hotty.TerminalKeys gives Bubble
// Tea's.
func (r *Rendition) SetKeys(keys string) { r.keys = keys }

// SetFit says, from the next Doc or Update on, whether the program sizes
// the surface to its document, by the host's fit (SPEC §5.2): the document
// is then as tall as its content, where it otherwise fills the surface,
// which would hold the fit at the surface's height.
func (r *Rendition) SetFit(fit bool) { r.fit = fit }

// SetSteps says, from the next Doc or Update on, whether the host says
// where in a dragged element the pointer is (SPEC §9.1's steps, `steps` in
// its capabilities). A Slider's track is then one element that a drag
// moves wherever the pointer goes, its notches only for taps; without
// steps, each notch is a drag target, and a drag off them stops.
func (r *Rendition) SetSteps(steps bool) { r.steps = steps }

// Name is the HOTTY surface's name.
func (r *Rendition) Name() string { return r.name }

// Doc is the whole document of the view as it is now, for a=doc; the host
// has it from then on. A document sent again (the host lost the surface,
// or the program deleted it off screen) has no focus, so the next Update
// gives the host the keyboard again if the view has it.
func (r *Rendition) Doc() string {
	m := r.markup()
	main, layer := surface(r.C.V, r.theme, r.keys, r.fit, m)
	r.sent, r.anim = []*node{main, layer}, m.anim
	r.host = keyboard{}
	return head + main.html() + layer.html()
}

// Update is the deltas that bring the host's document to the view as it
// is now: none when nothing changed, or before the document was sent.
func (r *Rendition) Update() []string {
	if r.sent == nil {
		return nil
	}
	m := r.markup()
	main, layer := surface(r.C.V, r.theme, r.keys, r.fit, m)
	r.anim = m.anim
	r.holdEdit(main)
	out := diff(r.name, r.sent[0], main, nil)
	out = diff(r.name, r.sent[1], layer, out)
	r.sent = []*node{main, layer}
	if r.pop != nil {
		if n := popover(r.C.V, r.theme, r.openList()); n != nil {
			out = diff(r.PopoverName(), r.pop, n, out)
			r.pop = n
		} else {
			r.pop = nil
		}
	}
	// The keyboard, when the controller moved it: autofocus, the focus
	// and blur functions, the end of a Shortcut (Key).
	// While a select's list is open, the program has the keyboard (Key):
	// a host scrolls with the arrows a focused button leaves (SPEC §5.3),
	// so they would never reach it.
	st := r.C.St
	want := keyboard{st.Keyboard && !r.Away && r.openList() == nil, r.focusOn(st.Focus)}
	if want != r.host && (want.on || r.host.on) {
		switch {
		case !want.on:
			out = append(out, hotty.Blur(r.name))
		case r.back && want.focus == r.host.focus:
			out = append(out, hotty.Focus(r.name, ""))
		case want.focus != st.Focus:
			// A part of the element: a HottyDiff's selected hunk.
			out = append(out, hotty.Focus(r.name, want.focus))
		case want.focus != "" && r.C.V.Find(want.focus) != nil:
			out = append(out, hotty.Focus(r.name, domID(want.focus)))
		default:
			out = append(out, hotty.Focus(r.name, ""))
		}
		r.host = want
	}
	r.back = false
	return out
}

// Event is what the user did in the surface: it acts on the controller,
// and Update then makes the deltas. Events of other surfaces, and of kinds
// it does not know, do nothing.
func (r *Rendition) Event(ev hotty.Event) error {
	if ev.Surface == r.PopoverName() {
		return r.popoverEvent(ev)
	}
	if ev.Surface != r.name {
		return nil
	}
	c := r.C
	pressed := r.pressed
	if ev.Kind != hotty.EventChange {
		r.pressed = false
	}
	if ev.Kind != hotty.EventClick {
		// The click a drag ends with comes right after its dragend; after
		// anything else (the blur of a new press), a click is a tap.
		r.dragged = ""
	}
	switch ev.Kind {
	case hotty.EventFocus:
		c.St.Keyboard, r.host.on = true, true
		// The user moved the focus (SPEC §10.1): t names the element, the
		// nearest one with an id. The program's own a=focus names none.
		if id, part, ok := viewID(ev.Target); ok && c.V.Find(id) != nil {
			r.host.focus = id
			if part != "" {
				r.host.focus = partID(id, part)
			}
			if !r.Away {
				c.St.Focus = id
			}
		}
		return nil
	case hotty.EventBlur:
		if r.openList() != nil && !r.host.on {
			// The blur an open list asked for (Update): the controller
			// keeps the keyboard, and the program works the list.
			return nil
		}
		if pressed && c.St.Keyboard {
			// The blur of the press that started a drag on a
			// HottyRangeSlider's track (rangeDrag): the knob it moves keeps
			// the keyboard, which Update gives the host back.
			r.host.on = false
			return nil
		}
		r.list = nil
		if r.Away {
			r.host.on = false
			return nil
		}
		c.St.Keyboard, r.host.on = false, false
		if r.pending == "" {
			return nil
		}
		// The field is committed: the Shortcut runs on current inputs.
		// The blur was the rendition's own, so the keyboard goes back
		// where the host had it (Update), unless the Shortcut moves it.
		key := r.pending
		r.pending = ""
		c.St.Keyboard, r.back = true, true
		_, err := c.Shortcut(key)
		return err
	}
	if ev.Target == dismissID && ev.Kind == hotty.EventClick {
		// The click focused nothing: Update gives the select the keyboard
		// back.
		r.list, r.host.focus = nil, ""
		return nil
	}
	if ev.Target == backdropID && ev.Kind == hotty.EventClick {
		r.list = nil
		c.CloseModal()
		return nil
	}
	if done, err := r.rangeDrag(ev); done {
		return err
	}
	if done, err := r.slideSteps(ev); done {
		return err
	}
	id, part, ok := viewID(ev.Target)
	if !ok {
		return nil
	}
	e := c.V.Find(id)
	if e == nil {
		return nil
	}
	switch ev.Kind {
	case hotty.EventClick:
		if isSelect(e) {
			return r.clickSelect(e, part, ev)
		}
		if e.Kind == view.Table || e.Kind == view.RichList || e.Kind == view.DiffView {
			// A row (a Table's), an item (a HottyList's) or a hunk (a
			// HottyDiff's) is selected by a click, and acted on by another.
			c.Focus(id)
			r.host = keyboard{true, id}
			prefix := partRow
			switch e.Kind {
			case view.RichList:
				prefix = partItem
			case view.DiffView:
				prefix = partHunk
			}
			if i, err := strconv.Atoi(strings.TrimPrefix(part, prefix)); err == nil && strings.HasPrefix(part, prefix) {
				if i == e.SelectedRow() {
					return c.Activate(id)
				}
				return c.SelectRow(id, i)
			}
			return nil
		}
		if e.Kind == view.Tree {
			// A node is selected by a click, a branch opened or closed by
			// it, and a leaf acted on by another (view.Controller.ClickNode).
			c.Focus(id)
			r.host = keyboard{true, id}
			if i, err := strconv.Atoi(strings.TrimPrefix(part, partNode)); err == nil && strings.HasPrefix(part, partNode) {
				return c.ClickNode(id, i)
			}
			return nil
		}
		if e.Kind == view.Slider && (part == partLess || part == partMore) {
			n := 1
			if part == partLess {
				n = -1
			}
			return c.StepSlider(id, n, "")
		}
		if e.Kind == view.Slider && len(part) > 1 && part[0] == partNotch[0] {
			// A tap on the track, or the click a drag ends with where it
			// began (SPEC §9.1).
			return r.notch(e, part)
		}
		if e.Kind == view.RangeSlider {
			// A tap on the track: the nearer knob goes there and takes the
			// keyboard, which the tap gave the terminal (Update gives it
			// back).
			if v, ok := notchAt(e, part); ok {
				_, err := c.PressKnob(id, v)
				return err
			}
			return nil
		}
		if part != "" {
			return nil
		}
		if e.Focusable() {
			c.Focus(id)
			r.host = keyboard{true, id}
		}
		return c.Activate(id)
	case hotty.EventInput, hotty.EventChange:
		// A text field's change is its commit, which comes as the host
		// gives its keyboard up: to where the program moved it, or to the
		// other rendition. So only typing, or a control's change, which
		// comes at once, says where the keyboard is.
		if ev.Kind == hotty.EventInput || e.Kind != view.TextField && e.Kind != view.DateTime {
			r.host.focus = id
			if !r.Away {
				c.St.Focus = id
			}
		}
		switch e.Kind {
		case view.CheckBox:
			on, _ := ev.Checked()
			return c.SetValue(id, on)
		case view.Option:
			if on, _ := ev.Checked(); on != e.Active {
				return c.Activate(id)
			}
			return nil
		}
		return c.SetValue(id, ev.Value())
	case hotty.EventSubmit:
		if e.Kind == view.Form {
			return c.Submit(id)
		}
	case hotty.EventDragStart, hotty.EventDrag, hotty.EventDragEnd:
		if e.Kind == view.Slider && len(part) > 1 && part[0] == partNotch[0] {
			return r.notch(e, part)
		}
	}
	return nil
}

// slideSteps takes a drag of a Slider's track on a host with steps (SPEC
// §9.1): dragstart names the track, and from then on x is the step under
// the pointer along it, wherever the pointer is and whatever t says, until
// dragend. The click a drag ends with, on the notch it was released on,
// is the drag's, which set the value already. done reports that ev was
// one of these.
func (r *Rendition) slideSteps(ev hotty.Event) (done bool, err error) {
	dragged := r.dragged
	r.dragged = ""
	switch ev.Kind {
	case hotty.EventDragStart:
		id, part, ok := viewID(ev.Target)
		if !ok || part != "" {
			return false, nil
		}
		e := r.C.V.Find(id)
		x, has := stepX(ev)
		if e == nil || e.Kind != view.Slider || !has {
			return false, nil
		}
		r.drag = id
		return true, r.C.SetValue(id, notchValue(e, x))
	case hotty.EventDrag, hotty.EventDragEnd:
		id := r.drag
		if id == "" {
			return false, nil
		}
		if ev.Kind == hotty.EventDragEnd {
			r.drag = ""
			if releasedOn(ev, id) {
				r.dragged = id
			}
		}
		e := r.C.V.Find(id)
		if x, has := stepX(ev); has && e != nil && e.Kind == view.Slider {
			return true, r.C.SetValue(id, notchValue(e, x))
		}
		return true, nil
	case hotty.EventClick:
		if id, _, ok := viewID(ev.Target); ok && dragged != "" && id == dragged {
			return true, nil
		}
	}
	return false, nil
}

// releasedOn reports that a dragend let go on the element with view id
// id, a part of it included: only then may the click a drag ends with
// follow (SPEC §9.1), which is the drag's. Let go elsewhere, the next
// click is a tap of its own.
func releasedOn(ev hotty.Event, id string) bool {
	t, _, ok := viewID(ev.Target)
	return ok && t == id
}

// stepX is a drag's step along its element, if the host said one (SPEC
// §9.1, data-steps).
func stepX(ev hotty.Event) (int, bool) {
	d, ok := ev.Drag()
	return d.X, ok && d.HasX
}

// notch sets a Slider to the value of its notch part names ("k3").
func (r *Rendition) notch(e *view.Element, part string) error {
	i, err := strconv.Atoi(part[1:])
	if err != nil {
		return nil
	}
	return r.C.SetValue(e.ID, notchValue(e, i))
}

// Key is a key the program read while the surface is the one keys apply
// to, named as SPEC §10.4 has it ("Control+s", "Escape", "Space"): one
// the focused field did not use (its keymap, SetKeys), or any while the
// surface does not have the keyboard. A Shortcut takes it. While the surface has the keyboard, a field may
// hold an edit the host has not committed, so Key sends a=blur and the
// Shortcut runs when the blur comes back (Event), after the field's
// change; then the keyboard goes back to the surface. Else Escape closes
// an open Modal, and ? switches a HottyKeyHints' views. ok reports whether the surface took the key.
func (r *Rendition) Key(key string) (cmds []string, ok bool, err error) {
	if k, named := hotty.ParseKey(key); named {
		key = k
	}
	c := r.C
	if r.pending != "" {
		// The blur never came: the host had given the keyboard back
		// already. Run what waited.
		pending := r.pending
		r.pending = ""
		if _, err = c.Shortcut(pending); err != nil {
			return nil, true, err
		}
	}
	if slices.ContainsFunc(c.V.Shortcuts, func(sc view.Shortcut) bool { return view.SameKey(sc.Key, key) }) {
		if c.St.Keyboard {
			r.pending = key
			return []string{hotty.Blur(r.name)}, true, nil
		}
		_, err = c.Shortcut(key)
		return nil, true, err
	}
	if l := r.openList(); l != nil {
		if ok, err := r.listKey(l, key); ok {
			return nil, true, err
		}
	}
	// A focused HottyList is a box on the host too: it moves, turns pages
	// and filters by the keys (view.Controller.ListKey). Its filter takes
	// Escape before an open Modal does.
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && e.Kind == view.RichList {
		if ok, err := c.ListKey(e.ID, key); ok {
			return nil, true, err
		}
	}
	if key == "Escape" && c.St.Modal != "" {
		c.CloseModal()
		return nil, true, nil
	}
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && isSelect(e) {
		if ok, err := r.selectKey(e, key); ok {
			return nil, true, err
		}
	}
	// A focused Table is a box on the host, which leaves every key to the
	// program (SPEC §10.2): the arrows, Page Up, Page Down, Home and End
	// move its selection, and Enter acts on it.
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && e.Kind == view.Table {
		if ok, err := c.TableKey(e.ID, key); ok {
			return nil, true, err
		}
	}
	// So is a HottyDiff, whose keys (diffKeys) move its selection.
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && e.Kind == view.DiffView {
		if ok, err := c.DiffKey(e.ID, key); ok {
			return nil, true, err
		}
	}
	// And a HottyTree, whose keys (treeKeys) move its selection and open
	// and close its branches.
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && e.Kind == view.Tree {
		if ok, err := c.TreeKey(e.ID, key); ok {
			return nil, true, err
		}
	}
	// A focused Slider is a button on the host, which leaves arrows, Home
	// and End to the program (SPEC §10.2); so is a HottyRangeSlider's knob.
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && (e.Kind == view.Slider || e.Kind == view.Knob) {
		switch key {
		case "ArrowLeft", "ArrowDown":
			return nil, true, c.StepSlider(e.ID, -1, "")
		case "ArrowRight", "ArrowUp":
			return nil, true, c.StepSlider(e.ID, 1, "")
		case "Home", "End":
			return nil, true, c.StepSlider(e.ID, 0, key)
		}
	}
	// ? that nothing took switches the key hints' views (HottyKeyHints): a
	// focused field types it, so it reaches the program only from
	// elsewhere.
	if view.SameKey(key, "?") && c.ToggleHints() {
		return nil, true, nil
	}
	return nil, false, nil
}

// ListOpen reports whether a select's list is open. It shows in a surface
// of its own (Popover).
func (r *Rendition) ListOpen() bool { return r.openList() != nil }

// Popover is where a select's open list goes, which the program shows as
// a surface of its own, PopoverName, above the others: a list inside the
// surface would be cut at its edges, and a surface grown to hold it would
// hide what it covered. Under the select, which covers At in the
// surface's cells when Placed (the click said where); Cols wide at least,
// for its longest option; Rows high, a guess until the host's fit says.
type Popover struct {
	At     hotty.Area
	Placed bool
	Cols   int
	Rows   int
}

// Popover is where the open list goes; ok is false while none is open.
func (r *Rendition) Popover() (p Popover, ok bool) {
	l := r.openList()
	if l == nil {
		return Popover{}, false
	}
	e := r.C.V.Find(l.id)
	p = Popover{At: l.at, Placed: l.placed, Rows: 2*len(e.Options) + 1}
	for _, o := range e.Options {
		p.Cols = max(p.Cols, utf8.RuneCountInString(o.Label)+6)
	}
	if l.placed {
		p.Cols = max(p.Cols, l.at.W)
	}
	return p, true
}

// PopoverName is the open list's surface's name: the surface's, and
// "-list".
func (r *Rendition) PopoverName() string { return r.name + "-list" }

// PopoverDoc is the open list's document, for a=doc of PopoverName; Update
// brings it up to date from then on, until the list closes.
func (r *Rendition) PopoverDoc() string {
	r.pop = popover(r.C.V, r.theme, r.openList())
	if r.pop == nil {
		r.pop = el("div", "id", popoverID, "class", "k-popover")
	}
	return head + r.pop.html()
}

// popoverEvent is what the user did in the open list's surface: a click
// on an option picks it, as in the surface (clickSelect), and Update gives
// the select the keyboard back, which the host has nowhere now. The host's
// focus and blur there are not the view's.
func (r *Rendition) popoverEvent(ev hotty.Event) error {
	l := r.openList()
	id, part, ok := viewID(ev.Target)
	if ev.Kind != hotty.EventClick || l == nil || !ok || id != l.id || part == "" {
		return nil
	}
	err := r.clickSelect(r.C.V.Find(id), part, ev)
	r.host = keyboard{}
	return err
}

// openList is the open select's list, while the select is still there,
// a select, and has the keyboard, as in cells (profile §3.7); else none.
func (r *Rendition) openList() *list {
	if r.list == nil {
		return nil
	}
	c := r.C
	if e := c.V.Find(r.list.id); e == nil || !isSelect(e) || !c.St.Keyboard || c.St.Focus != r.list.id {
		r.list = nil
	}
	return r.list
}

// clickSelect is a click on a select, which opens or closes its list (by
// the cells the click says the select covers), or on an option in its
// list, which picks it and closes the list.
func (r *Rendition) clickSelect(e *view.Element, part string, ev hotty.Event) error {
	c := r.C
	c.Focus(e.ID)
	// The host's keyboard is on what was clicked: the select (a view id),
	// or an option (a DOM id, as Update names one).
	r.host = keyboard{true, e.ID}
	if part != "" {
		r.host.focus = partID(e.ID, part)
	}
	if part == "" {
		if r.list != nil && r.list.id == e.ID {
			r.list = nil
			return nil
		}
		at, placed := ev.Area()
		r.list = &list{id: e.ID, at: at, placed: placed, hi: max(pickedIndex(e), 0)}
		return nil
	}
	if len(part) < 2 || part[0] != partOption[0] {
		return nil
	}
	i, err := strconv.Atoi(part[1:])
	if err != nil || i < 0 || i >= len(e.Options) {
		return nil
	}
	r.list = nil
	return c.SetValue(e.ID, e.Options[i].Value)
}

// listKey is a key while a list is open, which the program has: as in
// cells, the arrows, Home, End, Page Up, Page Down and a letter move the
// highlight, Space or Enter picks it, and Escape closes the list. Tab
// closes it and goes on.
func (r *Rendition) listKey(l *list, key string) (bool, error) {
	e := r.C.V.Find(l.id)
	n := len(e.Options)
	switch key {
	case "ArrowUp":
		l.hi = max(l.hi-1, 0)
	case "ArrowDown":
		l.hi = min(l.hi+1, n-1)
	case "Home", "PageUp":
		l.hi = 0
	case "End", "PageDown":
		l.hi = n - 1
	case "Space", "Enter":
		r.list = nil
		if l.hi < 0 || l.hi >= n {
			return true, nil
		}
		return true, r.C.SetValue(e.ID, e.Options[l.hi].Value)
	case "Escape":
		r.list = nil
	case "Tab", "Shift+Tab":
		r.list = nil
		return false, nil
	default:
		if i := nextByLetter(e, l.hi, key); i >= 0 {
			l.hi = i
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

// nextByLetter is the next option after from, wrapping, whose label
// starts with key, a printable character (case aside); -1 if none does,
// or key is not one.
func nextByLetter(e *view.Element, from int, key string) int {
	n := len(e.Options)
	if utf8.RuneCountInString(key) != 1 {
		return -1
	}
	for k := 1; k <= n; k++ {
		j := (max(from, -1) + k + n) % n
		if strings.HasPrefix(strings.ToLower(e.Options[j].Label), strings.ToLower(key)) {
			return j
		}
	}
	return -1
}

// selectKey is a key on a closed select, which a host leaves to the
// program as a focused button's (SPEC §10.2): as in cells (profile §3.7),
// the arrows, Home, End, Page Up and Page Down pick an option, and so does
// a letter, the next option that starts with it.
func (r *Rendition) selectKey(e *view.Element, key string) (bool, error) {
	n := len(e.Options)
	if n == 0 {
		return false, nil
	}
	cur := pickedIndex(e)
	i := -1
	switch key {
	case "ArrowUp":
		i = max(cur-1, 0)
	case "ArrowDown":
		i = min(cur+1, n-1)
	case "Home", "PageUp":
		i = 0
	case "End", "PageDown":
		i = n - 1
	default:
		if i = nextByLetter(e, cur, key); i < 0 {
			return utf8.RuneCountInString(key) == 1, nil
		}
	}
	if i == cur {
		return true, nil
	}
	return true, r.C.SetValue(e.ID, e.Options[i].Value)
}

// pickedIndex is the index of a select's picked option, -1 if none is.
func pickedIndex(e *view.Element) int {
	picked, _ := e.Value.([]string)
	for i, o := range e.Options {
		if contains(picked, o.Value) {
			return i
		}
	}
	return -1
}

// holdEdit keeps, in the new document main, the value of the text control
// the user is editing as the host was last sent it, so Update sends none:
// the host has what was typed. An echo comes a key late while the user
// types on, and a host that sets a focused field's value from it (against
// SPEC §6.2, as hotty-blitz's attr op does) drops the keys typed since and
// leaves the caret behind them. Once the field is left, the view's value
// goes out, and the program's wins, as §6.2 has it.
func (r *Rendition) holdEdit(main *node) {
	if !r.host.on || r.host.focus == "" {
		return
	}
	id := domID(r.host.focus)
	old, cur := r.sent[0].find(id), main.find(id)
	if old == nil || cur == nil || old.tag != cur.tag {
		return
	}
	switch cur.tag {
	case "textarea":
		cur.kids = old.kids
	case "input":
		if v, ok := old.attr("value"); ok {
			cur.set("value", v)
		}
	}
}
