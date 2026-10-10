package cells

import (
	"math"

	"github.com/neuroplastio/hotty-go/blocks"
	"github.com/neuroplastio/hotty-go/braille"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyChart in cells (profile §3.4), as ntcharts draws one for Bubble
// Tea: a line in braille dots (hotty-go's braille), two across and four
// down a cell, or bars in eighths of a row (hotty-go's blocks); the
// ticks' labels up its left side, an axis of box lines, the points'
// labels under it, and a legend. A HottySparkline is the bars alone, a
// column a value.

// seriesRoles are the series' colours, in order, round again past the
// fourth: not accent, which marks only focus (profile §3.6).
var seriesRoles = [...]Role{Info, Warning, Success, Error}

func seriesRole(i int) Role { return seriesRoles[i%len(seriesRoles)] }

// chartPlotWidth is a line chart's plot's natural width, and minPlot any
// chart's narrowest.
const (
	chartPlotWidth = 40
	minPlot        = 8
)

// noData is what an empty chart's plot says.
const noData = "No data"

// chartYWidth is the columns of a chart's tick labels: its widest, so that
// the plot stays put whichever of them show.
func chartYWidth(e *view.Element) int {
	n := 1
	for _, t := range e.Ticks() {
		n = max(n, Width(t.Text))
	}
	return n
}

// chartWidth is a chart's natural width: its labels, the axis and a space,
// and its plot: a line's chartPlotWidth; for bars, a group a point, as
// wide as its widest label or two columns a series, a column apart; and
// its legend, if that is wider.
func chartWidth(e *view.Element) int {
	plot := chartPlotWidth
	if e.Variant == view.ChartBar {
		n, gw := e.Points(), 2*max(len(e.Series), 1)
		for _, l := range e.Labels {
			gw = max(gw, Width(l))
		}
		plot = max(n*(gw+1)-1, minPlot, Width(noData))
	}
	return chartYWidth(e) + 2 + max(plot, width(legendGlyphs(e)))
}

// chartMinimum is the narrowest a chart gets: its labels, the axis, and a
// plot of minPlot.
func chartMinimum(e *view.Element) int { return chartYWidth(e) + 2 + minPlot }

// chartHeight is a chart's rows: its plot's, the axis, and a row for the
// points' labels and one for the legend, when it has them.
func chartHeight(e *view.Element) int {
	h := e.Height + 1
	if len(e.Labels) > 0 {
		h++
	}
	if e.Legend() {
		h++
	}
	return h
}

// paintChart paints a chart (profile §3.4): its ticks' labels in muted,
// right-aligned, each on the row its value is drawn in, and `┤` there on
// the axis, `│` elsewhere, and `└───` under the plot, in border; the plot;
// the points' labels; the legend.
func (l *layout) paintChart(cv *canvas, e *view.Element, x, y, w int) {
	yl := chartYWidth(e)
	px, pw, ph := x+yl+2, max(w-yl-2, 1), max(e.Height, 1)
	labels := chartTickRows(e, ph)
	axis, muted := style{role: Border}, style{role: Muted}
	for r := range ph {
		mark := "│"
		if t, ok := labels[r]; ok {
			cv.write(x+yl-Width(t), y+r, Width(t), line(t, muted))
			mark = "┤"
		}
		cv.set(x+yl+1, y+r, glyph{text: mark, width: 1, style: axis})
	}
	cv.write(x+yl+1, y+ph, pw+1, concat(glyphs("└", axis), repeat("─", pw, axis)))
	row := y + ph + 1
	switch {
	case e.Empty():
		msg := fit(line(noData, muted), pw)
		cv.write(px+(pw-width(msg))/2, y+ph/2, pw, msg)
		if e.Variant == view.ChartBar {
			row = paintCategories(cv, e, px, row, pw)
		} else {
			row = paintXLabels(cv, e, px, row, pw)
		}
	case e.Variant == view.ChartBar:
		paintBars(cv, e, px, y, pw, ph)
		row = paintCategories(cv, e, px, row, pw)
	default:
		paintLines(cv, e, px, y, pw, ph)
		row = paintXLabels(cv, e, px, row, pw)
	}
	if e.Legend() {
		cv.write(px, row, pw, fit(legendGlyphs(e), pw))
	}
}

// chartTickRows are the plot's rows that carry a tick's label, by row
// (view.TickRow): every tick's, as the axis is made to have them a row
// apart (view.ChartAxis). Where the plot is too short for that, as plothot
// labels its axis: 0's, the top tick's and the bottom one's first, then
// each other's whose row has a free row either side, since labels on
// neighbouring rows read as one.
func chartTickRows(e *view.Element, ph int) map[int]string {
	ticks := e.Ticks()
	if len(ticks) == 0 {
		return nil
	}
	rows := make([]int, len(ticks))
	var order []int
	for i, t := range ticks {
		rows[i] = view.TickRow(e.Variant, t.Frac, ph)
		if t.Value == 0 {
			order = append(order, i)
		}
	}
	firsts := len(order) + 2
	order = append(order, len(ticks)-1, 0)
	for i := len(ticks) - 2; i >= 1; i-- {
		order = append(order, i)
	}
	chosen := map[int]int{} // row: tick
	free := func(r int) bool {
		for have := range chosen {
			if r >= have-1 && r <= have+1 {
				return false
			}
		}
		return true
	}
	for k, i := range order {
		if _, taken := chosen[rows[i]]; taken || k >= firsts && !free(rows[i]) {
			continue
		}
		chosen[rows[i]] = i
	}
	out := map[int]string{}
	for r, i := range chosen {
		out[r] = ticks[i].Text
	}
	return out
}

func clamp01(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return math.Max(0, math.Min(1, f))
}

// paintLines draws each series as a line of braille dots across the plot,
// in its colour, each point at its slot (view.Element.Slots): a missing
// value breaks the line, and a lone value between two is a dot. A chart of
// one value and no window is a level line across, as on a host (hotty-go's
// chart.Line).
func paintLines(cv *canvas, e *view.Element, x, y, pw, ph int) {
	c := braille.New(pw, ph)
	dw, dh := c.Size()
	a := e.YAxis()
	slots, off := e.Slots(), e.Slots()-e.Points()
	xOf := func(j int) int {
		if slots <= 1 {
			return 0
		}
		return int(math.Round(float64(off+j) / float64(slots-1) * float64(dw-1)))
	}
	for i, s := range e.Series {
		ink := uint8(i%len(seriesRoles) + 1)
		px, py := -1, -1
		for j, v := range s.Values {
			if math.IsNaN(v) {
				px = -1
				continue
			}
			cx, cy := xOf(j), view.DotY(a.Frac(v), dh)
			if px < 0 {
				c.Set(cx, cy, ink)
			} else {
				c.Line(px, py, cx, cy, ink)
			}
			px, py = cx, cy
		}
		if slots == 1 && e.Window == 0 && px >= 0 {
			c.Line(0, py, dw-1, py, ink)
		}
	}
	for r := range ph {
		for col := range pw {
			if ch, ink := c.Cell(col, r); ch != ' ' {
				cv.set(x+col, y+r, glyph{text: string(ch), width: 1, style: style{role: seriesRole(int(ink) - 1)}})
			}
		}
	}
}

// barGroups is how a bar chart's points share its plot's pw columns: a
// group a point, from the first that fits, gw columns each and one apart,
// each series' bar bw of them, from lead in. When a column a bar does not
// fit them all, the last points that fit show.
func barGroups(e *view.Element, pw int) (first, gw, bw, lead int) {
	n, s := e.Points(), max(len(e.Series), 1)
	if n == 0 {
		return 0, 0, 0, 0
	}
	gw = (pw - (n - 1)) / n
	if gw < s {
		first, gw = max(n-max((pw+1)/(s+1), 1), 0), s
	}
	bw = gw / s
	return first, gw, bw, (gw - bw*s) / 2
}

// paintBars draws a bar for each series' value at each point, from the
// axis's 0 to the value, in the series' colour (barCell).
func paintBars(cv *canvas, e *view.Element, x, y, pw, ph int) {
	first, gw, bw, lead := barGroups(e, pw)
	a := e.YAxis()
	z := clamp01(a.Frac(0)) * float64(ph)
	for k := first; k < e.Points(); k++ {
		gx := x + (k-first)*(gw+1) + lead
		for i := range e.Series {
			v := e.At(i, k)
			if math.IsNaN(v) {
				continue
			}
			end := clamp01(a.Frac(v)) * float64(ph)
			st := style{role: seriesRole(i)}
			for r := range ph {
				if g := barCell(z, end, ph-1-r); g != " " {
					for c := range bw {
						if bx := gx + i*bw + c; bx < x+pw {
							cv.set(bx, y+r, glyph{text: g, width: 1, style: st})
						}
					}
				}
			}
		}
	}
}

// barEighths are a row filled from its bottom, by eighths.
var barEighths = []rune(" ▁▂▃▄▅▆▇█")

// barCell is the glyph, in the row up rows from the plot's bottom, of a
// bar from the axis's 0, z rows up, to its value, end rows up: the part
// of the row the bar covers. A part from the row's bottom is in eighths,
// as hotty-go's blocks draws a bar rising from a row's edge; a part from
// its top, of a bar that hangs below 0, or of one that rises from a 0
// inside the row, is in halves (`▀`, and `▔` for less), since a cell has
// no other blocks that hang. The row with the bar's 0 end shows it however
// short, so that a value is never taken for none.
func barCell(z, end float64, up int) string {
	lo, hi := math.Min(z, end), math.Max(z, end)
	bot, top := float64(up), float64(up+1)
	a, b := math.Max(lo, bot), math.Min(hi, top)
	if b <= a {
		return " "
	}
	f := b - a
	zero := z >= bot && z <= top && lo < hi
	fromBottom := a == bot || b < top && a-bot < top-b
	switch {
	case a == bot && b == top:
		return "█"
	case fromBottom:
		e := int(math.Round((b - bot) * 8))
		if e == 0 && zero {
			e = 1
		}
		return string(barEighths[min(e, 8)])
	case f >= 0.75:
		return "█"
	case f >= 0.25:
		return "▀"
	case f >= 1.0/16 || zero:
		return "▔"
	}
	return " "
}

// paintCategories writes a bar chart's points' labels under its axis, each
// centred under its group, cut to it with "…", in muted; it returns the
// next row.
func paintCategories(cv *canvas, e *view.Element, x, y, pw int) int {
	if len(e.Labels) == 0 {
		return y
	}
	first, gw, _, _ := barGroups(e, pw)
	for k := first; k < e.Points(); k++ {
		t := fit(line(e.PointLabel(k), style{role: Muted}), gw)
		gx := x + (k-first)*(gw+1)
		if gx >= x+pw {
			break
		}
		cv.write(gx+(gw-width(t))/2, y, min(gw, x+pw-gx), t)
	}
	return y + 1
}

// paintXLabels writes a line chart's labels under its axis: three of its
// points', or five where the plot is 60 columns or wider, evenly spread
// (view.Element.LabelPicks): the first starting at its slot, the last
// ending at its own, the others centred on theirs, and one that would
// touch the one before it left out; it returns the next row.
func paintXLabels(cv *canvas, e *view.Element, x, y, pw int) int {
	if len(e.Labels) == 0 {
		return y
	}
	n := 3
	if pw >= 60 {
		n = 5
	}
	picks := e.LabelPicks(n)
	end := -2 // the last column written
	for k, j := range picks {
		if j < 0 {
			continue
		}
		t := line(e.PointLabel(j), style{role: Muted})
		tw := width(t)
		at := 0
		if len(picks) > 1 {
			at = int(math.Round(float64(k) / float64(len(picks)-1) * float64(pw-1)))
		}
		start := at - tw/2
		switch {
		case k == 0:
			start = at
		case k == len(picks)-1:
			start = at - tw + 1
		}
		start = max(0, min(pw-tw, start))
		if start <= end+1 {
			continue
		}
		cv.write(x+start, y, pw-start, t)
		end = start + tw - 1
	}
	return y + 1
}

// legendGlyphs are a chart's legend: each series' key in its colour (`━`
// for a line, `■` for bars) and its label in muted, three columns apart.
func legendGlyphs(e *view.Element) []glyph {
	if !e.Legend() {
		return nil
	}
	key := "━"
	if e.Variant == view.ChartBar {
		key = "■"
	}
	var out []glyph
	for i, s := range e.Series {
		if i > 0 {
			out = append(out, repeat(" ", 3, style{})...)
		}
		out = append(out, glyphs(key, style{role: seriesRole(i)})...)
		if s.Label != "" {
			out = append(out, glyphs(" "+s.Label, style{role: Muted})...)
		}
	}
	return out
}

// sparkWidth is a sparkline's columns: its window, else a column a value.
func sparkWidth(e *view.Element) int {
	if e.Window > 0 {
		return e.Window
	}
	if len(e.Series) == 0 {
		return 0
	}
	return len(e.Series[0].Values)
}

// paintSpark paints a sparkline as ntcharts' does: a column a value, in
// info, filled from its bottom in eighths of a row (hotty-go's blocks),
// from Min at its bottom to Max at its top; the newest value in its last
// column, those before it to its left, as many as fit. A missing value is
// blank, and only it: a value at Min is an eighth (▁), so that the lowest
// is not taken for a gap.
func (l *layout) paintSpark(cv *canvas, e *view.Element, x, y, w int) {
	if len(e.Series) == 0 {
		return
	}
	vals := e.Series[0].Values
	cols := min(sparkWidth(e), w)
	n := min(len(vals), cols)
	vals = vals[len(vals)-n:]
	h := max(e.Height, 1)
	st := style{role: Info}
	for k, v := range vals {
		if math.IsNaN(v) {
			continue
		}
		f := clamp01((v - e.Min) / (e.Max - e.Min))
		for r, g := range blocks.Column(math.Max(f*float64(h), 1.0/8), h) {
			if g != ' ' {
				cv.set(x+cols-n+k, y+r, glyph{text: string(g), width: 1, style: st})
			}
		}
	}
}
