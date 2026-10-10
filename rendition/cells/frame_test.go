package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// TerminalQuery asks for the text's and the background's colours, and
// for each colour of the floor's and the two magentas by number, once.
func TestTerminalQuery(t *testing.T) {
	q := TerminalQuery()
	for _, ask := range []string{"\x1b]10;?\x1b\\", "\x1b]11;?\x1b\\", "\x1b]4;1;?", "\x1b]4;2;?", "\x1b]4;3;?", "\x1b]4;5;?", "\x1b]4;6;?", "\x1b]4;8;?", "\x1b]4;12;?", "\x1b]4;13;?"} {
		if n := strings.Count(q, ask); n != 1 {
			t.Errorf("%q asked %d times in %q", ask, n, q)
		}
	}
	if n := strings.Count(q, "\x1b]4;"); n != 8 {
		t.Errorf("%d OSC 4 questions, want 8", n)
	}
}

// Where the theme leaves the roles to the terminal and the terminal said
// its colours, a tint paints a background on the cells it changes, and
// nowhere else, so a transparent terminal stays so: a surface a little
// toward the text, a selection a fifth toward the accent, a role's tint
// toward its colour. Until it says them, nothing is tinted.
func TestTerminalTints(t *testing.T) {
	term := theme.Default
	term.Term = theme.Terminal{Fg: "#ffffff", Bg: "#000000"}
	term.Term.ANSI[1], term.Term.ANSI[12] = "#ff0000", "#0000ff"
	for _, c := range []struct {
		name string
		cell Cell
		want string
	}{
		{"a surface", Cell{Text: " ", Width: 1, Back: Surface, BackMix: 255}, "48;2;24;24;24"},
		{"a selection", Cell{Text: "a", Width: 1, Back: Selection, BackMix: 255}, "48;2;0;0;51"},
		{"an error's sixth", Cell{Text: "a", Width: 1, Role: Error, Back: Error, BackMix: 42}, "31;48;2;42;0;0"},
		{"no tint", Cell{Text: "a", Width: 1, Role: Error}, "31"},
	} {
		if got := c.cell.style(&term); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
		if got := c.cell.style(&theme.Default); strings.Contains(got, "48;") {
			t.Errorf("%s, the terminal's colours unknown: %q", c.name, got)
		}
	}
	// A row keeps a tinted blank at its end, and leaves out a plain one.
	f := newFrame(4, 1)
	f.Cells[0][0] = Cell{Text: "a", Width: 1}
	f.Cells[0][2] = Cell{Text: " ", Width: 1, Back: Surface, BackMix: 255}
	if got := f.Themed(term); got != "a \x1b[0;48;2;24;24;24m \x1b[0m" {
		t.Errorf("the row: %q", got)
	}
	if got := f.Themed(theme.Default); got != "a" {
		t.Errorf("the row, the terminal's colours unknown: %q", got)
	}
}
