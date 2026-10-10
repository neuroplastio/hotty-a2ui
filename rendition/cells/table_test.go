package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// cityTable is a surface with a Table of three rows, two at a time, and
// the actions it sent.
func cityTable(t *testing.T) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name+":"+a2ui.ToString(o.Action.Context["city"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"sel":"Tokyo","rows":[
	 {"rank":1,"city":"Tokyo","pop":"37,274,000"},{"rank":2,"city":"Delhi","pop":"32,065,760"},{"rank":3,"city":"São Paulo","pop":"22,429,800"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyTable","catalogId":"` + hotty.ID + `","columns":[{"key":"rank","header":"Rank","align":"end"},{"key":"city","header":"City"},{"key":"pop","header":"Population","align":"end"}],
	  "rows":{"@path":"/rows"},"rowKey":"city","selected":{"@path":"/sel"},"height":2,
	  "onActivate":{"event":{"name":"open","context":{"city":{"@path":"/sel"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// A Table is its header, a rule that says which rows show while the body
// scrolls, and the rows, each cell padded a column a side, cut with "…"
// when the columns do not fit; the widest columns give way first. While
// it scrolls, a scroll view's bar runs down the body a blank column past
// the last column's padding, and the rule runs over them.
func TestTableDraws(t *testing.T) {
	r, _, _ := cityTable(t)
	want := " Rank  City       Population\n" +
		"──────────────────── 1–2 of 3 ─\n" +
		"    1  Tokyo      37,274,000  ┃\n" +
		"    2  Delhi      32,065,760  │"
	if got := r.Draw(40).Plain(); got != want {
		t.Fatalf("at 40 columns:\n%s\nwant\n%s", got, want)
	}
	want = " Rank  City  Pop…\n" +
		"───────── 1–2 of 3 ─\n" +
		"    1  Tok…  37,…  ┃\n" +
		"    2  Del…  32,…  │"
	if got := r.Draw(20).Plain(); got != want {
		t.Fatalf("at 20 columns:\n%s\nwant\n%s", got, want)
	}
}

// A table that shows all its rows has no bar, and keeps its width.
func TestTableNoBar(t *testing.T) {
	r := coding(t, `{"id":"t","component":"HottyTable","catalogId":"`+hotty.ID+`","columns":[{"key":"city","header":"City"}],
		"rows":[{"city":"Tokyo"},{"city":"Delhi"}],"height":2}`)
	if got := r.Draw(40).Plain(); got != " City\n───────\n Tokyo\n Delhi" {
		t.Errorf("got\n%s", got)
	}
	if w := tableNatural(r.c.V.Find("t")); w != 7 {
		t.Errorf("natural width %d", w)
	}
}

// The selected row is reversed across the table: in the accent while the
// table has the keyboard, in muted otherwise.
func TestTableSelection(t *testing.T) {
	r, c, _ := cityTable(t)
	f := r.Draw(40)
	if cell := f.Cells[2][0]; cell.Role != Muted || cell.Attr&Reverse == 0 {
		t.Errorf("an unfocused selected row: %v %v", cell.Role, cell.Attr)
	}
	if cell := f.Cells[3][0]; cell.Attr&Reverse != 0 {
		t.Errorf("an unselected row is reversed")
	}
	if thumb, track := f.Cells[2][30], f.Cells[3][30]; thumb.Role != Muted || thumb.Attr != 0 || track.Role != Border || track.Attr != Faint {
		t.Errorf("the bar without the keyboard: %+v %+v", thumb, track)
	}
	c.Focus("root")
	f = r.Draw(40)
	for x := range 28 {
		if cell := f.Cells[2][x]; cell.Role != Accent || cell.Attr&Reverse == 0 {
			t.Fatalf("the focused selected row at %d: %v %v", x, cell.Role, cell.Attr)
		}
	}
	if thumb := f.Cells[2][30]; thumb.Text != "┃" || thumb.Role != Accent || thumb.Attr&Reverse != 0 {
		t.Errorf("the bar with the keyboard: %+v", thumb)
	}
}

// Keys move the selection and scroll the body; a click on a row selects
// it, and a click on the selected row, or Enter, acts on it.
func TestTableKeysAndClicks(t *testing.T) {
	r, c, actions := cityTable(t)
	r.Draw(40)
	if err := r.Click(5, 3); err != nil {
		t.Fatal(err)
	}
	if !r.focused("root") || c.S.Data.Value("/sel") != "Delhi" || len(*actions) != 0 {
		t.Fatalf("a click on Delhi's row: focused %v, selected %v, actions %v", r.focused("root"), c.S.Data.Value("/sel"), *actions)
	}
	r.Draw(40)
	if err := r.Click(5, 3); err != nil {
		t.Fatal(err)
	}
	if len(*actions) != 1 || (*actions)[0] != "open:Delhi" {
		t.Fatalf("a second click: %v", *actions)
	}
	keys(t, r, "ArrowDown")
	want := " Rank  City       Population\n" +
		"──────────────────── 2–3 of 3 ─\n" +
		"    2  Delhi      32,065,760  │\n" +
		"    3  São Paulo  22,429,800  ┃"
	if got := r.Draw(40).Plain(); got != want {
		t.Fatalf("after ArrowDown:\n%s\nwant\n%s", got, want)
	}
	keys(t, r, "Home", "Enter")
	if len(*actions) != 2 || (*actions)[1] != "open:Tokyo" {
		t.Fatalf("Home, Enter: %v", *actions)
	}
	if err := r.Click(5, 0); err != nil || c.S.Data.Value("/sel") != "Tokyo" || len(*actions) != 2 {
		t.Errorf("a click on the header changed something: %v %v %v", err, c.S.Data.Value("/sel"), *actions)
	}
}
