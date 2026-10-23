package cells

import (
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// hidden (io_neuroplast_hotty.hidden, profile §6.34) follows the data
// model: while true, the component shows nothing, takes no room and no
// Tab stop, and the keyboard on it moves on to what follows it.
// accessibility.hidden is a screen reader's alone: its icon is drawn.
func TestHidden(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"same":true}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["edit","reset","next","note"]},
	 {"id":"edit","component":"Button","child":"edit-t","action":{"event":{"name":"edit"}}},
	 {"id":"edit-t","component":"Text","text":"Edit"},
	 {"id":"reset","component":"Button","child":"reset-t","action":{"event":{"name":"reset"}},
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"hidden":{"@path":"/same"}}}}},
	 {"id":"reset-t","component":"Text","text":"Reset"},
	 {"id":"next","component":"Button","child":"next-t","action":{"event":{"name":"next"}}},
	 {"id":"next-t","component":"Text","text":"Next"},
	 {"id":"note","component":"Row","children":["mark","say"]},
	 {"id":"mark","component":"Icon","name":"check","accessibility":{"hidden":true}},
	 {"id":"say","component":"Text","text":"saved"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	r := New(c)
	same := func(v string) {
		t.Helper()
		if err := p.ProcessJSON([]byte(`{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/same","value":` + v + `}}`)); err != nil {
			t.Fatal(err)
		}
		c.Rebuild() // as a program does on the processor's Changed
	}

	hiddenFrame := r.Draw(20).Plain()
	if strings.Contains(hiddenFrame, "Reset") || !strings.Contains(hiddenFrame, "✓ saved") {
		t.Errorf("Reset shows, or the icon hidden from a screen reader does not:\n%s", hiddenFrame)
	}
	if _, _, _, _, ok := r.Box("reset"); ok {
		t.Error("the hidden Reset has a box")
	}
	if got := c.V.Focusables(); !slices.Equal(got, []string{"edit", "next"}) {
		t.Errorf("Tab stops while Reset is hidden: %v", got)
	}

	// Shown, it takes the room it had none of.
	same("false")
	shownFrame := r.Draw(20).Plain()
	if !strings.Contains(shownFrame, "Reset") || strings.Count(shownFrame, "\n") <= strings.Count(hiddenFrame, "\n") {
		t.Errorf("Reset shown takes no room:\n%s\nwas\n%s", shownFrame, hiddenFrame)
	}
	if got := c.V.Focusables(); !slices.Equal(got, []string{"edit", "reset", "next"}) {
		t.Errorf("Tab stops while Reset shows: %v", got)
	}

	// Pressed, Reset hides itself, and the keyboard goes on to Next.
	c.Focus("reset")
	same("true")
	if c.St.Focus != "next" || !c.St.Keyboard {
		t.Errorf("the keyboard after Reset hid: %q %v, want next", c.St.Focus, c.St.Keyboard)
	}
	if got := r.Draw(20).Plain(); got != hiddenFrame {
		t.Errorf("hidden again:\n%s\nwant\n%s", got, hiddenFrame)
	}
}
