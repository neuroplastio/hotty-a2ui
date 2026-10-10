package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottySwitch flips on activation: its bound value in the data model,
// an unbound one in the renderer's state. Its disabled, bound here, keeps
// it from flipping and from the keyboard, and its check's error shows once
// it is touched.
func TestSwitch(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"on":false,"off":false}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["bound","free"]},
	 {"id":"bound","component":"HottySwitch",` + h + `,"label":"Bound","value":{"@path":"/on"},"disabled":{"@path":"/off"},
	  "checks":[{"condition":{"@path":"/on"},"message":"Turn it on"}]},
	 {"id":"free","component":"HottySwitch",` + h + `,"value":true}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	e := c.V.Find("bound")
	if e.Kind != view.Switch || e.Label != "Bound" || e.On() || e.Disabled || e.Error != "" || !e.Focusable() {
		t.Fatalf("bound: %+v", e)
	}
	if err := c.Activate("bound"); err != nil {
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/on"); got != true {
		t.Errorf("/on is %v after a flip", got)
	}
	if err := c.Activate("bound"); err != nil {
		t.Fatal(err)
	}
	if e := c.V.Find("bound"); e.Error != "Turn it on" {
		t.Errorf("touched and off, its error is %q", e.Error)
	}
	if err := c.Activate("free"); err != nil {
		t.Fatal(err)
	}
	if c.V.Find("free").On() || c.St.Local["free"] != false {
		t.Errorf("an unbound switch did not flip in the renderer's state: %v", c.St.Local["free"])
	}
	if err := c.S.Write("/off", true); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if err := c.Activate("bound"); err != nil {
		t.Fatal(err)
	}
	if e := c.V.Find("bound"); !e.Disabled || e.Focusable() || e.On() {
		t.Errorf("disabled: %+v", e)
	}
	if hints, _ := c.KeyHints("", false); len(hints) != 0 {
		t.Errorf("hints without the keyboard: %v", hints)
	}
	c.Focus("free")
	if hints, _ := c.KeyHints("", false); len(hints) == 0 || hints[0] != (view.Hint{Key: "space/enter", Desc: "toggle"}) {
		t.Errorf("a switch's hints: %v", hints)
	}
}
