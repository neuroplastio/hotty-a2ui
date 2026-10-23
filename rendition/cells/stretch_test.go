package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// stretch (io_neuroplast_hotty.stretch, profile §6.35) on the root, with
// the rows SetRows gave, makes the root as tall as the pane, and a weighted
// HottyScrollView takes the rows the rest leaves, its height the least. A
// HottyScrollView in a Row keeps its height, as a host's does.
func TestStretch(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["head","body","side"],
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"stretch":true}}}},
	 {"id":"head","component":"Text","text":"Settings"},
	 {"id":"body","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["a","b","c"],"height":3,"weight":1},
	 {"id":"side","component":"Row","justify":"stretch","children":["nav","log"],"weight":1},
	 {"id":"nav","component":"Text","text":"nav"},
	 {"id":"log","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["x"],"height":2}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	rows := func(id string) int {
		t.Helper()
		_, _, _, h, ok := r.Box(id)
		if !ok {
			t.Fatalf("%s has no box", id)
		}
		return h
	}

	// Without SetRows, the frame is as tall as its content.
	f := r.Draw(40)
	if got := rows("body"); got != 3 {
		t.Errorf("body without SetRows: %d rows, want its height, 3", got)
	}
	natural := f.Rows

	// A pane of 30 rows: the 30 less the head, the Row and the gaps go to
	// the weighted body and Row by weight; the Row's scroll view keeps 2.
	r.SetRows(30)
	f = r.Draw(40)
	if f.Rows != 30 {
		t.Errorf("frame: %d rows, want the pane's 30", f.Rows)
	}
	spare := 30 - natural
	if got, want := rows("body"), 3+spare/2+spare%2; got != want {
		t.Errorf("body in a pane of 30: %d rows, want %d", got, want)
	}
	if got := rows("log"); got != 2 {
		t.Errorf("a scroll view in a Row: %d rows, want its height, 2", got)
	}

	// A pane shorter than the content: the content's rows, the height kept.
	r.SetRows(4)
	if f = r.Draw(40); f.Rows != natural || rows("body") != 3 {
		t.Errorf("a pane of 4: frame %d rows (want %d), body %d (want 3)", f.Rows, natural, rows("body"))
	}
}
