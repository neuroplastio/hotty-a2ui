package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// width (io_neuroplast_hotty.width, profile §6.37) is a TextField's or a
// DateTimeInput's: a whole number from 1 to MaxWidth, a fraction rounded
// down. One that is not a number is none, and so is one on anything else.
func TestWidth(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	width := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"width":` + v + `}}}` }
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["port","none","part","far","zero","neg","word","long","date","pick","col"]},
	 {"id":"port","component":"TextField","label":"Port",` + width("5") + `},
	 {"id":"none","component":"TextField","label":"Name"},
	 {"id":"part","component":"TextField","label":"Part",` + width("7.9") + `},
	 {"id":"far","component":"TextField","label":"Far",` + width("1000") + `},
	 {"id":"zero","component":"TextField","label":"Zero",` + width("0") + `},
	 {"id":"neg","component":"TextField","label":"Neg",` + width("-3") + `},
	 {"id":"word","component":"TextField","label":"Word",` + width(`"wide"`) + `},
	 {"id":"long","component":"TextField","label":"Notes","variant":"longText",` + width("40") + `},
	 {"id":"date","component":"DateTimeInput","label":"When","enableDate":true,` + width("12") + `},
	 {"id":"pick","component":"ChoicePicker","label":"Pick","variant":"mutuallyExclusive","options":[{"label":"a","value":"a"}],` + width("12") + `},
	 {"id":"col","component":"Column","children":[],` + width("12") + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for id, want := range map[string]int{"port": 5, "none": 0, "part": 7, "far": view.MaxWidth, "zero": 1, "neg": 1, "word": 0, "long": 40, "date": 12, "pick": 0, "col": 0} {
		if got := c.V.Find(id).Width; got != want {
			t.Errorf("%s's width: %d, want %d (0 for none)", id, got, want)
		}
	}
}
