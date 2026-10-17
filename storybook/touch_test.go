package storybook

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytea"
	"github.com/neuroplastio/hotty-go/hottytest"
)

// touches are the a=touch commands the program sent, in order.
func touches(h *hottytest.Host) []hotty.Message {
	var out []hotty.Message
	for _, m := range h.Commands() {
		if hotty.Get(m.Control, "a") == "touch" {
			out = append(out, m)
		}
	}
	return out
}

// offTrack says why an area is not a Slider's track on the screen, or ""
// for none: each is a row of ━, ⎯ and ■, and no more of it, from col on.
// The host hears of the areas as the frame is drawn, so a test waits for
// them to agree.
func offTrack(h *hottytest.Host, areas []hotty.TouchArea, col int) string {
	rows := strings.Split(h.Screen(), "\n")
	track := func(r rune) bool { return strings.ContainsRune("━⎯■", r) }
	for _, a := range areas {
		if a.H != 1 || a.Drag != "x" || a.C < col || a.R >= len(rows) {
			return fmt.Sprintf("area %+v", a)
		}
		row := []rune(rows[a.R])
		if a.C+a.W > len(row) {
			return fmt.Sprintf("area %+v past its row %q", a, rows[a.R])
		}
		cells := string(row[a.C : a.C+a.W])
		if strings.IndexFunc(cells, func(r rune) bool { return !track(r) }) >= 0 || !strings.Contains(cells, "■") {
			return fmt.Sprintf("area %+v is %q, not a track", a, cells)
		}
		if a.C > 0 && track(row[a.C-1]) || a.C+a.W < len(row) && track(row[a.C+a.W]) {
			return fmt.Sprintf("area %+v: the track goes on past it in %q", a, rows[a.R])
		}
	}
	return ""
}

// tracks waits for five areas, the sliders', that are their tracks on the
// screen, from col on, and for which more holds; and returns them.
func tracks(t *testing.T, h *hottytest.Host, what string, col int, more func([]hotty.TouchArea) bool) []hotty.TouchArea {
	t.Helper()
	var a []hotty.TouchArea
	why := "no areas"
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(5 * time.Millisecond) {
		a = h.TouchAreas()
		if len(a) != 5 {
			why = fmt.Sprintf("%d areas", len(a))
			continue
		}
		if why = offTrack(h, a, col); why == "" && (more == nil || more(a)) {
			return a
		}
	}
	t.Fatalf("never: %s: %s\n%s", what, why, h.Screen())
	return nil
}

// TestTouchAreas: on a host, each Slider the Book draws in cells is cells
// where a finger drags as the mouse does (SPEC §9.6). The Book names its
// track on the screen, under the program's own row; again where the
// layout moves it (the sidebar hidden), the screen changes size, or is
// drawn anew; and none once the story is text or surfaces, where it tells
// the host so. Side by side, the cells column's.
func TestTouchAreas(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	p := run(t, h, Options{First: "hotty/slider", Rendition: rendCells, Mouse: true})
	// Brightness, Speed, Minimum rating, Price and Volume.
	first := tracks(t, h, "the sliders' areas", 0, nil)
	if first[0].R < 1 {
		t.Errorf("an area on the program's own row: %+v", first[0])
	}

	h.Type("\x1bOP") // F1: the sidebar hidden, the tracks longer and further left
	moved := tracks(t, h, "the areas moved", 0, func(a []hotty.TouchArea) bool {
		return a[0].C < first[0].C && a[0].W > first[0].W
	})

	p.Send(tea.WindowSizeMsg{Width: 100, Height: 40})
	tracks(t, h, "the areas at 100 columns", 0, func(a []hotty.TouchArea) bool {
		return !slices.Equal(a, moved) && a[0].C+a[0].W < 100
	})

	// The screen drawn anew: the areas named again (TestTouchAgain has
	// them the same).
	n := len(touches(h))
	p.Send(hottytea.RelayoutMsg{})
	tracks(t, h, "the areas named again", 0, func([]hotty.TouchArea) bool { return len(touches(h)) > n })

	h.Type("\x1bOQ") // F2: text
	eventually(t, "no areas in text", func() bool { return h.TouchAreas() == nil })
	if ts := touches(h); len(ts[len(ts)-1].Payload) != 0 {
		t.Errorf("the last a=touch has %d bytes, want none", len(ts[len(ts)-1].Payload))
	}

	h.Type("\x1bOQ") // F2: side by side, surfaces then cells
	tracks(t, h, "the cells column's areas", 50, nil)

	h.Type("\x1bOQ") // F2: surfaces
	eventually(t, "no areas with surfaces", func() bool { return h.TouchAreas() == nil })
}

// TestNoTouchAreas: the Book names no touch areas where the program does
// not read the mouse (no Options.Mouse), nor in a terminal that is no
// host.
func TestNoTouchAreas(t *testing.T) {
	for _, c := range []struct {
		name string
		opts []hottytest.Option
		o    Options
	}{
		{"no mouse", nil, Options{First: "hotty/slider", Rendition: rendCells}},
		{"no host", []hottytest.Option{hottytest.Text()}, Options{First: "hotty/slider", Mouse: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := hottytest.New(t, append(c.opts, hottytest.Size(120, 40))...)
			run(t, h, c.o)
			eventually(t, "the sliders in cells", func() bool { return strings.Contains(h.Screen(), "Brightness") })
			h.Type("\x1bOP") // F1: another frame, so the first's commands are out
			eventually(t, "the sidebar hidden", func() bool { return !strings.Contains(h.Screen(), "Big text") })
			if ts := touches(h); len(ts) > 0 {
				t.Errorf("%d a=touch sent", len(ts))
			}
		})
	}
}

// TestTouchAgain: the Book names the touch areas when they change, and
// the same ones again only when the screen may have dropped them (a
// relayout, a resize, a resume); none when it has none to take back, and
// at most hotty.TouchMax. Quit takes them back, once.
func TestTouchAgain(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Native
	b := New(Options{Mouse: true})
	sent := func() string {
		if c := s.Flush(); c != nil {
			if m, ok := c().(tea.RawMsg); ok {
				return m.Msg.(string)
			}
		}
		return ""
	}
	a := []hotty.TouchArea{{C: 3, R: 4, W: 10, H: 1, Drag: "x"}}
	for i, c := range []struct {
		do   func()
		want string
	}{
		{func() { b.touch(s, a) }, hotty.Touch(a)},
		{func() { b.touch(s, a) }, ""},
		{func() { b.Update(hottytea.RelayoutMsg{}, s); b.touch(s, a) }, hotty.Touch(a)},
		{func() { b.Update(tea.ResumeMsg{}, s); b.touch(s, a) }, hotty.Touch(a)},
		{func() { b.Update(tea.WindowSizeMsg{Width: 80, Height: 24}, s); b.touch(s, a) }, hotty.Touch(a)},
		{func() { b.touch(s, nil) }, hotty.Touch(nil)},
		{func() { b.touch(s, nil) }, ""},
		{func() { b.Update(hottytea.RelayoutMsg{}, s); b.touch(s, []hotty.TouchArea{}) }, ""},
		{func() { b.Quit(s) }, ""},
		{func() { b.touch(s, a); b.Quit(s) }, hotty.Touch(a) + hotty.Touch(nil)},
		{func() { b.Quit(s) }, ""},
		{func() { b.touch(s, slices.Repeat(a, hotty.TouchMax+1)) }, hotty.Touch(slices.Repeat(a, hotty.TouchMax))},
	} {
		c.do()
		if got := sent(); got != c.want {
			t.Errorf("%d: sent %q, want %q", i, got, c.want)
		}
	}
	// Without the mouse, none.
	b = New(Options{})
	if b.touch(s, a); sent() != "" {
		t.Error("areas sent without Options.Mouse")
	}
}
