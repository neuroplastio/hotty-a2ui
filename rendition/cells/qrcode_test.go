package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/qr"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

func qrComp(id, props string) string {
	return `{"id":"` + id + `","component":"HottyQRCode","catalogId":"` + hotty.ID + `",` + props + `}`
}

// A HottyQRCode is its modules two to a cell in half blocks, a quiet zone
// of four modules round it, every cell of it ink on paper; its label
// centred under it. Read back from the frame's text, the modules are the
// code's, the quiet zone's light.
func TestQRCodeDraws(t *testing.T) {
	r := coding(t, qrComp("q", `"value":"https://hotty.neuroplast.io","label":"Scan me"`), `{"id":"after","component":"Text","text":"after"}`)
	f := r.Draw(60)
	c := r.c.V.Find("q").QR
	side := c.Size + 2*qr.Quiet
	rows := (side + 1) / 2
	if c.Size != 29 || f.Rows != rows+2 {
		t.Fatalf("%d modules, %d rows:\n%s", c.Size, f.Rows, f.Plain())
	}
	for y := range rows {
		for x := range side {
			cl := f.Cells[y][x]
			mx, my := x-qr.Quiet, 2*y-qr.Quiet
			if want := halfBlock(c.Dark(mx, my), c.Dark(mx, my+1)); cl.Text != want || !cl.Paper {
				t.Fatalf("cell %d,%d is %q (paper %v), want %q", x, y, cl.Text, cl.Paper, want)
			}
		}
		if f.Cells[y][side].Paper {
			t.Fatalf("row %d: paper past the code", y)
		}
	}
	lines := strings.Split(f.Plain(), "\n")
	if want := strings.Repeat(" ", (side-7)/2) + "Scan me"; lines[rows] != want || lines[rows+1] != "after" {
		t.Errorf("under the code:\n%s", strings.Join(lines[rows:], "\n"))
	}
	if lines[0] != "" || lines[1] != "" {
		t.Errorf("the quiet zone's rows: %q", lines[:2])
	}
}

// Ink on paper is black on white whatever the theme: at the floor the 256
// colours' darkest grey on their white, the quiet zone's blank cells kept
// to the row's end; in a theme that paints its background, truecolor;
// under NO_COLOR, the glyphs alone.
func TestQRCodePaper(t *testing.T) {
	r := coding(t, qrComp("q", `"value":"12345"`))
	f := r.Draw(40)
	floor := strings.Split(f.ANSI(true), "\n")
	if want := "\x1b[0;38;5;232;48;5;231m" + strings.Repeat(" ", 29) + "\x1b[0m"; floor[0] != want {
		t.Errorf("the floor's top row %q, want %q", floor[0], want)
	}
	dark, _ := theme.ByName("Dracula")
	if got := strings.Split(f.Themed(dark), "\n")[0]; !strings.HasPrefix(got, "\x1b[0;38;2;0;0;0;48;2;255;255;255m"+strings.Repeat(" ", 29)+"\x1b[0;") {
		t.Errorf("in a theme %q", got)
	}
	mono := f.ANSI(false)
	if strings.Contains(mono, "\x1b[") {
		t.Errorf("under NO_COLOR %q", mono)
	}
	if mono != f.Plain() {
		t.Errorf("under NO_COLOR, not the glyphs alone:\n%s", mono)
	}
}

// A box narrower than the code has the value as text instead, wrapped,
// under the label; a value too long for a code is an error; an empty one
// takes no rows.
func TestQRCodeFallbacks(t *testing.T) {
	r := coding(t, qrComp("q", `"value":"https://hotty.neuroplast.io/a2ui","label":"Scan me"`))
	if got := r.Draw(20).Plain(); got != "Scan me\nhttps://hotty.neurop\nlast.io/a2ui" {
		t.Errorf("at 20 columns:\n%s", got)
	}
	long := strings.Repeat("x", 3000)
	r = coding(t, qrComp("q", `"value":"`+long+`","label":"Scan me"`), `{"id":"after","component":"Text","text":"after"}`)
	if got := r.Draw(40).Plain(); got != "✗ "+view.QRTooLong+"\nafter" {
		t.Errorf("too long:\n%s", got)
	}
	r = coding(t, qrComp("q", `"value":"","label":"Scan me"`), `{"id":"after","component":"Text","text":"after"}`)
	if got := r.Draw(40).Plain(); got != "after" {
		t.Errorf("empty:\n%s", got)
	}
}

// Its natural width is its code's, quiet zone and all, and it does not
// shrink: in a Row, the text beside it wraps; in a Row too narrow for it,
// its value shows as text.
func TestQRCodeInRow(t *testing.T) {
	r := coding(t, `{"id":"row","component":"Row","children":["q","t"]}`, qrComp("q", `"value":"12345"`),
		`{"id":"t","component":"Text","text":"Scan it with a phone's camera","weight":1}`)
	lines := strings.Split(r.Draw(45).Plain(), "\n")
	if lines[0] != strings.Repeat(" ", 30)+"Scan it with a" || lines[1] != strings.Repeat(" ", 30)+"phone's camera" || !strings.HasPrefix(lines[2], "    █▀▀▀▀▀█ ") || !strings.HasSuffix(lines[2], " █▀▀▀▀▀█") {
		t.Errorf("in a Row:\n%s", strings.Join(lines[:3], "\n"))
	}
	if got := strings.Split(r.Draw(20).Plain(), "\n")[0]; got != "12345        Scan it" {
		t.Errorf("in a narrow Row: %q", got)
	}
}
