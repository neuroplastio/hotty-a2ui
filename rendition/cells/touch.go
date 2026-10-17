package cells

import (
	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// TouchAreas are the cells of the last frame Draw returned where a finger
// drags as the mouse does (SPEC §9.6): the track of each Slider and
// HottyRangeSlider that is not disabled, as much of it as is drawn, which
// drags across ("x"). A press there and a drag move it (Click, Drag), and
// a tap is a click. They are the frame's cells, from its top left, in the
// order painted; under an open Modal there are none.
//
// A program on a HOTTY host that has the terminal report the mouse names
// them, moved to where it shows the frame (hotty.Touch), so that a finger
// drags a slider in cells as the mouse does, where it would scroll the
// terminal otherwise.
func (r *Rendition) TouchAreas() []hotty.TouchArea {
	var out []hotty.TouchArea
	for _, h := range r.hits {
		t := h.track
		if t == nil {
			continue
		}
		// A knob's own hit is a cell of its range's track.
		e := r.c.V.Find(h.id)
		if e == nil || e.Disabled || e.Kind != view.Slider && e.Kind != view.RangeSlider {
			continue
		}
		// A scroll view cuts the hit to what it shows (clipHits), and the
		// track with it.
		x0, x1 := max(h.x, t.x), min(h.x+h.w, t.x+t.n)
		if x1 > x0 && h.h > 0 {
			out = append(out, hotty.TouchArea{C: x0, R: h.y, W: x1 - x0, H: h.h, Drag: "x"})
		}
	}
	return out
}
