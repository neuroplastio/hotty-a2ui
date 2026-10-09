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
	// To and Mix blend the cell's colour toward another role's: Mix/255 of
	// the way from Role's to To's, where the theme has both as "#rrggbb"
	// (a Progress bar's gradient); else Role's. Mix 0 is Role's alone.
	To  Role
	Mix uint8
	// Back and BackMix tint the cell's background: BackMix/255 of the way
	// from the theme's background to Back's colour (a marked line of
	// code), where the theme has both as "#rrggbb"; BackMix 0 is no tint,
	// and so is a theme that keeps the terminal's background.
	Back    Role
	BackMix uint8
	// Where there is no tint (a theme that keeps the terminal's
	// background, NO_COLOR), a cell with BackMix shows its Back another
	// way, as git's diff-highlight does a changed line: in Back's colour
	// with BackFg, and with BackAttr's attributes (its changed words
	// reversed).
	BackFg   bool
	BackAttr Attr
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
	attr, role := c.Attr, c.Role
	if c.BackMix > 0 && !c.tinted(th) {
		attr |= c.BackAttr
		if c.BackFg {
			role = c.Back
		}
	}
	var p []string
	for i, s := range sgr {
		if attr&(1<<i) != 0 {
			p = append(p, s)
		}
	}
	if th == nil {
		return strings.Join(p, ";")
	}
	if role != c.Role {
		c.Role, c.Mix = role, 0
	}
	hex := th.Colour(roleNames[c.Role])
	if c.Mix > 0 {
		hex = blend(hex, th.Colour(roleNames[c.To]), c.Mix)
	}
	if hex != "" {
		p = append(p, truecolour("38", hex))
	} else if int(c.Role) < len(ansi16) && ansi16[c.Role] != "" {
		p = append(p, ansi16[c.Role])
	}
	if th.Bg != "" {
		bg := th.Bg
		if c.BackMix > 0 {
			bg = blend(th.Bg, th.Colour(roleNames[c.Back]), c.BackMix)
		}
		p = append(p, truecolour("48", bg))
	}
	return strings.Join(p, ";")
}

// tinted reports whether the cell's background can be tinted (BackMix) in
// a theme: one whose background and Back's colour are both "#rrggbb".
func (c Cell) tinted(th *theme.Theme) bool {
	if th == nil {
		return false
	}
	_, okBg := rgb(th.Bg)
	_, okBack := rgb(th.Colour(roleNames[c.Back]))
	return okBg && okBack
}

// truecolour is the SGR for a "#rrggbb" colour, foreground (38) or
// background (48); a colour that is not one is the terminal's.
func truecolour(layer, hex string) string {
	v, ok := rgb(hex)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s;2;%d;%d;%d", layer, v>>16, v>>8&0xff, v&0xff)
}

// rgb is a "#rrggbb" colour's value, 0xrrggbb.
func rgb(hex string) (uint32, bool) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, false
	}
	v, err := strconv.ParseUint(hex[1:], 16, 24)
	return uint32(v), err == nil
}

// blend is the colour m/255 of the way from a to b, channel by channel;
// a when either is not a "#rrggbb".
func blend(a, b string, m uint8) string {
	x, okA := rgb(a)
	y, okB := rgb(b)
	if !okA || !okB {
		return a
	}
	ch := func(shift uint) uint32 {
		p, q := int(x>>shift&0xff), int(y>>shift&0xff)
		return uint32(p + (q-p)*int(m)/255)
	}
	return fmt.Sprintf("#%02x%02x%02x", ch(16), ch(8), ch(0))
}
