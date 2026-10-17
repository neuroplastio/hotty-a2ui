package html

import (
	"strconv"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// cycleKeys are the keys a cycle gives the program, which steps it
// (view.Controller.CycleKey): the arrows, Home and End, which a host that
// scrolls would take. A focused button uses no letters, and the host
// clicks it on Space and Enter, which steps it on.
const cycleKeys = "arrow_left=program arrow_right=program arrow_up=program arrow_down=program home=program end=program"

// cycle is a cycle on a host (io_neuroplast_hotty.cycle, profile §2): a
// field as wide as its widest option (box), holding its value, a button
// and the Tab stop (n), which a click steps on, between two chevrons,
// buttons out of the Tab order, as a Slider's − and +, that step it back
// and on.
func cycle(e *view.Element) (n, box *node) {
	n = el("button", "id", domID(e.ID), "type", "button", "class", "k-cycle-value", "role", "spinbutton",
		"aria-valuemin", "1", "aria-valuemax", strconv.Itoa(len(e.Options)), "data-keys", cycleKeys)
	widest := 1
	for _, o := range e.Options {
		widest = max(widest, utf8.RuneCountInString(o.Label))
	}
	value := el("span", "class", "k-value", "style", "min-width: "+strconv.Itoa(widest)+"ch")
	if i := e.Picked(); i >= 0 {
		o := e.Options[i]
		n.set("aria-valuenow", strconv.Itoa(i+1)).set("aria-valuetext", o.Label).add(beside(o.Icon))
		value.add(texts(o.Label)...)
	} else {
		value.set("class", "k-value k-none").add(txt("…"))
	}
	n.add(value)
	less := el("button", "id", partID(e.ID, partLess), "type", "button", "class", "k-cycle-step", "tabindex", "-1", "aria-label", "previous").
		add(beside("keyboard_arrow_left"))
	more := el("button", "id", partID(e.ID, partMore), "type", "button", "class", "k-cycle-step", "tabindex", "-1", "aria-label", "next").
		add(beside("keyboard_arrow_right"))
	return n, el("div", "id", partID(e.ID, partRange), "class", "k-input k-cycle").add(less, n, more)
}
