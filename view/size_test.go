package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// size (io_neuroplast_hotty.size, profile §6.38) is a Button's: "large",
// or "" for medium, the default, and anything else. One on anything else
// is none.
func TestSize(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	size := func(v string) string { return `"metadata":{"extensions":{"io_neuroplast_hotty":{"size":` + v + `}}}` }
	button := func(id, v string) string {
		return `{"id":"` + id + `","component":"Button","child":"` + id + `-t","action":{"event":{"name":"go"}},` + size(v) + `},
	 {"id":"` + id + `-t","component":"Text","text":"Go"}`
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["large","medium","huge","num","none","text"]},
	 ` + button("large", `"large"`) + `,
	 ` + button("medium", `"medium"`) + `,
	 ` + button("huge", `"huge"`) + `,
	 ` + button("num", `2`) + `,
	 {"id":"none","component":"Button","child":"none-t","action":{"event":{"name":"go"}}},
	 {"id":"none-t","component":"Text","text":"Go"},
	 {"id":"text","component":"Text","text":"Hi",` + size(`"large"`) + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for id, want := range map[string]string{"large": "large", "medium": "", "huge": "", "num": "", "none": "", "text": ""} {
		if got := c.V.Find(id).Size; got != want {
			t.Errorf("%s's size: %q, want %q", id, got, want)
		}
	}
}
