package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// ranges is a bound HottyRangeSlider, Price from 0 to 100 in steps of 5,
// at 20 to 70; one from 10 to 20 whose start is bound to a path with no
// value and whose end is a literal 18; and a Slider that fills from its
// end.
func ranges(t *testing.T) *view.Controller {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"lo":20,"hi":70,"off":false}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["price","free","rating"]},
	 {"id":"price","component":"HottyRangeSlider",` + h + `,"label":"Price","start":{"@path":"/lo"},"end":{"@path":"/hi"},"max":100,"steps":20,
	  "disabled":{"@path":"/off"},"checks":[{"condition":{"@path":"/off"},"message":"Not yet"}]},
	 {"id":"free","component":"HottyRangeSlider",` + h + `,"start":{"@path":"/none"},"end":18,"min":10,"max":20,"accessibility":{"label":"Year"}},
	 {"id":"rating","component":"Slider","label":"Rating","value":3,"min":1,"max":5,
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"fill":"end"}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s"))
}

// A HottyRangeSlider is a RangeSlider with two knobs, each a Tab stop
// named for a screen reader; an absent end is the range's own end. A
// Slider's fill extension says which side of its knob is filled.
func TestRangeSliderMaps(t *testing.T) {
	c := ranges(t)
	e := c.V.Find("price")
	if e.Kind != view.RangeSlider || e.Label != "Price" || e.Min != 0 || e.Max != 100 || e.Step != 5 || e.Focusable() {
		t.Fatalf("price: %+v", e)
	}
	if lo, hi := e.Range(); lo != 20 || hi != 70 || e.RangeText() != "20–70" || e.RangeWidth() != 7 {
		t.Errorf("price: %v–%v %q %d", lo, hi, e.RangeText(), e.RangeWidth())
	}
	k := c.V.Find("price/knob/1")
	if k == nil || k.Kind != view.Knob || k.Selected != 1 || k.Label != "Price, end" || k.Value != 70.0 || !k.Focusable() {
		t.Errorf("price's end: %+v", k)
	}
	if lo, hi := c.V.Find("free").Range(); lo != 10 || hi != 18 {
		t.Errorf("free: %v–%v, want 10–18", lo, hi)
	}
	if k := c.V.Find("free/knob/0"); k.Label != "Year, start" {
		t.Errorf("free's start is named %q", k.Label)
	}
	want := []string{"price/knob/0", "price/knob/1", "free/knob/0", "free/knob/1", "rating"}
	if got := c.V.Focusables(); len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[4] != want[4] {
		t.Errorf("focusables %v, want %v", got, want)
	}
	if f := c.V.Find("rating").Fill; f != "end" {
		t.Errorf("rating's fill %q", f)
	}
}

// The arrows move a knob a step and Home and End to the range's ends,
// each stopped where it meets the other knob; its value goes where its
// end is bound, or to the renderer's state.
func TestRangeSliderSteps(t *testing.T) {
	c := ranges(t)
	data := func(p string) any { return c.S.Data.Value(p) }
	for _, s := range []struct {
		knob   string
		n      int
		to     string
		lo, hi float64
	}{
		{"price/knob/0", 1, "", 25, 70},
		{"price/knob/1", 0, "Home", 25, 25},
		{"price/knob/1", -1, "", 25, 25},
		{"price/knob/0", 1, "", 25, 25},
		{"price/knob/0", 0, "End", 25, 25},
		{"price/knob/0", -2, "", 15, 25},
		{"price/knob/1", 0, "End", 15, 100},
		{"price/knob/0", 0, "Home", 0, 100},
	} {
		if err := c.StepSlider(s.knob, s.n, s.to); err != nil {
			t.Fatal(err)
		}
		if data("/lo") != s.lo || data("/hi") != s.hi {
			t.Errorf("%s %d %s: %v–%v, want %v–%v", s.knob, s.n, s.to, data("/lo"), data("/hi"), s.lo, s.hi)
		}
	}
	if c.V.Find("price").Error != "Not yet" {
		t.Errorf("touched, its check's error is %q", c.V.Find("price").Error)
	}
	if err := c.StepSlider("free/knob/1", -1, ""); err != nil {
		t.Fatal(err)
	}
	if lo, hi := c.V.Find("free").Range(); lo != 10 || hi != 17.5 || c.St.Local["free/knob/1"] != 17.5 {
		t.Errorf("free after a step down: %v–%v", lo, hi)
	}
	if err := c.StepSlider("free/knob/0", 1, ""); err != nil {
		t.Fatal(err)
	}
	if data("/none") != 10.5 {
		t.Errorf("free's start, bound, after a step up: %v", data("/none"))
	}
	if err := c.S.Write("/off", true); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if err := c.StepSlider("price/knob/0", 1, ""); err != nil {
		t.Fatal(err)
	}
	if k := c.V.Find("price/knob/0"); data("/lo") != 0.0 || k.Focusable() || !k.Disabled {
		t.Errorf("disabled: %v, %+v", data("/lo"), k)
	}
}

// A press on the track moves the nearer knob there, which takes the
// keyboard; midway, the start's. Where the knobs meet, the press moves
// neither, and a drag's first move picks the one it goes towards.
func TestRangeSliderPress(t *testing.T) {
	c := ranges(t)
	data := func(p string) any { return c.S.Data.Value(p) }
	k, err := c.PressKnob("price", 88)
	if err != nil || k != 1 || data("/hi") != 90.0 || c.St.Focus != "price/knob/1" || !c.St.Keyboard {
		t.Fatalf("a press at 88: knob %d, %v, focus %q (%v)", k, data("/hi"), c.St.Focus, err)
	}
	if k, _ := c.PressKnob("price", 55); k != 0 || data("/lo") != 55.0 || c.St.Focus != "price/knob/0" {
		t.Errorf("a press midway at 55: knob %d, %v, focus %q", k, data("/lo"), c.St.Focus)
	}
	if k, _ = c.DragKnob("price", 0, 95); k != 0 || data("/lo") != 90.0 {
		t.Errorf("a drag of the start past the end: knob %d, %v", k, data("/lo"))
	}
	// They meet at 90: a press there moves neither, and the end's takes
	// the keyboard, as it can move away.
	if k, _ = c.PressKnob("price", 90); k != -1 || c.St.Focus != "price/knob/1" {
		t.Errorf("a press where they meet: knob %d, focus %q", k, c.St.Focus)
	}
	if k, _ = c.DragKnob("price", -1, 90); k != -1 {
		t.Errorf("no move yet: knob %d", k)
	}
	if k, _ = c.DragKnob("price", -1, 80); k != 0 || data("/lo") != 80.0 || data("/hi") != 90.0 || c.St.Focus != "price/knob/0" {
		t.Errorf("a drag down from where they meet: knob %d, %v–%v, focus %q", k, data("/lo"), data("/hi"), c.St.Focus)
	}
	if err := c.StepSlider("price/knob/0", 0, "End"); err != nil {
		t.Fatal(err)
	}
	if k, _ = c.DragKnob("price", -1, 100); k != 1 || data("/hi") != 100.0 {
		t.Errorf("a drag up from where they meet: knob %d, %v", k, data("/hi"))
	}
	// Meeting at the max, the start's takes the keyboard.
	if err := c.StepSlider("price/knob/0", 0, "End"); err != nil {
		t.Fatal(err)
	}
	if k, _ = c.PressKnob("price", 100); k != -1 || c.St.Focus != "price/knob/0" {
		t.Errorf("a press where they meet at the max: knob %d, focus %q", k, c.St.Focus)
	}
	c.Focus("price/knob/0")
	if hints, _ := c.KeyHints("", true); len(hints) == 0 || hints[0] != (view.Hint{Key: "←/→", Desc: "adjust"}) {
		t.Errorf("a knob's hints on a host: %v", hints)
	}
}
