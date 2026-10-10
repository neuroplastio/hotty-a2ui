package storybook

import (
	"image/color"
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// TerminalColours takes the terminal's answers as Bubble Tea delivers
// them, and nothing else.
func TestTerminalColours(t *testing.T) {
	var term theme.Terminal
	for _, msg := range []tea.Msg{
		tea.ForegroundColorMsg{Color: color.RGBA{R: 0xcd, G: 0xd6, B: 0xf4, A: 0xff}},
		tea.BackgroundColorMsg{Color: color.RGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff}},
		uv.UnknownOscEvent("\x1b]4;12;rgb:8989/b4b4/fafa\x1b\\"),
	} {
		if !TerminalColours(msg, &term) {
			t.Errorf("did not take %#v", msg)
		}
	}
	if term.Fg != "#cdd6f4" || term.Bg != "#1e1e2e" || term.ANSI[12] != "#89b4fa" {
		t.Errorf("heard %+v", term)
	}
	if TerminalColours(tea.KeyPressMsg{Code: 'q', Text: "q"}, &term) || TerminalColours(uv.UnknownOscEvent("\x1b]7279;x\x1b\\"), &term) {
		t.Error("took a message that is not a colour")
	}
}
