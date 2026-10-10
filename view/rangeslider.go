package view

import (
	"math"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// knobProps are a HottyRangeSlider's props its knobs edit, the start's
// and the end's.
var knobProps = [2]string{"start", "end"}

// mapRange makes a HottyRangeSlider's element (profile §6.20): a range of
// numbers from min to max, as a Slider's, whose two ends are knobs of
// their own (Knob children), each a Tab stop, each bound to its own path.
// An absent end is the range's own end; an end below the start is drawn
// at the start, until the user moves it.
func mapRange(b *Builder, n *a2ui.Node) *Element {
	lo, hi := a2ui.ToNumber(b.Raw(n, "min")), a2ui.ToNumber(b.Raw(n, "max"))
	if math.IsNaN(lo) {
		lo = 0
	}
	if math.IsNaN(hi) {
		hi = 100
	}
	e := &Element{Kind: RangeSlider, Label: b.String(n, "label"), Min: lo, Max: hi, Disabled: b.Bool(n, "disabled"), Error: b.FieldError(n)}
	if steps := a2ui.ToNumber(b.Raw(n, "steps")); steps >= 1 && hi > lo {
		e.Step = (hi - lo) / math.Floor(steps)
	}
	vs := [2]float64{lo, hi}
	for i := range vs {
		if v := a2ui.ToNumber(b.knobValue(n, i)); !math.IsNaN(v) {
			vs[i] = math.Max(lo, math.Min(hi, v))
		}
	}
	vs[1] = math.Max(vs[0], vs[1])
	name := e.Label
	if name == "" {
		name = b.a11y(n).Label
	}
	for i, v := range vs {
		label := [2]string{"Start", "End"}[i]
		if name != "" {
			label = name + ", " + knobProps[i]
		}
		e.Children = append(e.Children, &Element{ID: SubID(n.Key, "knob", i), Kind: Knob, Type: "HottyRangeSlider.knob",
			Value: v, Selected: i, Label: label, Disabled: e.Disabled})
	}
	return e
}

// knobValue is a knob's value as Value has a control's: the user's, while
// its prop is not bound and the user moved it; else the prop's.
func (b *Builder) knobValue(n *a2ui.Node, i int) any {
	if bd, ok := n.Props[knobProps[i]].(a2ui.Bound); !ok || !bd.Writable() {
		if v, ok := b.St.Local[SubID(n.Key, "knob", i)]; ok {
			return v
		}
	}
	return b.Raw(n, knobProps[i])
}

// Range is a RangeSlider's start and end, its knobs' values.
func (e *Element) Range() (start, end float64) {
	var vs [2]float64
	for _, k := range e.Children {
		if k.Kind == Knob && k.Selected >= 0 && k.Selected < 2 {
			vs[k.Selected], _ = k.Value.(float64)
		}
	}
	return vs[0], vs[1]
}

// RangeWidth is the most characters a RangeSlider's value takes, its
// start and its end with a dash between ("20–70"): twice a Slider's
// (SliderWidth) and one, so that the track stays where it is as they move.
func (e *Element) RangeWidth() int {
	lo, hi := e.Range()
	return 2*max(e.valueWidth(lo), e.valueWidth(hi)) + 1
}

// RangeText is a RangeSlider's value as it reads: "20–70".
func (e *Element) RangeText() string {
	lo, hi := e.Range()
	return a2ui.NumberString(lo) + "–" + a2ui.NumberString(hi)
}

// NearerKnob is the knob of a RangeSlider a press at v moves: the nearer
// one, 0 the start's and 1 the end's, the start's where v is midway
// between them. Where the knobs meet, it is the start's below them and the
// end's above them, and -1 at them: neither, until the pointer moves one
// way (DragKnob).
func (e *Element) NearerKnob(v float64) int {
	lo, hi := e.Range()
	d0, d1 := math.Abs(v-lo), math.Abs(v-hi)
	switch {
	case d0 < d1:
		return 0
	case d1 < d0:
		return 1
	case lo < hi, v < lo:
		return 0
	case v > hi:
		return 1
	}
	return -1
}

// PressKnob is a press on a RangeSlider's track at v, a click's or the
// start of a drag (profile §3.7): the nearer knob takes the keyboard and
// goes to v, on its steps, stopped where it meets the other. It returns
// that knob, 0 or 1. Where the knobs meet at v, neither moves and it
// returns -1: the end's knob takes the keyboard, or the start's at the
// max, the one that can move away from the other; a drag's first move
// then picks the one it goes towards (DragKnob). A disabled one does
// nothing.
func (c *Controller) PressKnob(id string, v float64) (int, error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != RangeSlider || e.Disabled || math.IsNaN(v) {
		return -1, nil
	}
	v = e.snap(v)
	k := e.NearerKnob(v)
	if k < 0 {
		focus := 1
		if _, hi := e.Range(); hi >= e.Max {
			focus = 0
		}
		c.Focus(SubID(id, "knob", focus))
		return -1, nil
	}
	c.Focus(SubID(id, "knob", k))
	return k, c.SetValue(SubID(id, "knob", k), v)
}

// DragKnob moves knob k of a RangeSlider to v, as a drag does after
// PressKnob, stopped where it meets the other knob. Where the press found
// the knobs meeting (k < 0), the first v past them picks the knob it goes
// towards, which takes the keyboard; DragKnob returns the knob it moved,
// -1 until then.
func (c *Controller) DragKnob(id string, k int, v float64) (int, error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != RangeSlider || e.Disabled || math.IsNaN(v) {
		return k, nil
	}
	v = e.snap(v)
	if k < 0 {
		lo, hi := e.Range()
		switch {
		case v < lo:
			k = 0
		case v > hi:
			k = 1
		default:
			return -1, nil
		}
		c.Focus(SubID(id, "knob", k))
	}
	return k, c.SetValue(SubID(id, "knob", k), v)
}

// rangeOf is the RangeSlider a knob is one of, and the knob's index; nil
// when id is no knob.
func (c *Controller) rangeOf(id string) (*Element, int) {
	if k := c.V.Find(id); k == nil || k.Kind != Knob {
		return nil, 0
	}
	parent, i := parentAndIndex(id)
	r := c.V.Find(parent)
	if r == nil || r.Kind != RangeSlider || i < 0 || i > 1 {
		return nil, 0
	}
	return r, i
}

// setKnob writes knob i of a RangeSlider: v clamped to the range, on its
// steps, and stopped where it meets the other knob; to its prop's binding,
// else to the renderer's state.
func (c *Controller) setKnob(r *Element, i int, v float64) error {
	v = r.snap(v)
	lo, hi := r.Range()
	if i == 0 {
		v = math.Min(v, hi)
	} else {
		v = math.Max(v, lo)
	}
	c.St.Touched[r.ID] = true
	if bd, ok := c.V.Node(r.ID).Props[knobProps[i]].(a2ui.Bound); ok && bd.Writable() {
		return c.S.Write(bd.Path, v)
	}
	c.St.Local[SubID(r.ID, "knob", i)] = v
	return nil
}
