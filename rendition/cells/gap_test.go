package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// gap (io_neuroplast_hotty.gap, profile §6.36) puts its rows between a
// Column's children, whatever they are, in place of the blank rows cells
// would put there: two Texts, which stack tight, a row apart; a Button a
// Text would set apart, tight. In a Row it is the columns between the
// children, in the Row's natural width too.
func TestGap(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	gap := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"gap":` + v + `}}}` }
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","align":"start","children":["apart","tight","wide","none"]},
	 {"id":"apart","component":"Column","children":["a1","a2"],` + gap("1") + `},
	 {"id":"a1","component":"Text","text":"one"},
	 {"id":"a2","component":"Text","text":"two"},
	 {"id":"tight","component":"Column","children":["t1","t2"],` + gap("0") + `},
	 {"id":"t1","component":"Text","text":"Sure?"},
	 {"id":"t2","component":"Button","child":"t2-t","action":{"event":{"name":"ok"}}},
	 {"id":"t2-t","component":"Text","text":"OK"},
	 {"id":"wide","component":"Row","children":["w1","w2"],` + gap("3") + `},
	 {"id":"w1","component":"Text","text":"left"},
	 {"id":"w2","component":"Text","text":"right"},
	 {"id":"none","component":"Row","children":["n1","n2"],` + gap("0") + `},
	 {"id":"n1","component":"Text","text":"in"},
	 {"id":"n2","component":"Text","text":"to"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	frame := r.Draw(30).Plain()
	lines := strings.Split(frame, "\n")
	want := []string{"one", "", "two", "Sure?", " OK ", "left   right", "into"}
	if len(lines) < len(want) {
		t.Fatalf("frame:\n%s", frame)
	}
	for i, w := range want {
		if strings.TrimRight(lines[i], " ") != strings.TrimRight(w, " ") {
			t.Errorf("row %d: %q, want %q\n%s", i, lines[i], w, frame)
		}
	}
	for id, w := range map[string]int{"wide": 12, "none": 4} {
		if _, _, got, _, _ := r.Box(id); got != w {
			t.Errorf("%s is %d wide, want %d", id, got, w)
		}
	}
}
