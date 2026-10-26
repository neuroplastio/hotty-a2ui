package view

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-go"
)

// caps writes Caps' chords as a hint would draw them: "/" between
// alternatives, ">" between a sequence's chords, a chord's Sep inside one.
func caps(c *Controller, keys string, glyphs bool) string {
	var alts []string
	for _, seq := range c.Caps(keys, glyphs) {
		var cs []string
		for _, ch := range seq {
			var b strings.Builder
			for k, l := range ch.Labels {
				if ch.Joined(k) {
					b.WriteString(ch.Sep)
				}
				b.WriteString(l)
			}
			cs = append(cs, "["+b.String()+"]")
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
		{linux, "ctrl+b /", false, "[Ctrl+B]>[/]"},
		{linux, "/ x", false, "[/]>[X]"},
		{linux, "a//", false, "[A]/[/]"},
		{linux, "ctrl+//x", false, "[Ctrl+/]/[X]"},
		{linux, "+/-", false, "[+]/[-]"},
		{linux, "ctrl++", false, "[Ctrl++]"},
		{linux, "g g", false, "[G]>[G]"},
		{linux, "shift+ctrl+p", false, "[Ctrl+Shift+P]"},
		{linux, "super+space", false, "[Meta+Space]"},
		{linux, "esc", true, "[Esc]"},
		{linux, "f5/caps_lock", false, "[F5]/[Caps Lock]"},
		{mac, "super+shift+p", true, "[\U000F0633\U000F0636P]"},
		{mac, "super+shift+p", false, "[Command-Shift-P]"},
		{mac, "alt+enter", false, "[Option-Return]"},
		{mac, "ctrl+backspace/delete", false, "[Control-Delete]/[Forward Delete]"},
		{mac, "delete", true, "[\U000F0E7F]"},
		{mac, "esc", true, "[esc]"},
		{mac, "esc", false, "[Esc]"},
		{mac, "ctrl+x ctrl+s", true, "[\U000F0634X]>[\U000F0634S]"},
		{win, "super+e", true, "[⊞+E]"},
		{win, "super+e", false, "[Win+E]"},
		{win, "ctrl+alt+delete", false, "[Ctrl+Alt+Del]"},
		{linux, "ctrl+backspace", true, "[\U000F0634+\U000F0B5C]"},
		{linux, "super+ctrl+alt+shift+r", true, "[Meta+\U000F0634+Alt+\U000F0636+R]"},
		{win, "super+shift+s", true, "[⊞+\U000F0636+S]"},
		{linux, "ctrl+backspace", false, "[Ctrl+Backspace]"},
		{win, "backspace", true, "[\U000F0B5C]"},
		{mac, "backspace", true, "[\U000F0B5C]"},
	} {
		if got := caps(tc.c, tc.keys, tc.glyphs); got != tc.want {
			t.Errorf("%q (os %q, glyphs %v): %q, want %q", tc.keys, tc.c.Keycaps.OS, tc.glyphs, got, tc.want)
		}
	}
}

// Case, Sep and Glue change the system's: a user may write keys as the
// system they came from does.
func TestCapsStyle(t *testing.T) {
	for _, tc := range []struct {
		k      Keycaps
		keys   string
		glyphs bool
		want   string
	}{
		{Keycaps{Case: CaseLower}, "ctrl+shift+l", false, "[ctrl+shift+l]"},
		{Keycaps{Case: CaseLower}, "f5/caps_lock", false, "[f5]/[caps lock]"},
		{Keycaps{OS: "mac", Case: CaseTitle}, "super+esc", true, "[\U000F0633Esc]"},
		{Keycaps{OS: "mac", Case: CaseLower}, "super+shift+p", false, "[command-shift-p]"},
		{Keycaps{Sep: "-"}, "ctrl+x ctrl+s", false, "[Ctrl-X]>[Ctrl-S]"},
		{Keycaps{Sep: " "}, "alt+enter", false, "[Alt Enter]"},
		{Keycaps{Glue: GlueOn}, "ctrl+shift+p", true, "[\U000F0634\U000F0636P]"},
		{Keycaps{Glue: GlueOn}, "super+ctrl+r", true, "[Meta+\U000F0634R]"},
		{Keycaps{Glue: GlueOn}, "super+ctrl+alt+shift+r", true, "[Meta+Alt+\U000F0634\U000F0636R]"},
		{Keycaps{OS: "windows", Glue: GlueOn}, "super+shift+s", true, "[⊞\U000F0636S]"},
		{Keycaps{Glue: GlueOn, Labels: map[string]string{"super": "\U000F033D"}}, "super+alt+l", true, "[Alt+\U000F033DL]"},
		{Keycaps{Glue: GlueOn}, "ctrl+shift+p", false, "[Ctrl+Shift+P]"},
		{Keycaps{OS: "mac", Glue: GlueOff}, "super+shift+p", true, "[\U000F0633-\U000F0636-P]"},
		{Keycaps{OS: "mac", Labels: map[string]string{"super": "Super"}}, "super+shift+p", true, "[Super-\U000F0636P]"},
		{Keycaps{Case: CaseLower, Labels: map[string]string{"super": "Super"}}, "super+k", false, "[Super+k]"},
		{Keycaps{Case: CaseShifted}, "ctrl+shift+l", false, "[ctrl+L]"},
		{Keycaps{Case: CaseShifted}, "ctrl+l", false, "[ctrl+l]"},
		{Keycaps{Case: CaseShifted}, "G/g g", false, "[G]/[g]>[g]"},
		{Keycaps{Case: CaseShifted}, "shift+tab/shift+1/?", false, "[shift+tab]/[shift+1]/[?]"},
		{Keycaps{Case: CaseShifted}, "ctrl+shift+arrow_right/f5", false, "[ctrl+shift+→]/[f5]"},
		{Keycaps{OS: "mac", Case: CaseShifted}, "super+shift+p", true, "[\U000F0633P]"},
		{Keycaps{OS: "mac", Case: CaseShifted}, "super+shift+p", false, "[command-P]"},
		{Keycaps{Case: CaseShifted, Labels: map[string]string{"p": "π"}}, "ctrl+shift+p", false, "[ctrl+shift+π]"},
	} {
		c := &Controller{Keycaps: tc.k}
		if got := caps(c, tc.keys, tc.glyphs); got != tc.want {
			t.Errorf("%q %+v (glyphs %v): %q, want %q", tc.keys, tc.k, tc.glyphs, got, tc.want)
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

// Then is ">" unless set: a space is a gap alone.
func TestKeycapsBetween(t *testing.T) {
	for then, want := range map[string]string{"": ">", " ": " ", "→": "→"} {
		if got := (Keycaps{Then: then}).Between(); got != want {
			t.Errorf("Then %q: %q, want %q", then, got, want)
		}
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

// A label is UTF-8 whatever a key's name starts with: kit b538269 cased a
// name's first byte, so "⇧⇥" went out as U+FFFD and two stray bytes, and a
// host refused the whole document.
func TestCapsUTF8(t *testing.T) {
	for _, os := range []string{"linux", "mac", "windows"} {
		for _, how := range []string{"", CaseTitle, CaseLower} {
			for _, glue := range []string{"", GlueOn, GlueOff} {
				c := &Controller{Keycaps: Keycaps{OS: os, Case: how, Glue: glue}}
				for _, keys := range []string{"ctrl+b ⇧⇥", "ctrl+b shift+tab", "⇥/⌃é", "shift+é_ß", "ctrl+b ⇧"} {
					for _, glyphs := range []bool{false, true} {
						if got := caps(c, keys, glyphs); !utf8.ValidString(got) {
							t.Errorf("%q %s case %q glue %q (glyphs %v): %q", keys, os, how, glue, glyphs, got)
						}
					}
				}
			}
		}
	}
}
