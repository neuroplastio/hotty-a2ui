package html

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// toastHost is the toast story's surface on a host that says where the
// pointer is (SPEC §9.4: placed with v=1), its clock stopped at the time
// it returns, and pump, which plays the program: the rendition's deltas,
// and the host's events, until neither has more.
func toastHost(t *testing.T) (h *hottytest.Host, r *Rendition, now *time.Time, pump func()) {
	t.Helper()
	run, err := story.Start(story.Find("hotty/toast"))
	if err != nil {
		t.Fatal(err)
	}
	h = hottytest.New(t)
	r = New(run.Surfaces()[0].C, "toast")
	at := time.Unix(0, 0)
	r.Clock = func() time.Time { return at }
	send := func(cmds ...string) {
		t.Helper()
		for _, c := range cmds {
			_, _ = io.WriteString(h, c)
		}
		if errs := h.Errors(); len(errs) > 0 {
			t.Fatalf("the host refused: %v", errs)
		}
	}
	seen := 0
	pump = func() {
		t.Helper()
		send(r.Update()...)
		for evs := h.Events(); seen < len(evs); evs = h.Events() {
			for _, ev := range evs[seen:] {
				seen++
				if err := r.Event(ev); err != nil {
					t.Fatal(err)
				}
				send(r.Update()...)
			}
		}
	}
	send(hotty.Doc("toast", r.Doc()), hotty.Place("toast", hotty.Placement{Cols: 80, Rows: 24, Hover: true}))
	pump()
	return h, r, &at, pump
}

// The toasts are a region of their own, the document's last element, over
// the surface and the layer: a box each, the newest first, its kind's mark
// (an icon, hidden from a screen reader), its message, and its action, a
// button; an error or a warning is an alert, news and a success a status.
func TestToastsOnHost(t *testing.T) {
	h, r, now, pump := toastHost(t)
	s := h.Surface("toast")
	doc := s.HTML()
	if i, j := strings.Index(doc, `id="~o"`), strings.Index(doc, `id="~t"`); i < 0 || j < i {
		t.Fatalf("the toasts' region is not after the layer:\n%s", doc)
	}
	if !strings.Contains(doc, `<div id="~t" class="k-toasts" role="region" aria-label="Notifications">`) {
		t.Errorf("no region:\n%s", doc)
	}
	if got := s.TextOf(toastsID); got != "Indexing 1,204 filesDraft savedDisk almost full: 2 GB leftCould not reach the serverRetry" {
		t.Errorf("the toasts read %q", got)
	}
	for id, want := range map[string]string{"toast-3": "status", "toast-2": "status", "toast-1": "alert", "offline": "alert"} {
		if role, _ := s.Attr(DOMID(view.ToastID(id)), "role"); role != want {
			t.Errorf("%s: role %q, want %q", id, role, want)
		}
	}
	if !strings.Contains(doc, `<div id="hottyToast~3Aoffline" class="k-toast k-toast-error" role="alert" data-on="click"><span class="k-icon k-toast-mark" role="img" aria-label="error" aria-hidden="true">`) {
		t.Errorf("the error's mark:\n%s", doc)
	}
	if !strings.Contains(doc, `<button id="hottyToast~3Aoffline/action/0" type="button" class="k-btn k-borderless k-toast-action">Retry</button>`) {
		t.Errorf("the error's action:\n%s", doc)
	}
	// The toasts count down on the clock while the program draws, a step
	// at a time: the success goes after 5s.
	if d := r.Animating(); d != view.ToastStep {
		t.Errorf("Animating %v, want %v", d, view.ToastStep)
	}
	*now = now.Add(5 * time.Second)
	pump()
	if strings.Contains(s.TextOf(toastsID), "Draft saved") {
		t.Error("the success stayed past its time")
	}
	// The pointer on the info toast holds it, where the host says where
	// the pointer is; the warning goes on.
	must(t, h.Hover("toast", DOMID(view.ToastID("toast-3")), 70, 1))
	pump()
	*now = now.Add(5 * time.Second)
	pump()
	if got := s.TextOf(toastsID); got != "Indexing 1,204 filesCould not reach the serverRetry" {
		t.Errorf("after 10s with the pointer on the info: %q", got)
	}
	// Off it, its time goes on.
	must(t, h.Hover("toast", "", 0, 20))
	pump()
	*now = now.Add(3 * time.Second)
	pump()
	if got := s.TextOf(toastsID); got != "Could not reach the serverRetry" {
		t.Errorf("after 13s: %q", got)
	}
}

// The tooltip on a host: a row of the HottyKeyHints, over its keys, with
// the description of the element the host says the pointer is over, else
// of the one with the keyboard; a key or a click hides the pointer's
// until it goes elsewhere, as in cells.
func TestTooltipOnHost(t *testing.T) {
	h, r, _, pump := toastHost(t)
	s := h.Surface("toast")
	tip := partID("help", partTip)
	if e, ok := s.Element(tip); !ok || s.TextOf(tip) != "" {
		t.Fatalf("the tip row: %+v %v %q", e, ok, s.TextOf(tip))
	}
	must(t, h.Hover("toast", "star", 2, 9))
	pump()
	if got := s.TextOf(tip); got != "Keeps the conversation at the top of the list" {
		t.Errorf("star hovered: %q", got)
	}
	// The field's label has an id of its own; its description is the
	// field's, the nearest.
	must(t, h.Hover("toast", partID("email", partLabel), 2, 11))
	pump()
	if got := s.TextOf(tip); got != "Where receipts go; nobody else sees it" {
		t.Errorf("email's label hovered: %q", got)
	}
	// The program moving the keyboard leaves the pointer's; the user moving
	// it (Tab, a focus event) hides it, until the pointer goes elsewhere.
	r.C.Focus("mute")
	pump()
	if got := s.TextOf(tip); got != "Where receipts go; nobody else sees it" {
		t.Errorf("mute focused by the program: %q", got)
	}
	if !h.Key("Tab") {
		t.Fatal("the host did not take Tab")
	}
	pump()
	if got := s.TextOf(tip); got != "Keeps the conversation at the top of the list" {
		t.Errorf("after Tab: %q, want the star's, focused", got)
	}
	must(t, h.Hover("toast", "notify", 2, 14))
	pump()
	if got := s.TextOf(tip); got != "Toasts for builds and deploys while you work" {
		t.Errorf("notify hovered: %q", got)
	}
	// On a toast, which has none: the keyboard's.
	must(t, h.Hover("toast", DOMID(view.ToastID("offline")), 70, 10))
	pump()
	if got := s.TextOf(tip); got != "Keeps the conversation at the top of the list" {
		t.Errorf("a toast hovered: %q", got)
	}
}
