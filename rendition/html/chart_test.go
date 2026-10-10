package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// TestChartOnHost: a HottyChart on a host is an image to a screen reader,
// named by its kind and its series; its ticks' labels sit at their values;
// a line is an svg a series, a path made by hotty-go's chart (a lone point
// a dot, a gap where a value is missing); bars are boxes from the axis's 0,
// one below it hanging down. A new point is an attribute's delta a line,
// and what the host then has is the document. A HottySparkline is a box of
// bars, the newest first; an empty chart says so until a point comes.
func TestChartOnHost(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"a":[1,2,3],"b":[3,null,1],"v":[2,null,8]}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["load","bars","spark","none"]},
 {"id":"load","component":"HottyChart","catalogId":"`+hottycat.ID+`","series":[{"label":"cpu","values":{"@path":"/a"}},{"label":"mem","values":{"@path":"/b"}}],"window":4,"height":4},
 {"id":"bars","component":"HottyChart","catalogId":"`+hottycat.ID+`","kind":"bar","values":[3,-1,2],"labels":["a","b","c"],"height":4},
 {"id":"spark","component":"HottySparkline","catalogId":"`+hottycat.ID+`","values":{"@path":"/v"},"window":5,"max":10},
 {"id":"none","component":"HottyChart","catalogId":"`+hottycat.ID+`","values":{"@path":"/n"},"height":2}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	path := func(i string) string { return partID("load", partPath+i) }
	for _, want := range [][3]string{{"load", "role", "img"}, {"load", "aria-label", "Line chart: cpu, mem"}, {"load", "class", "k-chart k-chart-line"},
		{"load", "style", "--k-rows: 4"}, {partID("load", partSeries+"1"), "class", "k-series-1"}, {partID("load", partSeries+"1"), "viewBox", "0 0 480 80"},
		{path("0"), "d", "M160,60L320,40L480,20"}, {path("1"), "d", "M160,20L160,20M480,60L480,60"}, {path("0"), "stroke", "currentColor"},
		{"bars", "aria-label", "Bar chart"}, {"spark", "role", "img"}, {"spark", "style", "--k-rows: 1; --k-n: 5"}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	h := s.HTML()
	for _, want := range []string{
		`<div class="k-chart-legend"><span class="k-chart-key"><i class="k-key k-series-0"></i>cpu</span><span class="k-chart-key"><i class="k-key k-series-1"></i>mem</span></div>`,
		`<div class="k-chart-y" aria-hidden="true" style="min-width: 2ch"><span style="top: 100.00%">-2</span><span style="top: 66.67%">0</span>`,
		`<div class="k-bar-group"><span class="k-bar k-series-0"><i style="bottom: 16.67%; height: 16.67%"></i></span></div>`,
		`<div class="k-chart-cats" aria-hidden="true"><span>a</span><span>b</span><span>c</span></div>`,
		`<i style="height: 80.00%"></i><i></i><i style="height: 20.00%"></i></span>`,
		`<span class="k-chart-empty">No data</span>`} {
		if !strings.Contains(h, want) {
			t.Errorf("no %s in\n%s", want, h)
		}
	}

	var set []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/a/3","value":2}}]`), &set))
	before := x.deltas
	x.process(set...)
	x.check(r)
	if n := x.deltas - before; n != 2 {
		t.Errorf("a new point in a window made %d deltas, want the two lines' d", n)
	}
	if got, _ := s.Attr(path("0"), "d"); got != "M0,60L160,40L320,20L480,40" {
		t.Errorf("after a new point, the line is %q", got)
	}

	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/n/0","value":5}}]`), &set))
	x.process(set...)
	x.check(r)
	if strings.Contains(s.HTML(), "No data") {
		t.Error("a chart with a point still says it has none")
	}
}
