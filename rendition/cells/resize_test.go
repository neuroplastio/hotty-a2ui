package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/story"
)

// gripsOf are the cells of a frame's grips, top to bottom.
func gripsOf(f *Frame) [][2]int {
	var at [][2]int
	for y := range f.Rows {
		for x := range f.Cols {
			if f.Cells[y][x].Text == gripGlyph {
				at = append(at, [2]int{x, y})
			}
		}
	}
	return at
}

// TestResize: a longText's grip is in its value's last column on its
// last row, unless its resize is none; the pointer over it says the axes,
// and a drag resizes the value along those alone, the keyboard staying
// where it was.
func TestResize(t *testing.T) {
	run, err := story.Start(story.Find("hotty/resize"))
	if err != nil {
		t.Fatal(err)
	}
	c := run.Surfaces()[0].C
	r := New(c)
	f := r.Draw(60)
	grips := gripsOf(f)
	if len(grips) != 3 {
		t.Fatalf("%d grips, want 3 (none has none):\n%s", len(grips), f.Plain())
	}
	for i, want := range []string{"nwse-resize", "ns-resize", "ew-resize"} {
		if got := r.Pointer(grips[i][0], grips[i][1]); got != want {
			t.Errorf("grip %d: pointer %q, want %q", i, got, want)
		}
	}
	type size struct{ rows, cols int }
	drag := func(i, dx, dy int) {
		t.Helper()
		g := gripsOf(r.Draw(60))[i]
		if err := r.Click(g[0], g[1]); err != nil {
			t.Fatal(err)
		}
		if err := r.Drag(g[0]+dx, g[1]+dy); err != nil {
			t.Fatal(err)
		}
		f := r.Draw(60)
		if at := gripsOf(f)[i]; f.Cells[at[1]][at[0]].Role != Accent {
			t.Errorf("grip %d while dragged: %v, want accent", i, f.Cells[at[1]][at[0]].Role)
		}
		if err := r.Release(); err != nil {
			t.Fatal(err)
		}
	}
	areaOf := func(id string) size {
		t.Helper()
		r.Draw(60)
		for _, h := range r.hits {
			if h.id == id && h.field != nil {
				return size{h.field.rows, h.field.w + 1}
			}
		}
		t.Fatalf("no field %s", id)
		return size{}
	}
	was := map[string]size{}
	for _, id := range []string{"both", "vertical", "horizontal"} {
		was[id] = areaOf(id)
	}
	drag(0, -30, 0)
	drag(0, 0, 2)
	if got, want := areaOf("both"), (size{was["both"].rows + 2, was["both"].cols - 30}); got != want {
		t.Errorf("both: %v, want %v", got, want)
	}
	drag(1, -20, 1)
	if got, want := areaOf("vertical"), (size{was["vertical"].rows + 1, was["vertical"].cols}); got != want {
		t.Errorf("vertical: %v, want %v", got, want)
	}
	drag(2, -20, -2)
	if got, want := areaOf("horizontal"), (size{was["horizontal"].rows, was["horizontal"].cols - 20}); got != want {
		t.Errorf("horizontal: %v, want %v", got, want)
	}
	drag(2, -100, 0)
	if got := areaOf("horizontal").cols; got != minGripCols {
		t.Errorf("horizontal at its narrowest: %d columns, want %d", got, minGripCols)
	}
	if c.St.Focus != "" {
		t.Errorf("the grips took the keyboard: %q", c.St.Focus)
	}
}
