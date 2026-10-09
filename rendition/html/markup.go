package html

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
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
	// list is the select whose list is open, if one is.
	list *list
	// now is the time the markup shows (Rendition.Clock).
	now time.Time
	// anim is how soon what it made changes by itself (Rendition.Animating).
	anim time.Duration
}

// animate notes that the markup changes again after d.
func (m *markup) animate(d time.Duration) {
	if d > 0 && (m.anim == 0 || d < m.anim) {
		m.anim = d
	}
}

// list is an open select's list: the rendition's, as cells' is (profile
// §3.8). It goes in the layer, under the select where a click said the
// select is (SPEC §9), else in the middle; hi is the option focused.
type list struct {
	id     string
	at     hotty.Area
	placed bool
	hi     int
}

// surface is the surface's two top elements: the view, and the layer an
// open Modal's content shows in.
func surface(v *view.Surface, th theme.Theme, keys string, fit bool, m *markup) (main, layer *node) {
	l := m.list
	class := "k-surface"
	if fit {
		class += " k-fit"
	}
	main = el("div", "id", surfaceID, "class", class).add(m.element(v.Root))
	layer = el("div", "id", layerID, "class", "k-layer")
	if css := themeCSS(th); css != "" {
		main.set("style", css)
		layer.set("style", css)
	}
	// The surface's keymap, for every text field in it (SPEC §10.2): the
	// layer holds an open Modal's.
	if keys != "" {
		main.set("data-keys", keys)
		layer.set("data-keys", keys)
	}
	if v.Overlay != nil {
		main.flag("inert", true)
		layer.add(
			el("div", "id", backdropID, "class", "k-backdrop", "data-on", "click"),
			el("div", "id", dialogID, "class", "k-dialog", "role", "dialog", "aria-modal", "true").add(m.element(v.Overlay)),
		)
	}
	if l != nil {
		if e := v.Find(l.id); e != nil && isSelect(e) {
			layer.add(el("div", "id", dismissID, "class", "k-dismiss", "data-on", "click"))
		}
	}
	return main, layer
}

// popover is an open select's list as a document of its own, which the
// program places over the rest (Rendition.Popover): out of the surface,
// it is not cut at the surface's edges, and hides nothing it doesn't
// cover. Nil when no list is open.
func popover(v *view.Surface, th theme.Theme, l *list) *node {
	if l == nil {
		return nil
	}
	e := v.Find(l.id)
	if e == nil || !isSelect(e) {
		return nil
	}
	n := el("div", "id", popoverID, "class", "k-popover")
	if css := themeCSS(th); css != "" {
		n.set("style", css)
	}
	return n.add((&markup{list: l}).listbox(e, l))
}

// isSelect reports whether a Choice is a select: one value, its options
// in a list (the checkbox display).
func isSelect(e *view.Element) bool { return e.Kind == view.Choice && len(e.Children) == 0 }

// listbox is a select's open list: a row an option, the picked one filled
// and the one the keys are on (hi) marked, as focus would be. The options
// are out of the Tab order: the program has the keyboard (Rendition.Key).
// It fills its popover, and scrolls when the program makes that shorter.
func (m *markup) listbox(e *view.Element, l *list) *node {
	n := el("div", "id", partID(e.ID, partList), "class", "k-listbox", "role", "listbox")
	if e.Label != "" {
		n.set("aria-labelledby", partID(e.ID, partLabel))
	}
	picked, _ := e.Value.([]string)
	for i, o := range e.Options {
		class := "k-item k-btn k-borderless k-option"
		if i == l.hi {
			class += " k-hi"
			n.set("aria-activedescendant", partID(e.ID, partOption+strconv.Itoa(i)))
		}
		n.add(el("button", "id", partID(e.ID, partOption+strconv.Itoa(i)), "type", "button", "class", class, "tabindex", "-1",
			"role", "option", "aria-selected", boolString(contains(picked, o.Value))).add(texts(o.Label)...))
	}
	return n
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
		n.text = groupBreaks(strings.TrimSpace(view.PlainText(view.Markdown(e.Markdown))), n.text)
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
	case view.Progress:
		// A bar and a percentage under the label. Without a value, a
		// quarter of the bar sweeps across it on the view's clock, as in
		// cells (profile §3.4): the element's --k-at is where it starts,
		// so that a tick is one attribute's delta.
		f, known := e.Fraction()
		n = el("div", "id", id, "class", "k-progress", "role", "progressbar",
			"aria-valuemin", "0", "aria-valuemax", a2ui.NumberString(e.Max))
		if e.Label != "" {
			n.set("aria-label", e.Label)
			n.add(el("div", "class", "k-progress-label").add(texts(e.Label)...))
		}
		track := el("span", "class", "k-progress-track")
		bar := el("div", "class", "k-progress-bar").add(track)
		if known {
			n.set("aria-valuenow", a2ui.NumberString(f*e.Max))
			class := "k-progress-fill"
			if f >= 1 {
				class += " k-done"
			}
			track.add(el("span", "class", class, "style", "width: "+strconv.FormatFloat(f*100, 'f', 2, 64)+"%"))
			bar.add(el("output", "class", "k-progress-value").add(txt(fmt.Sprintf("%.0f%%", f*100))))
		} else {
			at := float64(view.ProgressStep(m.now))*125/view.ProgressSteps - 25
			n.set("style", "--k-at: "+strconv.FormatFloat(at, 'f', 2, 64)+"%")
			track.add(el("span", "class", "k-progress-fill k-indeterminate"))
			m.animate(view.ProgressInterval)
		}
		n.add(bar)
	case view.Spinner:
		// The frame the view's clock is at, as in cells, then its label;
		// the frame has an id, so that a tick is one text's delta. It
		// keeps its set's widest frame's cells (in the mono face, a ch
		// each), so that the label stays put.
		n = el("span", "id", id, "class", "k-spinner", "role", "status")
		frame, cols := "", 0
		set := e.SpinnerFrames()
		for _, f := range set.Frames {
			cols = max(cols, cells.Width(f))
		}
		if e.Active {
			frame = set.Frame(m.now)
			m.animate(set.Interval)
		}
		n.add(el("span", "id", partID(e.ID, partFrame), "class", "k-spinner-frame", "aria-hidden", "true",
			"style", "min-width: "+strconv.Itoa(cols)+"ch").add(texts(frame)...))
		if e.Label != "" {
			n.add(el("span", "class", "k-spinner-label").add(texts(e.Label)...))
		}
	case view.Table:
		n = table(e)
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
		if e.Item {
			class += " k-item"
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
		out := el("output", "id", partID(e.ID, partOutput), "for", id, "style", "min-width: "+strconv.Itoa(e.SliderWidth())+"ch").add(txt(a2ui.NumberString(f)))
		outer = m.field(e, "k-field", label(e), el("div", "id", partID(e.ID, partRange), "class", "k-slide").add(less, n, more, out))
	case view.Choice:
		picked, _ := e.Value.([]string)
		if len(e.Children) == 0 {
			// A select is a button with the picked option's label; its
			// list opens in the layer (listbox). Hosts draw select unevenly
			// (Blitz not at all), and the list is the program's to place
			// (SPEC §9).
			value := el("span", "class", "k-value")
			for _, o := range e.Options {
				if contains(picked, o.Value) {
					value.add(texts(o.Label)...)
					break
				}
			}
			if len(value.kids) == 0 {
				value.set("class", "k-value k-none").add(txt("…"))
			}
			open := m.list != nil && m.list.id == e.ID
			n = el("button", "id", id, "type", "button", "class", "k-input k-select", "aria-haspopup", "listbox",
				"aria-expanded", boolString(open)).add(value, el("span", "class", "k-caret", "aria-hidden", "true").add(txt("▾")))
			if open {
				n.set("aria-controls", partID(e.ID, partList))
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
	if e.Keys != "" {
		n.set("data-keys", e.Keys)
	}
	if outer == nil {
		outer = n
	}
	if e.Weight > 0 {
		// kit.css: in a Row, the weight is the child's share of it; in a
		// Column, of the rows to spare.
		class, _ := outer.attr("class")
		outer.set("class", strings.TrimSpace(class+" k-weighted")).set("style", "--k-w: "+a2ui.NumberString(e.Weight))
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
	put("--k-info", th.Info)
	put("--k-r-button", th.Shape.Button)
	put("--k-r-chip", th.Shape.Chip)
	put("--k-r-card", th.Shape.Card)
	put("--k-r-field", th.Shape.Field)
	put("--k-link", th.Accent)
	put("--k-focus", th.Accent)
	return b.String()
}

// tableKeys are the keys a Table gives the program, which works its
// selection (Rendition.Key): on a host that scrolls a surface with them,
// they would never reach it otherwise (SPEC §10.2, keys for the program).
const tableKeys = "ArrowUp=program ArrowDown=program PageUp=program PageDown=program Home=program End=program"

// table is a Table: a focusable box holding a table, whose keys reach the
// program (tableKeys; Enter reaches it anyway), and whose rows each report
// a click, which selects the row, or acts on it once selected (Event).
// The body holds the rows the view shows (Element.Top), as cells does: the
// selection never scrolls out of sight, which a host's own scrolling
// would let it do. Under it, while it scrolls, which rows show.
func table(e *view.Element) *node {
	box := el("div", "id", domID(e.ID), "class", "k-table", "tabindex", "0", "role", "grid",
		"aria-rowcount", strconv.Itoa(len(e.Cells)), "data-keys", tableKeys)
	head := el("tr")
	for _, c := range e.Columns {
		head.add(el("th", "class", "k-"+c.Align).add(texts(c.Header)...))
	}
	body := el("tbody")
	top, end := e.Top, len(e.Cells)
	if e.Height > 0 {
		end = min(top+e.Height, end)
	}
	sel := e.SelectedRow()
	for i := top; i < end; i++ {
		tr := el("tr", "id", partID(e.ID, partRow+strconv.Itoa(i)), "class", "k-row", "data-on", "click",
			"aria-rowindex", strconv.Itoa(i+2), "aria-selected", strconv.FormatBool(i == sel))
		if i == sel {
			tr.set("class", "k-row k-sel")
		}
		for j, c := range e.Columns {
			tr.add(el("td", "class", "k-"+c.Align).add(texts(e.Cells[i][j])...))
		}
		body.add(tr)
	}
	box.add(el("table").add(el("thead").add(head), body))
	switch n := len(e.Cells); {
	case n == 0:
		box.add(el("div", "class", "k-table-note").add(txt("No rows")))
	case e.Height > 0 && n > e.Height:
		box.add(el("div", "class", "k-table-note").add(txt(strconv.Itoa(top+1) + "–" + strconv.Itoa(end) + " of " + strconv.Itoa(n))))
	}
	return box
}
