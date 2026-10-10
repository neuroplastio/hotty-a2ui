package storybook

import (
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// TerminalColours takes into t what the terminal said of its colours, in
// answer to cells.TerminalQuery: Bubble Tea's ForegroundColorMsg and
// BackgroundColorMsg, and an OSC 4 reply, which reaches a program as an
// unknown OSC event. It reports whether msg was one, so that the program
// draws again: the cells it paints with t (theme.Theme.Term) tint and
// blend from then on (profile §3.6).
func TerminalColours(msg tea.Msg, t *theme.Terminal) bool {
	switch msg := msg.(type) {
	case tea.ForegroundColorMsg:
		t.Fg = theme.Hex(msg.Color)
	case tea.BackgroundColorMsg:
		t.Bg = theme.Hex(msg.Color)
	case uv.UnknownOscEvent:
		return t.Hear(string(msg))
	default:
		return false
	}
	return true
}
