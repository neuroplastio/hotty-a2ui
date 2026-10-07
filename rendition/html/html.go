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

	"github.com/neuroplastio/hotty-go"

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
}

type keyboard struct {
	on    bool
	focus string
}

// New is the rendition of a controller's surface, as the host surface
// name.
func New(c *view.Controller, name string) *Rendition { return &Rendition{C: c, name: name} }

// Name is the HOTTY surface's name.
func (r *Rendition) Name() string { return r.name }

// Doc is the whole document of the view as it is now, for a=doc; the host
// has it from then on.
func (r *Rendition) Doc() string {
	main, layer := surface(r.C.V)
	r.sent = []*node{main, layer}
	return head + main.html() + layer.html()
}

// Update is the deltas that bring the host's document to the view as it
// is now: none when nothing changed, or before the document was sent.
func (r *Rendition) Update() []string {
	if r.sent == nil {
		return nil
	}
	main, layer := surface(r.C.V)
	out := r.diff(r.sent[0], main, nil)
	out = r.diff(r.sent[1], layer, out)
	r.sent = []*node{main, layer}
	// The keyboard, when the controller moved it: autofocus, the focus
	// and blur functions, the end of a Shortcut (Key).
	st := r.C.St
	if want := (keyboard{st.Keyboard, st.Focus}); want != r.host && (want.on || r.host.on) {
		switch {
		case !want.on:
			out = append(out, hotty.Blur(r.name))
		case want.focus != "" && r.C.V.Find(want.focus) != nil:
			out = append(out, hotty.Focus(r.name, domID(want.focus)))
		default:
			out = append(out, hotty.Focus(r.name, ""))
		}
		r.host = want
	}
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
		c.St.Keyboard, r.host.on = false, false
		if r.pending == "" {
			return nil
		}
		// The field is committed: the Shortcut runs on current inputs.
		// The blur was the rendition's own, so the keyboard goes back
		// where it was (Update), unless the Shortcut moves it.
		key := r.pending
		r.pending = ""
		c.St.Keyboard = true
		_, err := c.Shortcut(key)
		return err
	}
	if ev.Target == backdropID && ev.Kind == hotty.EventClick {
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
		if part != "" {
			return nil
		}
		if e.Focusable() {
			c.Focus(id)
			r.host = keyboard{true, id}
		}
		return c.Activate(id)
	case hotty.EventInput, hotty.EventChange:
		c.St.Focus, r.host.focus = id, id
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
	if key == "Escape" && c.St.Modal != "" {
		c.CloseModal()
		return nil, true, nil
	}
	return nil, false, nil
}
