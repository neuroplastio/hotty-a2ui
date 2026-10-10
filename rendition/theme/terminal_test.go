package theme

import (
	"image/color"
	"testing"
)

// Hear takes OSC 10, 11 and 4 replies, ended by ST or BEL, with channels
// of 1 to 4 hex digits, and several numbers in one OSC 4; it leaves
// anything else alone.
func TestHear(t *testing.T) {
	var term Terminal
	for _, reply := range []string{
		"\x1b]10;rgb:cdcd/d6d6/f4f4\x1b\\",
		"\x1b]11;rgb:1e/1e/2e\x07",
		"\x1b]4;1;rgb:f3f3/8b8b/a8a8\x1b\\",
		"\x1b]4;6;rgb:9494/e2e2/d5d5;12;rgb:f/8/0\x1b\\",
		"\x1b]4;8;#585B70\x07",
	} {
		if !term.Hear(reply) {
			t.Errorf("did not take %q", reply)
		}
	}
	want := Terminal{Fg: "#cdd6f4", Bg: "#1e1e2e"}
	want.ANSI[1], want.ANSI[6], want.ANSI[12], want.ANSI[8] = "#f38ba8", "#94e2d5", "#ff8800", "#585b70"
	if term != want {
		t.Errorf("heard %+v, want %+v", term, want)
	}
	for _, other := range []string{
		"\x1b]7279;a=q\x1b\\",
		"\x1b]4;16;rgb:00/00/00\x1b\\",
		"\x1b]11;?\x1b\\",
		"\x1b]10;rgb:zz/00/00\x1b\\",
		"q",
	} {
		if term.Hear(other) {
			t.Errorf("took %q", other)
		}
	}
	if term != want {
		t.Errorf("changed by what it did not take: %+v", term)
	}
}

func TestHex(t *testing.T) {
	if got := Hex(color.RGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff}); got != "#1e1e2e" {
		t.Errorf("Hex = %q", got)
	}
	if got := Hex(nil); got != "" {
		t.Errorf("Hex(nil) = %q", got)
	}
}
