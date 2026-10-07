package cells

import (
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// fieldWidth is a text field's natural width when its label is shorter:
// an HTML input's default size.
const fieldWidth = 20

// trackWidth is a Slider's natural track width.
const trackWidth = 10

// Text-like controls (profile §3.5).

func isTextControl(e *view.Element) bool {
	return e != nil && (e.Kind == view.TextField || e.Kind == view.DateTime)
}

func isLongText(e *view.Element) bool { return e.Kind == view.TextField && e.Variant == "longText" }

// isSelect reports whether a Choice is a select: one value, from a list
// rather than chips.
func isSelect(e *view.Element) bool { return e != nil && e.Kind == view.Choice && len(e.Children) == 0 }

// controlWidth is a control's or a leaf's natural width.
func controlWidth(e *view.Element) int {
	switch e.Kind {
	case view.Button:
		return width(buttonFace(e, style{}))
	case view.TextField, view.DateTime:
		return max(Width(e.Label), fieldWidth)
	case view.CheckBox:
		return width(boxFace(false, e.Label, style{}))
	case view.Choice:
		if isSelect(e) {
			return width(selectLine(e, style{}, noWrap))
		}
		n := 0
		for i, ow := range optionWidths(e) {
			if i > 0 {
				n += optionGap
			}
			n += ow
		}
		return max(Width(e.Label), n)
	case view.Option:
		return Width(e.Label) + 4
	case view.Slider:
		n := trackWidth + 1 + sliderValueWidth(e)
		if e.Label != "" {
			n += Width(e.Label) + 1
		}
		return n
	case view.Image:
		return Width(imageText(e))
	case view.Icon:
		return Width(view.IconGlyph(e.Name))
	case view.Media:
		return Width("▶ " + e.Alt)
	case view.Placeholder:
		return Width(placeholderText(e))
	case view.Tab:
		return Width(e.Label)
	}
	return 0
}

// controlHeight is a control's or a leaf's height at width w.
func (l *layout) controlHeight(e *view.Element, w int) int {
	h := 1
	switch e.Kind {
	case view.TextField, view.DateTime:
		h = fieldRows(e)
		if e.Label != "" {
			h++
		}
	case view.Choice:
		switch {
		case isSelect(e):
			if l.r.listOpen(e) {
				h += len(e.Options)
			}
		default:
			h = 0
			if rows, _ := flow(optionWidths(e), optionGap, w); len(rows) > 0 {
				h = rows[len(rows)-1] + 1
			}
			if e.Label != "" {
				h++
			}
		}
	}
	return h + len(errorLines(e, w))
}

// errorLines are a control's error, "✗ message", wrapped.
func errorLines(e *view.Element, w int) [][]glyph {
	if e.Error == "" {
		return nil
	}
	return wrap(line("✗ "+e.Error, style{role: Error}), w, nil, repeat(" ", 2, style{}))
}

// fieldRows is how many rows a field's value takes: one, or a longText's
// lines, at least 3 and at most 8.
func fieldRows(e *view.Element) int {
	if !isLongText(e) {
		return 1
	}
	v, _ := e.Value.(string)
	return min(max(len(splitClusters(clusters(v))), 3), 8)
}

// buttonFace is a Button as drawn: "[ label ]", or the label alone when
// borderless.
func buttonFace(e *view.Element, st style) []glyph {
	label := line(buttonLabel(e), st)
	if e.Variant == "borderless" {
		return label
	}
	return concat(glyphs("[ ", st), label, glyphs(" ]", st))
}

// buttonLabel is a Button's content as one line: its Texts' plain text and
// its Icons' glyphs, a space apart.
func buttonLabel(e *view.Element) string {
	var parts []string
	var walk func(e *view.Element)
	walk = func(e *view.Element) {
		if e.A11y.Hidden {
			return
		}
		switch e.Kind {
		case view.Text:
			if s := strings.TrimSpace(view.PlainText(view.Markdown(e.Markdown))); s != "" {
				parts = append(parts, strings.ReplaceAll(s, "\n", " "))
			}
		case view.Icon:
			parts = append(parts, view.IconGlyph(e.Name))
		}
		for _, c := range e.Children {
			walk(c)
		}
	}
	for _, c := range e.Children {
		walk(c)
	}
	if len(parts) == 0 {
		return e.A11y.Label
	}
	return strings.Join(parts, " ")
}

// boxFace is "[x] label" or "[ ] label"; the box alone without a label.
func boxFace(on bool, label string, st style) []glyph {
	box := "[ ]"
	if on {
		box = "[x]"
	}
	gs := glyphs(box, st)
	if label != "" {
		gs = concat(gs, line(" "+label, style{}))
	}
	return gs
}

// optionFace is one of a Choice's options as drawn: a chip, "( label )"
// or "(● label)", or a box; reversed in the accent when it has the
// keyboard, the chip whole, else the box.
func optionFace(choice, o *view.Element, focused bool) []glyph {
	st := style{}
	if focused {
		st = style{role: Accent, attr: Reverse}
	}
	if choice.Variant != "chips" {
		return boxFace(o.Active, o.Label, st)
	}
	if o.Active {
		return line("(● "+o.Label+")", st)
	}
	return line("( "+o.Label+" )", st)
}

func optionWidths(e *view.Element) []int {
	ws := make([]int, len(e.Children))
	for i, o := range e.Children {
		ws[i] = width(optionFace(e, o, false))
	}
	return ws
}

// selectLine is a select's row: "label: value ▸", the value cut to fit w.
func selectLine(e *view.Element, valueSt style, w int) []glyph {
	var prefix []glyph
	if e.Label != "" {
		prefix = line(e.Label+": ", style{role: Muted})
	}
	value := line(pickedLabel(e), valueSt)
	if len(value) == 0 {
		value = line("…", style{role: Muted, attr: valueSt.attr})
	}
	if w == noWrap {
		for _, o := range e.Options {
			if Width(o.Label) > width(value) {
				value = line(o.Label, valueSt)
			}
		}
	}
	arrow := glyphs(" ▸", valueSt)
	room := w - width(prefix) - width(arrow)
	if room < 1 {
		return fit(concat(prefix, value, arrow), w)
	}
	return concat(prefix, fit(value, room), arrow)
}

// picked is the index of a select's value among its options, or -1.
func picked(e *view.Element) int {
	vs, _ := e.Value.([]string)
	for i, o := range e.Options {
		for _, v := range vs {
			if v == o.Value {
				return i
			}
		}
	}
	return -1
}

func pickedLabel(e *view.Element) string {
	if i := picked(e); i >= 0 {
		return e.Options[i].Label
	}
	return ""
}

func sliderValueWidth(e *view.Element) int {
	v, _ := e.Value.(float64)
	return max(Width(a2ui.NumberString(e.Min)), Width(a2ui.NumberString(e.Max)), Width(a2ui.NumberString(v)))
}

func imageText(e *view.Element) string {
	if e.Alt == "" {
		return "[image]"
	}
	return "[image: " + e.Alt + "]"
}

func placeholderText(e *view.Element) string {
	if e.State == a2ui.Pending {
		return "…"
	}
	return "! " + e.Type
}

// dateHint is a DateTime's placeholder: the form its value takes.
func dateHint(e *view.Element) string {
	switch {
	case e.Date && e.Time:
		return "YYYY-MM-DDTHH:MM"
	case e.Time:
		return "HH:MM"
	}
	return "YYYY-MM-DD"
}
