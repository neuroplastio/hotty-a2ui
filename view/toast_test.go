package view_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// toastRun is a run of a surface whose buttons show toasts, and the
// actions the agent got.
func toastRun(t *testing.T) (*story.Run, *view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	run := story.NewRun()
	var got []*a2ui.ActionMessage
	run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			got = append(got, o.Action)
		}
	}
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"items":[{"id":"a1","name":"Groceries"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["rows","mute"]},
	 {"id":"rows","component":"List","children":{"componentId":"del","path":"/items"}},
	 {"id":"del","component":"Button","child":"del_t","action":{"functionCall":{"@call":"hottyToast",` + h + `,"args":{
	   "message":{"@path":"name"},"kind":"success",
	   "actionLabel":"Undo","action":{"event":{"name":"undo","context":{"id":{"@path":"id"}}}}}}}},
	 {"id":"del_t","component":"Text","text":"Archive"},
	 {"id":"mute","component":"Button","child":"mute_i","accessibility":{"label":"Mute","description":"Silences this conversation"},
	  "action":{"event":{"name":"mute"}}},
	 {"id":"mute_i","component":"Icon","name":"volumeOff"}]}}]`
	if err := run.Feed(json.RawMessage(msgs)); err != nil {
		t.Fatal(err)
	}
	return run, run.Surfaces()[0].C, &got
}

// agentToast is the agent's callRendererFunction of hottyToast.
func agentToast(t *testing.T, run *story.Run, args string) {
	t.Helper()
	m := `{"version":"v1.0","callRendererFunction":{"functionCallId":"c","callFunction":{"@call":"hottyToast","catalogId":"` + hotty.ID + `","args":` + args + `}}}`
	if err := run.Feed(json.RawMessage(m)); err != nil {
		t.Fatal(err)
	}
}

// toastIDs are the toasts the view shows, the newest first.
func toastIDs(c *view.Controller) []string {
	var out []string
	for _, e := range c.V.Toasts {
		out = append(out, e.Name+":"+e.Variant+":"+e.Label)
	}
	return out
}

// A Button's hottyToast shows a toast whose message and action's context
// are read as it shows, in the Button's scope; the agent's shows one on the
// surface, read in its data model. They stack, the newest first, and a
// toast with an id there takes its place.
func TestToastShow(t *testing.T) {
	run, c, got := toastRun(t)
	del := c.FindComponent("del", a2ui.Scope{Path: "/items/0"})
	if err := c.Activate(del); err != nil {
		t.Fatal(err)
	}
	agentToast(t, run, `{"message":{"@path":"/items/0/name"},"kind":"warning","id":"w","timeout":0}`)
	agentToast(t, run, `{"message":"Hello","kind":"nonsense"}`) // refused: no such kind
	agentToast(t, run, `{"message":"Hello"}`)
	if ids := toastIDs(c); !slices.Equal(ids, []string{"toast-2:info:Hello", "w:warning:Groceries", "toast-1:success:Groceries"}) {
		t.Fatalf("toasts %q", ids)
	}
	a := c.V.Toasts[2].Children
	if len(a) != 1 || a[0].Kind != view.ToastAction || a[0].Label != "Undo" || !a[0].Focusable() {
		t.Fatalf("the action: %+v", a)
	}
	// The data moves on: the action still names what was archived.
	if err := c.S.Write("/items", []any{}); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if err := c.Activate(view.ToastActionID("toast-1")); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || (*got)[0].Name != "undo" || (*got)[0].Context["id"] != "a1" || (*got)[0].SourceComponentID != "toast-1" || (*got)[0].SurfaceID != "s" {
		t.Fatalf("the action sent %+v", *got)
	}
	if ids := toastIDs(c); len(ids) != 2 {
		t.Errorf("a picked toast stays: %q", ids)
	}
	agentToast(t, run, `{"message":"Again","id":"w","kind":"error"}`)
	if ids := toastIDs(c); !slices.Equal(ids, []string{"toast-2:info:Hello", "w:error:Again"}) {
		t.Errorf("replaced in its place: %q", ids)
	}
	for i := range view.MaxToasts {
		agentToast(t, run, `{"message":"n`+string(rune('0'+i))+`"}`)
	}
	if ids := toastIDs(c); len(ids) != view.MaxToasts || !strings.HasSuffix(ids[len(ids)-1], ":n0") {
		t.Errorf("past the most, the oldest go: %q", ids)
	}
}

// A toast's time counts at the draws, on the drawing rendition's clock,
// from the first that draws it, and not while the keyboard is on its
// action or the rendition holds it; one that has shown its timeout goes.
// A held toast's time counts again from the first draw that does not
// hold it.
func TestToastTime(t *testing.T) {
	run, c, _ := toastRun(t)
	agentToast(t, run, `{"message":"Short","id":"short","timeout":1000}`)
	agentToast(t, run, `{"message":"Long","id":"long","timeout":10000,"actionLabel":"View","action":{"event":{"name":"view"}}}`)
	agentToast(t, run, `{"message":"Sticky","id":"sticky","timeout":0}`)
	t0 := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	at := func(ms int) time.Time { return t0.Add(time.Duration(ms) * time.Millisecond) }
	if next := c.TickToasts(at(0), nil); next != view.ToastStep {
		t.Errorf("counting down, next %v", next)
	}
	c.TickToasts(at(900), nil)
	held := func(id string) bool { return id == "short" }
	c.TickToasts(at(950), held)
	c.TickToasts(at(5000), held)
	if ids := toastIDs(c); len(ids) != 3 {
		t.Fatalf("short held at 950 ms: %q", ids)
	}
	c.Focus(view.ToastActionID("long"))
	c.TickToasts(at(5100), nil)
	if ids := toastIDs(c); len(ids) != 3 {
		t.Fatalf("let go at 5100 ms, short has shown 950: %q", ids)
	}
	c.TickToasts(at(5200), nil)
	if ids := toastIDs(c); !slices.Equal(ids, []string{"sticky:info:Sticky", "long:info:Long"}) {
		t.Fatalf("at 5200 ms short has shown 1050: %q", ids)
	}
	c.TickToasts(at(30000), nil)
	if len(c.V.Toasts) != 2 {
		t.Fatalf("the keyboard on its action holds long: %q", toastIDs(c))
	}
	c.Focus("")
	c.TickToasts(at(30000), nil)
	if c.TickToasts(at(34000), nil); len(c.V.Toasts) != 2 {
		t.Fatalf("long has shown 9100 ms: %q", toastIDs(c))
	}
	if next := c.TickToasts(at(35000), nil); next != 0 || !slices.Equal(toastIDs(c), []string{"sticky:info:Sticky"}) {
		t.Errorf("long gone: %q, next %v", toastIDs(c), next)
	}
	if !c.DismissNewest() || len(c.V.Toasts) != 0 || c.DismissNewest() {
		t.Errorf("Escape's dismissal: %q", toastIDs(c))
	}
}

// The toasts' actions are the last Tab stops, the newest first; Escape
// dismisses the toast whose action has the keyboard, and the keyboard
// leaves the surface with it.
func TestToastKeys(t *testing.T) {
	run, c, _ := toastRun(t)
	for _, id := range []string{"a", "b"} {
		agentToast(t, run, `{"message":"`+id+`","id":"`+id+`","actionLabel":"Open","action":{"event":{"name":"open"}}}`)
	}
	f := c.V.Focusables()
	if want := []string{view.ToastActionID("b"), view.ToastActionID("a")}; !slices.Equal(f[len(f)-2:], want) {
		t.Fatalf("Tab stops %q", f)
	}
	c.Focus(view.ToastActionID("a"))
	if short, _ := c.KeyHints("", false); len(short) < 2 || short[0] != (view.Hint{Key: "enter", Desc: "open"}) || short[1] != (view.Hint{Key: "esc", Desc: "dismiss"}) {
		t.Errorf("hints %v", short)
	}
	if !c.DismissNewest() || !slices.Equal(toastIDs(c), []string{"b:info:b"}) || c.St.Keyboard {
		t.Errorf("Escape on a's action: %q, keyboard %v", toastIDs(c), c.St.Keyboard)
	}
	if err := c.Activate(view.ToastID("b")); err != nil || len(c.V.Toasts) != 0 {
		t.Errorf("a click on a toast dismisses it: %q", toastIDs(c))
	}
	m := `{"version":"v1.0","callRendererFunction":{"functionCallId":"d","callFunction":{"@call":"hottyDismissToast","catalogId":"` + hotty.ID + `","args":{"id":"gone"}}}}`
	if err := run.Feed(json.RawMessage(m)); err != nil {
		t.Errorf("dismissing a toast gone: %v", err)
	}
}

// A tooltip is the nearest accessibility.description: the hovered
// element's, else the focused one's.
func TestTooltip(t *testing.T) {
	_, c, _ := toastRun(t)
	if !c.V.Described() {
		t.Error("the surface has a description")
	}
	if got := c.Tooltip(""); got != "" {
		t.Errorf("no keyboard, no hover: %q", got)
	}
	c.Focus("mute")
	if got := c.Tooltip(""); got != "Silences this conversation" {
		t.Errorf("focused: %q", got)
	}
	if got := c.Tooltip("mute_i"); got != "Silences this conversation" {
		t.Errorf("its icon, hovered: %q", got)
	}
	del := c.FindComponent("del", a2ui.Scope{Path: "/items/0"})
	c.Focus(del)
	if got := c.Tooltip("mute"); got != "Silences this conversation" {
		t.Errorf("hover over focus: %q", got)
	}
	if got := c.Tooltip(""); got != "" {
		t.Errorf("no description: %q", got)
	}
}
