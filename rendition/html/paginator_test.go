package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// A HottyPaginator on a host is a focusable box, its child's page, then
// its dots, each a click target, or its numbers; a click on a dot shows
// its page and gives the box the keyboard, whose keys reach the program
// (SPEC §10.2), which turns the pages as in cells and runs onChange.
func TestPaginatorOnHost(t *testing.T) {
	x := newHarness(t)
	h := `"catalogId":"` + hottycat.ID + `"`
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"page":1,"fruit":[
 {"name":"Apple"},{"name":"Banana"},{"name":"Cherry"},{"name":"Date"},{"name":"Elderberry"},{"name":"Fig"},{"name":"Grape"}]}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["pager","results"]},
 {"id":"pager","component":"HottyPaginator",`+h+`,"child":"list","perPage":3,"page":{"@path":"/page"},
  "onChange":{"event":{"name":"turned","context":{"page":{"@path":"/page"}}}}},
 {"id":"list","component":"List","children":{"componentId":"fruit","path":"/fruit"}},
 {"id":"fruit","component":"Text","text":{"@path":"name"}},
 {"id":"results","component":"HottyPaginator",`+h+`,"pages":12,"page":3,"displayStyle":"numbers","accessibility":{"label":"Result pages"}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	dot := func(i string) string { return partID("pager", partDots+i) }
	for _, want := range [][3]string{{"pager", "class", "k-pager"}, {"pager", "tabindex", "0"}, {"pager", "role", "group"},
		{"pager", "aria-label", "Pages"}, {"pager", "data-keys", pagerKeys},
		{partID("pager", partDots), "role", "status"}, {partID("pager", partDots), "aria-label", "Page 1 of 3"},
		{dot("0"), "class", "k-dot k-on"}, {dot("2"), "class", "k-dot"}, {dot("2"), "data-on", "click"},
		{"results", "class", "k-pager k-bare"}, {"results", "aria-label", "Result pages"},
		{partID("results", partDots), "class", "k-pages k-page-n"}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	if keys, _ := s.Attr("pager", "data-keys"); len(strings.Fields(hotty.ParseKeymap(keys).Format())) != 6 {
		t.Errorf("the keys %q do not all parse", keys)
	}
	if got := s.TextOf("pager"); !strings.Contains(got, "Cherry") || strings.Contains(got, "Date") {
		t.Errorf("page 1 reads %q", got)
	}
	if got := s.TextOf(partID("results", partDots)); got != "3/12" {
		t.Errorf("the results' numbers %q", got)
	}

	must(t, x.h.Click(r.name, dot("2")))
	x.pump()
	if got := r.C.S.Data.Value("/page"); got != 3.0 || r.C.St.Focus != "pager" || !r.C.St.Keyboard || s.Focused() != "pager" {
		t.Fatalf("a click on the third dot: /page %v, the program's focus %q (%v), the host's %q", got, r.C.St.Focus, r.C.St.Keyboard, s.Focused())
	}
	if got := s.TextOf("pager"); !strings.Contains(got, "Grape") || strings.Contains(got, "Apple") {
		t.Errorf("page 3 reads %q", got)
	}
	if v, _ := s.Attr(dot("2"), "class"); v != "k-dot k-on" {
		t.Errorf("the third dot's class %q", v)
	}
	for _, step := range []struct {
		key  string
		page float64
	}{{"ArrowLeft", 2}, {"Home", 1}, {"l", 2}, {"End", 3}, {"h", 2}, {"ArrowRight", 3}} {
		// The host leaves the key to the program, as the program reads it.
		if x.h.Key(step.key) {
			t.Fatalf("the host used %s", step.key)
		}
		if _, ok, err := r.Key(step.key); !ok || err != nil {
			t.Fatalf("%s: %v %v", step.key, ok, err)
		}
		x.update(r)
		if got := r.C.S.Data.Value("/page"); got != step.page {
			t.Errorf("%s: /page %v, want %v", step.key, got, step.page)
		}
	}
	if len(x.actions) != 7 || x.actions[6].Name != "turned" || x.actions[6].Context["page"] != 3.0 {
		t.Errorf("%d actions, the last %+v", len(x.actions), x.actions[len(x.actions)-1])
	}
	if v, _ := s.Attr(partID("pager", partDots), "aria-label"); v != "Page 3 of 3" {
		t.Errorf("the dots say %q", v)
	}
	x.check(r)
}
