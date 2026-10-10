package html

import (
	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// toasts is the toasts' region (profile §2, §6.23), the document's last
// element, fixed at the surface's top right corner over the surface and
// an open Modal's layer (kit.css): a toast a box, the newest first, each
// its kind's mark, its message, and its action, a button. A toast takes a
// click (data-on), which dismisses it; its action a click of its own. An
// error or a warning is an alert, which a screen reader reads at once;
// news and a success are a status, read when it is idle. The real thing,
// a surface of its own at a higher z (SPEC §5.2), is vault KIT-13h: this
// one is cut at the surface's edges.
func toasts(v *view.Surface, css string) *node {
	n := el("div", "id", toastsID, "class", "k-toasts", "role", "region", "aria-label", "Notifications")
	if css != "" {
		n.set("style", css)
	}
	for _, t := range v.Toasts {
		role := "status"
		if t.Variant == view.ToastError || t.Variant == view.ToastWarning {
			role = "alert"
		}
		box := el("div", "id", domID(t.ID), "class", "k-toast k-toast-"+t.Variant, "role", role, "data-on", "click").add(
			beside(view.ToastIcons[t.Variant]).set("class", "k-icon k-toast-mark"),
			el("span", "class", "k-toast-msg").add(texts(t.Label)...))
		for _, a := range t.Children {
			if a.Kind == view.ToastAction {
				box.add(el("button", "id", domID(a.ID), "type", "button", "class", "k-btn k-borderless k-toast-action").add(texts(a.Label)...))
			}
		}
		n.add(box)
	}
	return n
}

// hovered takes a hover (SPEC §9.4), which a host sends to a surface
// placed with v=1 as the element under the pointer changes: over a toast,
// its time waits; over another element, its description, or the nearest
// one's it is in, shows in the HottyKeyHints in place of the keyboard's
// (view.Controller.Tooltip), as cells' Hover has it. A host that does not
// say leaves the tooltip the keyboard's.
func (r *Rendition) hovered(ev hotty.Event) {
	r.hover, r.held = "", ""
	if h, ok := ev.Hover(); ok && !h.Out {
		if id, _, ok := viewID(ev.Target); ok {
			switch e := r.C.V.Find(id); {
			case e == nil:
			case e.Kind == view.Toast || e.Kind == view.ToastAction:
				r.held = e.Name
			default:
				r.hover = id
			}
		}
	}
	if r.hover != r.quiet {
		r.quiet = ""
	}
	if r.quiet != "" {
		r.hover = ""
	}
}

// hush hides the hovered description, as a key or a click hides a
// tooltip, until the pointer leaves the element it is on: the host says
// so only when it goes to another.
func (r *Rendition) hush() {
	if r.hover != "" {
		r.quiet, r.hover = r.hover, ""
	}
}

// heldToast reports whether the pointer is on a toast (hovered), whose
// time then waits (view.Controller.TickToasts).
func (r *Rendition) heldToast(id string) bool { return id != "" && id == r.held }
