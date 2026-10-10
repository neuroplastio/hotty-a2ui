package cells

import (
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/icons"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// fieldWidth is a text field's natural width when its label is shorter:
// an HTML input's default size.
const fieldWidth = 20

// trackWidth is a Slider's natural track width.
const trackWidth = 10

// gutter is the columns left of a field (profile §3.4): a bar, "┃ ",
// while it has the keyboard, as huh draws its fields.
const gutter = 2

// prompt is the columns of a one-line text field's prompt, "> ".
const prompt = 2

// isField reports whether an element is a field: a control with a gutter
// and, when it has a label of its own, a title row (profile §3.4).
func isField(e *view.Element) bool {
	if e == nil {
		return false
	}
	switch e.Kind {
	case view.TextField, view.DateTime, view.CheckBox, view.Switch, view.Choice, view.Slider, view.RangeSlider:
		return true
	}
	return false
}

// isControlElement reports whether an element is a control a Column may
// set a blank row apart from (separator): a field, a Button that is not a
// List's row, a Progress, a Spinner, a Table, a HottyList, a
// HottyScrollView or a HottyTree.
func isControlElement(e *view.Element) bool {
	if isField(e) {
		return true
	}
	switch {
	case e == nil:
		return false
	case e.Kind == view.Button:
		return !e.Item
	}
	return e.Kind == view.Progress || e.Kind == view.Spinner || e.Kind == view.Table || e.Kind == view.RichList || e.Kind == view.ScrollView ||
		e.Kind == view.Tree
}

// hasTitle reports whether an element has a title row of its own: a text
// field, a DateTime, a Choice or a Progress with a label, a Table, its
// header, and a HottyList, its status line. A CheckBox's, a HottySwitch's,
// a Slider's and a Spinner's labels are on their one row.
func hasTitle(e *view.Element) bool {
	switch e.Kind {
	case view.Table, view.RichList:
		return true
	case view.TextField, view.DateTime, view.Choice, view.Progress:
		return e.Label != ""
	}
	return false
}

// progressWidth is a Progress bar's natural width; percentWidth, the
// columns of its percentage, " 100%".
const (
	progressWidth = 20
	percentWidth  = 5
)

// eighths are the glyphs of a Progress bar's last, partly filled cell,
// by eighths filled.
var eighths = [...]string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// spinnerWidth is the columns a Spinner's frames take: its widest frame,
// so that its label stays put while it spins.
func spinnerWidth(e *view.Element) int {
	n := 0
	for _, f := range e.SpinnerFrames().Frames {
		n = max(n, Width(f))
	}
	return n
}

// Text-like controls (profile §3.5).

func isTextControl(e *view.Element) bool {
	return e != nil && (e.Kind == view.TextField || e.Kind == view.DateTime)
}

func isLongText(e *view.Element) bool { return e.Kind == view.TextField && e.Variant == "longText" }

// isSelect reports whether a Choice is a select: one value, from a list
// rather than chips.
func isSelect(e *view.Element) bool { return e != nil && e.Kind == view.Choice && len(e.Children) == 0 }

// controlWidth is a control's or a leaf's natural width; a field's
// includes its gutter.
func controlWidth(e *view.Element) int {
	switch e.Kind {
	case view.Button:
		return width(buttonFace(e, style{}))
	case view.TextField, view.DateTime:
		n := prompt + fieldWidth
		if isLongText(e) {
			n = fieldWidth
		}
		return gutter + max(Width(e.Label), n)
	case view.CheckBox:
		return gutter + width(boxFace(false, e.Label, style{}))
	case view.Switch:
		return gutter + width(switchFace(e, false))
	case view.Choice:
		if isSelect(e) {
			return gutter + max(Width(e.Label), width(selectValue(e, noWrap)))
		}
		if e.Variant == "chips" {
			n := 0
			for i, ow := range optionWidths(e) {
				if i > 0 {
					n += optionGap
				}
				n += ow
			}
			return gutter + max(Width(e.Label), n)
		}
		n := 0
		for _, ow := range optionWidths(e) {
			n = max(n, ow)
		}
		return gutter + max(Width(e.Label), n)
	case view.Option:
		return Width(e.Label) + 4
	case view.Slider, view.RangeSlider:
		n := trackWidth + 1 + sliderValueWidth(e)
		if e.Label != "" {
			n += Width(e.Label) + 1
		}
		return gutter + n
	case view.Progress:
		return max(Width(e.Label), progressWidth+percentWidth)
	case view.Spinner:
		n := spinnerWidth(e)
		if e.Label != "" {
			n += 1 + Width(e.Label)
		}
		return n
	case view.Table:
		return tableWidth(columnWidths(e))
	case view.RichList:
		return listWidth(e)
	case view.Tree:
		return treeWidth(e)
	case view.Chart:
		return chartWidth(e)
	case view.Sparkline:
		return sparkWidth(e)
	case view.Image:
		return Width(imageText(e))
	case view.Icon:
		return Width(icons.Glyph(e.Name))
	case view.Media:
		return Width("▶ " + e.Alt)
	case view.Placeholder:
		return Width(placeholderText(e))
	case view.Tab:
		return Width(e.Label)
	}
	return 0
}

// controlHeight is a control's or a leaf's height at width w, a field's
// gutter included in w.
func (l *layout) controlHeight(e *view.Element, w int) int {
	if isField(e) {
		w = max(w-gutter, 1)
	}
	h := 1
	switch e.Kind {
	case view.TextField, view.DateTime:
		h = fieldRows(e)
		if e.Label != "" {
			h++
		}
	case view.Progress:
		if e.Label != "" {
			h++
		}
	case view.Table:
		h = 2 + bodyRows(e)
	case view.RichList:
		h = listHeight(e)
	case view.Tree:
		h = treeHeight(e)
	case view.Chart:
		h = chartHeight(e)
	case view.Sparkline:
		h = max(e.Height, 1)
	case view.KeyHints:
		h = len(l.r.keyHints(e, w))
	case view.Choice:
		switch {
		case isSelect(e):
			if l.r.listOpen(e) {
				h += len(e.Options)
			}
		case e.Variant == "chips":
			h = 0
			if rows, _ := flow(optionWidths(e), optionGap, w); len(rows) > 0 {
				h = rows[len(rows)-1] + 1
			}
		default:
			h = len(e.Children)
		}
		if e.Label != "" {
			h++
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

// buttonFace is a Button as drawn: "[ label ]", the label alone when
// borderless, and " label " as a List's item, whatever its variant.
func buttonFace(e *view.Element, st style) []glyph {
	label := line(buttonLabel(e), st)
	if e.Item {
		return concat(glyphs(" ", st), label, glyphs(" ", st))
	}
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
			parts = append(parts, icons.Glyph(e.Name))
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

// boxFace is "[•] label" or "[ ] label", as huh marks a choice; the box
// alone without a label.
func boxFace(on bool, label string, st style) []glyph {
	box := "[ ]"
	if on {
		box = "[•]"
	}
	gs := glyphs(box, st)
	if label != "" {
		gs = concat(gs, line(" "+label, style{}))
	}
	return gs
}

// switchFace is a HottySwitch as drawn past its gutter: a track with its
// knob at the end, "▬▬■" in the accent when it is on, a bar into a filled
// knob, and "□⎯⎯" in muted when it is off, a hollow knob on a thin line,
// so that the knob's end, its fill and the track's weight say which
// without colour; then a space and its label, in the accent while it has
// the keyboard. A disabled one is muted and faint throughout.
func switchFace(e *view.Element, focused bool) []glyph {
	track, label := style{role: Muted}, style{}
	switch {
	case e.Disabled:
		track = style{role: Muted, attr: Faint}
		label = track
	case focused:
		label = style{role: Accent}
	}
	// The thin line is ⎯, not ─: a terminal draws ─ itself, in the
	// middle of the cell, where the font's square need not be. ⎯ is
	// drawn by a font, a symbol font where the text's lacks it (Meslo
	// does), and met the square in the maintainer's terminal.
	face := "□⎯⎯"
	if e.On() {
		face = "▬▬■"
		if !e.Disabled {
			track.role = Accent
		}
	}
	gs := glyphs(face, track)
	if e.Label != "" {
		gs = concat(gs, line(" "+e.Label, label))
	}
	return gs
}

// optionFace is one of a Choice's options as drawn: a chip, "( label )"
// or "(● label)", reversed in the accent when it has the keyboard; or a
// row, "> [•] label", whose "> " marks the one with the keyboard.
func optionFace(choice, o *view.Element, focused bool) []glyph {
	if choice.Variant != "chips" {
		mark, box := glyphs("  ", style{}), style{}
		if focused {
			mark, box = glyphs("> ", style{role: Accent}), style{role: Accent}
		}
		return concat(mark, boxFace(o.Active, o.Label, box))
	}
	st := style{}
	if focused {
		st = style{role: Accent, attr: Reverse}
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

// selectValue is a select's value row: the picked option's label, or "…"
// in muted, then " ▾" in muted, the label cut to fit w. At noWrap it is as
// wide as the widest option makes it.
func selectValue(e *view.Element, w int) []glyph {
	value := line(pickedLabel(e), style{})
	if len(value) == 0 {
		value = line("…", style{role: Muted})
	}
	if w == noWrap {
		for _, o := range e.Options {
			if Width(o.Label) > width(value) {
				value = line(o.Label, style{})
			}
		}
	}
	arrow := glyphs(" ▾", style{role: Muted})
	if room := w - width(arrow); room >= 1 {
		return concat(fit(value, room), arrow)
	}
	return fit(concat(value, arrow), w)
}

// titleStyle is a field's title: bold, in the accent while the field has
// the keyboard.
func titleStyle(focused bool) style {
	if focused {
		return style{role: Accent, attr: Bold}
	}
	return style{attr: Bold}
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

// sliderValueWidth is the columns a Slider's value takes at most, or a
// HottyRangeSlider's two ("20–70").
func sliderValueWidth(e *view.Element) int {
	if e.Kind == view.RangeSlider {
		return e.RangeWidth()
	}
	return e.SliderWidth()
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
