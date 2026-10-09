package view_test

import (
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyScrollView's lines are text, a line each; one the agent left out
// (a null, padding past the end) is empty. It shows 10 rows without a
// height. With follow it follows the tail until it is scrolled; a child
// takes the place of lines.
func TestScrollView(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"log":["one",2,null,"four"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["log","doc"]},
	 {"id":"log","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":{"@path":"/log"},"follow":true,"wrap":true},
	 {"id":"doc","component":"HottyScrollView","catalogId":"` + hotty.ID + `","child":"text","height":4},
	 {"id":"text","component":"Text","text":"A document"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	e := c.V.Find("log")
	if e == nil || e.Kind != view.ScrollView || !e.Focusable() || !slices.Equal(e.Lines, []string{"one", "2", "", "four"}) {
		t.Fatalf("log: %+v", e)
	}
	if e.Height != 10 || !e.Active || !e.Tail || !e.Wrap {
		t.Errorf("log: height %d, follow %v, tail %v, wrap %v", e.Height, e.Active, e.Tail, e.Wrap)
	}
	doc := c.V.Find("doc")
	if len(doc.Children) != 1 || doc.Children[0].ID != "text" || doc.Lines != nil || doc.Height != 4 || doc.Tail {
		t.Errorf("doc: %+v", doc)
	}
	// A rendition scrolled the log up: it stops following.
	c.Scrolled("log", 1, 0, false)
	c.Rebuild()
	if e = c.V.Find("log"); e.Top != 1 || e.Tail {
		t.Errorf("scrolled: top %d, tail %v", e.Top, e.Tail)
	}
	c.ScrollTo("log", "end")
	if e = c.V.Find("log"); !e.Tail {
		t.Error("hottyScrollTo end does not follow the tail")
	}
	c.ScrollTo("log", "start")
	if e = c.V.Find("log"); e.Top != 0 || e.Tail {
		t.Errorf("hottyScrollTo start: top %d, tail %v", e.Top, e.Tail)
	}
}

// hottyScrollTo, as the agent calls it, scrolls the scroll view by its id.
func TestScrollToCall(t *testing.T) {
	cat := hotty.Catalog()
	var got []string
	hotty.Implement(cat, hotty.Renderer{
		Focus: func(*a2ui.Surface, a2ui.Scope, string) error { return nil },
		Blur:  func(*a2ui.Surface) error { return nil },
		ScrollTo: func(_ *a2ui.Surface, _ a2ui.Scope, id, to string) error {
			got = append(got, id+":"+to)
			return nil
		},
	})
	p := a2ui.NewProcessor(basic.Catalog(), cat)
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["a"]}]}},
	{"version":"v1.0","callRendererFunction":{"functionCallId":"c1","callFunction":{"@call":"hottyScrollTo","catalogId":"` + hotty.ID + `","args":{"id":"root","to":"start"}}}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"root:start"}) {
		t.Errorf("calls %q", got)
	}
}

// A HottyScrollView's hints are bubbles' viewport's: its short help, and
// in the full view its moves and its pages, and ← and → for lines that do
// not wrap. A host scrolls it itself, so it has none there.
func TestScrollHints(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["log","wrapped","hints"]},
	 {"id":"log","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["a"]},
	 {"id":"wrapped","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["a"],"wrap":true},
	 {"id":"hints","component":"HottyKeyHints","catalogId":"` + hotty.ID + `"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.Focus("log")
	short, full := c.KeyHints("", false)
	if got := hintKeys(short); !slices.Equal(got, []string{"↑/k up", "↓/j down", "f/pgdn page down", "b/pgup page up", "? more"}) {
		t.Errorf("short: %q", got)
	}
	if len(full) != 4 || !slices.Equal(hintKeys(full[2]), []string{"←/h move left", "→/l move right"}) {
		t.Errorf("full: %q", full)
	}
	c.Focus("wrapped")
	if _, full = c.KeyHints("", false); len(full) != 3 {
		t.Errorf("wrapped, the full view: %q", full)
	}
	if short, _ = c.KeyHints("", true); slices.Contains(hintKeys(short), "↑/k up") {
		t.Errorf("on a host: %q", short)
	}
}
