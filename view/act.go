package view

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// Controller is one surface as the user acts on it: the renditions turn
// what the user does (a click, a key, an edit) into its calls, which
// change the data model or the renderer's state, run actions, and build
// the view again.
type Controller struct {
	S  *a2ui.Surface
	St *State
	V  *Surface
}

// NewController is a surface's controller, its view built.
func NewController(s *a2ui.Surface) *Controller {
	c := &Controller{S: s, St: NewState()}
	c.Rebuild()
	return c
}

// Rebuild builds the view again, from the surface as it is resolved now.
// Until an element has had the keyboard, one with autofocus takes it as
// soon as it is there: a surface streams in after it is made.
func (c *Controller) Rebuild() {
	c.V = Build(c.S, c.St)
	if c.St.Focus != "" {
		return
	}
	c.V.Walk(func(e *Element) bool {
		if e.Autofocus && e.Focusable() {
			c.St.Focus, c.St.Keyboard = e.ID, true
			return false
		}
		return true
	})
}

// Ancestors are the elements that contain id, outermost first; nil when
// id is not in the view.
func (c *Controller) Ancestors(id string) []*Element {
	var path []*Element
	var walk func(e *Element) bool
	walk = func(e *Element) bool {
		if e == nil {
			return false
		}
		if e.ID == id {
			return true
		}
		path = append(path, e)
		for _, ch := range e.Children {
			if walk(ch) {
				return true
			}
		}
		path = path[:len(path)-1]
		return false
	}
	if walk(c.V.Root) || walk(c.V.Overlay) {
		return path
	}
	return nil
}

// FormOf is the Form that contains id, or nil.
func (c *Controller) FormOf(id string) *Element {
	anc := c.Ancestors(id)
	for i := len(anc) - 1; i >= 0; i-- {
		if anc[i].Kind == Form {
			return anc[i]
		}
	}
	return nil
}

// Activate is a click on an element, or Enter or Space on it (SPEC
// §10.2): a Button runs its action, a CheckBox or an Option toggles, a
// Tab is shown, a link opens. Inside a Modal's trigger it then opens the
// Modal.
func (c *Controller) Activate(id string) error {
	e := c.V.Find(id)
	if e == nil {
		return nil
	}
	var err error
	switch e.Kind {
	case Button:
		if e.Disabled {
			c.St.Submitted[id] = true
			c.Rebuild()
			return nil
		}
		err = c.S.Tree.Invoke(c.V.Node(id), "action", true)
	case CheckBox:
		v, _ := e.Value.(bool)
		err = c.set(e, !v)
	case Tab:
		tabs, i := parentAndIndex(id)
		c.St.Tabs[tabs] = i
	case Option:
		choiceID, _ := parentAndIndex(id)
		ch := c.V.Find(choiceID)
		picked, _ := ch.Value.([]string)
		v, _ := e.Value.(string)
		switch {
		case !ch.Multiple:
			picked = []string{v}
		case slices.Contains(picked, v):
			picked = slices.DeleteFunc(slices.Clone(picked), func(x string) bool { return x == v })
		default:
			picked = append(slices.Clone(picked), v)
		}
		err = c.set(ch, picked)
	case Media:
		if c.S.Env().OpenURL != nil && e.URL != "" {
			err = c.S.Env().OpenURL(e.URL)
		}
	}
	for _, a := range c.Ancestors(id) {
		if a.Kind == Modal && !a.Open {
			c.St.Modal = a.ID
		}
	}
	if e.Kind == Modal && e.Clickable {
		c.St.Modal = e.ID
	}
	c.Rebuild()
	return err
}

// CloseModal closes the open Modal.
func (c *Controller) CloseModal() {
	if c.St.Modal != "" {
		c.St.Modal = ""
		c.Rebuild()
	}
}

// SetValue is the user's edit of a control: a TextField's or a
// DateTime's string, a CheckBox's bool, a Slider's number (clamped to its
// range, on its steps), a Choice's value or values. A bound value goes to
// the data model; another stays the renderer's.
func (c *Controller) SetValue(id string, v any) error {
	e := c.V.Find(id)
	if e == nil {
		return nil
	}
	switch e.Kind {
	case Slider:
		f := a2ui.ToNumber(v)
		if math.IsNaN(f) {
			return nil
		}
		v = e.snap(f)
	case Choice:
		if s, ok := v.(string); ok {
			v = []string{s}
		}
	}
	err := c.set(e, v)
	c.Rebuild()
	return err
}

// StepSlider moves a Slider by n of its steps (back when n < 0), or to an
// end: "Home" its Min, "End" its Max. A Slider with no step moves by a
// twentieth of its range. It does nothing to any other element.
func (c *Controller) StepSlider(id string, n int, to string) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Slider {
		return nil
	}
	f, _ := e.Value.(float64)
	switch to {
	case "Home":
		f = e.Min
	case "End":
		f = e.Max
	default:
		f += float64(n) * e.SliderStep()
	}
	return c.SetValue(id, f)
}

// set writes a control's value: to its binding, else to the state.
func (c *Controller) set(e *Element, v any) error {
	c.St.Touched[e.ID] = true
	n := c.V.Node(e.ID)
	if l, ok := v.([]string); ok {
		a := make([]any, len(l))
		for i, s := range l {
			a[i] = s
		}
		v = a
	}
	if bd, ok := n.Props["value"].(a2ui.Bound); ok && bd.Writable() {
		return c.S.Write(bd.Path, v)
	}
	c.St.Local[e.ID] = v
	return nil
}

// Submit submits a Form: its checks and those of every control in it
// must pass, else their errors show; then its onSubmit runs.
func (c *Controller) Submit(formID string) error {
	f := c.V.Find(formID)
	if f == nil || f.Kind != Form {
		return nil
	}
	ok := !f.Disabled
	var walk func(e *Element)
	walk = func(e *Element) {
		if n := c.V.Node(e.ID); n != nil && !n.Valid() && e.ID != formID {
			ok = false
		}
		for _, ch := range e.Children {
			walk(ch)
		}
	}
	walk(f)
	if !ok {
		c.St.Submitted[formID] = true
		c.Rebuild()
		return nil
	}
	err := c.S.Tree.Invoke(c.V.Node(formID), "onSubmit", true)
	c.Rebuild()
	return err
}

// Enter is Enter in a text field: it submits the field's Form, if it is
// in one.
func (c *Controller) Enter(id string) error {
	if f := c.FormOf(id); f != nil {
		return c.Submit(f.ID)
	}
	return nil
}

// Shortcut runs the surface's Shortcut for a key ("Control+s"), if it has
// one: it presses its Button, as a click would and only if the Button's
// checks pass, or runs its action. ok reports whether one took the key.
func (c *Controller) Shortcut(key string) (ok bool, err error) {
	for _, sc := range c.V.Shortcuts {
		if !SameKey(sc.Key, key) {
			continue
		}
		if sc.Press != "" {
			if b := c.V.Find(sc.Press); b != nil && !b.Disabled {
				return true, c.Activate(sc.Press)
			}
			return true, nil
		}
		err = c.S.Tree.Invoke(c.V.Node(sc.ID), "action", true)
		c.Rebuild()
		return true, err
	}
	return false, nil
}

// SameKey compares two keys as HOTTY names them (SPEC §10.4): modifiers
// in any order, and a character with Shift in itself, so that Control+S
// is Control+Shift+s, and Space is " ". A key that is no name there is
// only itself.
func SameKey(a, b string) bool {
	ca, oka := hotty.ParseKey(a)
	cb, okb := hotty.ParseKey(b)
	if !oka || !okb {
		return a == b
	}
	return ca == cb
}

// Focus gives the keyboard to an element, or takes it from the surface
// with id "".
func (c *Controller) Focus(id string) {
	if id == "" {
		c.St.Keyboard = false
		return
	}
	c.St.Focus, c.St.Keyboard = id, true
}

// FindComponent is the element of a component, by its id: the instance
// in scope, for a template's, else the first in tree order; "" if none
// is in the view.
func (c *Controller) FindComponent(componentID string, scope a2ui.Scope) string {
	found, first := "", ""
	c.V.Walk(func(e *Element) bool {
		n := c.V.Node(e.ID)
		if n == nil || n.ComponentID != componentID {
			return true
		}
		if first == "" {
			first = e.ID
		}
		if n.Scope.Path == scope.Path {
			found = e.ID
			return false
		}
		return true
	})
	if found == "" {
		found = first
	}
	return found
}

// FocusNext moves the keyboard to the next focusable element (or the
// previous one, back), as Tab and Shift+Tab do: past the last, or before
// the first, the surface loses it (SPEC §10.2). It reports whether the
// surface still has the keyboard.
func (c *Controller) FocusNext(back bool) bool {
	ids := c.V.Focusables()
	if len(ids) == 0 {
		c.St.Keyboard = false
		return false
	}
	i := slices.Index(ids, c.St.Focus)
	switch {
	case !c.St.Keyboard || i < 0:
		if back {
			i = len(ids) - 1
		} else {
			i = 0
		}
	case back:
		i--
	default:
		i++
	}
	if i < 0 || i >= len(ids) {
		c.St.Keyboard = false
		return false
	}
	c.St.Focus, c.St.Keyboard = ids[i], true
	return true
}

// parentAndIndex reads a SubID: the element's id and the part's index.
func parentAndIndex(id string) (string, int) {
	i := strings.LastIndex(id, "/")
	n, _ := strconv.Atoi(id[i+1:])
	j := strings.LastIndex(id[:i], "/")
	return id[:j], n
}

// snap clamps a Slider's value to its range and puts it on its steps: its
// Step, else a hundredth of its range. It is rounded to as many decimals
// as the step and Min have, so that 0.45 less two steps of 0.05 is 0.35,
// not 0.35000000000000003.
func (e *Element) snap(f float64) float64 {
	f = math.Max(e.Min, math.Min(e.Max, f))
	q := e.Step
	if q <= 0 {
		q = (e.Max - e.Min) / 100
	}
	if q <= 0 || math.IsInf(q, 0) {
		return f
	}
	f = e.Min + math.Round((f-e.Min)/q)*q
	p := math.Pow(10, float64(e.sliderDecimals()))
	return math.Max(e.Min, math.Min(e.Max, math.Round(f*p)/p))
}

// sliderDecimals is how many digits a Slider's value has after the point
// once snapped: as many as its step (a hundredth of its range without
// one) and its Min have, at most 12.
func (e *Element) sliderDecimals() int {
	q := e.Step
	if q <= 0 {
		q = (e.Max - e.Min) / 100
	}
	if q <= 0 || math.IsInf(q, 0) {
		return 0
	}
	return min(max(decimals(q), decimals(e.Min)), 12)
}

// SliderWidth is the most characters a Slider's value takes: its ends',
// its value's, and a snapped value's whole part, point and decimals. A
// rendition that gives the value this much room keeps the track where it
// is as the value moves (0.45 is wider than 1).
func (e *Element) SliderWidth() int {
	v, _ := e.Value.(float64)
	w := max(len(a2ui.NumberString(e.Min)), len(a2ui.NumberString(e.Max)), len(a2ui.NumberString(v)))
	if d := e.sliderDecimals(); d > 0 {
		whole := 0
		for _, f := range []float64{e.Min, e.Max} {
			whole = max(whole, len(strconv.FormatFloat(math.Trunc(math.Abs(f)), 'f', 0, 64)))
		}
		if e.Min < 0 {
			whole++
		}
		w = max(w, whole+1+d)
	}
	return w
}

// decimals is how many digits a number has after the point, written as
// briefly as it reads back.
func decimals(f float64) int {
	s := strconv.FormatFloat(math.Abs(f), 'f', -1, 64)
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return len(s) - i - 1
	}
	return 0
}
