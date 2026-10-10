package view_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// timerRun is a run of a surface of timers, the actions the agent got,
// and a clock at its start.
func timerRun(t *testing.T, data, comps string) (*story.Run, *view.Controller, *[]*a2ui.ActionMessage, time.Time) {
	t.Helper()
	run := story.NewRun()
	var got []*a2ui.ActionMessage
	run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			got = append(got, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":` + data + `}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[` + comps + `]}}]`
	if err := run.Feed(json.RawMessage(msgs)); err != nil {
		t.Fatal(err)
	}
	return run, run.Surfaces()[0].C, &got, time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
}

// face is a timer's element as the view has it: its time, and whether it
// counts.
func face(t *testing.T, c *view.Controller, id string) (string, bool) {
	t.Helper()
	e := c.V.Find(id)
	if e == nil || e.Kind != view.Timer {
		t.Fatalf("%s: %+v", id, e)
	}
	s, _ := e.Value.(string)
	return s, e.Active
}

const h = `"catalogId":"` + hotty.ID + `"`

// A HottyTimer counts down from its first draw, its time left rounded up
// to its interval, and a HottyStopwatch up, rounded down; each draw asks
// for the next within a TimerStep. When the timer runs out it shows 0,
// stops, writes false where its running is bound, and runs its onTimeout
// once; started again, it starts over.
func TestTimerCounts(t *testing.T) {
	_, c, got, t0 := timerRun(t, `{"tea":true}`, `
	 {"id":"root","component":"Column","children":["tea","lap"]},
	 {"id":"tea","component":"HottyTimer",`+h+`,"label":"Tea","duration":3000,"running":{"@path":"/tea"},
	  "onTimeout":{"event":{"name":"ready","context":{"running":{"@path":"/tea"}}}}},
	 {"id":"lap","component":"HottyStopwatch",`+h+`,"interval":100}`)
	if e := c.V.Find("tea"); e.Variant != view.TimerCountdown || e.Label != "Tea" || e.Focusable() {
		t.Fatalf("tea: %+v", e)
	}
	if e := c.V.Find("lap"); e.Variant != view.TimerStopwatch {
		t.Fatalf("lap: %+v", e)
	}
	for _, at := range []struct {
		ms        int64
		tea, lap  string
		teaCounts bool
		step      time.Duration
	}{
		{0, "3s", "0s", true, view.TimerStep(100 * time.Millisecond)},
		{1, "3s", "0s", true, 100 * time.Millisecond},
		{999, "3s", "900ms", true, 100 * time.Millisecond},
		{1000, "2s", "1s", true, 100 * time.Millisecond},
		{1550, "2s", "1.5s", true, 100 * time.Millisecond},
		{2999, "1s", "2.9s", true, 100 * time.Millisecond},
		{3000, "0s", "3s", false, 100 * time.Millisecond},
		{4000, "0s", "4s", false, 100 * time.Millisecond},
	} {
		step := c.TickTimers(t0.Add(time.Duration(at.ms) * time.Millisecond))
		tea, counts := face(t, c, "tea")
		lap, _ := face(t, c, "lap")
		if tea != at.tea || counts != at.teaCounts || lap != at.lap || step != at.step {
			t.Errorf("at %dms: tea %q (counts %v), lap %q, step %v; want %q (%v), %q, %v", at.ms, tea, counts, lap, step, at.tea, at.teaCounts, at.lap, at.step)
		}
	}
	if len(*got) != 1 || (*got)[0].Name != "ready" || (*got)[0].SourceComponentID != "tea" || (*got)[0].Context["running"] != false {
		t.Fatalf("onTimeout sent %+v, want one ready, after running went false", *got)
	}
	if c.S.Data.Value("/tea") != false {
		t.Errorf("/tea is %v once the timer ran out", c.S.Data.Value("/tea"))
	}

	// Started again, from the agent or a Button, it starts over.
	if err := c.S.Write("/tea", true); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	c.TickTimers(t0.Add(5 * time.Second))
	if tea, counts := face(t, c, "tea"); tea != "3s" || !counts {
		t.Errorf("started once run out: %q (counts %v), want 3s", tea, counts)
	}
	c.TickTimers(t0.Add(6500 * time.Millisecond))
	if tea, _ := face(t, c, "tea"); tea != "2s" {
		t.Errorf("1.5s after it started over: %q", tea)
	}
	c.TickTimers(t0.Add(8 * time.Second))
	if len(*got) != 2 {
		t.Errorf("run out twice, onTimeout ran %d times", len(*got))
	}
	// Once none counts, it asks for no more draws.
	if err := c.TimerDo("lap", view.TimerStop); err != nil {
		t.Fatal(err)
	}
	if step := c.TickTimers(t0.Add(9 * time.Second)); step != 0 {
		t.Errorf("nothing counts, and it asks for a draw in %v", step)
	}
}

// hottyStartTimer, hottyStopTimer, hottyToggleTimer and hottyResetTimer,
// from Buttons: stop holds the time (the draw after counts up to itself),
// start counts on from there, toggle does the other, and reset takes it
// back to 0, running as it was. An unbound running is the renderer's.
func TestTimerFunctions(t *testing.T) {
	call := func(fn string) string {
		return `{"functionCall":{"@call":"` + fn + `",` + h + `,"args":{"id":"lap"}}}`
	}
	_, c, _, t0 := timerRun(t, `{}`, `
	 {"id":"root","component":"Column","children":["lap","start","stop","toggle","reset"]},
	 {"id":"lap","component":"HottyStopwatch",`+h+`},
	 {"id":"start","component":"Button","child":"t","action":`+call("hottyStartTimer")+`},
	 {"id":"stop","component":"Button","child":"t","action":`+call("hottyStopTimer")+`},
	 {"id":"toggle","component":"Button","child":"t","action":`+call("hottyToggleTimer")+`},
	 {"id":"reset","component":"Button","child":"t","action":`+call("hottyResetTimer")+`},
	 {"id":"t","component":"Text","text":"Go"}`)
	at := func(s float64) time.Time { return t0.Add(time.Duration(s * float64(time.Second))) }
	press := func(id string) {
		t.Helper()
		if err := c.Activate(c.FindComponent(id, a2ui.RootScope)); err != nil {
			t.Fatal(err)
		}
	}
	c.TickTimers(at(0))
	c.TickTimers(at(2.5))
	press("stop")
	if c.St.Local["lap"] != false {
		t.Errorf("an unbound running, stopped: %v", c.St.Local["lap"])
	}
	c.TickTimers(at(3))
	c.TickTimers(at(10))
	if lap, counts := face(t, c, "lap"); lap != "3s" || counts {
		t.Errorf("stopped at 3s, 7s later: %q (counts %v)", lap, counts)
	}
	press("toggle")
	c.TickTimers(at(11))
	c.TickTimers(at(12.2))
	if lap, counts := face(t, c, "lap"); lap != "4s" || !counts {
		t.Errorf("toggled on at 11s, at 12.2s: %q (counts %v), want 4s", lap, counts)
	}
	press("reset")
	if lap, _ := face(t, c, "lap"); lap != "0s" {
		t.Errorf("reset: %q", lap)
	}
	c.TickTimers(at(13))
	c.TickTimers(at(14.5))
	if lap, counts := face(t, c, "lap"); lap != "1s" || !counts {
		t.Errorf("reset at 12.2s, running, at 14.5s: %q (counts %v), want 1s from the draw at 13s", lap, counts)
	}
	press("toggle")
	press("reset")
	c.TickTimers(at(20))
	if lap, counts := face(t, c, "lap"); lap != "0s" || counts {
		t.Errorf("stopped and reset: %q (counts %v)", lap, counts)
	}
	press("start")
	c.TickTimers(at(21))
	c.TickTimers(at(23))
	if lap, _ := face(t, c, "lap"); lap != "2s" {
		t.Errorf("started at 21s, at 23s: %q", lap)
	}
	if err := c.TimerDo("lap", "pause"); err == nil {
		t.Error("a timer paused")
	}
	if err := c.TimerDo("start", view.TimerStart); err == nil {
		t.Error("a Button started as a timer")
	}
}

// A HottyStopwatch's time is written where its elapsed is bound, in whole
// milliseconds, as the last draw counted it: before an action runs, so a
// lap's context reads the time the user sees, when it stops, and when it
// is reset; not at a draw while it counts. One not bound writes nothing.
func TestStopwatchElapsed(t *testing.T) {
	call := func(fn string) string {
		return `{"functionCall":{"@call":"` + fn + `",` + h + `,"args":{"id":"lap"}}}`
	}
	_, c, got, t0 := timerRun(t, `{}`, `
	 {"id":"root","component":"Column","children":["lap","plain","mark","stop","reset"]},
	 {"id":"lap","component":"HottyStopwatch",`+h+`,"interval":100,"elapsed":{"@path":"/ms"}},
	 {"id":"plain","component":"HottyStopwatch",`+h+`},
	 {"id":"mark","component":"Button","child":"t","action":{"event":{"name":"lap","context":{"ms":{"@path":"/ms"}}}}},
	 {"id":"stop","component":"Button","child":"t","action":`+call("hottyStopTimer")+`},
	 {"id":"reset","component":"Button","child":"t","action":`+call("hottyResetTimer")+`},
	 {"id":"t","component":"Text","text":"Go"}`)
	at := func(s float64) time.Time { return t0.Add(time.Duration(s * float64(time.Second))) }
	press := func(id string) {
		t.Helper()
		if err := c.Activate(c.FindComponent(id, a2ui.RootScope)); err != nil {
			t.Fatal(err)
		}
	}
	ms := func() any { v, _, _ := c.S.Data.Get("/ms"); return v }
	c.TickTimers(at(0))
	c.TickTimers(at(1.25))
	if v := ms(); v != nil {
		t.Errorf("written at a draw while it counts: %v", v)
	}
	press("mark")
	if len(*got) != 1 || (*got)[0].Context["ms"] != 1250.0 {
		t.Fatalf("the lap's context: %+v, want ms 1250", *got)
	}
	press("stop")
	c.TickTimers(at(2.5))
	if v := ms(); v != 2500.0 {
		t.Errorf("stopped at the draw at 2.5s: %v, want 2500", v)
	}
	press("reset")
	if v := ms(); v != 0.0 {
		t.Errorf("reset: %v, want 0", v)
	}
	if v, ok, _ := c.S.Data.Get("/plain"); ok {
		t.Errorf("a stopwatch with no elapsed wrote %v", v)
	}
}

// A timer the view does not show, in a tab not shown, counts all the same,
// runs out on time, and the agent's hottyResetTimer finds it; one whose
// duration is a path that holds nothing does not count, nor run out; a
// timer removed loses its time.
func TestTimerUnseen(t *testing.T) {
	run, c, got, t0 := timerRun(t, `{}`, `
	 {"id":"root","component":"Tabs","tabs":[{"title":"One","child":"one"},{"title":"Two","child":"egg"}]},
	 {"id":"one","component":"HottyTimer",`+h+`,"duration":{"@path":"/later"},"onTimeout":{"event":{"name":"never"}}},
	 {"id":"egg","component":"HottyTimer",`+h+`,"label":"Egg","duration":2000,"format":"clock","onTimeout":{"event":{"name":"boiled"}}}`)
	if c.V.Find("egg") != nil {
		t.Fatal("the second tab shows")
	}
	if step := c.TickTimers(t0); step != view.TimerStep(time.Second) {
		t.Errorf("an unseen timer counts, and asks for a draw in %v", step)
	}
	if one, counts := face(t, c, "one"); one != "0s" || counts {
		t.Errorf("no duration yet: %q (counts %v)", one, counts)
	}
	c.TickTimers(t0.Add(2 * time.Second))
	if len(*got) != 1 || (*got)[0].Name != "boiled" {
		t.Fatalf("the unseen timer's onTimeout: %+v", *got)
	}
	if c.St.Local["egg"] != false {
		t.Errorf("run out, an unbound running is %v", c.St.Local["egg"])
	}
	m := `{"version":"v1.0","callRendererFunction":{"functionCallId":"c","callFunction":{"@call":"hottyResetTimer",` + h + `,"args":{"id":"egg"}}}}`
	if err := run.Feed(json.RawMessage(m)); err != nil {
		t.Fatal(err)
	}
	c.St.Tabs["root"] = 1
	c.Rebuild()
	if egg, counts := face(t, c, "egg"); egg != "0:02" || counts {
		t.Errorf("reset by the agent, stopped: %q (counts %v)", egg, counts)
	}
	if len(*got) != 1 {
		t.Errorf("never ran out, and sent %+v", *got)
	}
	upd := `{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Tabs","tabs":[{"title":"One","child":"one"}]}]}}`
	if err := run.Feed(json.RawMessage(upd)); err != nil {
		t.Fatal(err)
	}
	c.TickTimers(t0.Add(3 * time.Second))
	if _, ok := c.St.Timers["egg"]; ok {
		t.Error("a timer gone keeps its time")
	}
}

// FormatTimer writes a time as Go writes a duration, bubbles' way, or as a
// clock does, to the interval's fraction of a second.
func TestFormatTimer(t *testing.T) {
	ms := time.Millisecond
	for _, tc := range []struct {
		d, interval time.Duration
		clock       bool
		want        string
	}{
		{0, time.Second, false, "0s"},
		{90 * time.Second, time.Second, false, "1m30s"},
		{4500 * ms, 100 * ms, false, "4.5s"},
		{300 * ms, 100 * ms, false, "300ms"},
		{3723 * time.Second, time.Second, false, "1h2m3s"},
		{0, time.Second, true, "0:00"},
		{90 * time.Second, time.Second, true, "1:30"},
		{25 * time.Minute, time.Minute, true, "25:00"},
		{3723 * time.Second, time.Second, true, "1:02:03"},
		{4500 * ms, 100 * ms, true, "0:04.5"},
		{4560 * ms, 10 * ms, true, "0:04.56"},
		{4567 * ms, ms, true, "0:04.567"},
		{4500 * ms, 250 * ms, true, "0:04.50"},
		{4500 * ms, 1500 * ms, true, "0:04.5"},
	} {
		if got := view.FormatTimer(tc.d, tc.interval, tc.clock); got != tc.want {
			t.Errorf("%v in %v, clock %v: %q, want %q", tc.d, tc.interval, tc.clock, got, tc.want)
		}
	}
	for interval, want := range map[time.Duration]time.Duration{
		time.Second: time.Second / 10, time.Minute: time.Second / 10, 50 * ms: 50 * ms, ms: time.Second / 30,
	} {
		if got := view.TimerStep(interval); got != want {
			t.Errorf("TimerStep(%v) is %v, want %v", interval, got, want)
		}
	}
}
