package view_test

import (
	"math"
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// charts is a surface of the given hotty components, each a JSON object
// without its catalogId, over a data model.
func charts(t *testing.T, data string, comps ...string) *view.Controller {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	ids := ""
	list := ""
	for i, c := range comps {
		id := string(rune('a' + i))
		if i > 0 {
			ids += ","
			list += ","
		}
		ids += `"` + id + `"`
		list += `{"id":"` + id + `","catalogId":"` + hotty.ID + `",` + c + `}`
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":` + data + `}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":[` + ids + `]},` + list + `]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s"))
}

// same reports whether two lists of values are equal, NaN equal to NaN.
func same(a, b []float64) bool {
	return slices.EqualFunc(a, b, func(x, y float64) bool { return x == y || math.IsNaN(x) && math.IsNaN(y) })
}

var nan = math.NaN()

// A HottyChart's element has its kind, its series (bound, or its values
// as one), its labels, its height (8 without one) and its axis, fitted to
// the values from 0, its ends on round ticks; a window keeps the last
// points and their labels; a null, or what is not a number, is a gap.
func TestChartBuilds(t *testing.T) {
	c := charts(t, `{"cpu":[12,18,61,57],"mem":[40,41,null,54],"t":["10:00","10:01","10:02","10:03"],"w":[1,2,"x",4,5,6]}`,
		`"component":"HottyChart","series":[{"label":"cpu","values":{"@path":"/cpu"}},{"label":"mem","values":{"@path":"/mem"}}],"labels":{"@path":"/t"}`,
		`"component":"HottyChart","kind":"bar","values":[420,310,-180,-60,250,530],"height":6`,
		`"component":"HottyChart","values":{"@path":"/w"},"labels":["a","b","c","d","e","f"],"window":4`,
		`"component":"HottyChart","values":[3,7],"min":0,"max":100,"height":3`,
		`"component":"HottyChart","values":{"@path":"/none"},"max":-5`)
	a := c.V.Find("a")
	if a.Kind != view.Chart || a.Variant != view.ChartLine || a.Height != view.ChartRows || a.Focusable() {
		t.Fatalf("a: kind %s %s, height %d, focusable %v", a.Kind, a.Variant, a.Height, a.Focusable())
	}
	if len(a.Series) != 2 || a.Series[1].Label != "mem" || !same(a.Series[1].Values, []float64{40, 41, nan, 54}) || !slices.Equal(a.Labels, []string{"10:00", "10:01", "10:02", "10:03"}) {
		t.Errorf("a: series %+v, labels %v", a.Series, a.Labels)
	}
	if a.Min != 0 || a.Max != 75 || a.Step != 25 || !a.Legend() || a.Points() != 4 {
		t.Errorf("a: axis %v..%v by %v, legend %v, points %d", a.Min, a.Max, a.Step, a.Legend(), a.Points())
	}
	b := c.V.Find("b")
	if b.Variant != view.ChartBar || b.Height != 6 || b.Min != -250 || b.Max != 750 || b.Step != 250 || b.Legend() {
		t.Errorf("b: %s, height %d, axis %v..%v by %v, legend %v", b.Variant, b.Height, b.Min, b.Max, b.Step, b.Legend())
	}
	var ticks []string
	for _, tk := range b.Ticks() {
		ticks = append(ticks, tk.Text)
	}
	if !slices.Equal(ticks, []string{"-250", "0", "250", "500", "750"}) {
		t.Errorf("b's ticks: %v", ticks)
	}
	w := c.V.Find("c")
	if !same(w.Series[0].Values, []float64{nan, 4, 5, 6}) || !slices.Equal(w.Labels, []string{"c", "d", "e", "f"}) || w.Window != 4 || w.Slots() != 4 {
		t.Errorf("c: window %d, values %v, labels %v", w.Window, w.Series[0].Values, w.Labels)
	}
	if d := c.V.Find("d"); d.Min != 0 || d.Max != 100 || d.Height != 3 {
		t.Errorf("d: axis %v..%v, height %d", d.Min, d.Max, d.Height)
	}
	e := c.V.Find("e")
	if e.Variant != view.ChartLine || !e.Empty() || e.Ticks() != nil || !(e.Max > e.Min) {
		t.Errorf("e: %s, empty %v, ticks %v, axis %v..%v", e.Variant, e.Empty(), e.Ticks(), e.Min, e.Max)
	}
}

// A HottySparkline is one series, a row tall without a height, from 0 (or
// its lowest value, below 0) to its highest unless min and max say; a
// window keeps the last values.
func TestSparklineBuilds(t *testing.T) {
	c := charts(t, `{"v":[3,9,null,6]}`,
		`"component":"HottySparkline","values":{"@path":"/v"}`,
		`"component":"HottySparkline","values":[2,-4,1],"height":2`,
		`"component":"HottySparkline","values":[50,60,70],"min":40,"max":100,"window":2`,
		`"component":"HottySparkline","values":{"@path":"/none"},"window":20`,
		`"component":"HottySparkline","values":[5,5]`)
	for _, want := range []struct {
		id         string
		vals       []float64
		lo, hi     float64
		height, wn int
	}{
		{"a", []float64{3, 9, nan, 6}, 0, 9, 1, 0},
		{"b", []float64{2, -4, 1}, -4, 2, 2, 0},
		{"c", []float64{60, 70}, 40, 100, 1, 2},
		{"d", []float64{}, 0, 1, 1, 20},
		{"e", []float64{5, 5}, 0, 5, 1, 0},
	} {
		e := c.V.Find(want.id)
		if e.Kind != view.Sparkline || e.Focusable() || len(e.Series) != 1 || !same(e.Series[0].Values, want.vals) ||
			e.Min != want.lo || e.Max != want.hi || e.Height != want.height || e.Window != want.wn {
			t.Errorf("%s: values %v, %v..%v, height %d, window %d; want %+v", want.id, e.Series, e.Min, e.Max, e.Height, e.Window, want)
		}
	}
}

// A chart's axis has its ticks a row apart at least, so that cells label
// each: where the fitted axis's are closer, it takes fewer, unless that
// would squeeze the plot; in 3 rows none does, and it is the tightest.
func TestChartAxis(t *testing.T) {
	vals := []view.Series{{Values: []float64{12, 61}}}
	for _, want := range []struct {
		kind         string
		rows         int
		lo, hi, step float64
		apart        bool
	}{
		{view.ChartLine, 8, 0, 75, 25, true},
		{view.ChartBar, 8, 0, 75, 25, true},
		{view.ChartLine, 3, 0, 75, 25, false},
		{view.ChartLine, 20, 0, 70, 10, true},
	} {
		a := view.ChartAxis(want.kind, vals, nan, nan, want.rows)
		if a.Lo != want.lo || a.Hi != want.hi || a.Step != want.step {
			t.Errorf("%s, %d rows: %v..%v by %v, want %v..%v by %v", want.kind, want.rows, a.Lo, a.Hi, a.Step, want.lo, want.hi, want.step)
		}
		rows := map[int]bool{}
		for _, v := range a.Ticks() {
			if !want.apart {
				break
			}
			r := view.TickRow(want.kind, a.Frac(v), want.rows)
			if rows[r-1] || rows[r] || rows[r+1] {
				t.Errorf("%s, %d rows: tick %v's row %d is next to another's", want.kind, want.rows, v, r)
			}
			rows[r] = true
		}
	}
	if a := view.ChartAxis(view.ChartLine, vals, 10, 5, 8); a.Lo != 10 || a.Hi <= 61 {
		t.Errorf("a max under the min: %v..%v", a.Lo, a.Hi)
	}
}

// A tick's label has the decimals its step needs, and is in k, M or G once
// the step is.
func TestTickLabel(t *testing.T) {
	for _, c := range []struct {
		v, step float64
		want    string
	}{
		{0, 25, "0"}, {-250, 250, "-250"}, {1500, 500, "1500"}, {2500, 500, "2500"}, {2500, 2500, "2.5k"},
		{125000, 25000, "125k"}, {1e6, 5e5, "1000k"}, {1.5e6, 1.5e6, "1.5M"}, {3e9, 1e9, "3G"}, {0.25, 0.05, "0.25"}, {0.5, 0.5, "0.5"},
	} {
		if got := view.TickLabel(c.v, c.step); got != c.want {
			t.Errorf("TickLabel(%v, %v) = %q, want %q", c.v, c.step, got, c.want)
		}
	}
}

// A line chart labels a few points evenly along its slots, the first and
// the last among them; in a window still filling, a slot with no point yet
// has none.
func TestLabelPicks(t *testing.T) {
	c := charts(t, `{}`,
		`"component":"HottyChart","values":[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19],"labels":["0","1","2","3","4","5","6","7","8","9","10","11","12","13","14","15","16","17","18","19"]`,
		`"component":"HottyChart","values":[1,2,3,4],"labels":["a","b","c","d"],"window":10`,
		`"component":"HottyChart","values":[1,2]`)
	if got := c.V.Find("a").LabelPicks(3); !slices.Equal(got, []int{0, 10, 19}) {
		t.Errorf("20 points: %v", got)
	}
	if got := c.V.Find("b").LabelPicks(3); !slices.Equal(got, []int{-1, -1, 3}) {
		t.Errorf("4 points in a window of 10: %v", got)
	}
	if got := c.V.Find("c").LabelPicks(3); got != nil {
		t.Errorf("no labels: %v", got)
	}
}
