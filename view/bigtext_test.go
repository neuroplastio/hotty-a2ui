package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyBigText is its text as it is, its size (medium without one, or
// with one the catalog lacks) and its align (start likewise); bound, its
// text follows the data model. It takes no keyboard.
func TestBigText(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"score":"3 : 1"}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["title","score","odd"]},
	 {"id":"title","component":"HottyBigText",` + h + `,"text":"Hello\nworld","size":"large","align":"center"},
	 {"id":"score","component":"HottyBigText",` + h + `,"text":{"@path":"/score"}},
	 {"id":"odd","component":"HottyBigText",` + h + `,"text":"x","size":"huge","align":"middle"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for _, want := range []view.Element{
		{ID: "title", Label: "Hello\nworld", Variant: view.BigLarge, Align: "center"},
		{ID: "score", Label: "3 : 1", Variant: view.BigMedium, Align: "start"},
		{ID: "odd", Label: "x", Variant: view.BigMedium, Align: "start"},
	} {
		e := c.V.Find(want.ID)
		if e.Kind != view.BigText || e.Label != want.Label || e.Variant != want.Variant || e.Align != want.Align || e.Focusable() {
			t.Errorf("%s: %+v", want.ID, e)
		}
	}
	if err := c.S.Write("/score", "4 : 1"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if got := c.V.Find("score").Label; got != "4 : 1" {
		t.Errorf("the score is %q once written", got)
	}
	if got := c.V.Focusables(); len(got) != 0 {
		t.Errorf("focusables %v", got)
	}
}
