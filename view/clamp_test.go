package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// lines (io_neuroplast_hotty.lines, profile §6.40) is a Text's Clamp: a
// whole number from 1 to MaxLines, a fraction rounded down. One that is
// not a number is none, and so is one on anything else.
func TestClamp(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	lines := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"lines":` + v + `}}}` }
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["one","none","part","far","zero","word","field"]},
	 {"id":"one","component":"Text","text":"Title",` + lines("1") + `},
	 {"id":"none","component":"Text","text":"Prose"},
	 {"id":"part","component":"Text","text":"Two",` + lines("2.9") + `},
	 {"id":"far","component":"Text","text":"Far",` + lines("1e300") + `},
	 {"id":"zero","component":"Text","text":"Zero",` + lines("0") + `},
	 {"id":"word","component":"Text","text":"Word",` + lines(`"one"`) + `},
	 {"id":"field","component":"TextField","label":"Name",` + lines("1") + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for id, want := range map[string]int{"one": 1, "none": 0, "part": 2, "far": view.MaxLines, "zero": 1, "word": 0, "field": 0} {
		if got := c.V.Find(id).Clamp; got != want {
			t.Errorf("%s's clamp: %d, want %d (0 for none)", id, got, want)
		}
	}
}
