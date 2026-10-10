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

// ansi256 is an ANSI-16 foreground's colour as an index of the 256, for
// an underline's colour (SGR 58), which has no 16-colour form. It is also
// the number the terminal's palette knows the colour by (OSC 4).
var ansi256 = map[string]string{"90": "8", "94": "12", "32": "2", "33": "3", "31": "1", "36": "6"}

// Where a theme leaves the roles to the terminal, the selection and the
// surface have no number in its palette: they are worked out from the
// colours it said (profile §3.6). selectionMix is how far a selection is
// from the background toward the accent; surfaceMix, a surface toward the
// text.
const (
	selectionMix = 51 // 20%
	surfaceMix   = 24 // 9%
)

// TerminalQuery asks the terminal for the colours cells works out blends
// and tints from where the theme leaves the roles to it: its text's (OSC
// 10), its background's (OSC 11), the floor's and its two magentas (OSC 4,
// by number). A
// program writes it once, at the start, and gives the answers to
// theme.Terminal (Hear, Hex) and the Terminal to the theme it paints
// with (theme.Theme.Term). Nothing waits on them: until they come, and
// on a terminal that never answers, cells draws at the floor.
func TerminalQuery() string {
	b := strings.Builder{}
	b.WriteString("\x1b]10;?\x1b\\\x1b]11;?\x1b\\")
	seen := map[string]bool{}
	for _, sgr := range ansi16 {
		if n, ok := ansi256[sgr]; ok && !seen[n] {
			seen[n] = true
			b.WriteString("\x1b]4;" + n + ";?\x1b\\")
		}
	}
	// Its magentas, which no role is, for a HottyProgress's fill alone
	// (Cell.Fill).
	b.WriteString("\x1b]4;5;?\x1b\\\x1b]4;13;?\x1b\\")
	return b.String()
}

// colour is a role's colour as "#rrggbb", for a blend or a tint: the
// theme's, else what the terminal said (theme.Theme.Term), else "".
func colour(th *theme.Theme, r Role) string {
	if hex := th.Colour(roleNames[r]); hex != "" {
		return hex
	}
	t := th.Term
	switch r {
	case Fg:
		return t.Fg
	case Bg:
		return t.Bg
	case Selection:
		return mix(colour(th, Bg), colour(th, Accent), selectionMix)
	case Surface:
		return mix(colour(th, Bg), colour(th, Fg), surfaceMix)
	}
	if int(r) < len(ansi16) {
		if n, err := strconv.Atoi(ansi256[ansi16[r]]); err == nil && n < len(t.ANSI) {
			return t.ANSI[n]
		}
	}
	return ""
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
	// Overline is a line at the cell's top (SGR 53): a drag's line at the
	// frame's top, where there is no row above to underline (profile
	// §3.4).
	Overline
)

// sgr are the attributes' SGR parameters, in Attr's order.
var sgr = [...]string{"1", "2", "3", "4", "7", "9", "53"}

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
	// the way from Role's to To's, where both are known as "#rrggbb", the
	// theme's or the terminal's (a Progress bar's gradient); else Role's.
	// Mix 0 is Role's alone.
	To  Role
	Mix uint8
	// Fill marks a HottyProgress's fill: where the theme leaves the accent
	// to the terminal and the terminal said its accent and its magenta,
	// Mix/255 of the way from the one to the other (fillColour); elsewhere
	// Role into To, as any blend.
	Fill bool
	// Back and BackMix tint the cell's background: BackMix/255 of the way
	// from the background to Back's colour (a marked line of code), where
	// both are known, the theme's or the terminal's; BackMix 0 is no tint.
	Back    Role
	BackMix uint8
	// Where there is no tint (a terminal that has not said its colours),
	// a cell with BackMix shows its Back by BackAttr's attributes instead
	// (a changed word underlined), and under NO_COLOR by MonoAttr's where
	// it has some (that word reversed), as nothing else reads there.
	BackAttr Attr
	MonoAttr Attr
	// BackShade shows Back, where there is no tint but the terminal has
	// colours, as the terminal's bright black under the cell (SGR 100): a
	// Button's fill, which an attribute would not read as.
	BackShade bool
	// Line colours the cell's underline in a role's colour (SGR 58) when
	// LineSet, so that a drag's line is in the accent under text in its
	// own colour.
	Line    Role
	LineSet bool
}

var blank = Cell{Text: " ", Width: 1}

// Frame is a surface laid out and painted in cells.
type Frame struct {
	Cols, Rows int
	// Cells are the rows, each Cols cells.
	Cells [][]Cell

	cursorCol, cursorRow int
	cursor, cursorBlock  bool
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

// grow makes the frame at least rows tall, with blank rows: for what is
// drawn over it past its end (a field's suggestions).
func (f *Frame) grow(rows int) {
	for f.Rows < rows {
		row := make([]Cell, f.Cols)
		for x := range row {
			row[x] = blank
		}
		f.Cells = append(f.Cells, row)
		f.Rows++
	}
}

// Cursor is where the terminal's cursor goes: in the text control that
// has the keyboard; ok is false when none does.
func (f *Frame) Cursor() (col, row int, ok bool) {
	return f.cursorCol, f.cursorRow, f.cursor
}

// BlockCursor reports whether the terminal's cursor is a block, as under a
// terminal's keymap, the frame having reversed the cell it is on; else it
// is a line, a GUI field's, which the terminal draws (profile §3.5).
func (f *Frame) BlockCursor() bool { return f.cursorBlock }

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
		for end > 0 && row[end-1].bare(th) {
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

// bare reports whether a row may leave the cell out at its end: a blank
// with nothing that shows, no background under it.
func (c Cell) bare(th *theme.Theme) bool {
	if c.Text != " " || c.Attr != 0 || c.Link != "" {
		return false
	}
	if th == nil {
		return c.BackMix == 0 || c.MonoAttr == 0
	}
	return th.Bg == "" && !(c.BackMix > 0 && (c.tinted(th) || c.BackShade))
}

// style is a cell's SGR parameters; th nil is no colour at all.
func (c Cell) style(th *theme.Theme) string {
	attr := c.Attr
	tinted := c.BackMix > 0 && c.tinted(th)
	if c.BackMix > 0 && !tinted {
		if th == nil && c.MonoAttr != 0 {
			attr |= c.MonoAttr
		} else {
			attr |= c.BackAttr
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
	hex := th.Colour(roleNames[c.Role])
	if f := c.fillColour(th); f != "" {
		hex = f
	} else if c.Mix > 0 {
		if m := mix(colour(th, c.Role), colour(th, c.To), c.Mix); m != "" {
			hex = m
		}
	}
	if hex != "" {
		p = append(p, truecolour("38", hex))
	} else if int(c.Role) < len(ansi16) && ansi16[c.Role] != "" {
		p = append(p, ansi16[c.Role])
	}
	if c.LineSet && attr&Underline != 0 {
		if line := th.Colour(roleNames[c.Line]); line != "" {
			p = append(p, truecolour("58", line))
		} else if n, ok := ansi256[ansi16[c.Line]]; ok {
			p = append(p, "58;5;"+n)
		}
	}
	// A theme's background goes under every cell; a terminal's, which it
	// paints itself, only where a tint changes it.
	switch {
	case tinted:
		p = append(p, truecolour("48", mix(colour(th, Bg), colour(th, c.Back), c.BackMix)))
	case c.BackMix > 0 && c.BackShade:
		p = append(p, "100")
	case th.Bg != "":
		p = append(p, truecolour("48", th.Bg))
	}
	return strings.Join(p, ";")
}

// fillColour is a HottyProgress fill cell's colour in the terminal's own
// gradient (profile §3.4): Mix/255 of the way from its accent (OSC 4;12)
// into its bright magenta (4;13, else 4;5), blue into pink in most
// palettes, as bubbles' default gradient goes; "" where the theme colours
// the accent itself, or the terminal has not said both.
func (c Cell) fillColour(th *theme.Theme) string {
	if !c.Fill || th.Accent != "" {
		return ""
	}
	magenta := th.Term.ANSI[13]
	if magenta == "" {
		magenta = th.Term.ANSI[5]
	}
	return mix(colour(th, Accent), magenta, c.Mix)
}

// tinted reports whether the cell's background can be tinted (BackMix):
// whether the background and Back's colour are both known, the theme's
// or the terminal's.
func (c Cell) tinted(th *theme.Theme) bool {
	if th == nil {
		return false
	}
	_, okBg := rgb(colour(th, Bg))
	_, okBack := rgb(colour(th, c.Back))
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

// mix is the colour m/255 of the way from a to b, channel by channel; ""
// when either is not a "#rrggbb".
func mix(a, b string, m uint8) string {
	x, okA := rgb(a)
	y, okB := rgb(b)
	if !okA || !okB {
		return ""
	}
	ch := func(shift uint) uint32 {
		p, q := int(x>>shift&0xff), int(y>>shift&0xff)
		return uint32(p + (q-p)*int(m)/255)
	}
	return fmt.Sprintf("#%02x%02x%02x", ch(16), ch(8), ch(0))
}
