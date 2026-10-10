package view

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// Toasts and tooltips (profile §6.23) in the view. A toast is a notice the
// renderer shows over its surface, in a corner, for a while: hottyToast
// shows one, hottyDismissToast takes it away. Toasts are the surface's
// state, not its components' (A2UI has no time, vault a2ui-limits L6), so
// each rendition counts their time on its own clock as it draws them
// (TickToasts), as it does a Spinner's frames. A tooltip is a component's
// accessibility.description, which a HottyKeyHints shows (Tooltip).

// The kinds of toast: what its glyph and its colour say.
const (
	ToastInfo    = "info"
	ToastSuccess = "success"
	ToastWarning = "warning"
	ToastError   = "error"
)

// ToastIcons are the kinds' marks: the basic catalog's icons for them,
// which a host draws and cells shows as their glyphs (icons.Glyph), so
// that the kind reads without colour.
var ToastIcons = map[string]string{
	ToastInfo:    "info",
	ToastSuccess: "check",
	ToastWarning: "warning",
	ToastError:   "error",
}

const (
	// ToastTimeout is how long a toast shows when its call gives no
	// timeout: Textual's, five seconds.
	ToastTimeout = 5 * time.Second
	// ToastStep is how often a rendition draws again while a toast counts
	// down (Animating): one goes within a step of its time.
	ToastStep = time.Second / 4
	// MaxToasts is how many toasts a surface shows at once: a newer one
	// takes the oldest's place.
	MaxToasts = 5
)

// ToastState is a toast as its surface keeps it (hottyToast).
type ToastState struct {
	// ID names it: the caller's, or one the renderer made ("toast-1"). A
	// toast shown with the ID of one there takes its place.
	ID      string
	Message string
	// Kind is info, success, warning or error.
	Kind string
	// Timeout is how long it shows; 0 until it is dismissed.
	Timeout time.Duration
	// Label is its action's label, and Event what the action sends the
	// agent, its context read when the toast showed; nil without one.
	Label string
	Event *a2ui.ActionMessage

	// shown is how long it has shown, counted at the renditions' draws
	// (TickToasts); last is the last of those, zero before the first, and
	// held whether it was held then, its time not counting.
	shown time.Duration
	last  time.Time
	held  bool
}

// ToastID is the view id of a toast's element: a prefix no component of
// the kit's has, then its ID.
func ToastID(id string) string { return toastPrefix + id }

const toastPrefix = "hottyToast:"

// ToastActionID is the view id of a toast's action.
func ToastActionID(id string) string { return SubID(ToastID(id), "action", 0) }

// mapToasts makes the toasts' elements, the newest first, as they stack
// from the corner: a Toast (Label its message, Variant its kind, Name its
// ID), holding its action, a ToastAction (Label its label, Name the
// toast's ID), when it has one.
func mapToasts(st *State) []*Element {
	var out []*Element
	for i := len(st.Toasts) - 1; i >= 0; i-- {
		t := st.Toasts[i]
		e := &Element{ID: ToastID(t.ID), Kind: Toast, Type: "hottyToast", Label: t.Message, Variant: t.Kind, Name: t.ID}
		if t.Event != nil {
			e.Children = []*Element{{ID: ToastActionID(t.ID), Kind: ToastAction, Type: "hottyToast.action", Label: t.Label, Name: t.ID}}
		}
		out = append(out, e)
	}
	return out
}

// ShowToast shows a toast from hottyToast's arguments, resolved: message,
// kind, timeout (milliseconds), id, actionLabel and action ({"event": …},
// its context resolved already). It takes the place of a toast with the
// same id, and its time starts again; else it goes on the stack, nearest
// the corner, and the oldest goes once there are MaxToasts. It returns the
// toast's id.
func (c *Controller) ShowToast(args map[string]any) (string, error) {
	t := &ToastState{Message: a2ui.ToString(args["message"]), Kind: ToastInfo, Timeout: ToastTimeout, Label: a2ui.ToString(args["actionLabel"])}
	switch k, _ := args["kind"].(string); k {
	case ToastSuccess, ToastWarning, ToastError:
		t.Kind = k
	}
	if v, ok := args["timeout"]; ok && v != nil {
		ms := a2ui.ToNumber(v)
		if math.IsNaN(ms) || ms < 0 {
			return "", fmt.Errorf("a toast's timeout is a number of milliseconds, not %v", v)
		}
		t.Timeout = time.Duration(ms * float64(time.Millisecond))
	}
	t.ID = a2ui.ToString(args["id"])
	if t.ID == "" {
		t.ID = c.newToastID()
	}
	if a, ok := args["action"].(map[string]any); ok {
		ev, _ := a["event"].(map[string]any)
		name, _ := ev["name"].(string)
		if strings.TrimSpace(name) == "" {
			return "", fmt.Errorf("a toast's action is an event with a name")
		}
		if t.Label == "" {
			return "", fmt.Errorf("a toast's action needs an actionLabel")
		}
		// Its context was read as the toast showed: it names what the toast
		// is about, wherever the data has moved since.
		t.Event = &a2ui.ActionMessage{Name: name, SurfaceID: c.S.ID, SourceComponentID: t.ID, Context: map[string]any{}}
		if ctx, ok := ev["context"].(map[string]any); ok {
			t.Event.Context = ctx
		}
		if um, ok := ev["userMessage"].(string); ok {
			t.Event.UserMessage = um
		}
	} else {
		t.Label = ""
	}
	if i := slices.IndexFunc(c.St.Toasts, func(x *ToastState) bool { return x.ID == t.ID }); i >= 0 {
		c.St.Toasts[i] = t
	} else {
		c.St.Toasts = append(c.St.Toasts, t)
		if n := len(c.St.Toasts) - MaxToasts; n > 0 {
			c.leave(c.St.Toasts[:n]...)
			c.St.Toasts = slices.Delete(c.St.Toasts, 0, n)
		}
	}
	c.Rebuild()
	return t.ID, nil
}

// newToastID is an id for a toast its caller named none: "toast-" and a
// number, one no toast there has.
func (c *Controller) newToastID() string {
	for {
		c.St.ToastSeq++
		id := "toast-" + strconv.Itoa(c.St.ToastSeq)
		if !slices.ContainsFunc(c.St.Toasts, func(t *ToastState) bool { return t.ID == id }) {
			return id
		}
	}
}

// DismissToast takes a toast away; ok is false when none has the id,
// gone already. A toast whose action had the keyboard takes it from the
// surface, as a host's focus goes with the element it was on.
func (c *Controller) DismissToast(id string) (ok bool) {
	i := slices.IndexFunc(c.St.Toasts, func(t *ToastState) bool { return t.ID == id })
	if i < 0 {
		return false
	}
	c.leave(c.St.Toasts[i])
	c.St.Toasts = slices.Delete(c.St.Toasts, i, i+1)
	c.Rebuild()
	return true
}

// DismissNewest takes away the toast whose action has the keyboard, else
// the newest: what Escape does once nothing else took it. ok is false when
// there is none.
func (c *Controller) DismissNewest() (ok bool) {
	if t := c.focusedToast(); t != nil {
		return c.DismissToast(t.ID)
	}
	if n := len(c.St.Toasts); n > 0 {
		return c.DismissToast(c.St.Toasts[n-1].ID)
	}
	return false
}

// PickToast is the user picking a toast's action: its event goes to the
// agent, stamped now, and the toast goes.
func (c *Controller) PickToast(id string) {
	i := slices.IndexFunc(c.St.Toasts, func(t *ToastState) bool { return t.ID == id })
	if i < 0 {
		return
	}
	if ev := c.St.Toasts[i].Event; ev != nil && c.S.Env().Action != nil {
		a := *ev
		a.Timestamp = a2ui.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		a.Context, _ = a2ui.Clone(ev.Context).(map[string]any)
		c.S.Env().Action(c.S, &a)
	}
	c.DismissToast(id)
}

// TickToasts counts the toasts' time at a rendition's draw, on its clock,
// before it draws them: since the draw before, each one not held then has
// shown that much longer, and one that has shown its timeout goes. A
// toast is held while the keyboard is on its action, or while held says
// so (the pointer on it, in cells), and not counted from then until the
// next draw. It returns how soon to draw again: ToastStep while a toast
// counts down, 0 while none does.
func (c *Controller) TickToasts(now time.Time, held func(id string) bool) time.Duration {
	focused := c.focusedToast()
	var gone []*ToastState
	next := time.Duration(0)
	for _, t := range c.St.Toasts {
		if !t.last.IsZero() && !t.held && now.After(t.last) {
			t.shown += now.Sub(t.last)
		}
		t.last, t.held = now, t == focused || held != nil && held(t.ID)
		switch {
		case t.Timeout <= 0:
		case t.shown >= t.Timeout:
			gone = append(gone, t)
		case !t.held:
			next = ToastStep
		}
	}
	if len(gone) > 0 {
		c.leave(gone...)
		c.St.Toasts = slices.DeleteFunc(c.St.Toasts, func(t *ToastState) bool { return slices.Contains(gone, t) })
		c.Rebuild()
	}
	return next
}

// focusedToast is the toast whose action has the keyboard, or nil.
func (c *Controller) focusedToast() *ToastState {
	if !c.St.Keyboard || !strings.HasPrefix(c.St.Focus, toastPrefix) {
		return nil
	}
	for _, t := range c.St.Toasts {
		if c.St.Focus == ToastActionID(t.ID) {
			return t
		}
	}
	return nil
}

// leave takes the keyboard from the surface when it is on one of the
// toasts' actions, which are going.
func (c *Controller) leave(ts ...*ToastState) {
	if f := c.focusedToast(); f != nil && slices.Contains(ts, f) {
		c.Focus("")
	}
}

// Tooltip is the description to show in a status line (a HottyKeyHints):
// that of the element at hover, the pointer's (a rendition's; "" for
// none), or else that of the element with the keyboard; each the nearest
// accessibility.description from the element outward, so that a knob
// shows its slider's and a tab its Tabs'. "" when there is none.
func (c *Controller) Tooltip(hover string) string {
	if d := c.Description(hover); d != "" {
		return d
	}
	if c.St.Keyboard {
		return c.Description(c.St.Focus)
	}
	return ""
}

// Description is the accessibility.description nearest an element: its
// own, else that of the nearest element it is in with one; "" for none.
func (c *Controller) Description(id string) string {
	if id == "" {
		return ""
	}
	e := c.V.Find(id)
	if e == nil {
		return ""
	}
	if e.A11y.Description != "" {
		return e.A11y.Description
	}
	anc := c.Ancestors(id)
	for i := len(anc) - 1; i >= 0; i-- {
		if d := anc[i].A11y.Description; d != "" {
			return d
		}
	}
	return ""
}

// Described reports whether any element of the surface has a
// description: a HottyKeyHints then keeps a row for it, so that nothing
// moves as the keyboard does.
func (s *Surface) Described() bool {
	found := false
	s.Walk(func(e *Element) bool {
		found = e.A11y.Description != ""
		return !found
	})
	return found
}
