package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

func bigComp(id, props string) string {
	return `{"id":"` + id + `","component":"HottyBigText","catalogId":"` + hotty.ID + `",` + props + `}`
}

// Every font has every character the profile lists, each glyph as many
// rows as the font's pixels and every row as wide, and its digits all as
// wide, so that a number that changes stays put.
func TestBigFonts(t *testing.T) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !\"#$%&'()*+,-./:;<=>?@[\\]_"
	for name, f := range bigFonts {
		for _, r := range chars {
			g, ok := f.glyphs[r]
			if !ok {
				t.Errorf("%s: no %q", name, r)
				continue
			}
			if len(g) != f.px {
				t.Errorf("%s: %q has %d rows", name, r, len(g))
			}
			for _, row := range g {
				if len(row) != len(g[0]) || strings.Trim(row, "#.") != "" {
					t.Errorf("%s: %q's row %q", name, r, row)
				}
			}
		}
		if len(f.glyphs) != len(chars) {
			t.Errorf("%s has %d glyphs, want %d", name, len(f.glyphs), len(chars))
		}
		for _, d := range "123456789" {
			if len(f.glyphs[d][0]) != len(f.glyphs['0'][0]) {
				t.Errorf("%s: %q is not as wide as 0", name, d)
			}
		}
	}
}

// A HottyBigText is its letters in blocks, a column apart: medium 5×7
// pixels in half blocks, 4 rows; small 3×5, 3 rows; large 3×5 in full
// blocks two columns a pixel, 5 rows. Lowercase is drawn in capitals, a
// letter's accent is left off, and what the font lacks is a question
// mark. A blank row comes after it.
func TestBigTextDraws(t *testing.T) {
	for _, tc := range []struct{ props, want string }{
		{`"text":"HI"`, "█   █ ▀█▀\n█▄▄▄█  █\n█   █  █\n▀   ▀ ▀▀▀"},
		{`"text":"hi"`, "█   █ ▀█▀\n█▄▄▄█  █\n█   █  █\n▀   ▀ ▀▀▀"},
		{`"text":"HI","size":"small"`, "█ █ ▀█▀\n█▀█  █\n▀ ▀ ▀▀▀"},
		{`"text":"HI","size":"large"`, "██  ██ ██████\n██  ██   ██\n██████   ██\n██  ██   ██\n██  ██ ██████"},
		{`"text":"Ì€","size":"small"`, "▀█▀ ▀▀▄\n █   ▀\n▀▀▀  ▀"},
	} {
		r := coding(t, bigComp("b", tc.props), `{"id":"after","component":"Text","text":"after"}`)
		f := r.Draw(40)
		checkFrame(t, f)
		if got, want := f.Plain(), tc.want+"\n\nafter"; got != want {
			t.Errorf("%s:\n%s\nwant\n%s", tc.props, got, want)
		}
		if f.Cells[0][0].Role != Fg || f.Cells[0][0].Attr != 0 {
			t.Errorf("%s: its letters are %v %v", tc.props, f.Cells[0][0].Role, f.Cells[0][0].Attr)
		}
	}
}

// A line too long for its box wraps between words, the lines a blank row
// apart; a word too long for a line of its own breaks between letters;
// each line goes where align says. Its natural width is its widest line,
// and a Row gives it that.
func TestBigTextWraps(t *testing.T) {
	r := coding(t, bigComp("b", `"text":"HI HI\nI"`))
	f := r.Draw(9)
	checkFrame(t, f)
	hi := "█   █ ▀█▀\n█▄▄▄█  █\n█   █  █\n▀   ▀ ▀▀▀"
	if got, want := f.Plain(), hi+"\n\n"+hi+"\n\n▀█▀\n █\n █\n▀▀▀"; got != want {
		t.Errorf("at 9:\n%s\nwant\n%s", got, want)
	}
	if got := r.Draw(40).Plain(); !strings.HasPrefix(got, "█   █ ▀█▀     █   █ ▀█▀\n") {
		t.Errorf("at 40, one line, the words a space's 3 pixels and a column each side apart:\n%s", got)
	}
	f = r.Draw(6)
	checkFrame(t, f)
	if got := strings.Split(f.Plain(), "\n"); len(got) != 5*4+4 || got[0] != "█   █" || got[5] != "▀█▀" {
		t.Errorf("at 6, each word breaks between its letters:\n%s", f.Plain())
	}

	r = coding(t, bigComp("c", `"text":"HI","align":"center"`), bigComp("e", `"text":"HI","align":"end"`))
	f = r.Draw(21)
	if got := strings.Split(f.Plain(), "\n"); got[0] != "      █   █ ▀█▀" || got[5] != "            █   █ ▀█▀" {
		t.Errorf("centred and at the end:\n%s", f.Plain())
	}

	r = coding(t, `{"id":"row","component":"Row","children":["name","b"]}`, `{"id":"name","component":"Text","text":"Score"}`,
		bigComp("b", `"text":"10"`))
	if got := strings.Split(r.Draw(40).Plain(), "\n")[0]; got != "Score  ▄█   ▄▀▀▀▄" {
		t.Errorf("in a Row: %q", got)
	}
	if e := r.c.V.Find("b"); e.Kind != view.BigText {
		t.Fatalf("b is %v", e.Kind)
	}
}

// A HottyBigText with no text takes no rows.
func TestBigTextEmpty(t *testing.T) {
	r := coding(t, bigComp("b", `"text":""`), `{"id":"after","component":"Text","text":"after"}`)
	if got := r.Draw(20).Plain(); got != "after" {
		t.Errorf("got %q", got)
	}
}
