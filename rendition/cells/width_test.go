package cells

import (
	"strings"
	"testing"
)

func TestWidth(t *testing.T) {
	for _, tc := range []struct {
		name, s  string
		clusters int
		width    int
	}{
		{"ASCII", "Ada", 3, 3},
		{"CJK", "日本語", 3, 6},
		{"fullwidth", "ＡＢ", 2, 4},
		{"Hangul", "한국", 2, 4},
		{"combining acute", "e\u0301", 1, 1},
		{"combining marks", "a\u0300\u0316b", 2, 2},
		{"emoji", "\U0001F600", 1, 2},
		{"emoji ZWJ", "\U0001F469\u200D\U0001F4BB", 1, 2},
		{"family ZWJ", "\U0001F468\u200D\U0001F469\u200D\U0001F467", 1, 2},
		{"skin tone", "\U0001F44D\U0001F3FD", 1, 2},
		{"flag", "\U0001F1EF\U0001F1F5", 1, 2},
		{"two flags", "\U0001F1EF\U0001F1F5\U0001F1FA\U0001F1F8", 2, 4},
		{"lone regional indicator", "\U0001F1EF", 1, 1},
		{"tag flag", "\U0001F3F4\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F", 1, 2},
		{"text heart", "❤", 1, 1},
		{"VS16 heart", "❤\uFE0F", 1, 2},
		{"VS16 smiley", "☺\uFE0F", 1, 2},
		{"VS15 watch", "⌚\uFE0E", 1, 2},
		{"keycap", "1\uFE0F\u20E3", 1, 2},
		{"ZWJ alone", "\u200D", 1, 0},
		{"soft hyphen", "a\u00ADb", 3, 2},
		{"box drawing", "╭─╮", 3, 3},
		{"ambiguous", "±·", 2, 2},
	} {
		if got := len(clusters(tc.s)); got != tc.clusters {
			t.Errorf("%s: %d clusters, want %d", tc.name, got, tc.clusters)
		}
		if got := Width(tc.s); got != tc.width {
			t.Errorf("%s: width %d, want %d", tc.name, got, tc.width)
		}
	}
}

// TestWideCells checks wide clusters in a frame: two cells, the second
// empty; a wide cluster that does not fit before the edge is a space.
func TestWideCells(t *testing.T) {
	f := newFrame(5, 1)
	cv := &canvas{f: f}
	cv.write(0, 0, 5, glyphs("日\U0001F469\u200D\U0001F4BB", style{}))
	if got := f.Plain(); got != "日\U0001F469\u200D\U0001F4BB" {
		t.Errorf("plain %q", got)
	}
	if c := f.Cells[0][1]; c.Text != "" || c.Width != 0 {
		t.Errorf("second half %+v", c)
	}
	cv.set(4, 0, glyphs("語", style{})[0])
	if c := f.Cells[0][4]; c.Text != " " || c.Width != 1 {
		t.Errorf("wide cluster at the edge: %+v", c)
	}
	cv.set(1, 0, glyphs("x", style{})[0])
	if got := f.Plain(); got != " x\U0001F469\u200D\U0001F4BB" {
		t.Errorf("a wide cluster cut in half: %q", got)
	}
	checkFrame(t, f)
}

func TestWrap(t *testing.T) {
	for _, tc := range []struct {
		s    string
		w    int
		want string
	}{
		{"the quick brown fox", 9, "the quick|brown fox"},
		{"the quick brown fox", 10, "the quick|brown fox"},
		{"a supercalifragilistic word", 8, "a|supercal|ifragili|stic|word"},
		{"line one\nline two", 20, "line one|line two"},
		{"日本語のテキスト", 6, "日本語|のテキ|スト"},
		{"  indented", 20, "  indented"},
		{"spaces   between", 20, "spaces   between"},
	} {
		var lines []string
		for _, l := range wrap(glyphs(tc.s, style{}), tc.w, nil, nil) {
			var b strings.Builder
			for _, g := range l {
				b.WriteString(g.text)
			}
			lines = append(lines, b.String())
		}
		if got := strings.Join(lines, "|"); got != tc.want {
			t.Errorf("wrap(%q, %d) = %q, want %q", tc.s, tc.w, got, tc.want)
		}
	}
}

func TestANSI(t *testing.T) {
	f := newFrame(12, 1)
	cv := &canvas{f: f}
	n := cv.write(0, 0, 12, glyphs("✗ ", style{role: Error}))
	cv.write(n, 0, 12-n, glyphs("link", style{attr: Underline, link: "https://neuroplast.io"}))
	got := f.ANSI(true)
	want := "\x1b[0;31m✗ \x1b[0;4m\x1b]8;;https://neuroplast.io\x1b\\link\x1b]8;;\x1b\\\x1b[0m"
	if got != want {
		t.Errorf("ANSI(true)\n%q\nwant\n%q", got, want)
	}
	got = f.ANSI(false)
	want = "✗ \x1b[0;4m\x1b]8;;https://neuroplast.io\x1b\\link\x1b]8;;\x1b\\\x1b[0m"
	if got != want {
		t.Errorf("ANSI(false)\n%q\nwant\n%q", got, want)
	}
}
