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

// The short view is bubbles' help line: the keys muted, what they do
// fainter, the separators faint border; what does not fit is cut, " …"
// ending the line where it fits.
func TestKeyHintsLine(t *testing.T) {
	r, c := helped(t)
	if got := last(r.Draw(60)); got != "? more" {
		t.Errorf("without the keyboard: %q", got)
	}
	c.Focus("l")
	f := r.Draw(60)
	if got := last(f); got != "↑/k up • ↓/j down • / filter • enter choose • ? more" {
		t.Fatalf("the list focused: %q", got)
	}
	row := f.Cells[len(f.Cells)-1]
	for i, want := range map[int]style{0: hintKey, 2: hintKey, 3: hintDesc, 4: hintDesc, 7: hintDim, 9: hintKey} {
		if row[i].Role != want.role || row[i].Attr != want.attr {
			t.Errorf("column %d (%q): %v %v, want %v %v", i, row[i].Text, row[i].Role, row[i].Attr, want.role, want.attr)
		}
	}
	if got := last(r.Draw(30)); got != "↑/k up • ↓/j down • / filter …" {
		t.Errorf("at 30 columns: %q", got)
	}
	if got := last(r.Draw(28)); got != "↑/k up • ↓/j down • / filter" {
		t.Errorf("at 28 columns: %q", got)
	}
}

// ? opens the full view, in columns four apart, each a group: the list's
// moves, its filter and Enter, then Tab, Shift+Tab and ?; ? again closes
// it. While the list's
// filter or a field is typed, ? is typed.
func TestKeyHintsFull(t *testing.T) {
	r, c := helped(t)
	c.Focus("l")
	r.Draw(60)
	keys(t, r, "?")
	want := "↑/k      up             /     filter    tab       next\n" +
		"↓/j      down           enter choose    shift+tab back\n" +
		"→/l/pgdn next page                      ?         close help\n" +
		"←/h/pgup prev page\n" +
		"g/home   go to start\n" +
		"G/end    go to end"
	if got := r.Draw(60).Plain(); !strings.HasSuffix(got, "\n"+want) {
		t.Fatalf("the full view:\n%s\nwant it to end\n%s", got, want)
	}
	if got := r.Draw(30).Plain(); !strings.Contains(got, "\n↑/k      up          …\n") {
		t.Errorf("at 30 columns, the second group does not fit:\n%s", got)
	}
	keys(t, r, "?")
	if got := last(r.Draw(60)); got != "↑/k up • ↓/j down • / filter • enter choose • ? more" {
		t.Fatalf("? again: %q", got)
	}
	keys(t, r, "/", "?")
	if e := c.V.Find("l"); e.Query.Text != "?" || c.St.FullHints {
		t.Errorf("? in the filter: %+v, full %v", e.Query, c.St.FullHints)
	}
	if got := last(r.Draw(60)); got != "enter apply filter • esc cancel • ? more" {
		t.Errorf("while the filter is typed: %q", got)
	}
	c.Focus("title")
	keys(t, r, "?")
	if got := c.S.Data.Value("/title"); got != "?" || c.St.FullHints {
		t.Errorf("? in a field: %v, full %v", got, c.St.FullHints)
	}
}
