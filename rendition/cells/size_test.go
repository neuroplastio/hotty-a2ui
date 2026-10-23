package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// size (io_neuroplast_hotty.size, profile §6.38) changes nothing in cells:
// a large Button is a medium one's row, as wide.
func TestSize(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	large := `"metadata":{"extensions":{"io_neuroplast_hotty":{"size":"large"}}}`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","align":"start","children":["medium","large","bare"]},
	 {"id":"medium","component":"Button","variant":"primary","child":"medium-t","action":{"event":{"name":"go"}}},
	 {"id":"large","component":"Button","variant":"primary","child":"large-t","action":{"event":{"name":"go"}},` + large + `},
	 {"id":"bare","component":"Button","variant":"borderless","child":"bare-t","action":{"event":{"name":"go"}},` + large + `},
	 {"id":"medium-t","component":"Text","text":"Continue"},
	 {"id":"large-t","component":"Text","text":"Continue"},
	 {"id":"bare-t","component":"Text","text":"Continue"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	r.Draw(40)
	for id, w := range map[string]int{"medium": 10, "large": 10, "bare": 8} {
		if _, _, gw, gh, ok := r.Box(id); !ok || gw != w || gh != 1 {
			t.Errorf("%s: %d×%d (%v), want %d×1", id, gw, gh, ok, w)
		}
	}
}
