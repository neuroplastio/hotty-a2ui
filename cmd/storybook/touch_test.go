package main

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"
)

// TestTouchUntilQuit: on a host, the storybook names no touch areas while
// the story is in surfaces; in cells, each Slider's track, where a finger
// drags as the mouse does (SPEC §9.6), since it reads the mouse; and when
// it quits, it takes them back, with its surfaces' deletes. Bubble Tea
// writes those after it leaves the alternate screen, which drops the
// screen's areas anyway, so none is left on either screen.
func TestTouchUntilQuit(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	runBook(t, h, "hotty/slider", false)
	eventually(t, "the story on the host", func() bool {
		s := h.Surface("s1-0-h")
		return s != nil && strings.Contains(s.Text(), "Brightness")
	})
	if a := h.TouchAreas(); a != nil {
		t.Errorf("areas with the story in surfaces: %+v", a)
	}
	h.Type("\x1bOQ") // F2: cells
	eventually(t, "the sliders' areas", func() bool { return len(h.TouchAreas()) == 5 })
	h.Type("\x11") // ctrl+q
	eventually(t, "the storybook's surfaces deleted", func() bool { return h.Surface(navID) == nil })
	out := h.Output()
	took, del := strings.LastIndex(out, hotty.Touch(nil)), strings.LastIndex(out, hotty.Del(navID))
	if took < 0 || took > del {
		t.Errorf("no a=touch with no areas before the deletes: at %d, the deletes at %d", took, del)
	}
	if a := h.TouchAreas(); a != nil {
		t.Errorf("areas after quitting: %+v", a)
	}
}
