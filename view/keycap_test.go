package view

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
)

// caps writes Caps' chords as a hint would draw them: "/" between
// alternatives, ">" between a sequence's chords, "+" inside one.
func caps(c *Controller, keys string, glyphs bool) string {
	var alts []string
	for _, seq := range c.Caps(keys, glyphs) {
		var cs []string
		for _, ch := range seq {
			sep := "+"
			if ch.Tight {
				sep = ""
			}
			cs = append(cs, "["+strings.Join(ch.Labels, sep)+"]")
		}
		alts = append(alts, strings.Join(cs, ">"))
	}
	return strings.Join(alts, "/")
}

func TestCaps(t *testing.T) {
	linux := &Controller{}
	mac := &Controller{Platform: hotty.Platform{Mod: "super"}}
	win := &Controller{Keycaps: Keycaps{OS: "windows"}}
	for _, tc := range []struct {
		c      *Controller
		keys   string
		glyphs bool
		want   string
	}{
		{linux, "↑/k", false, "[↑]/[K]"},
		{linux, "↑/k", true, "[\U000F005D]/[K]"},
		{linux, "alt+↑/↓", false, "[Alt+↑]/[Alt+↓]"},
		{linux, "G/end", false, "[Shift+G]/[End]"},
		{linux, "→/l/pgdn", false, "[→]/[L]/[PgDn]"},
		{linux, "/", false, "[/]"},
		{linux, "ctrl+/", false, "[Ctrl+/]"},
		{linux, "+/-", false, "[+]/[-]"},
		{linux, "ctrl++", false, "[Ctrl++]"},
		{linux, "g g", false, "[G]>[G]"},
		{linux, "shift+ctrl+p", false, "[Ctrl+Shift+P]"},
		{linux, "super+space", false, "[Meta+Space]"},
		{linux, "esc", true, "[Esc]"},
		{linux, "f5/caps_lock", false, "[F5]/[Caps Lock]"},
		{mac, "super+shift+p", true, "[\U000F0636\U000F0633P]"},
		{mac, "super+shift+p", false, "[Shift+Cmd+P]"},
		{mac, "alt+enter", false, "[Option+Return]"},
		{mac, "esc", true, "[esc]"},
		{mac, "ctrl+x ctrl+s", true, "[\U000F0634X]>[\U000F0634S]"},
		{win, "super+e", true, "[⊞+E]"},
		{win, "super+e", false, "[Win+E]"},
		{win, "ctrl+alt+delete", false, "[Ctrl+Alt+Del]"},
		{linux, "ctrl+backspace", true, "[Ctrl+\U000F0B5C]"},
		{linux, "ctrl+backspace", false, "[Ctrl+Backspace]"},
		{win, "backspace", true, "[\U000F0B5C]"},
		{mac, "backspace", true, "[\U000F0B5C]"},
	} {
		if got := caps(tc.c, tc.keys, tc.glyphs); got != tc.want {
			t.Errorf("%q (os %q, glyphs %v): %q, want %q", tc.keys, tc.c.Keycaps.OS, tc.glyphs, got, tc.want)
		}
	}
}

// A user's labels win over the system's, in either presentation.
func TestCapsLabels(t *testing.T) {
	c := &Controller{Keycaps: Keycaps{Labels: map[string]string{"super": "◆"}}}
	if got := caps(c, "super+k", false); got != "[◆+K]" {
		t.Errorf("relabelled super: %q", got)
	}
	c.Keycaps.OS = "mac"
	if got := caps(c, "super+k", true); got != "[◆K]" {
		t.Errorf("relabelled super on macOS: %q", got)
	}
}

func TestKeycapsGlyphs(t *testing.T) {
	for _, tc := range []struct {
		present    string
		host, want bool
	}{
		{"", true, true}, {"", false, false},
		{PresentText, true, false}, {PresentGlyphs, false, true},
	} {
		if got := (Keycaps{Present: tc.present}).Glyphs(tc.host); got != tc.want {
			t.Errorf("Present %q on a host %v: %v", tc.present, tc.host, got)
		}
	}
}
