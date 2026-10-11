package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// colIn is the column of s's first rune in a frame's row.
func colIn(t *testing.T, f *Frame, row int, s string) int {
	t.Helper()
	i := strings.Index(rowText(f, row), s)
	if i < 0 {
		t.Fatalf("row %d has no %q: %q", row, s, rowText(f, row))
	}
	return len([]rune(rowText(f, row)[:i]))
}

// TestPointerShapes: a hand over what a click acts on, as a browser
// points over a button, a checkbox and a select; an I-beam over a text
// field's value; the terminal's own over a label, a disabled control and
// nothing.
func TestPointerShapes(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	f := r.Draw(60)
	agree, size, def := rowOf(t, f, "Agree"), rowOf(t, f, "Size"), rowOf(t, f, "Go")
	note := rowOf(t, f, "Note")
	for _, at := range []struct {
		what     string
		col, row int
		want     string
	}{
		{"the Name field's value", nameAt + 1, 0, "text"},
		{"the Name field's label", gutter, 0, ""},
		{"the Note field's value", gutter + 1, note + 1, "text"},
		{"the CheckBox's box", colIn(t, f, agree, "Agree") - 2, agree, "pointer"},
		{"the CheckBox's label", colIn(t, f, agree, "Agree"), agree, "pointer"},
		{"the select", colIn(t, f, size, "Medium"), size, "pointer"},
		{"the Button", colIn(t, f, def, "Go"), def, "pointer"},
		{"past the Button", 50, def, ""},
		{"off the frame", -1, -1, ""},
	} {
		if got := r.Pointer(at.col, at.row); got != at.want {
			t.Errorf("%s (%d,%d): %q, want %q\n%s", at.what, at.col, at.row, got, at.want, f.Plain())
		}
	}
	r.c.V.Find("go").Disabled = true
	f = r.Draw(60)
	if got := r.Pointer(colIn(t, f, def, "Go"), def); got != "" {
		t.Errorf("a disabled Button: %q, want the terminal's own", got)
	}
}

// TestHover: a hover shows where the pointer is, a step past the look of
// what it is over: a field's underline brighter, a Button's fill a step
// lighter, a checkbox's row on a faint tint; moving off draws it as it
// was, and Hover says when the frame is to be drawn again.
func TestHover(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	plain := r.Draw(60)
	at := func(s string) (int, int) {
		row := rowOf(t, plain, s)
		return colIn(t, plain, row, strings.TrimSpace(s)), row
	}
	if r.Hover(50, 0) {
		t.Error("a move over nothing hovered: changed")
	}
	gx, gy := at("Go")
	if !r.Hover(gx, gy) {
		t.Fatal("onto the Button: not changed")
	}
	f := r.Draw(60)
	if got := f.Cells[gy][gx]; got.BackMix != hoverButton {
		t.Errorf("the hovered Button's fill: BackMix %d, want %d", got.BackMix, hoverButton)
	}
	if r.Hover(gx+1, gy) {
		t.Error("along the same Button: changed")
	}
	if !r.Hover(nameAt+1, 0) {
		t.Fatal("onto the Name field: not changed")
	}
	f = r.Draw(60)
	if got := f.Cells[gy][gx]; got.BackMix != buttonTint {
		t.Errorf("the Button, the pointer gone: BackMix %d, want %d", got.BackMix, buttonTint)
	}
	if got := f.Cells[0][nameAt]; !got.LineSet || got.LineMix != hoverLineMix {
		t.Errorf("the hovered field's underline: LineMix %d, want %d", got.LineMix, hoverLineMix)
	}
	ax, ay := at("Agree")
	r.Hover(ax, ay)
	f = r.Draw(60)
	if got := f.Cells[ay][ax]; got.Back != Fg || got.BackMix != hoverTint {
		t.Errorf("the hovered CheckBox: Back %v, BackMix %d, want a tint of %d", got.Back, got.BackMix, hoverTint)
	}
	if got := f.Cells[ay][0]; got.BackMix != 0 {
		t.Errorf("the hovered CheckBox's gutter: tinted %d", got.BackMix)
	}
	if !r.Hover(-1, -1) {
		t.Error("off the frame: not changed")
	}
	th := theme.Theme{Bg: "#000000", Fg: "#ffffff", Border: "#444444", Accent: "#3366ff"}
	if got := r.Draw(60).Themed(th); got != plain.Themed(th) {
		t.Error("the pointer gone: the frame is not as it was")
	}
}
