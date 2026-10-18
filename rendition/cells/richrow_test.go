package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A vertical List's Button whose label is a Row is that Row laid out
// across the List, a column of padding each side, each Text in its own
// tone: a mark, a name that takes the room, a word at the end. The whole
// row takes the click, and the keyboard reverses it from end to end, as a
// plain row's.
func TestRichRow(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	var acts []string
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			acts = append(acts, o.Action.Name+" "+a2ui.ToString(o.Action.Context["who"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"rows":[
	  {"who":"alpha","word":"working","tone":"warning"},{"who":"beta","word":"idle","tone":"success"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"List","children":{"componentId":"row","path":"/rows"}},
	 {"id":"row","component":"Button","variant":"borderless","child":"line",
	  "action":{"event":{"name":"go","context":{"who":{"@path":"who"}}}}},
	 {"id":"line","component":"Row","align":"center","children":["dot","name","word"]},
	 {"id":"dot","component":"Text","text":"●","metadata":{"extensions":{"io_neuroplast_hotty":{"tone":{"@path":"tone"}}}}},
	 {"id":"name","component":"Text","text":{"@path":"who"},"weight":1},
	 {"id":"word","component":"Text","text":{"@path":"word"},"metadata":{"extensions":{"io_neuroplast_hotty":{"tone":{"@path":"tone"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	r := New(c)
	f := r.Draw(24)
	rows := strings.Split(f.Plain(), "\n")
	if want := []string{" ● alpha        working", " ● beta            idle"}; strings.Join(rows, "|") != strings.Join(want, "|") {
		t.Fatalf("rows:\n%s\nwant\n%s", f.Plain(), strings.Join(want, "\n"))
	}
	for _, at := range []struct {
		y, x int
		want Role
	}{{0, 1, Warning}, {0, 3, Fg}, {0, 16, Warning}, {1, 1, Success}, {1, 19, Success}} {
		if g := f.Cells[at.y][at.x]; g.Role != at.want || g.Attr&Underline != 0 {
			t.Errorf("row %d col %d %q: %v, %v; want %v, not underlined", at.y, at.x, g.Text, g.Role, g.Attr, at.want)
		}
	}

	// A click on the word at the row's end is the row's.
	if err := r.Click(20, 1); err != nil {
		t.Fatal(err)
	}
	if strings.Join(acts, "|") != "go beta" {
		t.Fatalf("a click on beta's word: %q", acts)
	}
	// With the keyboard, the row is reversed in the accent end to end.
	f = r.Draw(24)
	for x := 0; x < 24; x++ {
		if g := f.Cells[1][x]; g.Role != Accent || g.Attr&Reverse == 0 {
			t.Fatalf("beta's row with the keyboard, col %d %q: %v, %v", x, g.Text, g.Role, g.Attr)
		}
	}
}
