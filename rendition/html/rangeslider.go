package html

import (
	"math"
	"strconv"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// rangeSlider is a HottyRangeSlider's track: a div holding its rail, the
// fill between its knobs, the notches a tap sets a knob by (a Slider's),
// and its two knobs, each a button role=slider and a Tab stop, whose
// arrows reach the program (SPEC §10.2), which moves it (Key). On a host
// with steps (SPEC §9.1) the track is the drag target, as a Slider's is,
// and a press anywhere on it, a knob included, moves the nearer knob
// (view.Controller.PressKnob); the notches take taps alone. Without
// steps, each notch and each knob is a drag target. A disabled one takes
// neither. A finger drags along the track (touch-action, kit.css).
func (m *markup) rangeSlider(e *view.Element) *node {
	at := func(v float64) float64 {
		if e.Max <= e.Min {
			return 0
		}
		return math.Max(0, math.Min(1, (v-e.Min)/(e.Max-e.Min)))
	}
	lo, hi := e.Range()
	n := el("div", "id", domID(e.ID), "class", "k-track k-range", "role", "group").add(
		el("span", "class", "k-rail"),
		el("span", "class", "k-fill", "style", "left: "+pct(at(lo))+"; width: "+pct(at(hi)-at(lo))),
	)
	notch := "drag click"
	switch {
	case e.Disabled:
		notch = ""
		n.set("aria-disabled", "true")
	case m.steps:
		n.set("data-on", "drag")
		n.set("data-steps", strconv.Itoa(notches(e)-1))
		notch = "click"
	}
	n.add(notchSpans(e, notch)...)
	// Each knob's own range is up to the other, as a multi-thumb slider
	// says to a screen reader: the start's from min to the end, the end's
	// from the start to max.
	for _, k := range e.Children {
		v, _ := k.Value.(float64)
		vmin, vmax := e.Min, hi
		if k.Selected == 1 {
			vmin, vmax = lo, e.Max
		}
		b := el("button", "id", domID(k.ID), "type", "button", "class", "k-thumb", "role", "slider", "aria-label", k.Label,
			"aria-valuemin", a2ui.NumberString(vmin), "aria-valuemax", a2ui.NumberString(vmax), "aria-valuenow", a2ui.NumberString(v),
			"style", "left: "+pct(at(v))).flag("disabled", e.Disabled)
		if !e.Disabled && !m.steps {
			b.set("data-on", "drag")
		}
		n.add(b)
	}
	return n
}

// notchSpans are the notches a Slider's track is cut into (notches), each
// centred on where the knob stands at its value, half one at either end;
// on is their data-on, none when "".
func notchSpans(e *view.Element, on string) []*node {
	k := notches(e)
	out := make([]*node, 0, k)
	for i := range k {
		lo := math.Max(0, (float64(i)-0.5)/float64(k-1))
		hi := math.Min(1, (float64(i)+0.5)/float64(k-1))
		s := el("span", "id", partID(e.ID, partNotch+strconv.Itoa(i)), "class", "k-notch")
		if on != "" {
			s.set("data-on", on)
		}
		out = append(out, s.set("style", "left: "+strconv.FormatFloat(lo*100, 'f', 3, 64)+"%; width: "+strconv.FormatFloat((hi-lo)*100, 'f', 3, 64)+"%"))
	}
	return out
}

// rangeDrag takes a drag of a HottyRangeSlider (SPEC §9.1). Its dragstart
// picks the knob the drag moves: on the track, the nearer one, which takes
// the keyboard and goes to the pointer (view.Controller.PressKnob); on a
// knob, as a host without steps reports one, that knob, where it is. Each
// drag after moves it (DragKnob), to the step under the pointer on a host
// with steps, else to the notch it is over, until dragend. The click the
// drag ends with is the drag's (slideSteps). done reports that ev was one
// of these.
func (r *Rendition) rangeDrag(ev hotty.Event) (done bool, err error) {
	c := r.C
	switch ev.Kind {
	case hotty.EventDragStart:
		id, part, ok := viewID(ev.Target)
		e := c.V.Find(id)
		if !ok || e == nil {
			return false, nil
		}
		if e.Kind == view.Knob {
			if anc := c.Ancestors(id); len(anc) > 0 && anc[len(anc)-1].Kind == view.RangeSlider {
				rg := anc[len(anc)-1]
				r.drag, r.knob = rg.ID, e.Selected
				if lo, hi := rg.Range(); lo == hi {
					r.knob = -1
				}
				return true, nil
			}
			return false, nil
		}
		if e.Kind != view.RangeSlider {
			return false, nil
		}
		v, has := rangeAt(e, part, ev)
		if !has {
			return true, nil
		}
		// The press took the keyboard from the host, which sends blur next,
		// a track being no element that takes focus (SPEC §10.1): the knob
		// gets it back (Event, Update).
		r.drag, r.pressed = id, true
		r.knob, err = c.PressKnob(id, v)
		return true, err
	case hotty.EventDrag, hotty.EventDragEnd:
		e := c.V.Find(r.drag)
		if r.drag == "" || e == nil || e.Kind != view.RangeSlider {
			return false, nil
		}
		if ev.Kind == hotty.EventDragEnd {
			r.drag = ""
			if releasedOn(ev, e.ID) {
				r.dragged = e.ID
			}
		}
		_, part, _ := viewID(ev.Target)
		if v, has := rangeAt(e, part, ev); has {
			r.knob, err = c.DragKnob(e.ID, r.knob, v)
		}
		return true, err
	}
	return false, nil
}

// rangeAt is the value a drag of a HottyRangeSlider points at: the step
// under the pointer on a host with steps, else the notch it is over (part,
// "k3"). ok is false where it is neither.
func rangeAt(e *view.Element, part string, ev hotty.Event) (v float64, ok bool) {
	if x, has := stepX(ev); has {
		return notchValue(e, x), true
	}
	return notchAt(e, part)
}

// notchAt is the value of the notch of a track part names ("k3"); ok is
// false for a part that is no notch.
func notchAt(e *view.Element, part string) (v float64, ok bool) {
	if len(part) > 1 && part[0] == partNotch[0] {
		if i, err := strconv.Atoi(part[1:]); err == nil {
			return notchValue(e, i), true
		}
	}
	return 0, false
}

// TrackStep is where on a host a Slider's or a HottyRangeSlider's track
// stands for v: its notch's DOM id, which a tap there clicks, and its step
// (SPEC §9.1's x), which a drag there reports. ok is false for any other
// element.
func (r *Rendition) TrackStep(id string, v float64) (notch string, step int, ok bool) {
	e := r.C.V.Find(id)
	if e == nil || e.Kind != view.Slider && e.Kind != view.RangeSlider {
		return "", 0, false
	}
	k := notches(e)
	if e.Max > e.Min {
		step = int(math.Round((v - e.Min) / (e.Max - e.Min) * float64(k-1)))
	}
	step = min(max(step, 0), k-1)
	return partID(id, partNotch+strconv.Itoa(step)), step, true
}
