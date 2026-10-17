package cells

import (
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
)

// TouchAreas are the tracks of the Sliders and HottyRangeSliders that are
// not disabled, as the last frame drew them: a row each, across, the
// track's cells and no more, in the order painted (SPEC §9.6). A range's
// knobs are cells of its track, not areas of their own.
func TestTouchAreas(t *testing.T) {
	r, _ := ranges(t)
	if got := r.TouchAreas(); got != nil {
		t.Errorf("areas before a Draw: %+v", got)
	}
	f := r.Draw(40)
	var want []hotty.TouchArea
	for _, s := range []struct {
		id       string
		min, max float64
	}{{"price", 0, 100}, {"rating", 1, 5}} {
		c0, row, ok := r.TrackCell(s.id, s.min)
		c1, _, _ := r.TrackCell(s.id, s.max)
		if !ok {
			t.Fatalf("no track for %s", s.id)
		}
		want = append(want, hotty.TouchArea{C: c0, R: row, W: c1 - c0 + 1, H: 1, Drag: "x"})
	}
	got := r.TouchAreas()
	if !slices.Equal(got, want) {
		t.Fatalf("areas %+v, want %+v (the disabled Year has none)", got, want)
	}
	// Each covers the track's glyphs, and what is beside it is not track.
	track := func(c Cell) bool { return strings.Contains("━⎯■", c.Text) && c.Text != "" }
	for _, a := range got {
		row := f.Cells[a.R]
		for x := a.C; x < a.C+a.W; x++ {
			if !track(row[x]) {
				t.Errorf("%+v: column %d is %q", a, x, row[x].Text)
			}
		}
		if a.C > 0 && track(row[a.C-1]) || a.C+a.W < len(row) && track(row[a.C+a.W]) {
			t.Errorf("%+v: the track goes on past it", a)
		}
	}
}
