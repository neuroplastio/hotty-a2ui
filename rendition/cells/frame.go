package cells

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// Role is what a cell's colour means: NEIO-4's role names, which a theme
// colours. A cell names its foreground's role.
type Role uint8

// The roles.
const (
	Fg Role = iota
	Bg
	Muted
	Accent
	Selection
	Surface
	Border
	Success
	Warning
	Error
	Info
)

var roleNames = [...]string{"fg", "bg", "muted", "accent", "selection", "surface", "border", "success", "warning", "error", "info"}

func (r Role) String() string {
	if int(r) < len(roleNames) {
		return roleNames[r]
	}
	return "role(" + strconv.Itoa(int(r)) + ")"
}

// ansi16 is each role's SGR foreground at the ANSI-16 floor (profile
// §3.6): "" keeps the terminal's.
var ansi16 = [...]string{
	Fg: "", Bg: "", Surface: "",
	Muted: "90", Border: "90",
	Accent: "94", Selection: "94",
	Success: "32", Warning: "33", Error: "31", Info: "36",
}

// Attr is a cell's attributes, a set of flags: what NO_COLOR keeps.
type Attr uint8

// The attributes.
const (
	Bold Attr = 1 << iota
	Faint
	Italic
	Underline
	Reverse
	Strike
)

// sgr are the attributes' SGR parameters, in Attr's order.
var sgr = [...]string{"1", "2", "3", "4", "7", "9"}

// Cell is one cell of a frame.
type Cell struct {
	// Text is one grapheme cluster; "" in the second column of a wide
	// one.
	Text  string
	Width int
	Role  Role
	Attr  Attr
	// Link is the URL the cell links to (OSC 8), or "".
	Link string
}

var blank = Cell{Text: " ", Width: 1}

// Frame is a surface laid out and painted in cells.
type Frame struct {
	Cols, Rows int
	// Cells are the rows, each Cols cells.
	Cells [][]Cell

	cursorCol, cursorRow int
	cursor               bool
}

func newFrame(cols, rows int) *Frame {
	f := &Frame{Cols: cols, Rows: rows, Cells: make([][]Cell, rows)}
	for y := range f.Cells {
		row := make([]Cell, cols)
		for x := range row {
			row[x] = blank
		}
		f.Cells[y] = row
	}
	return f
}

// Cursor is where the terminal's cursor goes: in the text control that
// has the keyboard; ok is false when none does.
func (f *Frame) Cursor() (col, row int, ok bool) {
	return f.cursorCol, f.cursorRow, f.cursor
}

// Plain is the frame's text: its rows joined by "\n", trailing spaces
// trimmed, with no styles.
func (f *Frame) Plain() string {
	var b strings.Builder
	for y, row := range f.Cells {
		if y > 0 {
			b.WriteByte('\n')
		}
		var line strings.Builder
		for _, c := range row {
			line.WriteString(c.Text)
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
	}
	return b.String()
}

// ANSI is the frame as terminal output: its rows joined by "\n", styled
// with SGR at the ANSI-16 floor, links as OSC 8 hyperlinks. With colors
// false (NO_COLOR) only the attributes are written. Trailing blank cells
// with no attributes are left out; every row ends with its styles reset.
func (f *Frame) ANSI(colors bool) string {
	if colors {
		return f.Themed(theme.Default)
	}
	return f.render(nil)
}

// Themed is the frame as terminal output in a theme's colours: each role
// its theme colour (truecolor), a theme's background under every cell, so
// nothing is left out at the end of a row. A role the theme leaves to the
// terminal keeps the ANSI-16 floor.
func (f *Frame) Themed(th theme.Theme) string { return f.render(&th) }

func (f *Frame) render(th *theme.Theme) string {
	var b strings.Builder
	for y, row := range f.Cells {
		if y > 0 {
			b.WriteByte('\n')
		}
		end := len(row)
		for end > 0 && row[end-1].Text == " " && row[end-1].Attr == 0 && row[end-1].Link == "" && (th == nil || th.Bg == "") {
			end--
		}
		style, link := "", ""
		for _, c := range row[:end] {
			if c.Width == 0 && c.Text == "" {
				continue
			}
			if s := c.style(th); s != style {
				if s == "" {
					b.WriteString("\x1b[0m")
				} else {
					b.WriteString("\x1b[0;" + s + "m")
				}
				style = s
			}
			if c.Link != link {
				b.WriteString("\x1b]8;;" + c.Link + "\x1b\\")
				link = c.Link
			}
			b.WriteString(c.Text)
		}
		if link != "" {
			b.WriteString("\x1b]8;;\x1b\\")
		}
		if style != "" {
			b.WriteString("\x1b[0m")
		}
	}
	return b.String()
}

// style is a cell's SGR parameters; th nil is no colour at all.
func (c Cell) style(th *theme.Theme) string {
	var p []string
	for i, s := range sgr {
		if c.Attr&(1<<i) != 0 {
			p = append(p, s)
		}
	}
	if th == nil {
		return strings.Join(p, ";")
	}
	if hex := th.Colour(roleNames[c.Role]); hex != "" {
		p = append(p, truecolour("38", hex))
	} else if int(c.Role) < len(ansi16) && ansi16[c.Role] != "" {
		p = append(p, ansi16[c.Role])
	}
	if th.Bg != "" {
		p = append(p, truecolour("48", th.Bg))
	}
	return strings.Join(p, ";")
}

// truecolour is the SGR for a "#rrggbb" colour, foreground (38) or
// background (48); a colour that is not one is the terminal's.
func truecolour(layer, hex string) string {
	if len(hex) != 7 || hex[0] != '#' {
		return ""
	}
	v, err := strconv.ParseUint(hex[1:], 16, 24)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s;2;%d;%d;%d", layer, v>>16, v>>8&0xff, v&0xff)
}
