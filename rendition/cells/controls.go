package cells

import (
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/icons"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// fieldWidth is a longText's natural width when its label is shorter: an
// HTML textarea's default size.
const fieldWidth = 20

// textInput is a one-line text field's input: the underlined columns its
// value shows in, fixed as a GUI form's input is, whatever is typed, and
// wide enough for an email address.
const textInput = 32

// trackWidth is a Slider's natural track width.
const trackWidth = 10

// gutter is the columns left of a field (profile §3.4): a bar, "┃ ",
// while it has the keyboard, as huh draws its fields.
const gutter = 2

// inset is the blank column before a one-line field's input, between it
// and its label or gutter. The underline starts after it, at the value's
// first column (the maintainer, round 2: "keeping underline only on the
// input area").
const inset = 1

// inlineRow reports whether a field's label is on its value's row: a
// one-line text field's, a DateTime's or a select's, the labels of a run
// of them in a Column padded to the widest (labelRuns). A longText's and
// a Choice's options' labels are on a title row above.
func inlineRow(e *view.Element) bool {
	if e == nil || e.Label == "" {
		return false
	}
	return isOneLine(e)
}

// isOneLine reports whether a field has an input on one row: a one-line
// text field, a DateTime or a select.
func isOneLine(e *view.Element) bool {
	return isTextControl(e) && !isLongText(e) || isSelect(e)
}

// inputWidth is the columns of a one-line field's input (profile §3.5),
// at most the room it is given: a text field's textInput; a DateTime's,
// its value's form or its value, whichever is wider, and a column for the
// caret after it; a select's, its widest option, a space and "▾".
func inputWidth(e *view.Element) int {
	switch {
	case isSelect(e):
		return width(selectValue(e, noWrap))
	case e.Kind == view.DateTime:
		v, _ := e.Value.(string)
		return max(Width(e.DateHint()), Width(v)) + 1
	}
	return textInput
}

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

// hasTitle reports whether an element has a title row of its own: a
// longText, a Choice's options or a Progress with a label, a Table, its
// header, and a HottyList, its status line. A one-line field's, a
// CheckBox's, a HottySwitch's, a Slider's and a Spinner's labels are on
// their one row.
func hasTitle(e *view.Element) bool {
	switch e.Kind {
	case view.Table, view.RichList:
		return true
	case view.TextField, view.DateTime, view.Choice, view.Progress:
		return e.Label != "" && !inlineRow(e)
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
		if isLongText(e) {
			return gutter + max(Width(e.Label), fieldWidth)
		}
		return gutter + labelled(e, inset+inputWidth(e))
	case view.CheckBox:
		return gutter + width(boxFace(false, e.Label, style{}))
	case view.Switch:
		return gutter + width(switchFace(e, false))
	case view.Choice:
		if isSelect(e) {
			return gutter + labelled(e, inset+inputWidth(e))
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
		return tableNatural(e)
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
		if e.Label != "" && !inlineRow(e) {
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
		if isSelect(e) {
			if l.r.listOpen(e) {
				h += len(e.Options)
			}
		} else {
			h = len(e.Children)
		}
		if e.Label != "" && !inlineRow(e) {
			h++
		}
	}
	return h + len(errorLines(e, w-l.errorIndent(e, w)))
}

// errorLines are a control's error, "✗ message", wrapped.
func errorLines(e *view.Element, w int) [][]glyph {
	if e.Error == "" {
		return nil
	}
	return wrap(line("✗ "+e.Error, style{role: Error}), w, nil, repeat(" ", 2, style{}))
}

// errorIndent is the columns before a field's error in its box of w past
// the gutter: a one-line field's input's, its label, a space and the
// inset, so that the error starts under the value it is about (the
// maintainer, round 3); none for other controls, or where that would
// leave the error fewer than minErrorWidth columns.
func (l *layout) errorIndent(e *view.Element, w int) int {
	if !isOneLine(e) {
		return 0
	}
	n := inset
	if inlineRow(e) {
		n += min(l.labelWidth(e), max(w-4, 0)) + 1
	}
	if w-n < minErrorWidth {
		return 0
	}
	return n
}

// minErrorWidth is the fewest columns an error is indented into.
const minErrorWidth = 16

// fieldRows is how many rows a field's value takes: one, or a longText's
// lines, at least 3 and at most 8.
func fieldRows(e *view.Element) int {
	if !isLongText(e) {
		return 1
	}
	v, _ := e.Value.(string)
	return min(max(len(splitClusters(clusters(v))), 3), 8)
}

// labelled is a one-line field's width past its gutter for a value n
// wide: its label and a space before the value, on one row.
func labelled(e *view.Element, n int) int {
	if inlineRow(e) {
		return Width(e.Label) + 1 + n
	}
	return n
}

// buttonFace is a Button as drawn: " label ", a column of padding either
// side on its fill (paint's buttonFill), and so a List's item, unfilled,
// whatever its variant; the label alone when borderless.
func buttonFace(e *view.Element, st style) []glyph {
	label := line(buttonLabel(e), st)
	if e.Variant == "borderless" && !e.Item {
		return label
	}
	return concat(glyphs(" ", st), label, glyphs(" ", st))
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

// optionFace is one of a Choice's options as drawn, a row each, as huh's
// multiselect has them: "> [•] label", whose "> " marks the one with the
// keyboard. Chips that pick one are radio buttons, "> (•) label": pills
// wrapped across the field read as neither a form nor a list.
func optionFace(choice, o *view.Element, focused bool) []glyph {
	mark, box := glyphs("  ", style{}), style{}
	if focused {
		mark, box = glyphs("> ", style{role: Accent}), style{role: Accent}
	}
	if choice.Variant == "chips" && !choice.Multiple {
		return concat(mark, radioFace(o.Active, o.Label, box))
	}
	return concat(mark, boxFace(o.Active, o.Label, box))
}

// radioFace is "(•) label" or "( ) label": one of options that pick one.
func radioFace(on bool, label string, st style) []glyph {
	dot := "( )"
	if on {
		dot = "(•)"
	}
	return concat(glyphs(dot, st), line(" "+label, style{}))
}

func optionWidths(e *view.Element) []int {
	ws := make([]int, len(e.Children))
	for i, o := range e.Children {
		ws[i] = width(optionFace(e, o, false))
	}
	return ws
}

// selectValue is a select's input, w wide, as a one-line text field's is:
// the picked option's label, or "…" in muted, cut to fit, and "▾" in muted
// at the input's end, a space after the label at least. At noWrap it is
// as wide as the widest option makes it, so the chevron sits a space after
// the widest.
func selectValue(e *view.Element, w int) []glyph {
	value := line(pickedLabel(e), style{})
	if len(value) == 0 {
		value = line("…", style{role: Muted})
	}
	if w == noWrap {
		n := width(value)
		for _, o := range e.Options {
			n = max(n, Width(o.Label))
		}
		w = n + 2
	}
	arrow := glyphs("▾", style{role: Muted})
	room := w - 2
	if room < 1 {
		return fit(concat(value, arrow), w)
	}
	value = fit(value, room)
	return concat(value, repeat(" ", w-width(value)-1, style{}), arrow)
}

// titleStyle is a field's label: muted, in the accent while the field has
// the keyboard, as a GUI form's label over its input.
func titleStyle(focused bool) style {
	if focused {
		return style{role: Accent}
	}
	return style{role: Muted}
}

// fieldLineMix is how far a field's underline is from the background
// toward the accent: the accent, toned down, so that it is there without
// drawing the eye from the text (the maintainer, round 3).
const fieldLineMix = 128 // 50%

// underline rules a one-line field's input, x to x+w, as a GUI form's
// input is outlined: the line in the accent toned toward the background,
// with the keyboard or without it, the gutter and the label saying which
// field has it (SGR 58; a terminal without it draws the line in the text's
// colour).
func underline(cv *canvas, x, y, w int) {
	cv.restyle(x, y, w, 1, func(c *Cell) {
		c.Attr, c.Line, c.LineSet, c.LineMix = c.Attr|Underline, Accent, true, fieldLineMix
	})
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
