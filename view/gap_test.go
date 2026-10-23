package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// gap (io_neuroplast_hotty.gap, profile §6.36) is a Column's, a Row's or a
// List's: a whole number from 0 to MaxGap, a fraction rounded down. One
// that is not a number is none, and so is one on anything else.
func TestGap(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	gap := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"gap":` + v + `}}}` }
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["one","none","half","far","neg","word","row","list","text"]},
	 {"id":"one","component":"Column","children":[],` + gap("1") + `},
	 {"id":"none","component":"Column","children":[]},
	 {"id":"half","component":"Column","children":[],` + gap("1.7") + `},
	 {"id":"far","component":"Column","children":[],` + gap("20") + `},
	 {"id":"neg","component":"Column","children":[],` + gap("-2") + `},
	 {"id":"word","component":"Column","children":[],` + gap(`"wide"`) + `},
	 {"id":"row","component":"Row","children":[],` + gap("0") + `},
	 {"id":"list","component":"List","children":[],` + gap("2") + `},
	 {"id":"text","component":"Text","text":"x",` + gap("2") + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for id, want := range map[string]int{"one": 1, "none": -1, "half": 1, "far": view.MaxGap, "neg": 0, "word": -1, "row": 0, "list": 2, "text": -1} {
		got := -1
		if g := c.V.Find(id).Gap; g != nil {
			got = *g
		}
		if got != want {
			t.Errorf("%s's gap: %d, want %d (-1 for none)", id, got, want)
		}
	}
}
