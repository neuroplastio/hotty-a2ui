package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// width (io_neuroplast_hotty.width, profile §6.37) is a field's input's
// columns, in place of a text field's 32 and a DateTime's form: its value
// shows in that many, and its natural width counts that many. A longText
// is that wide, its grip's column among them.
func TestFieldWidth(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	width := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"width":` + v + `}}}` }
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"port":"8080123","name":"8080123"}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","align":"start","children":["port","name","date","long"]},
	 {"id":"port","component":"TextField","value":{"@path":"/port"},"accessibility":{"label":"Port"},` + width("5") + `},
	 {"id":"name","component":"TextField","value":{"@path":"/name"},"accessibility":{"label":"Name"}},
	 {"id":"date","component":"DateTimeInput","enableDate":true,"accessibility":{"label":"When"},` + width("4") + `},
	 {"id":"long","component":"TextField","variant":"longText","accessibility":{"label":"Notes"},` + width("12") + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	frame := r.Draw(60).Plain()
	lines := strings.Split(frame, "\n")
	for i, want := range []string{"8080…", "8080123"} {
		if i >= len(lines) || strings.TrimSpace(lines[i]) != want {
			t.Errorf("row %d shows %q, want %q\n%s", i, lines[min(i, len(lines)-1)], want, frame)
		}
	}
	// The gutter and the inset, then the input; a longText has no inset.
	for id, w := range map[string]int{"port": 2 + 1 + 5, "name": 2 + 1 + textInput, "date": 2 + 1 + 4, "long": 2 + 12} {
		if _, _, got, _, _ := r.Box(id); got != w {
			t.Errorf("%s is %d wide, want %d", id, got, w)
		}
	}
}
