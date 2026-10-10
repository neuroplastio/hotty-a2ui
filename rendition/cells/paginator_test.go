package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// pagers is a Column of HottyPaginators: one over a List of a template,
// twelve fruit five a page, its page bound and its turns sent; the agent's
// twelve pages as numbers, the third shown; and thirty pages of dots,
// more than fit; and the actions it sent.
func pagers(t *testing.T) (*Rendition, *view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"page":1,"fruit":[
	 {"name":"Apple"},{"name":"Banana"},{"name":"Cherry"},{"name":"Date"},{"name":"Elderberry"},{"name":"Fig"},
	 {"name":"Grape"},{"name":"Honeydew"},{"name":"Kiwi"},{"name":"Lemon"},{"name":"Mango"},{"name":"Nectarine"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["pager","results","many"]},
	 {"id":"pager","component":"HottyPaginator",` + h + `,"child":"list","perPage":5,"page":{"@path":"/page"},
	  "onChange":{"event":{"name":"turned","context":{"page":{"@path":"/page"}}}}},
	 {"id":"list","component":"List","children":{"componentId":"fruit","path":"/fruit"}},
	 {"id":"fruit","component":"Text","text":{"@path":"name"}},
	 {"id":"results","component":"HottyPaginator",` + h + `,"pages":12,"page":3,"displayStyle":"numbers"},
	 {"id":"many","component":"HottyPaginator",` + h + `,"pages":30}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// A HottyPaginator is its child's page past a field's gutter, a blank row,
// then its dots, the page shown's in the foreground and the others border
// and faint, and a blank row after it; one without a child is its dots
// alone, its numbers "3/12" with displayStyle numbers or where its dots do
// not fit. A focused one has the gutter's bar down its rows and the page
// shown in the accent.
func TestPaginatorDraws(t *testing.T) {
	r, c, _ := pagers(t)
	want := "  Apple\n" +
		"  Banana\n" +
		"  Cherry\n" +
		"  Date\n" +
		"  Elderberry\n" +
		"\n" +
		"  •••\n" +
		"\n" +
		"  3/12\n" +
		"\n" +
		"  1/30"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, tc := range []struct {
		row, col int
		role     Role
		attr     Attr
	}{
		{6, 2, Fg, 0}, {6, 3, Border, Faint}, {6, 4, Border, Faint}, {8, 2, Fg, 0},
	} {
		if cell := f.Cells[tc.row][tc.col]; cell.Role != tc.role || cell.Attr != tc.attr {
			t.Errorf("row %d col %d %q: %v %v, want %v %v", tc.row, tc.col, cell.Text, cell.Role, cell.Attr, tc.role, tc.attr)
		}
	}
	if _, y, _, h, ok := r.Box("pager"); !ok || y != 0 || h != 7 {
		t.Errorf("pager's box at row %d, %d tall", y, h)
	}
	c.Focus("pager")
	f = r.Draw(30)
	for row := range 7 {
		if bar := f.Cells[row][0]; bar.Text != "┃" || bar.Role != Accent {
			t.Errorf("focused: row %d's bar %q %v", row, bar.Text, bar.Role)
		}
	}
	if dot, other := f.Cells[6][2], f.Cells[6][3]; dot.Role != Accent || other.Role != Border || other.Attr != Faint {
		t.Errorf("focused: the page shown %v, another %v %v", dot.Role, other.Role, other.Attr)
	}
	if f.Cells[7][0].Text != " " {
		t.Errorf("the bar runs past the dots: %q", f.Cells[7][0].Text)
	}
	c.Focus("results")
	if n := r.Draw(30).Cells[8][2]; n.Text != "3" || n.Role != Accent {
		t.Errorf("focused numbers: %q %v", n.Text, n.Role)
	}
}

// The last page, shorter, keeps the paginator as tall as its tallest, so
// that the dots stay put as the pages turn, as they do not in bubbles.
func TestPaginatorKeepsItsHeight(t *testing.T) {
	r, c, _ := pagers(t)
	r.Draw(30)
	if err := c.TurnPage("pager", 2); err != nil {
		t.Fatal(err)
	}
	want := "  Mango\n" +
		"  Nectarine\n" +
		"\n" +
		"\n" +
		"\n" +
		"\n" +
		"  •••\n"
	if got := r.Draw(30).Plain(); got[:len(want)] != want {
		t.Errorf("got\n%s\nwant it to start\n%s", got, want)
	}
}

// Keys turn a focused paginator's pages, ← → and h l as bubbles' do, and a
// click on a dot shows its page, giving the paginator the keyboard; a click
// on its page gives it the keyboard alone.
func TestPaginatorKeysAndClicks(t *testing.T) {
	r, c, actions := pagers(t)
	r.Draw(30)
	key := func(k string) {
		t.Helper()
		if _, err := r.Key(k); err != nil {
			t.Fatal(err)
		}
		r.Draw(30)
	}
	c.Focus("pager")
	key("ArrowRight")
	key("l")
	if got := c.S.Data.Value("/page"); got != 3.0 || len(*actions) != 2 {
		t.Errorf("→ l: /page %v, %d actions", got, len(*actions))
	}
	key("h")
	if got := c.S.Data.Value("/page"); got != 2.0 {
		t.Errorf("h: /page %v", got)
	}
	c.Focus("")
	r.Draw(30)
	if err := r.Click(3, 1); err != nil { // Banana's row, on page 2 Grape's
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/page"); got != 2.0 || c.St.Focus != "pager" || !c.St.Keyboard {
		t.Errorf("a click on the page: /page %v, keyboard %v on %q", got, c.St.Keyboard, c.St.Focus)
	}
	r.Draw(30)
	col, row, ok := r.dotCell("pager", 1)
	if !ok || col != 2 || row != 6 {
		t.Fatalf("page 1's dot at %d,%d (%v)", col, row, ok)
	}
	if err := r.Click(col, row); err != nil {
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/page"); got != 1.0 || len(*actions) != 4 {
		t.Errorf("a click on the first dot: /page %v, %d actions", got, len(*actions))
	}
	if _, _, ok := r.dotCell("results", 1); ok {
		t.Error("numbers have dots")
	}
}
