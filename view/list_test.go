package view_test

import (
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// faveThings are bubbles' list example's items, as HottyList items.
const faveThings = `[
 {"label":"Raspberry Pi’s","description":"I have ’em all over my house","value":"pi"},
 {"label":"Nutella","description":"It's good on toast","value":"nutella"},
 {"label":"Bitter melon","description":"It cools you down","value":"melon"},
 {"label":"Nice socks","description":"And by that I mean socks without holes","value":"socks"},
 {"label":"Eight hours of sleep","description":"I had this once","value":"sleep"},
 {"label":"Cats","description":"Usually","value":"cats"},
 {"label":"Plantasia, the album","description":"My plants love it too","value":"plantasia"},
 {"label":"Pour over coffee","description":"It takes forever to make though","value":"coffee"},
 {"label":"VR","description":"Virtual reality...what is there to say?","value":"vr"},
 {"label":"Noguchi Lamps","description":"Such pleasing organic forms","value":"lamps"},
 {"label":"Linux","description":"Pretty much the best OS","value":"linux"},
 {"label":"Business school","description":"Just kidding","value":"school"}]`

// faveList is a surface with bubbles' list, five items a page, filterable,
// and a plain list beside it; and the actions it sent.
func faveList(t *testing.T) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"sel":"pi","things":` + faveThings + `}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["l","plain"]},
	 {"id":"l","component":"HottyList","catalogId":"` + hotty.ID + `","title":"My Fave Things","items":{"@path":"/things"},
	  "selected":{"@path":"/sel"},"filterable":true,"height":5,"onActivate":{"event":{"name":"open","context":{"thing":{"@path":"/sel"}}}}},
	 {"id":"plain","component":"HottyList","catalogId":"` + hotty.ID + `","items":[{"label":"One"},{"label":"Two"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s")), actions
}

// A HottyList's items are its labels and descriptions; an item's id is its
// value, else its index. It counts them in its status line, as bubbles'
// list does.
func TestListBuilds(t *testing.T) {
	c, _ := faveList(t)
	e := c.V.Find("l")
	if e == nil || e.Kind != view.RichList || !e.Focusable() || e.Label != "My Fave Things" || !e.Filter {
		t.Fatalf("l: %+v, want a focusable, filterable HottyList", e)
	}
	if len(e.Items) != 12 || e.Items[1] != (view.Entry{Label: "Nutella", Description: "It's good on toast"}) || e.RowIDs[1] != "nutella" {
		t.Errorf("items %+v, ids %q", e.Items, e.RowIDs)
	}
	if got := e.ListStatus(); got != "12 items" {
		t.Errorf("status %q", got)
	}
	if n, at := e.Pages(); n != 3 || at != 0 || e.SelectedRow() != 0 {
		t.Errorf("pages %d, at %d, selected %d", n, at, e.SelectedRow())
	}
	plain := c.V.Find("plain")
	if got := plain.RowIDs; !slices.Equal(got, []string{"0", "1"}) || plain.Placeholder != "No items." || plain.Filter {
		t.Errorf("a plain list: ids %q, empty text %q, filter %v", got, plain.Placeholder, plain.Filter)
	}
	if ok, _ := c.ListKey("plain", "/"); ok {
		t.Error("/ filters a list that is not filterable")
	}
}

// Keys move the selection a row, a page (to the same place, none past the
// ends) and to the ends, as bubbles' list does; the page follows the
// selection, whoever moves it.
func TestListKeys(t *testing.T) {
	c, actions := faveList(t)
	c.Focus("l")
	for _, step := range []struct {
		key      string
		sel, top int
	}{
		{"ArrowDown", 1, 0}, {"j", 2, 0}, {"k", 1, 0}, {"End", 11, 10}, {"PageDown", 11, 10},
		{"PageUp", 6, 5}, {"h", 1, 0}, {"ArrowLeft", 1, 0}, {"l", 6, 5}, {"ArrowRight", 11, 10},
		{"Home", 0, 0}, {"ArrowUp", 0, 0}, {"G", 11, 10}, {"g", 0, 0}, {"ArrowDown", 1, 0},
	} {
		if ok, err := c.ListKey("l", step.key); !ok || err != nil {
			t.Fatalf("%s: %v %v", step.key, ok, err)
		}
		if e := c.V.Find("l"); e.SelectedRow() != step.sel || e.Top != step.top {
			t.Fatalf("after %s: item %d, top %d; want %d, %d", step.key, e.SelectedRow(), e.Top, step.sel, step.top)
		}
	}
	if got := c.S.Data.Value("/sel"); got != "nutella" {
		t.Errorf("the selection wrote %v", got)
	}
	if ok, err := c.ListKey("l", "Enter"); !ok || err != nil || len(*actions) != 1 || (*actions)[0].Context["thing"] != "nutella" {
		t.Fatalf("Enter: %v %v %+v", ok, err, *actions)
	}
	if err := c.S.Write("/sel", "linux"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if e := c.V.Find("l"); e.SelectedRow() != 10 || e.Top != 10 {
		t.Errorf("selected by the agent: item %d, top %d", e.SelectedRow(), e.Top)
	}
	if ok, _ := c.ListKey("l", "x"); ok {
		t.Error("x is a key of a list that is not filtering")
	}
}

// / starts a filter: what the user types narrows the items to those whose
// label holds its characters in order, best first, as bubbles ranks them,
// and selects the first. Enter applies it, Escape drops it; Enter on a
// filter that leaves nothing drops it too.
func TestListFilter(t *testing.T) {
	c, actions := faveList(t)
	c.Focus("l")
	keys := func(ks ...string) *view.Element {
		t.Helper()
		for _, k := range ks {
			if ok, err := c.ListKey("l", k); !ok || err != nil {
				t.Fatalf("%s: %v %v", k, ok, err)
			}
		}
		return c.V.Find("l")
	}
	e := keys("/", "n")
	if !e.Query.Editing || e.Query.Text != "n" || !slices.Equal(e.Shown[:5], []int{1, 3, 9, 10, 2}) || len(e.Shown) != 7 {
		t.Fatalf("after / n: %+v, shown %v", e.Query, e.Shown)
	}
	if got := e.ListStatus(); got != "7 items • 5 filtered" {
		t.Errorf("status %q", got)
	}
	if e.SelectedRow() != 1 || !slices.Equal(e.Matched[0], []int{0}) || len(e.Matched) != 7 {
		t.Errorf("selected %d, matched %v", e.SelectedRow(), e.Matched)
	}
	if e = keys("ArrowDown"); e.SelectedRow() != 3 || e.Query.Text != "n" {
		t.Fatalf("ArrowDown while typing moves: item %d, %+v", e.SelectedRow(), e.Query)
	}
	if e = keys("j"); e.Query.Text != "nj" {
		t.Fatalf("j while typing is typed: %+v", e.Query)
	}
	e = keys("Backspace", "Enter")
	if e.Query.Editing || e.Query.Text != "n" || e.ListStatus() != "“n” 7 items • 5 filtered" {
		t.Fatalf("applied: %+v %q", e.Query, e.ListStatus())
	}
	// The agent selects an item the filter leaves out: Enter does nothing.
	if err := c.S.Write("/sel", "cats"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if ok, err := c.ListKey("l", "Enter"); !ok || err != nil || len(*actions) != 0 {
		t.Fatalf("Enter on a hidden item: %v %v %+v", ok, err, *actions)
	}
	if e = keys("Escape"); e.Query != (view.Query{}) || len(e.Shown) != 12 {
		t.Fatalf("Escape: %+v, %d shown", e.Query, len(e.Shown))
	}
	if ok, _ := c.ListKey("l", "Escape"); ok {
		t.Error("Escape without a filter is the list's")
	}
	e = keys("/", "z", "z", "q")
	if len(e.Shown) != 0 || e.ListStatus() != "Nothing matched • 12 filtered" {
		t.Fatalf("zzq: %v %q", e.Shown, e.ListStatus())
	}
	if e = keys("Enter"); e.Query != (view.Query{}) || len(e.Shown) != 12 {
		t.Fatalf("Enter on nothing matched: %+v", e.Query)
	}
	e = keys("/", "l", "i", "Space", "Escape")
	if e.Query != (view.Query{}) {
		t.Fatalf("Escape while typing: %+v", e.Query)
	}
	if e = keys("/", "l", "i", "Enter"); !slices.Equal(e.Shown, []int{10, 6}) || e.SelectedRow() != 10 {
		t.Errorf("li: shown %v, selected %d", e.Shown, e.SelectedRow())
	}
}
