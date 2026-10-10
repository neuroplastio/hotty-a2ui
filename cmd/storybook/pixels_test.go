package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/neuroplastio/hotty-a2ui/storybook"
)

// raw is what a command writes to the terminal, "" for none.
func raw(cmd tea.Cmd) string {
	if cmd == nil {
		return ""
	}
	if m, ok := cmd().(tea.RawMsg); ok {
		s, _ := m.Msg.(string)
		return s
	}
	return ""
}

// The pointer goes in pixels once the terminal said it knows SGR-Pixels
// and how big its cells are, in either order; reports go back into cells,
// a drag's with how far down its cell. Until then they stay as they come.
func TestPixels(t *testing.T) {
	var p pixels
	if got := raw(p.ask()); got != ansi.RequestModeMouseExtSgrPixel+"\x1b[16t" {
		t.Errorf("ask: %q", got)
	}
	click := tea.MouseClickMsg{X: 25, Y: 41, Button: tea.MouseLeft}
	if m, _ := p.update(click); m != click {
		t.Errorf("before the answers: %v", m)
	}
	if m, cmd := p.update(uv.CellSizeEvent{Width: 10, Height: 20}); m != nil || cmd != nil {
		t.Errorf("the cell's size alone: %v, %q", m, raw(cmd))
	}
	m, cmd := p.update(tea.ModeReportMsg{Mode: ansi.ModeMouseExtSgrPixel, Value: ansi.ModeReset})
	if m != nil || raw(cmd) != ansi.SetModeMouseExtSgrPixel {
		t.Fatalf("both answers: %v, %q", m, raw(cmd))
	}
	if m, _ := p.update(click); m != (tea.MouseClickMsg{X: 2, Y: 2, Button: tea.MouseLeft}) {
		t.Errorf("a click at (25, 41) px: %v", m)
	}
	m, _ = p.update(tea.MouseMotionMsg{X: 25, Y: 54, Button: tea.MouseLeft})
	if d, ok := m.(storybook.DragMsg); !ok || d.X != 2 || d.Y != 2 || d.Sub < 0.7 || d.Sub > 0.75 {
		t.Errorf("a drag at (25, 54) px: %#v", m)
	}
	if raw(p.off()) != ansi.ResetModeMouseExtSgrPixel {
		t.Errorf("off: %q", raw(p.off()))
	}
	var q pixels
	q.update(tea.ModeReportMsg{Mode: ansi.ModeMouseExtSgrPixel, Value: ansi.ModeNotRecognized})
	if _, cmd := q.update(uv.CellSizeEvent{Width: 10, Height: 20}); cmd != nil || q.off() != nil {
		t.Errorf("a terminal without the mode: %q", raw(cmd))
	}
}
