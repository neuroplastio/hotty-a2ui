package html

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// TestTimer: a HottyTimer or a HottyStopwatch is a timer (role=timer):
// its label, then its time, whose own id makes a tick one text's delta,
// and which is k-still, muted, while it does not count. The time moves on
// the view's clock, as in cells: a delta when it shows another, none in
// between. The rendition asks for an Update every TimerStep while one
// counts, and none once none does. A timer that runs out sends its
// onTimeout.
func TestTimer(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"lap":true}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["tea","lap"]},
 {"id":"tea","component":"HottyTimer","catalogId":"`+hottycat.ID+`","label":"Tea","duration":2000,"onTimeout":{"event":{"name":"ready"}}},
 {"id":"lap","component":"HottyStopwatch","catalogId":"`+hottycat.ID+`","interval":100,"running":{"@path":"/lap"}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	tea, lap := partID("tea", partFrame), partID("lap", partFrame)
	if role, _ := s.Attr("tea", "role"); role != "timer" {
		t.Errorf("the timer's role is %q", role)
	}
	if got := s.TextOf("tea"); got != "Tea2s" {
		t.Errorf("the timer reads %q", got)
	}
	if got := s.TextOf(lap); got != "0s" {
		t.Errorf("the stopwatch reads %q", got)
	}
	if d := r.Animating(); d != view.TimerStep(100*time.Millisecond) {
		t.Errorf("counting, it animates every %v", d)
	}
	for _, at := range []struct {
		ms   int64
		want []string
	}{
		{50, nil},
		{500, []string{hotty.SetText(r.name, lap, "500ms")}},
		{1000, []string{hotty.SetText(r.name, tea, "1s"), hotty.SetText(r.name, lap, "1s")}},
	} {
		x.now = time.Unix(0, at.ms*int64(time.Millisecond))
		d := r.Update()
		if strings.Join(d, "") != strings.Join(at.want, "") {
			t.Errorf("at %dms: deltas %q, want %q", at.ms, d, at.want)
		}
		x.send(d...)
	}
	x.now = time.Unix(2, 0)
	x.send(r.Update()...)
	x.check(r)
	if class, _ := s.Attr(tea, "class"); class != "k-timer-time k-still" || s.TextOf(tea) != "0s" {
		t.Errorf("run out, the time is %q, class %q", s.TextOf(tea), class)
	}
	if len(x.actions) != 1 || x.actions[0].Name != "ready" {
		t.Errorf("run out, the agent got %+v", x.actions)
	}

	var stop []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/lap","value":false}}]`), &stop))
	x.process(stop...)
	x.check(r)
	if class, _ := s.Attr(lap, "class"); class != "k-timer-time k-still" || s.TextOf(lap) != "2s" {
		t.Errorf("stopped, the stopwatch is %q, class %q", s.TextOf(lap), class)
	}
	if d := r.Animating(); d != 0 {
		t.Errorf("none counts, and it animates every %v", d)
	}
	x.now = time.Unix(5, 0)
	if d := r.Update(); len(d) != 0 {
		t.Errorf("none counts, and a tick sends %q", d)
	}
}
