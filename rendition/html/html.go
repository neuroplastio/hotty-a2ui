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
	// Away is set while the keyboard is in another rendition of the same
	// surface (cells beside it): the host's surface gives it up, and its
	// blur leaves the controller's keyboard where it is.
	Away bool
	// back is set when the keyboard goes back after a Shortcut's blur:
	// to where the host had it, which Tab may have moved unseen.
	back bool
	// list is the select whose list is open, nil when none is.
	list *list
}

type keyboard struct {
	on    bool
	focus string
}

// New is the rendition of a controller's surface, as the host surface
// name.
func New(c *view.Controller, name string) *Rendition { return &Rendition{C: c, name: name} }

// SetTheme paints the surface in a theme's colours from the next Doc or
// Update on; the zero theme is the host's own.
func (r *Rendition) SetTheme(th theme.Theme) { r.theme = th }

// Name is the HOTTY surface's name.
func (r *Rendition) Name() string { return r.name }

// Doc is the whole document of the view as it is now, for a=doc; the host
// has it from then on.
func (r *Rendition) Doc() string {
	main, layer := surface(r.C.V, r.theme, r.openList())
	r.sent = []*node{main, layer}
	return head + main.html() + layer.html()
}

// Update is the deltas that bring the host's document to the view as it
// is now: none when nothing changed, or before the document was sent.
func (r *Rendition) Update() []string {
	if r.sent == nil {
		return nil
	}
	main, layer := surface(r.C.V, r.theme, r.openList())
	r.holdEdit(main)
	out := r.diff(r.sent[0], main, nil)
	out = r.diff(r.sent[1], layer, out)
	r.sent = []*node{main, layer}
	// The keyboard, when the controller moved it: autofocus, the focus
	// and blur functions, the end of a Shortcut (Key).
	// While a select's list is open, the program has the keyboard (Key):
	// a host scrolls with the arrows a focused button leaves (SPEC §5.3),
	// so they would never reach it.
	st := r.C.St
	want := keyboard{st.Keyboard && !r.Away && r.openList() == nil, st.Focus}
	if want != r.host && (want.on || r.host.on) {
		switch {
		case !want.on:
			out = append(out, hotty.Blur(r.name))
		case r.back && want.focus == r.host.focus:
			out = append(out, hotty.Focus(r.name, ""))
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
	if ev.Surface != r.name {
		return nil
	}
	c := r.C
	switch ev.Kind {
	case hotty.EventFocus:
		c.St.Keyboard, r.host.on = true, true
		return nil
	case hotty.EventBlur:
		if r.openList() != nil && !r.host.on {
			// The blur an open list asked for (Update): the controller
			// keeps the keyboard, and the program works the list.
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
		if e.Kind == view.Slider && (part == partLess || part == partMore) {
			n := 1
			if part == partLess {
				n = -1
			}
			return c.StepSlider(id, n, "")
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
		// A field the host commits as it gives the keyboard up keeps its
		// value, not the focus: that is in the other rendition now.
		r.host.focus = id
		if !r.Away {
			c.St.Focus = id
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
			i, err := strconv.Atoi(part[1:])
			if err != nil {
				return nil
			}
			return c.SetValue(id, notchValue(e, i))
		}
	}
	return nil
}

// Key is a key the program read while the surface is the one keys apply
// to, as a W3C key value with its modifiers ("Control+s", "Escape"). A
// Shortcut takes it. While the surface has the keyboard, a field may
// hold an edit the host has not committed, so Key sends a=blur and the
// Shortcut runs when the blur comes back (Event), after the field's
// change; then the keyboard goes back to the surface. Else Escape closes
// an open Modal. ok reports whether the surface took the key.
func (r *Rendition) Key(key string) (cmds []string, ok bool, err error) {
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
	if key == "Escape" && c.St.Modal != "" {
		c.CloseModal()
		return nil, true, nil
	}
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && isSelect(e) {
		if ok, err := r.selectKey(e, key); ok {
			return nil, true, err
		}
	}
	// A focused Slider is a button on the host, which leaves arrows, Home
	// and End to the program (SPEC §10.2).
	if e := c.V.Find(c.St.Focus); c.St.Keyboard && e != nil && e.Kind == view.Slider {
		switch key {
		case "ArrowLeft", "ArrowDown":
			return nil, true, c.StepSlider(e.ID, -1, "")
		case "ArrowRight", "ArrowUp":
			return nil, true, c.StepSlider(e.ID, 1, "")
		case "Home", "End":
			return nil, true, c.StepSlider(e.ID, 0, key)
		}
	}
	return nil, false, nil
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
	case " ", "Enter":
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
	if utf8.RuneCountInString(key) != 1 || key == " " {
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
			return utf8.RuneCountInString(key) == 1 && key != " ", nil
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
