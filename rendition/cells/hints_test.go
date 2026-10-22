package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// helped is a surface with a filterable HottyList of four items, two a
// page, a title field, and a HottyKeyHints under them.
func helped(t *testing.T) (*Rendition, *view.Controller) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"title":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["l","title","help"]},
	 {"id":"l","component":"HottyList","catalogId":"` + hotty.ID + `","filterable":true,"height":2,
	  "items":[{"label":"Nutella"},{"label":"Bitter melon"},{"label":"Popcorn"},{"label":"Nuts"}],"onActivate":{"event":{"name":"eat"}}},
	 {"id":"title","component":"TextField","label":"Title","value":{"@path":"/title"}},
	 {"id":"help","component":"HottyKeyHints","catalogId":"` + hotty.ID + `"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c
}

// last is a frame's last row as text.
func last(f *Frame) string {
	rows := strings.Split(f.Plain(), "\n")
	return rows[len(rows)-1]
}

// Cells draw keys as words by default, but as glyphs where the user says
// their terminal's font has them (view.Keycaps.Present): cells don't
// force text. The labels follow the system either way.
func TestKeyHintsKeycaps(t *testing.T) {
	r, c := helped(t)
	c.Focus("l")
	for _, tc := range []struct {
		caps view.Keycaps
		want string
	}{
		{view.Keycaps{OS: "mac"}, " ↑ / K  up •  ↓ / J  down •  /  filter •  Return  choose •  ?  more"},
		{view.Keycaps{OS: "mac", Present: view.PresentGlyphs}, " \U000F005D / K  up •  \U000F0045 / J  down •  /  filter •  \U000F0311  choose •  ?  more"},
		{view.Keycaps{Present: view.PresentGlyphs}, " \U000F005D / K  up •  \U000F0045 / J  down •  /  filter •  Enter  choose •  ?  more"},
	} {
		c.Keycaps = tc.caps
		if got := last(r.Draw(80)); got != tc.want {
			t.Errorf("%+v: %q, want %q", tc.caps, got, tc.want)
		}
	}
}

// The short view is bubbles' help line: the keys caps, what they do
// muted and faint, the separators faint border; what does not fit is cut,
// " …" ending the line where it fits.
func TestKeyHintsLine(t *testing.T) {
	r, c := helped(t)
	if got := last(r.Draw(66)); got != " ?  more" {
		t.Errorf("without the keyboard: %q", got)
	}
	c.Focus("l")
	f := r.Draw(66)
	if got := last(f); got != " ↑ / K  up •  ↓ / J  down •  /  filter •  Enter  choose •  ?  more" {
		t.Fatalf("the list focused: %q", got)
	}
	row := f.Cells[len(f.Cells)-1]
	for i, want := range map[int]style{0: capStyle, 1: capStyle, 3: capSep, 5: capStyle, 7: hintDesc, 8: hintDesc, 11: hintDim, 13: capStyle} {
		if row[i].Role != want.role || row[i].Attr != want.attr || row[i].BackMix != want.backMix {
			t.Errorf("column %d (%q): %v %v %d, want %v %v %d", i, row[i].Text, row[i].Role, row[i].Attr, row[i].BackMix, want.role, want.attr, want.backMix)
		}
	}
	if got := last(r.Draw(40)); got != " ↑ / K  up •  ↓ / J  down •  /  filter …" {
		t.Errorf("at 40 columns: %q", got)
	}
	if got := last(r.Draw(39)); got != " ↑ / K  up •  ↓ / J  down •  /  filter" {
		t.Errorf("at 39 columns: %q", got)
	}
}

// ? opens the full view, in columns four apart, each a group: the list's
// moves, its filter and Enter, then Tab, shift+tab and ?; ? again closes
// it. While the list's
// filter or a field is typed, ? is typed.
func TestKeyHintsFull(t *testing.T) {
	r, c := helped(t)
	c.Focus("l")
	r.Draw(60)
	keys(t, r, "?")
	want := " ↑ / K          up              /      filter     Tab        next\n" +
		" ↓ / J          down            Enter  choose     Shift+Tab  back\n" +
		" → / L / PgDn   next page                         ?          close help\n" +
		" ← / H / PgUp   prev page\n" +
		" G / Home       go to start\n" +
		" Shift+G / End  go to end"
	if got := r.Draw(80).Plain(); !strings.HasSuffix(got, "\n"+want) {
		t.Fatalf("the full view:\n%s\nwant it to end\n%s", got, want)
	}
	if got := r.Draw(34).Plain(); !strings.Contains(got, "\n ↑ / K          up          …\n") {
		t.Errorf("at 34 columns, the second group does not fit:\n%s", got)
	}
	keys(t, r, "?")
	if got := last(r.Draw(66)); got != " ↑ / K  up •  ↓ / J  down •  /  filter •  Enter  choose •  ?  more" {
		t.Fatalf("? again: %q", got)
	}
	keys(t, r, "/", "?")
	if e := c.V.Find("l"); e.Query.Text != "?" || c.St.FullHints {
		t.Errorf("? in the filter: %+v, full %v", e.Query, c.St.FullHints)
	}
	if got := last(r.Draw(60)); got != " Enter  apply filter •  Esc  cancel •  ?  more" {
		t.Errorf("while the filter is typed: %q", got)
	}
	c.Focus("title")
	keys(t, r, "?")
	if got := c.S.Data.Value("/title"); got != "?" || c.St.FullHints {
		t.Errorf("? in a field: %v, full %v", got, c.St.FullHints)
	}
}
