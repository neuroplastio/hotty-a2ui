package html

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/qr"
)

// A HottyQRCode on a host is an svg, a white square with the quiet zone
// and a black path of its modules, a run a rectangle: read back, the path
// is the code's dark modules. Its label is under it, an image's name is
// its value, and a new value bound reaches the host as the path's delta.
// A value no code holds is its error.
func TestQRCode(t *testing.T) {
	x := newHarness(t)
	h := `"catalogId":"` + hottycat.ID + `"`
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"url":"https://hotty.neuroplast.io"}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["q","long"]},
 {"id":"q","component":"HottyQRCode",`+h+`,"value":{"@path":"/url"},"label":"Scan to open"},
 {"id":"long","component":"HottyQRCode",`+h+`,"value":"`+strings.Repeat("x", 3000)+`"}]}}]`), &msgs))
	x.process(msgs...)
	s := x.h.Surface(x.rs["s"].name)
	if !strings.Contains(s.HTML(), `<svg class="k-qr-code" viewBox="0 0 37 37" shape-rendering="crispEdges" role="img" aria-label="QR code: https://hotty.neuroplast.io" style="--k-modules: 37"><rect width="37" height="37" fill="#fff"></rect>`) {
		t.Errorf("the svg:\n%s", s.HTML())
	}
	path := partID("q", partPath)
	read := func(value string) {
		t.Helper()
		c, err := qr.Encode(value, qr.M)
		must(t, err)
		d, _ := s.Attr(path, "d")
		if got, want := modules(t, d, c.Size), dark(c); got != want {
			t.Errorf("%s: the path's modules\n%s\nwant\n%s", value, got, want)
		}
	}
	read("https://hotty.neuroplast.io")
	if got := s.TextOf(partID("q", partLabel)); got != "Scan to open" {
		t.Errorf("the label reads %q", got)
	}
	if got := s.TextOf(partID("long", partError)); got != "✗ Too long for a QR code" {
		t.Errorf("too long reads %q", got)
	}
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/url","value":"https://neuroplast.io"}}]`), &msgs))
	x.process(msgs...)
	read("https://neuroplast.io")
	if got, _ := s.Attr("q", "class"); got != "k-qr" {
		t.Errorf("q's class %q", got)
	}
}

// modules reads path data of "Mx yhNv1h-Nz" rectangles back into rows of
// '#' and '.', the quiet zone left off.
func modules(t *testing.T, d string, size int) string {
	t.Helper()
	grid := make([][]byte, size)
	for y := range grid {
		grid[y] = []byte(strings.Repeat(".", size))
	}
	for _, r := range strings.Split(strings.TrimSuffix(d, "z"), "z") {
		var x, y, n, m int
		if _, err := fmt.Sscanf(r, "M%d %dh%dv1h-%d", &x, &y, &n, &m); err != nil || n != m {
			t.Fatalf("%q: %v", r, err)
		}
		for i := range n {
			grid[y-qr.Quiet][x-qr.Quiet+i] = '#'
		}
	}
	var b strings.Builder
	for _, row := range grid {
		b.Write(row)
		b.WriteByte('\n')
	}
	return b.String()
}

func dark(c *qr.Code) string {
	var b strings.Builder
	for y := range c.Size {
		for x := range c.Size {
			if c.Dark(x, y) {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
