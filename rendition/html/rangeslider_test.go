package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// rangesOnHost is a Price HottyRangeSlider from 0 to 100 in steps of 5,
// bound at 20 to 70, a Rating Slider that fills from its end, and a
// disabled HottyRangeSlider, on a host with steps (SPEC §9.1) or without.
func rangesOnHost(t *testing.T, steps bool) (*harness, *Rendition) {
	t.Helper()
	var opts []hottytest.Option
	if steps {
		caps := hottytest.DefaultCaps()
		caps.Steps = true
		opts = append(opts, hottytest.Caps(caps))
	}
	x := newHarness(t, opts...)
	h := `"catalogId":"` + hottycat.ID + `"`
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"lo":20,"hi":70}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["price","rating","year"]},
 {"id":"price","component":"HottyRangeSlider",`+h+`,"label":"Price","start":{"@path":"/lo"},"end":{"@path":"/hi"},"max":100,"steps":20},
 {"id":"rating","component":"Slider","label":"Rating","value":3,"min":1,"max":5,
  "metadata":{"extensions":{"io_neuroplast_hotty":{"fill":"end"}}}},
 {"id":"year","component":"HottyRangeSlider",`+h+`,"label":"Year","start":2,"end":4,"max":10,"disabled":true}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	if steps {
		r.SetSteps(true)
		x.send(hotty.Doc(r.name, r.Doc()))
	}
	return x, r
}

// A HottyRangeSlider on a host is a group: its label, a track holding the
// fill between its knobs and its notches, its two knobs, each a button
// role=slider whose range is up to the other, and its value. A Slider
// with fill "end" fills from its knob to max. A disabled one's knobs are
// disabled buttons, its notches take nothing, and it is faint throughout.
func TestRangeSliderMarkup(t *testing.T) {
	x, r := rangesOnHost(t, false)
	s := x.h.Surface(r.name)
	k0, k1 := DOMID("price/knob/0"), DOMID("price/knob/1")
	for _, a := range []struct{ id, attr, want string }{
		{"price", "role", "group"},
		{"price", "aria-labelledby", partID("price", partLabel)},
		{k0, "role", "slider"},
		{k0, "aria-label", "Price, start"},
		{k0, "aria-valuemin", "0"},
		{k0, "aria-valuemax", "70"},
		{k0, "aria-valuenow", "20"},
		{k0, "style", "left: 20.00%"},
		{k0, "data-on", "drag"},
		{k1, "aria-label", "Price, end"},
		{k1, "aria-valuemin", "20"},
		{k1, "aria-valuemax", "100"},
		{k1, "aria-valuenow", "70"},
		{partID("price", partNotch+"4"), "data-on", "drag click"},
		{partID("price", partOutput), "for", k0 + " " + k1},
		{DOMID("year/knob/0"), "aria-valuenow", "2"},
	} {
		if got, _ := s.Attr(a.id, a.attr); got != a.want {
			t.Errorf("%s %s: %q, want %q", a.id, a.attr, got, a.want)
		}
	}
	if got := s.TextOf(partID("price", partOutput)); got != "20–70" {
		t.Errorf("price's value: %q", got)
	}
	if got := s.TextOf(partID("price", partLabel)); got != "Price" {
		t.Errorf("price's label: %q", got)
	}
	html := s.HTML()
	for _, want := range []string{
		`class="k-fill" style="left: 20.00%; width: 50.00%"`,
		`class="k-fill" style="left: 50.00%; width: 50.00%"`, // rating at 3 of 1–5
	} {
		if !strings.Contains(html, want) {
			t.Errorf("no %s", want)
		}
	}
	if _, ok := s.Attr(DOMID("year/knob/1"), "disabled"); !ok {
		t.Error("the disabled range's end is no disabled button")
	}
	if on, ok := s.Attr(partID("year", partNotch+"4"), "data-on"); ok {
		t.Errorf("the disabled range's notch takes %q", on)
	}
	if !strings.Contains(html, `class="k-field k-off"`) {
		t.Error("the disabled range is not faint")
	}
}

// The keyboard on the start's knob, Tab goes on to the end's; the arrows,
// which a focused button leaves to the program (SPEC §10.2), move a knob,
// stopped where it meets the other. A tap on the track moves the nearer
// knob, which takes the keyboard back from the terminal, where the tap
// left it.
func TestRangeSliderKeysOnHost(t *testing.T) {
	x, r := rangesOnHost(t, false)
	s := x.h.Surface(r.name)
	data := func(p string) any { return r.C.S.Data.Value(p) }
	key := func(k string) {
		t.Helper()
		if !x.h.Key(k) {
			cmds, _, err := r.Key(k)
			must(t, err)
			x.send(cmds...)
		}
		x.pump()
		x.update(r)
	}
	r.C.Focus("price/knob/0")
	x.update(r)
	if s.Focused() != DOMID("price/knob/0") {
		t.Fatalf("the program's focus: the host's %q", s.Focused())
	}
	key("ArrowRight")
	if data("/lo") != 25.0 {
		t.Errorf("ArrowRight on the start: %v", data("/lo"))
	}
	if v, _ := s.Attr(DOMID("price/knob/0"), "aria-valuenow"); v != "25" {
		t.Errorf("the start's aria-valuenow %q", v)
	}
	if v, _ := s.Attr(DOMID("price/knob/1"), "aria-valuemin"); v != "25" {
		t.Errorf("the end's aria-valuemin %q", v)
	}
	key("End")
	if data("/lo") != 70.0 {
		t.Errorf("End on the start stops at the end: %v", data("/lo"))
	}
	key("Tab")
	key("Home")
	if data("/hi") != 70.0 || s.Focused() != DOMID("price/knob/1") {
		t.Errorf("Home on the end stops at the start: %v, focus %q", data("/hi"), s.Focused())
	}
	key("End")
	x.check(r)

	// A tap on notch 2 (10): the start's is nearer.
	must(t, x.h.Click(r.name, partID("price", partNotch+"2")))
	x.pump()
	if data("/lo") != 10.0 || s.Focused() != DOMID("price/knob/0") || !r.C.St.Keyboard {
		t.Errorf("a tap at 10: %v, the host's focus %q", data("/lo"), s.Focused())
	}
	// A click on the end's knob gives it the keyboard where it is.
	must(t, x.h.Click(r.name, DOMID("price/knob/1")))
	x.pump()
	if data("/hi") != 100.0 || r.C.St.Focus != "price/knob/1" || s.Focused() != DOMID("price/knob/1") {
		t.Errorf("a click on the end: %v, focus %q", data("/hi"), r.C.St.Focus)
	}
	x.check(r)
}

// Without steps, a knob is the drag target, and the notches it crosses say
// where it goes; the click it ends with on itself moves nothing.
func TestRangeSliderDragOnHost(t *testing.T) {
	x, r := rangesOnHost(t, false)
	s := x.h.Surface(r.name)
	data := func(p string) any { return r.C.S.Data.Value(p) }
	must(t, x.h.DragStart(r.name, DOMID("price/knob/1"), 30, 0))
	x.pump()
	if r.C.St.Focus != "price/knob/1" || s.Focused() != DOMID("price/knob/1") {
		t.Errorf("pressed on the end: focus %q, the host's %q", r.C.St.Focus, s.Focused())
	}
	must(t, x.h.DragMove(partID("price", partNotch+"18"), 40, 0))
	x.pump()
	if data("/hi") != 90.0 {
		t.Errorf("dragged onto notch 18: %v", data("/hi"))
	}
	must(t, x.h.DragMove(partID("price", partNotch+"1"), 10, 0))
	x.pump()
	if data("/hi") != 20.0 || data("/lo") != 20.0 {
		t.Errorf("dragged past the start: %v–%v", data("/lo"), data("/hi"))
	}
	must(t, x.h.DragMove(DOMID("price/knob/1"), 12, 0))
	must(t, x.h.DragEnd(DOMID("price/knob/1"), 12, 0))
	x.pump()
	if data("/hi") != 20.0 || x.h.Dragging() {
		t.Errorf("let go on the knob: %v", data("/hi"))
	}
	x.check(r)
}

// With steps (SPEC §9.1), the track is the drag target: a press moves the
// nearer knob to the step under the pointer and gives it the keyboard,
// which the host's blur after the press (the track takes no focus) does
// not take away; each step after moves it. Where the knobs meet, the first
// move picks the one it goes towards. A tap on a notch moves the nearer
// knob there.
func TestRangeSliderStepsOnHost(t *testing.T) {
	x, r := rangesOnHost(t, true)
	s := x.h.Surface(r.name)
	data := func(p string) any { return r.C.S.Data.Value(p) }
	for _, a := range []struct{ id, attr, want string }{
		{"price", "data-on", "drag"},
		{"price", "data-steps", "20"},
		{partID("price", partNotch+"4"), "data-on", "click"},
	} {
		if got, _ := s.Attr(a.id, a.attr); got != a.want {
			t.Errorf("%s %s: %q, want %q", a.id, a.attr, got, a.want)
		}
	}
	if on, ok := s.Attr(DOMID("price/knob/0"), "data-on"); ok {
		t.Errorf("a knob takes %q on a host with steps", on)
	}
	// The keyboard is on the rating first, so that the press's blur shows.
	must(t, x.h.Click(r.name, DOMID("rating")))
	x.pump()
	must(t, x.h.DragStartStep(r.name, partID("price", partNotch+"17"), 40, 0, 17, 0))
	x.pump()
	if data("/hi") != 85.0 || r.C.St.Focus != "price/knob/1" || !r.C.St.Keyboard || s.Focused() != DOMID("price/knob/1") {
		t.Errorf("pressed at step 17: %v, focus %q (keyboard %v), the host's %q", data("/hi"), r.C.St.Focus, r.C.St.Keyboard, s.Focused())
	}
	for _, step := range []struct {
		what   string
		id     string
		x      int
		lo, hi float64
	}{
		{"off the track at step 20", "", 20, 20, 100},
		{"over the start at step 2", DOMID("price/knob/0"), 2, 20, 20},
		{"over the rating at step 4", DOMID("rating"), 4, 20, 20},
	} {
		must(t, x.h.DragMoveStep(step.id, 30, 0, step.x, 0))
		x.pump()
		if data("/lo") != step.lo || data("/hi") != step.hi {
			t.Errorf("%s: %v–%v, want %v–%v", step.what, data("/lo"), data("/hi"), step.lo, step.hi)
		}
	}
	must(t, x.h.DragEndStep("", 30, 0, 4, 0))
	x.pump()
	// They meet at 20: a press there moves neither; the move picks.
	must(t, x.h.DragStartStep(r.name, DOMID("price/knob/1"), 10, 0, 4, 0))
	x.pump()
	must(t, x.h.DragMoveStep("", 10, 0, 1, 0))
	x.pump()
	if data("/lo") != 5.0 || data("/hi") != 20.0 || r.C.St.Focus != "price/knob/0" || s.Focused() != DOMID("price/knob/0") {
		t.Errorf("from where they meet, down: %v–%v, focus %q, the host's %q", data("/lo"), data("/hi"), r.C.St.Focus, s.Focused())
	}
	// Let go on the track; the tap after it, its press a blur, is a tap.
	must(t, x.h.DragEndStep(partID("price", partNotch+"1"), 10, 0, 1, 0))
	x.pump()
	must(t, x.h.Click(r.name, partID("price", partNotch+"16")))
	x.pump()
	if data("/hi") != 80.0 || s.Focused() != DOMID("price/knob/1") {
		t.Errorf("a tap at 80 after a drag: %v, the host's focus %q", data("/hi"), s.Focused())
	}
	x.check(r)
}
