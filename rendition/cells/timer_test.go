package cells

import (
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// timing is the timer story's surface in cells, its clock stopped at the
// time it returns.
func timing(t *testing.T) (*Rendition, *view.Controller, *time.Time) {
	t.Helper()
	st := story.Find("hotty/timer")
	if st == nil {
		t.Fatal("no story hotty/timer")
	}
	run, err := story.Start(st)
	if err != nil {
		t.Fatal(err)
	}
	c := run.Surfaces()[0].C
	r := New(c)
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	r.Clock = func() time.Time { return now }
	return r, c, &now
}

// A timer is its label, a space and its time on one row, as bubbles'
// examples draw them, counting from the first draw on the rendition's
// clock: the stopwatch in tenths, Focus as a clock, the tea Go's way. The
// time is in the text's colour while it counts, and the rendition asks for
// a draw every tenth of a second.
func TestTimerRows(t *testing.T) {
	r, _, now := timing(t)
	at := func(ms int) string {
		*now = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC).Add(time.Duration(ms) * time.Millisecond)
		return r.Draw(60).Plain()
	}
	want := "Timer and stopwatch\n\n" +
		"Lap 0s\n" +
		"/lapMs: 0 ms\n" +
		"Focus 25:00\n" +
		"Tea 10s\n" +
		" Start   Stop   Reset\n\n" +
		"s start/stop lap • r reset lap • ? more"
	if got := at(0); got != want {
		t.Fatalf("at 0:\n%s\nwant\n%s", got, want)
	}
	got := at(3250)
	for _, row := range []string{"Lap 3.2s", "Focus 24:57", "Tea 7s"} {
		if !strings.Contains(got, row+"\n") {
			t.Errorf("at 3.25s, no %q in\n%s", row, got)
		}
	}
	if d := r.Animating(); d != view.TimerStep(100*time.Millisecond) {
		t.Errorf("counting, it animates every %v", d)
	}
	f := r.Draw(60)
	y := rowOf(t, f, "Tea 7s")
	for x, c := range f.Cells[y][:6] {
		if c.Role != Fg {
			t.Errorf("Tea's cell %d (%q) is %v while it counts", x, c.Text, c.Role)
		}
	}
}

// The tea runs out at 10 seconds: it shows 0s in muted, its onTimeout's
// toast shows, and the Buttons' checks, which read the path its running
// is bound to, turn Start on and Stop off. Start brews another. s stops the
// stopwatch, a HottyShortcut's hottyToggleTimer, and it stays where it
// stopped, in muted.
func TestTimerRunsOut(t *testing.T) {
	r, c, now := timing(t)
	t0 := *now
	f := r.Draw(60)
	y, ly := rowOf(t, f, "Tea 10s"), rowOf(t, f, "Lap 0s")
	*now = t0.Add(10 * time.Second)
	f = r.Draw(60)
	if l := strings.Split(f.Plain(), "\n")[y]; !strings.HasPrefix(l, "Tea 0s") {
		t.Errorf("run out: %q", l)
	}
	if c := f.Cells[y][4]; c.Role != Muted {
		t.Errorf("a timer run out is %v", c.Role)
	}
	if c.S.Data.Value("/tea") != false {
		t.Errorf("/tea is %v", c.S.Data.Value("/tea"))
	}
	if !strings.Contains(f.Plain(), "✓ Tea is ready") {
		t.Errorf("no toast:\n%s", f.Plain())
	}
	if c.V.Find("start").Disabled || !c.V.Find("stop").Disabled {
		t.Errorf("run out: Start disabled %v, Stop disabled %v", c.V.Find("start").Disabled, c.V.Find("stop").Disabled)
	}
	col, row, _, _, ok := r.Box("start")
	if !ok {
		t.Fatal("Start not drawn")
	}
	if err := r.Click(col+1, row); err != nil {
		t.Fatal(err)
	}
	if err := r.Release(); err != nil {
		t.Fatal(err)
	}
	// A program draws after each click and each key: the draw after one
	// counts the time up to it, and a timer started counts from it.
	if l := strings.Split(r.Draw(60).Plain(), "\n")[y]; !strings.HasPrefix(l, "Tea 10s") {
		t.Errorf("started over: %q", l)
	}
	*now = t0.Add(13600 * time.Millisecond)
	f = r.Draw(60)
	if l := strings.Split(f.Plain(), "\n")[y]; !strings.HasPrefix(l, "Tea 7s") {
		t.Errorf("started over at 10s, at 13.6s: %q", l)
	}

	c.Focus("reset")
	if ok, err := r.Key("s"); !ok || err != nil {
		t.Fatalf("s: %v %v", ok, err)
	}
	r.Draw(60)
	*now = t0.Add(20 * time.Second)
	f = r.Draw(60)
	if l := strings.Split(f.Plain(), "\n")[ly]; !strings.HasPrefix(l, "Lap 13.6s") {
		t.Errorf("stopped at 13.6s, at 20s: %q", l)
	}
	if c := f.Cells[ly][4]; c.Role != Muted {
		t.Errorf("a stopwatch stopped is %v", c.Role)
	}
	if ok, err := r.Key("r"); !ok || err != nil {
		t.Fatalf("r: %v %v", ok, err)
	}
	if l := strings.Split(r.Draw(60).Plain(), "\n")[ly]; !strings.HasPrefix(l, "Lap 0s") {
		t.Errorf("reset: %q", l)
	}
}
