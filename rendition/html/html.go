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
	return out
}

// Autofocus is the a=focus that gives the surface the keyboard at the
// element the view starts focused (io_neuroplast_hotty.autofocus), to send
// after the document; "" when there is none.
func (r *Rendition) Autofocus() string {
	if !r.C.St.Keyboard || r.C.St.Focus == "" {
		return ""
	}
	return hotty.Focus(r.name, domID(r.C.St.Focus))
}

// Focus gives the surface the keyboard at an element, or with id "" where
// it was (the focus renderer function).
func (r *Rendition) Focus(id string) string {
	r.C.Focus(id)
	if id == "" {
		r.C.St.Keyboard = true
		return hotty.Focus(r.name, "")
	}
	return hotty.Focus(r.name, domID(id))
}

// Blur takes the keyboard back from the surface (the blur renderer
// function). The host commits the focused field first.
func (r *Rendition) Blur() string { return hotty.Blur(r.name) }

// Event is what the user did in the surface: it acts on the controller,
// and returns the commands to send in turn, before the deltas Update then
// makes. Events of other surfaces, and of kinds it does not know, do
// nothing.
func (r *Rendition) Event(ev hotty.Event) (cmds []string, err error) {
	if ev.Surface != r.name {
		return nil, nil
	}
	c := r.C
	switch ev.Kind {
	case hotty.EventFocus:
		c.St.Keyboard = true
		return nil, nil
	case hotty.EventBlur:
		c.St.Keyboard = false
		if r.pending == "" {
			return nil, nil
		}
		// The field is committed: the Shortcut runs on current inputs,
		// and the keyboard goes back where it was.
		key := r.pending
		r.pending = ""
		modal := c.St.Modal
		_, err = c.Shortcut(key)
		if c.St.Modal == modal {
			c.St.Keyboard = true
			cmds = append(cmds, hotty.Focus(r.name, ""))
		}
		return cmds, err
	}
	if ev.Target == backdropID && ev.Kind == hotty.EventClick {
		c.CloseModal()
		return nil, nil
	}
	id, part, ok := viewID(ev.Target)
	if !ok {
		return nil, nil
	}
	e := c.V.Find(id)
	if e == nil {
		return nil, nil
	}
	switch ev.Kind {
	case hotty.EventClick:
		if part != "" {
			return nil, nil
		}
		if e.Focusable() {
			c.Focus(id)
		}
		return nil, c.Activate(id)
	case hotty.EventInput, hotty.EventChange:
		c.St.Focus = id
		switch e.Kind {
		case view.CheckBox:
			on, _ := ev.Checked()
			return nil, c.SetValue(id, on)
		case view.Option:
			if on, _ := ev.Checked(); on != e.Active {
				return nil, c.Activate(id)
			}
			return nil, nil
		}
		return nil, c.SetValue(id, ev.Value())
	case hotty.EventSubmit:
		if e.Kind == view.Form {
			return nil, c.Submit(id)
		}
	}
	return nil, nil
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
