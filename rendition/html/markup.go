package html

import (
	"math"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// markup makes a view's elements into the document's. Each element is one
// node with its DOM id, the one its events name; a control is wrapped with
// its label and its error, which have ids of their own so that a change
// to them is a small delta.
type markup struct {
	// forms is how many forms enclose the element: HTML cannot nest them,
	// so an inner Form is a group, and Enter submits the outer one.
	forms int
}

// surface is the surface's two top elements: the view, and the layer an
// open Modal's content shows in.
func surface(v *view.Surface, th theme.Theme) (main, layer *node) {
	var m markup
	main = el("div", "id", surfaceID, "class", "k-surface").add(m.element(v.Root))
	layer = el("div", "id", layerID, "class", "k-layer")
	if css := themeCSS(th); css != "" {
		main.set("style", css)
		layer.set("style", css)
	}
	if v.Overlay != nil {
		main.flag("inert", true)
		layer.add(
			el("div", "id", backdropID, "class", "k-backdrop", "data-on", "click"),
			el("div", "id", dialogID, "class", "k-dialog", "role", "dialog", "aria-modal", "true").add(m.element(v.Overlay)),
		)
	}
	return main, layer
}

func (m *markup) all(es []*view.Element) []*node {
	var out []*node
	for _, e := range es {
		if n := m.element(e); n != nil {
			out = append(out, n)
		}
	}
	return out
}

func (m *markup) element(e *view.Element) *node {
	if e == nil {
		return nil
	}
	id := domID(e.ID)
	// outer is what the element adds to its parent; n is the element
	// itself, which its id and its accessibility go on.
	var n, outer *node
	switch e.Kind {
	case view.Stack:
		n = el("div", "id", id, "class", stackClass(e)).add(m.all(e.Children)...)
	case view.Card:
		n = el("div", "id", id, "class", "k-card k-stack k-col").add(m.all(e.Children)...)
	case view.Text:
		n = el("div", "id", id, "class", "k-text k-v-"+e.Variant)
		n.text, n.raw = emojiHTML(view.MarkdownHTML(e.Markdown)), true
	case view.Image:
		n = el("img", "id", id, "class", "k-img k-v-"+e.Variant+" k-fit-"+e.Fit, "src", e.URL, "alt", e.Alt)
	case view.Icon:
		g, emoji := iconText(e.Name)
		class := "k-icon"
		if emoji {
			class += " k-emoji"
		}
		n = el("span", "id", id, "class", class, "role", "img", "aria-label", e.Name).add(txt(g))
	case view.Media:
		// A link, not a hyperlink: it takes the keyboard as it does in
		// cells, and its click is the program's, which opens it.
		n = el("a", "id", id, "class", "k-media", "href", e.URL).add(texts("▶ " + e.Alt)...)
	case view.Divider:
		if e.Dir == view.Vertical {
			n = el("div", "id", id, "class", "k-vr", "role", "separator", "aria-orientation", "vertical")
		} else {
			n = el("hr", "id", id, "class", "k-hr")
		}
	case view.Button:
		class := "k-btn k-" + e.Variant
		if len(e.Children) == 1 && e.Children[0].Kind == view.Icon {
			class += " k-icon-btn"
		}
		n = el("button", "id", id, "type", "button", "class", class).flag("disabled", e.Disabled).add(m.all(e.Children)...)
	case view.TextField:
		v, _ := e.Value.(string)
		switch e.Variant {
		case "longText":
			n = el("textarea", "id", id, "class", "k-input", "data-on", "input")
			if v != "" {
				n.add(txt(v))
			}
		default:
			typ := map[string]string{"obscured": "password", "number": "number"}[e.Variant]
			n = el("input", "id", id, "class", "k-input", "type", fallback(typ, "text"), "value", v, "data-on", "input")
		}
		if e.Placeholder != "" {
			n.set("placeholder", e.Placeholder)
		}
		outer = m.field(e, "k-field", label(e), n)
	case view.CheckBox:
		n = el("input", "id", id, "type", "checkbox").flag("checked", e.Value == true)
		outer = m.field(e, "k-check", n, label(e))
	case view.DateTime:
		// A text field, as in cells: a value in ISO 8601, its form the
		// placeholder. Hosts draw date and time inputs unevenly (Blitz
		// not at all), and none takes an offset such as "Z".
		v, _ := e.Value.(string)
		hint := e.DateHint()
		n = el("input", "id", id, "class", "k-input k-date", "type", "text", "value", v, "data-on", "input",
			"placeholder", hint, "size", strconv.Itoa(max(len(hint), len(v))), "inputmode", "numeric")
		outer = m.field(e, "k-field", label(e), n)
	case view.Slider:
		// One button is the slider, as one focusable: arrow keys reach the
		// program (SPEC §10.2), which steps it (Key). Its track shows the
		// value; − and + step it by a click, outside the Tab order. Hosts
		// draw a range input unevenly (Blitz not at all).
		f, _ := e.Value.(float64)
		pct := 0.0
		if e.Max > e.Min {
			pct = math.Max(0, math.Min(100, (f-e.Min)/(e.Max-e.Min)*100))
		}
		at := strconv.FormatFloat(pct, 'f', 2, 64) + "%"
		n = el("button", "id", id, "type", "button", "class", "k-track", "role", "slider",
			"aria-valuemin", a2ui.NumberString(e.Min), "aria-valuemax", a2ui.NumberString(e.Max),
			"aria-valuenow", a2ui.NumberString(f)).add(
			el("span", "class", "k-rail"),
			el("span", "class", "k-fill", "style", "width: "+at),
			el("span", "class", "k-knob", "style", "left: "+at),
		)
		// A drag reports the element under the pointer, not where on it
		// (SPEC §9.1), so the track is cut into notches across it, each
		// a drag target: pressing one sets its value, crossing them moves
		// it.
		k := notches(e)
		w := 100 / float64(k)
		for i := range k {
			n.add(el("span", "id", partID(e.ID, partNotch+strconv.Itoa(i)), "class", "k-notch", "data-on", "drag",
				"style", "left: "+strconv.FormatFloat(float64(i)*w, 'f', 3, 64)+"%; width: "+strconv.FormatFloat(w, 'f', 3, 64)+"%"))
		}
		less := el("button", "id", partID(e.ID, partLess), "type", "button", "class", "k-step", "tabindex", "-1", "aria-label", "less").add(txt("−"))
		more := el("button", "id", partID(e.ID, partMore), "type", "button", "class", "k-step", "tabindex", "-1", "aria-label", "more").add(txt("+"))
		out := el("output", "id", partID(e.ID, partOutput), "for", id).add(txt(a2ui.NumberString(f)))
		outer = m.field(e, "k-field", label(e), el("div", "id", partID(e.ID, partRange), "class", "k-slide").add(less, n, more, out))
	case view.Choice:
		picked, _ := e.Value.([]string)
		if len(e.Children) == 0 {
			n = el("select", "id", id, "class", "k-input")
			if len(picked) == 0 {
				// Nothing is picked yet: a select shows its first option,
				// which the user could then not pick.
				n.add(el("option", "value", "", "disabled", "", "selected", "").add(txt("…")))
			}
			for _, o := range e.Options {
				n.add(el("option", "value", o.Value).flag("selected", contains(picked, o.Value)).add(txt(o.Label)))
			}
			outer = m.field(e, "k-field", label(e), n)
			break
		}
		n = el("div", "id", id, "class", "k-options", "role", "group")
		for _, o := range e.Children {
			n.add(m.option(e, o))
		}
		var head *node
		if e.Label != "" {
			head = el("div", "id", partID(e.ID, partLabel), "class", "k-label").add(texts(e.Label)...)
			n.set("aria-labelledby", partID(e.ID, partLabel))
		}
		outer = m.field(e, "k-field", head, n)
	case view.Tabs:
		bar := el("div", "id", partID(e.ID, partTabs), "class", "k-tablist", "role", "tablist")
		panel := el("div", "id", partID(e.ID, partPanel), "class", "k-tabpanel k-stack k-col", "role", "tabpanel")
		for _, c := range e.Children {
			if c.Kind == view.Tab {
				bar.add(el("button", "id", domID(c.ID), "type", "button", "class", "k-tab", "role", "tab",
					"aria-selected", boolString(c.Active)).add(texts(c.Label)...))
				continue
			}
			panel.add(m.element(c))
		}
		n = el("div", "id", id, "class", "k-tabs k-stack k-col").add(bar, panel)
	case view.Modal:
		if e.Clickable {
			n = el("button", "id", id, "type", "button", "class", "k-trigger")
		} else {
			n = el("div", "id", id, "class", "k-modal k-stack k-col")
		}
		n.set("aria-haspopup", "dialog").set("aria-expanded", boolString(e.Open)).add(m.all(e.Children)...)
	case view.Form:
		if m.forms > 0 {
			n = el("div", "id", id, "class", "k-form", "role", "form")
			m.forms++
			n.add(m.all(e.Children)...)
			m.forms--
			break
		}
		m.forms++
		n = el("form", "id", id, "class", "k-form").add(m.all(e.Children)...)
		m.forms--
		// Enter in a field submits a form that has a submit button; this
		// one is out of sight and out of the Tab order.
		n.add(el("button", "id", partID(e.ID, partSubmit), "type", "submit", "class", "k-submit", "tabindex", "-1", "aria-hidden", "true"))
	case view.Placeholder:
		n = el("span", "id", id, "class", "k-ph k-warn").add(txt("! " + e.Type))
		if e.State == a2ui.Pending {
			n = el("span", "id", id, "class", "k-ph").add(txt("…"))
		}
	default:
		n = el("div", "id", id, "class", "k-stack k-col").add(m.all(e.Children)...)
	}
	accessible(n, e)
	if outer == nil {
		outer = n
	}
	if e.Weight > 0 {
		outer.set("style", "flex-grow: "+a2ui.NumberString(e.Weight))
	}
	return outer
}

// option is one of a Choice's options: a checkbox among several picked,
// else a chip that is pressed or not.
func (m *markup) option(choice, o *view.Element) *node {
	id := domID(o.ID)
	if choice.Multiple && choice.Variant != "chips" {
		in := el("input", "id", id, "type", "checkbox").flag("checked", o.Active)
		return el("label", "id", partID(o.ID, partWrap), "class", "k-opt").add(in).add(texts(o.Label)...)
	}
	return el("button", "id", id, "type", "button", "class", "k-chip", "aria-pressed", boolString(o.Active)).add(texts(o.Label)...)
}

// field wraps a control with what goes with it, and its error, which is
// always there (empty, and not shown, while there is none) so that it
// comes and goes by a text delta.
func (m *markup) field(e *view.Element, class string, parts ...*node) *node {
	w := el("div", "id", partID(e.ID, partWrap), "class", class).add(parts...)
	errID := partID(e.ID, partError)
	msg := el("div", "id", errID, "class", "k-error", "aria-live", "polite")
	if e.Error != "" {
		msg.add(txt("✗ " + e.Error))
	}
	return w.add(msg)
}

// label is a control's label, if it has one.
func label(e *view.Element) *node {
	if e.Label == "" {
		return nil
	}
	return el("label", "id", partID(e.ID, partLabel), "class", "k-label", "for", domID(e.ID)).add(texts(e.Label)...)
}

func accessible(n *node, e *view.Element) {
	if e.Error != "" {
		n.set("aria-invalid", "true").set("aria-describedby", partID(e.ID, partError))
	}
	a := e.A11y
	if a.Label != "" {
		n.set("aria-label", a.Label)
	}
	if a.Description != "" {
		n.set("aria-description", a.Description)
	}
	if a.Live != "" {
		n.set("aria-live", a.Live)
	}
	if a.Hidden {
		n.set("aria-hidden", "true")
	}
}

func stackClass(e *view.Element) string {
	c := []string{"k-stack", "k-col"}
	if e.Dir == view.Horizontal {
		c[1] = "k-row"
	}
	if e.Justify != "" {
		c = append(c, "k-j-"+e.Justify)
	}
	if e.Align != "" {
		c = append(c, "k-a-"+e.Align)
	}
	if e.Scroll {
		c = append(c, "k-scroll")
	}
	if e.Dir == view.Horizontal && wraps(e) {
		c = append(c, "k-wrap")
	}
	return strings.Join(c, " ")
}

// wraps: a Row wraps, as a page's inline content does, rather than cut
// what it cannot fit, when it holds a form field (squeezed, its value
// would be cut) or holds only inline things: texts, icons, buttons, small
// pictures. A Row of columns or cards shrinks them instead.
func wraps(e *view.Element) bool {
	field, inline := false, true
	for _, k := range e.Children {
		switch k.Kind {
		case view.TextField, view.DateTime, view.Slider:
			field = true
		case view.Choice:
			field = field || len(k.Children) == 0
		case view.Text, view.Icon, view.Button, view.Media, view.Placeholder:
		case view.Image:
			inline = inline && (k.Variant == "icon" || k.Variant == "avatar")
		default:
			inline = false
		}
	}
	return field || inline && len(e.Children) > 1
}

// notches is how many a Slider's track is cut into: one for each of its
// steps and its two ends, at most 41 (a twentieth of the range when it
// has no step).
func notches(e *view.Element) int {
	steps := 20.0
	if q := e.SliderStep(); q > 0 {
		steps = math.Round((e.Max - e.Min) / q)
	}
	return int(math.Min(math.Max(steps, 1), 40)) + 1
}

// notchValue is the value a Slider's notch sets.
func notchValue(e *view.Element, i int) float64 {
	k := notches(e)
	return e.Min + float64(min(max(i, 0), k-1))/float64(k-1)*(e.Max-e.Min)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func fallback(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

// themeCSS is a theme's colours as the kit's variables, and the surface's
// own colours; "" for the host's palette, which the zero theme keeps.
func themeCSS(th theme.Theme) string {
	if th == (theme.Theme{}) || th.Name == theme.Default.Name {
		return ""
	}
	var b strings.Builder
	put := func(k, v string) {
		if v != "" {
			b.WriteString(k + ":" + v + ";")
		}
	}
	put("--k-bg", th.Bg)
	put("--k-fg", th.Fg)
	put("--k-muted", th.Muted)
	put("--k-border", th.Border)
	put("--k-accent", th.Accent)
	put("--k-on-accent", th.Bg)
	put("--k-selection", th.Selection)
	put("--k-surface", th.Surface)
	put("--k-error", th.Error)
	put("--k-warning", th.Warning)
	put("--k-success", th.Success)
	put("--k-r-button", th.Shape.Button)
	put("--k-r-chip", th.Shape.Chip)
	put("--k-r-card", th.Shape.Card)
	put("--k-r-field", th.Shape.Field)
	put("--k-link", th.Accent)
	put("--k-focus", th.Accent)
	return b.String()
}
