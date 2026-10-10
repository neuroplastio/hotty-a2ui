package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// pagers is a surface with two HottyPaginators: one over a List of a
// template, twelve fruit five a page, its page bound and its turns sent;
// one of the agent's four pages, its page a literal; and the actions it
// sent.
func pagers(t *testing.T) (*view.Controller, *[]*a2ui.ActionMessage) {
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
	 {"id":"root","component":"Column","children":["pager","results"]},
	 {"id":"pager","component":"HottyPaginator",` + h + `,"child":"list","perPage":5,"page":{"@path":"/page"},
	  "onChange":{"event":{"name":"turned","context":{"page":{"@path":"/page"}}}}},
	 {"id":"list","component":"List","children":{"componentId":"fruit","path":"/fruit"}},
	 {"id":"fruit","component":"Text","text":{"@path":"name"}},
	 {"id":"results","component":"HottyPaginator",` + h + `,"pages":4,"page":9,"displayStyle":"numbers"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s")), actions
}

// A HottyPaginator with a child cuts the child's items into pages and
// gives the child the page shown; one without shows the agent's pages,
// the page held to them.
func TestPaginatorBuilds(t *testing.T) {
	c, _ := pagers(t)
	e := c.V.Find("pager")
	if e == nil || e.Kind != view.Paginator || !e.Focusable() || e.Variant != "dots" || e.Height != 5 {
		t.Fatalf("pager: %+v, want a focusable paginator of dots, 5 a page", e)
	}
	if n, at := e.Pages(); n != 3 || at != 0 || len(e.Paged) != 3 || len(e.Paged[2]) != 2 {
		t.Errorf("pages %d, at %d, paged %d", n, at, len(e.Paged))
	}
	list := c.V.Find("list")
	if len(list.Children) != 5 || list.Children[0].Markdown != "Apple" || c.V.Find("fruit[/fruit/5]") != nil {
		t.Errorf("the list shows %d items, the first %q, and Fig is in the view", len(list.Children), list.Children[0].Markdown)
	}
	if ids := c.V.Focusables(); len(ids) != 2 || ids[0] != "pager" || ids[1] != "results" {
		t.Errorf("focusables %q", ids)
	}
	r := c.V.Find("results")
	if n, at := r.Pages(); n != 4 || at != 3 || r.Variant != "numbers" || r.PageText() != "4/4" || len(r.Children) != 0 {
		t.Errorf("results: pages %d, at %d, %q, %q", n, at, r.Variant, r.PageText())
	}
}

// Keys turn pages as bubbles' paginator does, none past the ends, and Home
// and End go to them: each turn is written where page is bound, from 1,
// then onChange runs; a key at an end is taken and sends nothing. An
// unbound page is the renderer's.
func TestPaginatorKeys(t *testing.T) {
	c, actions := pagers(t)
	for _, step := range []struct {
		key  string
		page float64
		sent int
	}{
		{"ArrowLeft", 1, 0}, {"ArrowRight", 2, 1}, {"l", 3, 2}, {"PageDown", 3, 2}, {"h", 2, 3}, {"Home", 1, 4},
		{"End", 3, 5}, {"PageUp", 2, 6},
	} {
		if ok, err := c.PageKey("pager", step.key); !ok || err != nil {
			t.Fatalf("%s: ok %v, %v", step.key, ok, err)
		}
		if got := c.S.Data.Value("/page"); got != step.page || len(*actions) != step.sent {
			t.Errorf("%s: /page %v, %d actions; want %v, %d", step.key, got, len(*actions), step.page, step.sent)
		}
	}
	if a := (*actions)[len(*actions)-1]; a.Name != "turned" || a.Context["page"] != 2.0 {
		t.Errorf("the last action %s %v", a.Name, a.Context)
	}
	if list := c.V.Find("list"); len(list.Children) != 5 || list.Children[0].Markdown != "Fig" {
		t.Errorf("page 2 starts with %q", list.Children[0].Markdown)
	}
	if ok, _ := c.PageKey("pager", "ArrowUp"); ok {
		t.Error("the paginator took ArrowUp")
	}
	// The agent writes the page: the child follows.
	if err := c.S.Write("/page", 3.0); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if list := c.V.Find("list"); len(list.Children) != 2 || list.Children[1].Markdown != "Nectarine" {
		t.Errorf("page 3: %d items", len(list.Children))
	}
	if ok, err := c.PageKey("results", "Home"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if _, at := c.V.Find("results").Pages(); at != 0 || c.St.Local["results"] != 1.0 {
		t.Errorf("results at %d, its state %v", at, c.St.Local["results"])
	}
	if err := c.TurnPage("results", 7); err != nil {
		t.Fatal(err)
	}
	if _, at := c.V.Find("results").Pages(); at != 3 {
		t.Errorf("a turn past the end shows page %d", at+1)
	}
}

// A paginator's hints are bubbles' paginator example's, "←/→ page"; a host
// leaves its keys to the program, so they show there too.
func TestPaginatorHints(t *testing.T) {
	c, _ := pagers(t)
	c.Focus("pager")
	for _, host := range []bool{false, true} {
		short, full := c.KeyHints("", host)
		if len(short) == 0 || short[0] != (view.Hint{Key: "←/→", Desc: "page"}) || len(full) == 0 || len(full[0]) != 4 {
			t.Errorf("host %v: short %v, full %v", host, short, full)
		}
	}
}
