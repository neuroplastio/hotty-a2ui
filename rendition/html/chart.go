package html

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-go/chart"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyChart and a HottySparkline on a host (profile §2): HTML's boxes
// for the axes, the bars and the labels, and for a line an inline svg a
// series, its path made by hotty-go's chart.Line (its gaps, its slots, a
// value held to the plot) and drawn with presentation attributes only, in
// currentColor, which the svg's class sets: what every host's SVG takes,
// and the theme's colours, which the kit knows only as CSS variables.

// The box a line's svg is drawn in, in CSS pixels, which preserveAspectRatio
// none stretches over the plot. The markup does not know the plot's width:
// 480 is about 50 columns of a terminal's text, a storybook pane's plot,
// and a row is 20 (the kit's --hotty-cell-h without a host's). A plot much
// wider or narrower stretches the line's stroke where it is steep, which
// the plot's real size, the program's to send, mends (vault KIT-10h).
const (
	lineBoxW = 480
	lineBoxH = 20
)

// chartView is a HottyChart: a grid of its ticks' labels, each at its
// value's height, and its plot, --k-rows terminal rows tall, whose left
// and bottom borders are its axes; under the plot, its points' labels (a
// line's few, evenly spread; a bar chart's each under its bars) and its
// legend. It is an image to a screen reader, named by its accessibility
// label, else by its series.
func chartView(e *view.Element) *node {
	n := el("div", "id", domID(e.ID), "class", "k-chart k-chart-"+e.Variant, "role", "img", "aria-label", chartName(e),
		"style", "--k-rows: "+strconv.Itoa(max(e.Height, 1)))
	// The labels are placed at their values, out of the flow, so their
	// column is as wide as the widest, in figures (tabular, a ch each).
	ys := el("div", "class", "k-chart-y", "aria-hidden", "true")
	cols := 1
	for _, t := range e.Ticks() {
		ys.add(el("span", "style", "top: "+pct(1-t.Frac)).add(txt(t.Text)))
		cols = max(cols, utf8.RuneCountInString(t.Text))
	}
	ys.set("style", "min-width: "+strconv.Itoa(cols)+"ch")
	plot := el("div", "id", partID(e.ID, partPlot), "class", "k-chart-plot")
	switch {
	case e.Empty():
		plot.add(el("span", "class", "k-chart-empty").add(txt("No data")))
	case e.Variant == view.ChartBar:
		bars(plot, e)
	default:
		lines(plot, e)
	}
	n.add(ys, plot)
	if len(e.Labels) > 0 {
		row := el("div", "class", "k-chart-x", "aria-hidden", "true")
		if e.Variant == view.ChartBar {
			row.set("class", "k-chart-cats")
			for k := range e.Points() {
				row.add(el("span").add(texts(e.PointLabel(k))...))
			}
		} else {
			for _, j := range e.LabelPicks(3) {
				s := el("span")
				if j >= 0 {
					s.add(texts(e.PointLabel(j))...)
				}
				row.add(s)
			}
		}
		n.add(row)
	}
	if e.Legend() {
		legend := el("div", "class", "k-chart-legend")
		for i, s := range e.Series {
			legend.add(el("span", "class", "k-chart-key").add(el("i", "class", "k-key "+seriesClass(i))).add(texts(s.Label)...))
		}
		n.add(legend)
	}
	return n
}

// lines draws each series as a line across the plot: an svg of one path,
// whose id lets a new point be one attribute's delta.
func lines(plot *node, e *view.Element) {
	var xs []float64
	if e.Window > 0 {
		// The newest at the right edge, as in cells (view.Element.Slots).
		slots, off := e.Slots(), e.Slots()-e.Points()
		for j := range e.Points() {
			xs = append(xs, float64(off+j)/float64(max(slots-1, 1)))
		}
	}
	h := float64(max(e.Height, 1) * lineBoxH)
	box := "0 0 " + strconv.Itoa(lineBoxW) + " " + strconv.FormatFloat(h, 'f', -1, 64)
	for i, s := range e.Series {
		c := chart.Line{W: lineBoxW, H: h, Lo: e.Min, Hi: e.Max, Xs: xs}
		plot.add(el("svg", "id", partID(e.ID, partSeries+strconv.Itoa(i)), "class", seriesClass(i), "viewBox", box,
			"preserveAspectRatio", "none", "overflow", "visible", "aria-hidden", "true").add(
			el("path", "id", partID(e.ID, partPath+strconv.Itoa(i)), "d", c.Path(s.Values), "fill", "none", "stroke", "currentColor",
				"stroke-width", strconv.FormatFloat(chart.DefaultStroke, 'f', -1, 64), "stroke-linejoin", "round", "stroke-linecap", "round")))
	}
}

// bars draws a group a point, a bar a series in it, each from the axis's 0
// to its value: a box placed by its bottom and height, in percent of the
// plot.
func bars(plot *node, e *view.Element) {
	a := e.YAxis()
	z := clamp01(a.Frac(0))
	for k := range e.Points() {
		g := el("div", "class", "k-bar-group")
		for i := range e.Series {
			b := el("span", "class", "k-bar "+seriesClass(i))
			if v := e.At(i, k); !math.IsNaN(v) {
				f := clamp01(a.Frac(v))
				b.add(el("i", "style", "bottom: "+pct(math.Min(z, f))+"; height: "+pct(math.Abs(f-z))))
			}
			g.add(b)
		}
		plot.add(g)
	}
}

// sparkline is a HottySparkline: a column a value, a terminal column wide,
// from the bottom of a box --k-rows terminal rows tall and --k-n columns
// wide, the newest first in a row that runs right to left, so that it
// ends at the box's right edge and a box too narrow loses the oldest, as
// in cells.
func sparkline(e *view.Element) *node {
	var vals []float64
	if len(e.Series) > 0 {
		vals = e.Series[0].Values
	}
	cols := e.Window
	if cols == 0 {
		cols = len(vals)
	}
	n := el("span", "id", domID(e.ID), "class", "k-spark", "role", "img", "aria-label", "Sparkline",
		"style", "--k-rows: "+strconv.Itoa(max(e.Height, 1))+"; --k-n: "+strconv.Itoa(cols))
	// A value at Min is an eighth of a row, as in cells, so that only a
	// missing value is blank.
	least := 1 / float64(8*max(e.Height, 1))
	for k := len(vals) - 1; k >= 0; k-- {
		b := el("i")
		if v := vals[k]; !math.IsNaN(v) {
			b.set("style", "height: "+pct(math.Max(least, clamp01((v-e.Min)/(e.Max-e.Min)))))
		}
		n.add(b)
	}
	return n
}

// chartName is what a chart is to a screen reader without an
// accessibility label: its kind, and its series' labels.
func chartName(e *view.Element) string {
	name := "Line chart"
	if e.Variant == view.ChartBar {
		name = "Bar chart"
	}
	var labels []string
	for _, s := range e.Series {
		if s.Label != "" {
			labels = append(labels, s.Label)
		}
	}
	if len(labels) > 0 {
		name += ": " + strings.Join(labels, ", ")
	}
	return name
}

// seriesClass colours series i (kit.css): info, warning, success, error,
// and round again, as cells colours them.
func seriesClass(i int) string { return "k-series-" + strconv.Itoa(i%4) }

// pct is a fraction as a CSS percentage.
func pct(f float64) string { return strconv.FormatFloat(f*100, 'f', 2, 64) + "%" }

func clamp01(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return math.Max(0, math.Min(1, f))
}
