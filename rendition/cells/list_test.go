package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// snacks is a surface with a HottyList of four items, two a page, and the
// actions it sent.
func snacks(t *testing.T) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name+":"+a2ui.ToString(o.Action.Context["snack"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"sel":"nutella"}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyList","catalogId":"` + hotty.ID + `","title":"Snacks","selected":{"@path":"/sel"},"filterable":true,"height":2,
	  "items":[{"label":"Nutella","description":"It's good on toast","value":"nutella"},{"label":"Bitter melon","description":"It cools you down","value":"melon"},
	   {"label":"Popcorn","description":"Salted","value":"popcorn"},{"label":"Nuts","value":"nuts"}],
	  "onActivate":{"event":{"name":"eat","context":{"snack":{"@path":"/sel"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// A HottyList is laid out as bubbles' list: the title, the status line,
// a page of items, a label and a description each, a blank row between,
// then the page's dots; every line two columns in, the selected item's
// columns a "│" bar. A label too long is cut with "…".
func TestListDraws(t *testing.T) {
	r, _, _ := snacks(t)
	want := "   Snacks\n" +
		"\n" +
		"  4 items\n" +
		"\n" +
		"│ Nutella\n" +
		"│ It's good on toast\n" +
		"\n" +
		"  Bitter melon\n" +
		"  It cools you down\n" +
		"\n" +
		"  ••"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if c := f.Cells[0][2]; c.Role != Accent || c.Attr&Reverse == 0 {
		t.Errorf("the title: %v %v, want the accent reversed", c.Role, c.Attr)
	}
	if c := f.Cells[2][2]; c.Role != Muted {
		t.Errorf("the status line: %v", c.Role)
	}
	if on, off := f.Cells[10][2], f.Cells[10][3]; on.Role != Fg || off.Role != Border {
		t.Errorf("the dots: %v then %v, want the page's in fg, the others in border", on.Role, off.Role)
	}
	if got := r.Draw(12).Plain(); got != "   Snacks\n\n  4 items\n\n│ Nutella\n│ It's good…\n\n  Bitter me…\n  It cools …\n\n  ••" {
		t.Errorf("at 12 columns:\n%s", got)
	}
}

// The selected item's bar is muted while the list does not have the
// keyboard, its label and description plain; with the keyboard, all three
// are in the accent.
func TestListSelection(t *testing.T) {
	r, c, _ := snacks(t)
	f := r.Draw(30)
	if bar, label := f.Cells[4][0], f.Cells[4][2]; bar.Role != Muted || label.Role != Fg || label.Attr&Bold == 0 {
		t.Errorf("unfocused: bar %v, label %v %v, want its label bold", bar.Role, label.Role, label.Attr)
	}
	c.Focus("root")
	f = r.Draw(30)
	for _, at := range [][2]int{{4, 0}, {4, 2}, {5, 2}} {
		if cell := f.Cells[at[0]][at[1]]; cell.Role != Accent {
			t.Errorf("focused, at %v: %v", at, cell.Role)
		}
	}
	if cell := f.Cells[7][2]; cell.Role != Fg {
		t.Errorf("an item not selected: %v", cell.Role)
	}
}

// While the filter is typed it takes the title's row, with the cursor
// after it; the items it leaves come best first (Nuts, the shorter), the
// first selected, and the characters of a label that matched it are
// underlined. An item without a description keeps a blank second row, and
// the page's dots keep their rows while one page is left.
func TestListFilterLine(t *testing.T) {
	r, c, _ := snacks(t)
	c.Focus("root")
	keys(t, r, "/", "n", "u")
	want := "  Filter: nu\n" +
		"\n" +
		"  2 items • 2 filtered\n" +
		"\n" +
		"│ Nuts\n" +
		"│\n" +
		"\n" +
		"  Nutella\n" +
		"  It's good on toast\n" +
		"\n"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if col, row, ok := f.Cursor(); !ok || col != 12 || row != 0 {
		t.Errorf("the cursor at %d,%d (%v), want after the filter", col, row, ok)
	}
	for _, at := range []struct {
		row, col int
		under    bool
	}{{4, 2, true}, {4, 3, true}, {4, 4, false}, {7, 2, true}, {7, 3, true}, {7, 4, false}} {
		if got := f.Cells[at.row][at.col].Attr&Underline != 0; got != at.under {
			t.Errorf("row %d, column %d underlined: %v", at.row, at.col, got)
		}
	}
	keys(t, r, "Enter")
	if got := r.Draw(30).Plain(); got[:len("   Snacks\n\n  “nu” 2 items • 2 filtered")] != "   Snacks\n\n  “nu” 2 items • 2 filtered" {
		t.Errorf("applied:\n%s", got)
	}
}

// A click on an item selects it, and gives the list the keyboard; a click
// on the selected item, or Enter, acts on it; a click on the title does
// nothing more.
func TestListClicks(t *testing.T) {
	r, c, actions := snacks(t)
	r.Draw(30)
	if err := r.Click(4, 8); err != nil {
		t.Fatal(err)
	}
	if !r.focused("root") || c.S.Data.Value("/sel") != "melon" || len(*actions) != 0 {
		t.Fatalf("a click on Bitter melon's description: focused %v, selected %v, actions %v", r.focused("root"), c.S.Data.Value("/sel"), *actions)
	}
	r.Draw(30)
	if err := r.Click(4, 7); err != nil || len(*actions) != 1 || (*actions)[0] != "eat:melon" {
		t.Fatalf("a second click: %v %v", err, *actions)
	}
	keys(t, r, "ArrowRight")
	if got := c.S.Data.Value("/sel"); got != "nuts" {
		t.Fatalf("→ turned to %v, want the second page's second item", got)
	}
	r.Draw(30)
	if err := r.Click(4, 0); err != nil || c.S.Data.Value("/sel") != "nuts" || len(*actions) != 1 {
		t.Errorf("a click on the title changed something: %v %v %v", err, c.S.Data.Value("/sel"), *actions)
	}
}
