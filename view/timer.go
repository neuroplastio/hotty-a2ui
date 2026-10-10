package view

import (
	"fmt"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// Timers and stopwatches (profile §6.25) in the view. A HottyTimer counts
// down from its duration and runs its onTimeout when it gets to 0; a
// HottyStopwatch counts up from 0. A2UI has no time (vault a2ui-limits
// L6), so the time a timer has counted is the surface's state, not the
// data model's, counted on each rendition's clock as it draws
// (TickTimers), as a toast's is. Its running prop says whether it counts;
// hottyStartTimer, hottyStopTimer, hottyToggleTimer and hottyResetTimer
// work it from a Button or the agent (TimerDo), as bubbles' timer and
// stopwatch have Start, Stop, Toggle and Reset.

// Both are one kind of element, Timer, its Variant one of these.
const (
	TimerCountdown = "timer"
	TimerStopwatch = "stopwatch"
)

func init() {
	Register(hotty.ID, "HottyTimer", mapTimer)
	Register(hotty.ID, "HottyStopwatch", mapTimer)
}

// TimerState is a HottyTimer's or a HottyStopwatch's time as its surface
// keeps it (State.Timers), by its node's key.
type TimerState struct {
	// Elapsed is how long it has counted: a HottyTimer has Elapsed less
	// than its duration left, and has run out once Elapsed is its duration.
	Elapsed time.Duration

	// last is the draw it was counted at last (TickTimers), zero before
	// its first and after a reset; counting, whether it counted from then
	// on; props, its props then.
	last     time.Time
	counting bool
	props    timerProps
}

// timerProps are a timer's props, resolved.
type timerProps struct {
	// countdown: a HottyTimer, which counts down from duration (-1 while
	// that is not known: it does not count); else a HottyStopwatch.
	countdown bool
	duration  time.Duration
	// interval is the step its time shows in: 1s by default, as bubbles'.
	interval time.Duration
	// clock: its format is "clock" (1:30), else "duration" (1m30s).
	clock bool
	// running is its running prop: true by default; where it is not
	// bound, the user's start or stop (State.Local).
	running bool
}

// readTimer reads a timer node's props.
func readTimer(b *Builder, n *a2ui.Node) timerProps {
	t := timerProps{countdown: n.Type == "HottyTimer", duration: -1, interval: time.Second}
	// NaN, from a path that holds nothing, is no duration.
	if ms := a2ui.ToNumber(b.Raw(n, "duration")); t.countdown && ms >= 0 {
		t.duration = time.Duration(ms * float64(time.Millisecond))
	}
	if ms := a2ui.ToNumber(b.Raw(n, "interval")); ms >= 1 {
		t.interval = time.Duration(ms * float64(time.Millisecond))
	}
	t.clock = b.Enum(n, "format", "duration") == "clock"
	// running is true by default: absent, not bound to nothing.
	v := b.Value(n, "running")
	_, set := n.Props["running"]
	t.running = a2ui.Truthy(v) || v == nil && !set
	return t
}

// done reports whether a HottyTimer has run out, after elapsed.
func (t timerProps) done(elapsed time.Duration) bool {
	return t.countdown && t.duration >= 0 && elapsed >= t.duration
}

// counts reports whether it counts on after elapsed: it runs, and a
// HottyTimer knows its duration and has time left.
func (t timerProps) counts(elapsed time.Duration) bool {
	return t.running && (!t.countdown || t.duration >= 0 && !t.done(elapsed))
}

// shown is the time it shows after elapsed, a whole number of intervals:
// a stopwatch's elapsed rounded down; a timer's time left rounded up, no
// more than its duration, so that it shows 0 just as it runs out, as
// bubbles' timer does.
func (t timerProps) shown(elapsed time.Duration) time.Duration {
	switch {
	case !t.countdown:
		return elapsed / t.interval * t.interval
	case t.duration < 0:
		return 0
	}
	left := max(t.duration-elapsed, 0)
	return min((left+t.interval-1)/t.interval*t.interval, t.duration)
}

// face is its element's Value and Active after elapsed: the time it
// shows, formatted, and whether it counts.
func (t timerProps) face(elapsed time.Duration) (string, bool) {
	return FormatTimer(t.shown(elapsed), t.interval, t.clock), t.counts(elapsed)
}

// mapTimer makes a HottyTimer's or a HottyStopwatch's element: a Timer,
// its Variant TimerCountdown or TimerStopwatch, its Label, its Value the
// time it shows as text (FormatTimer), Active while it counts. The time is
// as the last draw counted it (TickTimers).
func mapTimer(b *Builder, n *a2ui.Node) *Element {
	t := readTimer(b, n)
	var elapsed time.Duration
	if st := b.St.Timers[n.Key]; st != nil {
		elapsed = st.Elapsed
	}
	e := &Element{Kind: Timer, Variant: TimerStopwatch, Label: b.String(n, "label")}
	if t.countdown {
		e.Variant = TimerCountdown
	}
	e.Value, e.Active = t.face(elapsed)
	return e
}

// FormatTimer is a time as a timer shows it, in its interval's step.
// "duration", the default, writes it as Go does and bubbles' timer and
// stopwatch show it: 1m30s, 4.5s, 300ms, 0s. "clock" (clock true) writes
// it as a clock does, minutes and seconds, the hours before them from an
// hour: 1:30, 0:04, 1:02:03; and the fraction of a second the interval
// has, to milliseconds: 0:04.5 in tenths.
func FormatTimer(d, interval time.Duration, clock bool) string {
	if !clock {
		return d.String()
	}
	s := fmt.Sprintf("%d:%02d", d/time.Minute, d%time.Minute/time.Second)
	if d >= time.Hour {
		s = fmt.Sprintf("%d:%02d:%02d", d/time.Hour, d%time.Hour/time.Minute, d%time.Minute/time.Second)
	}
	digits := 0
	switch frac := interval % time.Second; {
	case frac == 0:
	case frac%(100*time.Millisecond) == 0:
		digits = 1
	case frac%(10*time.Millisecond) == 0:
		digits = 2
	default:
		digits = 3
	}
	if digits > 0 {
		ms := int(d % time.Second / time.Millisecond)
		for range 3 - digits {
			ms /= 10
		}
		s += "." + fmt.Sprintf("%0*d", digits, ms)
	}
	return s
}

// TimerStep is how soon a rendition draws again while a timer counts
// (Animating): a tenth of a second, or the timer's interval where that is
// shorter, down to a thirtieth. Its time so shows within a step of when it
// changes, and its onTimeout runs within a step of its end.
func TimerStep(interval time.Duration) time.Duration {
	return min(max(interval, time.Second/30), time.Second/10)
}

// isTimer reports whether a node is a HottyTimer or a HottyStopwatch.
func isTimer(n *a2ui.Node) bool {
	return n.State == a2ui.Resolved && n.Catalog == hotty.ID && (n.Type == "HottyTimer" || n.Type == "HottyStopwatch")
}

// TickTimers counts the timers' time at a rendition's draw, on its clock,
// before it draws them, as TickToasts counts the toasts'. Each timer of the
// surface counts, whether the view shows it or not (one in a tab not
// shown): since the draw before, one that counted then has counted that
// much longer; one that runs from now on counts from now. A HottyTimer
// that has run out stops: its running is set false (written where it is
// bound), and its onTimeout runs, once, without the user's activation. A
// HottyTimer started once it has run out starts over. It returns how soon
// to draw again: the shortest TimerStep of those that count, 0 while none
// does.
func (c *Controller) TickTimers(now time.Time) time.Duration {
	b := &Builder{S: c.S, St: c.St}
	seen := map[string]bool{}
	var ended []*a2ui.Node
	next := time.Duration(0)
	for _, n := range c.S.Tree.Nodes() {
		if !isTimer(n) {
			continue
		}
		seen[n.Key] = true
		st := c.St.Timers[n.Key]
		if st == nil {
			st = &TimerState{}
			c.St.Timers[n.Key] = st
		}
		t := readTimer(b, n)
		if st.counting && !st.last.IsZero() && now.After(st.last) {
			st.Elapsed += now.Sub(st.last)
		}
		st.last, st.props = now, t
		if t.running && !st.counting && t.done(st.Elapsed) {
			st.Elapsed = 0
		}
		if t.done(st.Elapsed) {
			st.Elapsed = t.duration
			if t.running {
				ended = append(ended, n)
			}
		}
		st.counting = t.counts(st.Elapsed)
		if st.counting && (next == 0 || TimerStep(t.interval) < next) {
			next = TimerStep(t.interval)
		}
	}
	for k := range c.St.Timers {
		if !seen[k] {
			delete(c.St.Timers, k)
		}
	}
	if len(ended) == 0 {
		// Nothing else changed: the elements show the new time.
		c.V.Walk(func(e *Element) bool {
			if st := c.St.Timers[e.ID]; e.Kind == Timer && st != nil {
				e.Value, e.Active = st.props.face(st.Elapsed)
			}
			return true
		})
		return next
	}
	for _, n := range ended {
		// It stops first, so that what its onTimeout reads, and the Buttons
		// that start it, see it stopped.
		c.setRunning(n, false)
		if n.Props["onTimeout"] != nil {
			c.S.Tree.Invoke(n, "onTimeout", false)
		}
	}
	c.Rebuild()
	return next
}

// setRunning sets a timer's running: to the data model where it is bound,
// else as the renderer's (State.Local), as a control's value.
func (c *Controller) setRunning(n *a2ui.Node, on bool) error {
	if bd, ok := n.Props["running"].(a2ui.Bound); ok && bd.Writable() {
		return c.S.Write(bd.Path, on)
	}
	c.St.Local[n.Key] = on
	return nil
}

// The ways TimerDo works a timer: hottyStartTimer's, hottyStopTimer's,
// hottyToggleTimer's and hottyResetTimer's.
const (
	TimerStart  = "start"
	TimerStop   = "stop"
	TimerToggle = "toggle"
	TimerReset  = "reset"
)

// TimerDo works a timer by its node's key (FindTimer), as bubbles' Start,
// Stop, Toggle and Reset do: start sets its running true, and it counts on
// from where it was, a HottyTimer that ran out from its duration again
// (TickTimers); stop sets it false, where it is; toggle does the one it is
// not doing; reset takes it back to 0 counted, a HottyTimer to its
// duration left, running or not as it was.
func (c *Controller) TimerDo(key, do string) error {
	n := c.S.Tree.Node(key)
	if n == nil || !isTimer(n) {
		return fmt.Errorf("no timer '%s'", key)
	}
	var err error
	switch do {
	case TimerStart:
		err = c.setRunning(n, true)
	case TimerStop:
		err = c.setRunning(n, false)
	case TimerToggle:
		err = c.setRunning(n, !readTimer(&Builder{S: c.S, St: c.St}, n).running)
	case TimerReset:
		// The next draw counts from itself: the time from the last one to
		// now is gone with the rest.
		if st := c.St.Timers[key]; st != nil {
			st.Elapsed, st.last = 0, time.Time{}
		}
	default:
		return fmt.Errorf("a timer does not %q", do)
	}
	c.Rebuild()
	return err
}

// FindTimer is the node key of a HottyTimer or HottyStopwatch by its
// component id: the instance in scope, for a template's, else the first in
// tree order; "" if the surface has none. Unlike FindComponent it finds one
// the view does not show, which counts all the same (TickTimers).
func (c *Controller) FindTimer(componentID string, scope a2ui.Scope) string {
	found := ""
	for _, n := range c.S.Tree.Nodes() {
		if !isTimer(n) || n.ComponentID != componentID {
			continue
		}
		if n.Scope.Path == scope.Path {
			return n.Key
		}
		if found == "" {
			found = n.Key
		}
	}
	return found
}
