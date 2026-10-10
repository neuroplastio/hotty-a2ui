package view

import (
	"math"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-go/series"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// HottyChart and HottySparkline (profile §6.18, §6.19) in the view: series
// of numbers, the points the window keeps, and the scale they are drawn
// on, which every rendition shares, so that cells and a host put a value
// at the same height and label the same ticks.

// Series is one series of a chart: its label, for the legend, and its
// values, oldest first, NaN where one is missing (a null, or what is not a
// number).
type Series struct {
	Label  string    `json:"label,omitempty"`
	Values []float64 `json:"-"`
}

// The kinds of HottyChart (its Element's Variant).
const (
	ChartLine = "line"
	ChartBar  = "bar"
)

// ChartRows is a HottyChart's plot rows without a height; a sparkline's
// is one row.
const ChartRows = 8

// mapChart makes a HottyChart's element: Variant, its kind (line or bar);
// Series, its series, or its values as one; Labels, a label a point;
// Height, its plot's rows; Window, the points its x axis holds (0: as
// many as it has), the last of them kept; Min, Max and Step, its axis
// (ChartAxis).
func mapChart(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Chart, Variant: ChartLine, Height: ChartRows}
	if b.Enum(n, "kind", ChartLine) == ChartBar {
		e.Variant = ChartBar
	}
	if l, ok := b.Raw(n, "series").([]any); ok && len(l) > 0 {
		for _, s := range l {
			m, _ := s.(map[string]any)
			e.Series = append(e.Series, Series{Label: a2ui.ToString(m["label"]), Values: chartValues(m["values"])})
		}
	} else if v := b.Raw(n, "values"); v != nil {
		e.Series = []Series{{Values: chartValues(v)}}
	}
	e.Labels = strs(b.Raw(n, "labels"))
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
	}
	if w := a2ui.ToNumber(b.Raw(n, "window")); w >= 2 {
		e.Window = int(w)
		if drop := e.Points() - e.Window; drop > 0 {
			for i := range e.Series {
				v := e.Series[i].Values
				e.Series[i].Values = v[min(drop, len(v)):]
			}
			e.Labels = e.Labels[min(drop, len(e.Labels)):]
		}
	}
	lo, hi := bound(b, n, "min"), bound(b, n, "max")
	a := ChartAxis(e.Variant, e.Series, lo, hi, e.Height)
	e.Min, e.Max, e.Step = a.Lo, a.Hi, a.Step
	return e
}

// mapSparkline makes a HottySparkline's element: its values as one
// series; Height, its rows (1 without one); Window, the values it shows,
// the last of them; Min and Max, the values at its bottom and its top:
// those given, else 0 (or the lowest value, below 0) and the highest.
func mapSparkline(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Sparkline, Height: 1, Series: []Series{{Values: chartValues(b.Raw(n, "values"))}}}
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
	}
	if w := a2ui.ToNumber(b.Raw(n, "window")); w >= 1 {
		e.Window = int(w)
		if v := e.Series[0].Values; len(v) > e.Window {
			e.Series[0].Values = v[len(v)-e.Window:]
		}
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range e.Series[0].Values {
		if !math.IsNaN(v) {
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
	}
	lo = math.Min(lo, 0)
	if math.IsInf(hi, -1) {
		hi = 1
	}
	if v := bound(b, n, "min"); !math.IsNaN(v) {
		lo = v
	}
	if v := bound(b, n, "max"); !math.IsNaN(v) {
		hi = v
	}
	if !(hi > lo) {
		hi = lo + 1
	}
	e.Min, e.Max = lo, hi
	return e
}

// bound is a chart's min or max: NaN when it has none, or none that is a
// finite number.
func bound(b *Builder, n *a2ui.Node, prop string) float64 {
	v := b.Raw(n, prop)
	if v == nil {
		return math.NaN()
	}
	f := a2ui.ToNumber(v)
	if math.IsInf(f, 0) {
		return math.NaN()
	}
	return f
}

// chartValues are a list's values as numbers: NaN for a null (an index
// past the end pads with them, vault a2ui-limits L4) and for what is not a
// number.
func chartValues(v any) []float64 {
	l, _ := v.([]any)
	out := make([]float64, len(l))
	for i, x := range l {
		out[i] = a2ui.ToNumber(x)
		if math.IsInf(out[i], 0) {
			out[i] = math.NaN()
		}
	}
	return out
}

// ChartAxis is the axis a HottyChart of a kind (line or bar) draws its
// values on, its plot rows tall: its ends where lo and hi (NaN for none)
// fix them, else fitted to the values as hotty-go's series.Scale fits a
// chart's: from 0 when they are all of one sign, with some headroom, its
// ends on round ticks, about max(3, min(6, rows/2)) steps of them. When
// those ticks do not each fall in a row of their own with a free row
// between (TickRow), so that cells can label them all, evenly spread, it
// is the axis Scale makes with the most steps that does, if that spans no
// more than half as much again: a plot squeezed into part of its rows
// would read worse than a tick left unlabelled. In a plot too short for
// any, it is the tightest of them, with the fewest ticks, which cells
// labels in part. A max not above the min is dropped. Without values it
// is 0 to 1.
func ChartAxis(kind string, ss []Series, lo, hi float64, rows int) series.Axis {
	if !math.IsNaN(lo) && !math.IsNaN(hi) && hi <= lo {
		hi = math.NaN()
	}
	dlo, dhi := math.Inf(1), math.Inf(-1)
	for _, sr := range ss {
		for _, v := range sr.Values {
			if !math.IsNaN(v) {
				dlo, dhi = math.Min(dlo, v), math.Max(dhi, v)
			}
		}
	}
	t := max(3, min(6, rows/2))
	first := (&series.Scale{Min: lo, Max: hi, Ticks: t}).Fit(dlo, dhi)
	if ticksApart(kind, first, rows) {
		return first
	}
	best := first
	for t--; t >= 1; t-- {
		a := (&series.Scale{Min: lo, Max: hi, Ticks: t}).Fit(dlo, dhi)
		if ticksApart(kind, a, rows) && a.Hi-a.Lo <= 1.5*(first.Hi-first.Lo) {
			return a
		}
		if span, bs := a.Hi-a.Lo, best.Hi-best.Lo; span < bs || span == bs && a.Step > best.Step {
			best = a
		}
	}
	return best
}

// ticksApart reports whether each of an axis's ticks falls in a row of its
// own, with a free row between any two.
func ticksApart(kind string, a series.Axis, rows int) bool {
	used := map[int]bool{}
	for _, v := range a.Ticks() {
		r := TickRow(kind, a.Frac(v), rows)
		if used[r-1] || used[r] || used[r+1] {
			return false
		}
		used[r] = true
	}
	return true
}

// TickRow is the row, of a plot rows tall (0 at the top), that a value at
// frac of the axis is drawn in: for a line, the row its dots are in
// (DotY); for bars, which rise from a row's bottom edge, the row their top
// is in.
func TickRow(kind string, frac float64, rows int) int {
	if kind == ChartBar {
		up := int(math.Ceil(clamp01(frac)*float64(rows))) - 1
		return rows - 1 - max(0, up)
	}
	return DotY(frac, 4*rows) / 4
}

// DotY is the dot row (0 at the top) of a value at frac of the axis, in a
// plot dh braille dots high (four a row): from the middle of the bottom
// row to the middle of the top one, so that a tick's label, set in its
// row, is level with its value.
func DotY(frac float64, dh int) int {
	return int(math.Round((1-clamp01(frac))*float64(dh-4))) + 2
}

func clamp01(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return math.Max(0, math.Min(1, f))
}

// YAxis is a HottyChart's axis: its Min, Max and Step.
func (e *Element) YAxis() series.Axis { return series.Axis{Lo: e.Min, Hi: e.Max, Step: e.Step} }

// Points is how many points a chart has: its longest series' values.
func (e *Element) Points() int {
	n := 0
	for _, s := range e.Series {
		n = max(n, len(s.Values))
	}
	return n
}

// Slots is how many points a chart's x axis holds: its window, else as
// many as it has. Point j of Points is at slot Slots - Points + j: in a
// window, the newest is at the right edge, and the points before the
// first are blank, as a stream fills it.
func (e *Element) Slots() int {
	if e.Window > 0 {
		return max(e.Window, e.Points())
	}
	return e.Points()
}

// Empty reports whether a chart has no value to draw.
func (e *Element) Empty() bool {
	for _, s := range e.Series {
		for _, v := range s.Values {
			if !math.IsNaN(v) {
				return false
			}
		}
	}
	return true
}

// Legend reports whether a chart shows its series' labels: when it has
// several series, or one with a label.
func (e *Element) Legend() bool {
	return len(e.Series) > 1 || len(e.Series) == 1 && e.Series[0].Label != ""
}

// PointLabel is the label of a chart's point j, "" without one.
func (e *Element) PointLabel(j int) string {
	if j < 0 || j >= len(e.Labels) {
		return ""
	}
	return e.Labels[j]
}

// At is series i's value at point j: NaN where it has none.
func (e *Element) At(i, j int) float64 {
	if i < 0 || i >= len(e.Series) || j < 0 || j >= len(e.Series[i].Values) {
		return math.NaN()
	}
	return e.Series[i].Values[j]
}

// LabelPicks are the points a line chart labels along its x axis, n of
// them at most: evenly spread over its slots, the first and the last
// slot among them, each the point at that slot (-1 where the slot has
// none yet, a window filling).
func (e *Element) LabelPicks(n int) []int {
	slots := e.Slots()
	if len(e.Labels) == 0 || slots == 0 {
		return nil
	}
	n = min(n, slots)
	off := slots - e.Points()
	picks := make([]int, 0, n)
	for k := range n {
		slot := 0
		if n > 1 {
			slot = int(math.Round(float64(k) / float64(n-1) * float64(slots-1)))
		}
		j := slot - off
		if j < 0 || j >= len(e.Labels) {
			j = -1
		}
		picks = append(picks, j)
	}
	return picks
}

// Ticks are a chart's ticks, each with its label (TickLabel); none while
// it is empty, whose axis says nothing.
func (e *Element) Ticks() []Tick {
	if e.Empty() {
		return nil
	}
	a := e.YAxis()
	var out []Tick
	for _, v := range a.Ticks() {
		out = append(out, Tick{Value: v, Frac: a.Frac(v), Text: TickLabel(v, a.Step)})
	}
	return out
}

// Tick is a value on a chart's axis: where it is (Frac, 0 at the bottom,
// 1 at the top) and its label.
type Tick struct {
	Value float64
	Frac  float64
	Text  string
}

// TickLabel is a tick's label, with as many decimals as its step needs
// (0.25, 0.50): in thousands (k), millions (M) or billions (G) once the
// step is one (2.5k, 125k), so that the labels stay narrow; 0 alone.
func TickLabel(v, step float64) string {
	if v == 0 {
		return "0"
	}
	for _, u := range []struct {
		n float64
		s string
	}{{1e9, "G"}, {1e6, "M"}, {1e3, "k"}} {
		if step >= u.n {
			return num(v/u.n, stepDecimals(step/u.n)) + u.s
		}
	}
	return num(v, stepDecimals(step))
}

// stepDecimals is how many decimals a step needs: 0 for 5, 1 for 0.5.
func stepDecimals(step float64) int { return series.Axis{Step: step}.Decimals() }

func num(v float64, d int) string {
	s := strconv.FormatFloat(v, 'f', d, 64)
	if strings.Trim(s, "-0.") == "" {
		return strings.TrimPrefix(s, "-") // no -0
	}
	return s
}
