package theme

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Terminal is what a terminal says its colours are: its text's (OSC 10),
// its background's (OSC 11) and those its palette numbers (OSC 4), each
// "#rrggbb", or "" where it has not said. A theme that leaves a role to
// the terminal (Default leaves them all) still paints it in the
// terminal's palette, by number, and paints no background; these are for
// what has to be worked out from the terminal's colours: one between two
// (a HottyProgress's fill), a background tinted toward one (a changed
// line, a selection, a surface). Cells says which (profile §3.6).
type Terminal struct {
	Fg, Bg string
	ANSI   [16]string
}

// Hear takes a terminal's answer to a question about its colours, as the
// bytes came: an OSC 10, 11 or 4 reply, `ESC ] 4 ; 1 ; rgb:f3f3/8b8b/a8a8`
// ended by ST or BEL. It reports whether it was one it took.
func (t *Terminal) Hear(reply string) bool {
	body, ok := strings.CutPrefix(reply, "\x1b]")
	if !ok {
		return false
	}
	body = strings.TrimSuffix(strings.TrimSuffix(body, "\x07"), "\x1b\\")
	cmd, rest, ok := strings.Cut(body, ";")
	if !ok {
		return false
	}
	switch cmd {
	case "10", "11":
		hex := xColour(rest)
		if hex == "" {
			return false
		}
		if cmd == "10" {
			t.Fg = hex
		} else {
			t.Bg = hex
		}
		return true
	case "4":
		// One reply may carry several numbers: 4;1;rgb:…;2;rgb:….
		parts := strings.Split(rest, ";")
		took := false
		for i := 0; i+1 < len(parts); i += 2 {
			n, err := strconv.Atoi(parts[i])
			hex := xColour(parts[i+1])
			if err != nil || n < 0 || n >= len(t.ANSI) || hex == "" {
				continue
			}
			t.ANSI[n], took = hex, true
		}
		return took
	}
	return false
}

// xColour is an X11 colour spec as terminals answer with it,
// `rgb:r/g/b` with 1 to 4 hex digits a channel (`rgba:` too, its alpha
// left out), or `#rrggbb`, as "#rrggbb"; "" for anything else.
func xColour(spec string) string {
	if strings.HasPrefix(spec, "#") && len(spec) == 7 {
		if _, err := strconv.ParseUint(spec[1:], 16, 24); err == nil {
			return strings.ToLower(spec)
		}
		return ""
	}
	_, chans, ok := strings.Cut(spec, ":")
	if !ok || !strings.HasPrefix(spec, "rgb") {
		return ""
	}
	parts := strings.Split(chans, "/")
	if len(parts) < 3 {
		return ""
	}
	var v [3]uint64
	for i, p := range parts[:3] {
		n, err := strconv.ParseUint(p, 16, 16)
		if err != nil || len(p) == 0 || len(p) > 4 {
			return ""
		}
		// A channel of k digits is n/(16^k − 1) of the way: scale to 8 bits.
		v[i] = n * 255 / (1<<(4*len(p)) - 1)
	}
	return fmt.Sprintf("#%02x%02x%02x", v[0], v[1], v[2])
}

// Hex is a colour as "#rrggbb", as a program hears the terminal's text
// and background colours from a terminal library (Bubble Tea's
// ForegroundColorMsg and BackgroundColorMsg are colours); "" for nil.
func Hex(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// RGB is a "#rrggbb" colour as a colour a terminal library takes (Bubble
// Tea's cursor colour, say); nil for "" or anything else.
func RGB(hex string) color.Color {
	if len(hex) != 7 || hex[0] != '#' {
		return nil
	}
	n, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return nil
	}
	return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 0xff}
}
