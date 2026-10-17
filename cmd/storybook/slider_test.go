package main

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go/hottytest"
)

// TestSliderDragInCells: in cells, a press on a Slider's track and a drag
// with the button held, as a terminal reports them (SGR, motion with the
// button as 32), move it with the pointer, past either end of the track,
// and the release leaves it there: a Slider in a label run (Replicas), one
// on its own (Brightness), one with no label (Volume), and each of a
// HottyRangeSlider's knobs (Price).
func TestSliderDragInCells(t *testing.T) {
	for _, c := range []struct {
		story, row string
		knob       int // which ■ of the row the press is on
		to         int // the column the drag ends at, past an end
		want       string
	}{
		{"hotty/choice", "Replicas", 0, 119, " 10"},
		{"hotty/slider", "Brightness", 0, 0, " 0"},
		{"hotty/slider", "│◂", 0, 119, " 100 ▸"},
		{"hotty/slider", "Price", 1, 119, " 20–100"},
		{"hotty/slider", "Price", 0, 0, " 0–70"},
	} {
		t.Run(c.story+"/"+c.row, func(t *testing.T) {
			h := hottytest.New(t, hottytest.Text(), hottytest.Size(120, 40))
			runBook(t, h, c.story, false)
			eventually(t, "the story in cells", func() bool { return lineOf(h.Screen(), c.row) >= 0 })
			y := lineOf(h.Screen(), c.row)
			line := func() string { return strings.TrimRight(strings.Split(h.Screen(), "\n")[y], " ") }
			x := knobAt(line(), c.knob)
			if x < 0 {
				t.Fatalf("no knob in %q", line())
			}
			at := func(x int) string { return itoa(x+1) + ";" + itoa(y+1) }
			h.Type("\x1b[<0;" + at(x) + "M")
			for col := x; col != c.to; {
				col += min(max(c.to-col, -3), 3)
				h.Type("\x1b[<32;" + at(col) + "M")
			}
			h.Type("\x1b[<0;" + at(c.to) + "m")
			eventually(t, "the value at the end, "+c.want, func() bool { return strings.HasSuffix(line(), c.want) })
			// The pointer goes back with no button held, then F1 hides the
			// sidebar: the value stayed.
			h.Type("\x1b[<35;" + at(x) + "M\x1bOP")
			eventually(t, "the sidebar hidden", func() bool { return !strings.Contains(h.Screen(), "Big text") })
			y = lineOf(h.Screen(), c.row)
			if !strings.HasSuffix(line(), c.want) {
				t.Errorf("after the release, %q", line())
			}
		})
	}
}

// knobAt is the column of a slider row's nth knob (■).
func knobAt(line string, n int) int {
	for i, r := range []rune(line) {
		if r == '■' {
			if n == 0 {
				return i
			}
			n--
		}
	}
	return -1
}
