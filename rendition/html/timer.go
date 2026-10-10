package html

import "github.com/neuroplastio/hotty-a2ui/view"

// timer is a HottyTimer or a HottyStopwatch, as cells draws it: a span
// role=timer (which a screen reader does not read out at each tick)
// holding its label and its time. The time has an id of its own (~f), so
// that a tick is one text's delta, and k-still while it does not count, in
// muted (kit.css), one attribute's delta. The program's draws move it, on
// the view's clock (view.Controller.TickTimers), as a Spinner's frames: a
// host runs no animation of its own here (vault knowledge/host-motion.md).
func timer(e *view.Element) *node {
	n := el("span", "id", domID(e.ID), "class", "k-timer", "role", "timer")
	if e.Label != "" {
		n.add(el("span", "class", "k-timer-label").add(texts(e.Label)...))
	}
	t, _ := e.Value.(string)
	class := "k-timer-time"
	if !e.Active {
		class += " k-still"
	}
	return n.add(el("span", "id", partID(e.ID, partFrame), "class", class).add(txt(t)))
}
