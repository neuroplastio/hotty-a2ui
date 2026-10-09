package cells

import (
	"fmt"
	"math"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A hit is where a click lands on something: an element's cells, as last
// drawn.
type hit struct {
	x, y, w, h int
	id         string
	// opt is the option of an open select's row, a Table's row, or a
	// HottyList's item; -1 otherwise.
	opt int
	// field is a text control's value area; track a Slider's.
	field *fieldArea
	track *trackArea
	// disabled: a disabled Button, activated but not focused.
	disabled bool
}

type fieldArea struct {
	x, y       int
	hoff, voff int
	rows       int
}

type trackArea struct{ x, n int }

func (h hit) contains(col, row int) bool {
	return col >= h.x && col < h.x+h.w && row >= h.y && row < h.y+h.h
}

// paint paints an element in its box, at (x, y), w wide and h tall.
func (l *layout) paint(cv *canvas, e *view.Element, x, y, w, h int) {
	if w <= 0 || h <= 0 || e == nil || e.A11y.Hidden {
		return
	}
	l.r.boxes[e.ID] = box{x, y, w, h}
	switch e.Kind {
	case view.Stack:
		kids := shown(e.Children)
		if e.Dir != view.Horizontal {
			l.paintColumn(cv, kids, e.Justify, e.Align, x, y, w, h)
			return
		}
		xs, ws := l.row(kids, e.Justify, w)
		for i, k := range kids {
			kh := l.height(k, ws[i])
			dy, bh := 0, kh
			switch e.Align {
			case "center":
				dy = (h - kh) / 2
			case "end":
				dy = h - kh
			case "stretch", "":
				bh = h
			}
			l.paint(cv, k, x+xs[i], y+dy, ws[i], bh)
		}
	case view.Card:
		cv.box(x, y, w, h)
		l.paintColumn(cv, shown(e.Children), "start", "stretch", x+2, y+1, w-4, h-2)
	case view.Form:
		l.paintColumn(cv, shown(e.Children), "start", "stretch", x, y, w, h)
	case view.Modal:
		l.paintColumn(cv, shown(e.Children), "start", "stretch", x, y, w, h)
		if e.Clickable {
			if l.r.focused(e.ID) {
				cv.restyle(x, y, w, h, func(c *Cell) { c.Role, c.Attr = Accent, c.Attr|Reverse })
			}
			l.r.hits = append(l.r.hits, hit{x: x, y: y, w: w, h: h, id: e.ID, opt: -1})
		}
	case view.Tabs:
		bar, content := tabParts(e)
		n := l.paintTabBar(cv, bar, x, y, w)
		l.paintColumn(cv, content, "start", "stretch", x, y+n, w, h-n)
	case view.Text:
		for i, t := range l.lines(e, w) {
			if i >= h {
				break
			}
			if t.rule {
				cv.write(x, y+i, w, repeat("─", w, style{role: Border}))
				continue
			}
			cv.write(x, y+i, w, t.gs)
		}
	case view.Divider:
		b := style{role: Border}
		if e.Dir == view.Vertical {
			for j := y; j < y+h; j++ {
				cv.write(x, j, 1, glyphs("│", b))
			}
			return
		}
		cv.write(x, y, w, repeat("─", w, b))
	case view.Image:
		cv.write(x, y, w, fit(line(imageText(e), style{role: Muted}), w))
	case view.Icon:
		cv.write(x, y, w, glyphs(view.IconGlyph(e.Name), style{}))
	case view.Media:
		st := style{attr: Underline, link: e.URL}
		if l.r.focused(e.ID) {
			st.role, st.attr = Accent, st.attr|Reverse
		}
		n := cv.write(x, y, w, fit(line("▶ "+e.Alt, st), w))
		l.r.boxes[e.ID] = box{x, y, n, 1}
		if e.URL != "" {
			l.r.hits = append(l.r.hits, hit{x: x, y: y, w: n, h: 1, id: e.ID, opt: -1})
		}
	case view.Placeholder:
		st := style{role: Warning}
		if e.State == a2ui.Pending {
			st.role = Muted
		}
		cv.write(x, y, w, fit(line(placeholderText(e), st), w))
	case view.Button:
		st := style{}
		switch {
		case e.Disabled:
			st = style{role: Muted, attr: Faint}
		case l.r.focused(e.ID):
			st = style{role: Accent, attr: Reverse}
		}
		switch {
		case e.Variant == "borderless" && !e.Item:
			st.attr |= Underline
		case e.Variant == "primary", e.Item && e.Variant != "borderless":
			st.attr |= Bold
		}
		face := buttonFace(e, st)
		if width(face) > w {
			face = fitButton(e, st, w)
		}
		if e.Item {
			// A row: its style (focus, say) goes across the List.
			face = concat(face, glyphs(strings.Repeat(" ", max(w-width(face), 0)), st))
		}
		n := cv.write(x, y, w, face)
		l.r.boxes[e.ID] = box{x, y, n, 1}
		l.r.hits = append(l.r.hits, hit{x: x, y: y, w: n, h: 1, id: e.ID, opt: -1, disabled: e.Disabled})
	case view.TextField, view.DateTime, view.CheckBox, view.Choice, view.Slider:
		if w <= gutter {
			return
		}
		fx, fw := x+gutter, w-gutter
		switch e.Kind {
		case view.TextField, view.DateTime:
			l.paintField(cv, e, fx, y, fw)
		case view.CheckBox:
			on, _ := e.Value.(bool)
			n := l.paintBox(cv, e.ID, on, e.Label, fx, y, fw)
			l.r.boxes[e.ID] = box{fx, y, n, 1}
			l.r.hits = append(l.r.hits, hit{x: x, y: y, w: gutter + n, h: 1, id: e.ID, opt: -1})
			l.paintError(cv, e, fx, y+1, fw)
		case view.Choice:
			l.paintChoice(cv, e, fx, y, fw)
		case view.Slider:
			l.paintSlider(cv, e, fx, y, fw)
		}
		if l.hasKeyboard(e) {
			bar := style{role: Accent}
			for j := y; j < y+min(h, l.controlHeight(e, w)); j++ {
				cv.set(x, j, glyph{text: "┃", width: 1, style: bar})
			}
		}
	case view.Progress:
		l.paintProgress(cv, e, x, y, w)
	case view.Spinner:
		l.paintSpinner(cv, e, x, y, w)
	case view.Table:
		l.paintTable(cv, e, x, y, w)
	case view.RichList:
		l.paintList(cv, e, x, y, w)
	case view.KeyHints:
		l.paintKeyHints(cv, e, x, y, w)
	default:
		l.paintColumn(cv, shown(e.Children), "start", "stretch", x, y, w, h)
	}
}

// paintProgress paints a Progress: its label on a row, then the bar and
// the percentage, " 42%", as bubbles' progress has them. The bar is filled
// in eighths with "█" and a partial block, blending from info into accent
// (success alone once done), the rest "░" in border. An indeterminate bar is a quarter of the bar
// filled, sliding across with the clock, and the percentage's columns are
// blank.
func (l *layout) paintProgress(cv *canvas, e *view.Element, x, y, w int) {
	row := y
	if e.Label != "" {
		cv.write(x, row, w, fit(line(e.Label, style{}), w))
		row++
	}
	f, known := e.Fraction()
	pct := strings.Repeat(" ", percentWidth)
	if known {
		pct = fmt.Sprintf(" %3.0f%%", f*100)
	}
	n := w - Width(pct)
	if n < 3 {
		n, pct = w, ""
	}
	fill, empty := style{role: Info}, style{role: Border}
	var bar []glyph
	if known {
		// Across the bar, info blends into accent, as bubbles' default
		// blend does; a done bar is success alone.
		at := func(i int) style {
			if f >= 1 {
				return style{role: Success}
			}
			s := fill
			if n > 1 {
				s.to, s.mix = Accent, uint8(255*i/(n-1))
			}
			return s
		}
		eighth := int(f * float64(n*8))
		for i := range eighth / 8 {
			bar = append(bar, glyph{text: "█", width: 1, style: at(i)})
		}
		if p := eighth % 8; p > 0 {
			bar = append(bar, glyphs(eighths[p], at(eighth/8))...)
		}
	} else {
		seg := max(n/4, 1)
		at := view.ProgressStep(l.r.Clock())*(n+seg)/view.ProgressSteps - seg
		for i := range n {
			if i >= at && i < at+seg {
				bar = append(bar, glyph{text: "█", width: 1, style: fill})
			} else {
				bar = append(bar, glyph{text: "░", width: 1, style: empty})
			}
		}
		l.r.animate(view.ProgressInterval)
	}
	bar = append(bar, repeat("░", n-width(bar), empty)...)
	cv.write(x, row, n, bar)
	cv.write(x+n, row, w-n, line(pct, style{}))
}

// paintSpinner paints a Spinner: the frame the clock is at, in info, then
// a space and its label. One that does not spin keeps its frame's columns
// blank, so that its label stays where it was.
func (l *layout) paintSpinner(cv *canvas, e *view.Element, x, y, w int) {
	set := e.SpinnerFrames()
	if e.Active && len(set.Frames) > 0 {
		cv.write(x, y, w, line(set.Frame(l.r.Clock()), style{role: Info}))
		l.r.animate(set.Interval)
	}
	if e.Label != "" {
		col := spinnerWidth(e) + 1
		cv.write(x+col, y, w-col, fit(line(e.Label, style{}), w-col))
	}
}

// paintColumn lays children out down a box: each as tall as it is, spare
// rows to the weighted ones or as justify says; across, as align says.
func (l *layout) paintColumn(cv *canvas, kids []*view.Element, justify, align string, x, y, w, h int) {
	if w <= 0 {
		return
	}
	n := len(kids)
	ws, hs, weights := make([]int, n), make([]int, n), make([]float64, n)
	sum := 0
	for i, k := range kids {
		ws[i] = l.columnWidth(k, align, w)
		hs[i] = l.height(k, ws[i])
		weights[i] = k.Weight
		sum += hs[i] + separator(kids, i)
	}
	gaps := l.spread(hs, weights, justify, h-sum)
	yy := y
	for i, k := range kids {
		yy += gaps[i] + separator(kids, i)
		dx := 0
		switch align {
		case "center":
			dx = (w - ws[i]) / 2
		case "end":
			dx = w - ws[i]
		}
		l.paint(cv, k, x+dx, yy, ws[i], hs[i])
		yy += hs[i]
	}
}

// fitButton is a Button's face cut to w: the label loses columns first.
func fitButton(e *view.Element, st style, w int) []glyph {
	label := line(buttonLabel(e), st)
	if e.Variant == "borderless" || e.Item || w < 5 {
		return fit(buttonFace(e, st), w)
	}
	return concat(glyphs("[ ", st), fit(label, w-4), glyphs(" ]", st))
}

// paintTabBar paints a Tabs' titles, two columns apart, wrapping, and a
// rule under them; it returns the rows they take.
func (l *layout) paintTabBar(cv *canvas, bar []*view.Element, x, y, w int) int {
	if len(bar) == 0 {
		return 0
	}
	ws := make([]int, len(bar))
	for i, t := range bar {
		ws[i] = Width(t.Label)
	}
	rows, xs := flow(ws, 2, w)
	for i, t := range bar {
		st := style{}
		if t.Active {
			st.attr = Bold | Underline
		}
		if l.r.focused(t.ID) {
			st.role, st.attr = Accent, st.attr|Reverse
		}
		n := cv.write(x+xs[i], y+rows[i], w-xs[i], fit(line(t.Label, st), w-xs[i]))
		l.r.boxes[t.ID] = box{x + xs[i], y + rows[i], n, 1}
		l.r.hits = append(l.r.hits, hit{x: x + xs[i], y: y + rows[i], w: n, h: 1, id: t.ID, opt: -1})
	}
	n := rows[len(rows)-1] + 1
	cv.write(x, y+n, w, repeat("─", w, style{role: Border}))
	return n + 1
}

// hasKeyboard reports whether a field has the keyboard: itself, or for
// a Choice's options, one of them.
func (l *layout) hasKeyboard(e *view.Element) bool {
	if l.r.focused(e.ID) {
		return true
	}
	for _, o := range e.Children {
		if o.Kind == view.Option && l.r.focused(o.ID) {
			return true
		}
	}
	return false
}

// paintBox paints "[•] label" or "[ ] label", the box in the accent when
// id has the keyboard (the gutter's bar marks it); it returns the columns
// painted.
func (l *layout) paintBox(cv *canvas, id string, on bool, label string, x, y, w int) int {
	st := style{}
	if l.r.focused(id) {
		st = style{role: Accent}
	}
	return cv.write(x, y, w, fit(boxFace(on, label, st), w))
}

func (l *layout) paintError(cv *canvas, e *view.Element, x, y, w int) {
	for i, gs := range errorLines(e, w) {
		cv.write(x, y+i, w, gs)
	}
}

// paintChoice paints a Choice in its field's box (past the gutter): its
// title, then a select's value row and, while open, its list, "> ● label"
// on the highlighted row; or the options, chips in a flow or a row each.
func (l *layout) paintChoice(cv *canvas, e *view.Element, x, y, w int) {
	row := y
	if e.Label != "" {
		cv.write(x, row, w, fit(line(e.Label, titleStyle(l.hasKeyboard(e))), w))
		row++
	}
	if isSelect(e) {
		cv.write(x, row, w, selectValue(e, w))
		l.r.hits = append(l.r.hits, hit{x: x - gutter, y: y, w: w + gutter, h: row - y + 1, id: e.ID, opt: -1})
		row++
		if l.r.listOpen(e) {
			cur := picked(e)
			for i, o := range e.Options {
				mark, dot, st := glyphs("  ", style{}), "○ ", style{}
				if i == cur {
					dot = "● "
				}
				if i == l.r.hi {
					mark, st = glyphs("> ", style{role: Accent}), style{role: Accent, attr: Bold}
				}
				gs := concat(mark, glyphs(dot, st), line(o.Label, st))
				cv.write(x, row, w, fit(gs, w))
				l.r.hits = append(l.r.hits, hit{x: x - gutter, y: row, w: w + gutter, h: 1, id: e.ID, opt: i})
				row++
			}
		}
		l.paintError(cv, e, x, row, w)
		return
	}
	if e.Variant == "chips" {
		rows, xs := flow(optionWidths(e), optionGap, w)
		for i, o := range e.Children {
			face := optionFace(e, o, l.r.focused(o.ID))
			n := cv.write(x+xs[i], row+rows[i], w-xs[i], fit(face, w-xs[i]))
			l.r.boxes[o.ID] = box{x + xs[i], row + rows[i], n, 1}
			l.r.hits = append(l.r.hits, hit{x: x + xs[i], y: row + rows[i], w: n, h: 1, id: o.ID, opt: -1})
		}
		if len(rows) > 0 {
			row += rows[len(rows)-1] + 1
		}
		l.paintError(cv, e, x, row, w)
		return
	}
	for _, o := range e.Children {
		n := cv.write(x, row, w, fit(optionFace(e, o, l.r.focused(o.ID)), w))
		l.r.boxes[o.ID] = box{x, row, n, 1}
		l.r.hits = append(l.r.hits, hit{x: x - gutter, y: row, w: gutter + n, h: 1, id: o.ID, opt: -1})
		row++
	}
	l.paintError(cv, e, x, row, w)
}

// paintSlider paints "label ━━━━●──── 50": the track fills the room the
// label and the value leave, the knob where the value is.
func (l *layout) paintSlider(cv *canvas, e *view.Element, x, y, w int) {
	v, _ := e.Value.(float64)
	vw := sliderValueWidth(e)
	label := e.Label
	n := w - 1 - vw - Width(label) - 1
	if label == "" || n < 3 {
		label, n = "", w-1-vw
	}
	focused := l.r.focused(e.ID)
	col := x
	if label != "" {
		col += cv.write(col, y, w, line(label+" ", titleStyle(focused)))
	}
	if n >= 1 {
		k := 0
		if e.Max > e.Min && n > 1 {
			k = int(math.Round((v - e.Min) / (e.Max - e.Min) * float64(n-1)))
			k = min(max(k, 0), n-1)
		}
		done, knob, rest := style{}, style{}, style{role: Border}
		if focused {
			done, knob = style{role: Accent}, style{role: Accent, attr: Reverse}
		}
		track := concat(repeat("━", k, done), glyphs("●", knob), repeat("─", n-1-k, rest))
		cv.write(col, y, n, track)
		l.r.hits = append(l.r.hits, hit{x: x - gutter, y: y, w: w + gutter, h: 1, id: e.ID, opt: -1, track: &trackArea{x: col, n: n}})
		col += n + 1
	}
	cv.write(col, y, x+w-col, line(a2ui.NumberString(v), style{}))
	l.paintError(cv, e, x, y+1, w)
}
