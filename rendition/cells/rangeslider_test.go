package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// ranges is a Price HottyRangeSlider from 0 to 100 in steps of 5, bound
// at 20 to 70; a Rating Slider from 1 to 5 at 3 that fills from its end;
// and a disabled HottyRangeSlider, Year, 0 to 10 at 2 to 4.
func ranges(t *testing.T) (*Rendition, *view.Controller) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"lo":20,"hi":70}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["price","rating","year"]},
	 {"id":"price","component":"HottyRangeSlider",` + h + `,"label":"Price","start":{"@path":"/lo"},"end":{"@path":"/hi"},"max":100,"steps":20},
	 {"id":"rating","component":"Slider","label":"Rating","value":3,"min":1,"max":5,
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"fill":"end"}}}},
	 {"id":"year","component":"HottyRangeSlider",` + h + `,"label":"Year","start":2,"end":4,"max":10,"disabled":true}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c
}

// rolesOf is a row's roles from col on, one letter a cell: A accent, b
// border, m muted, f muted and faint, . anything else.
func rolesOf(f *Frame, row, col, n int) string {
	var s []byte
	for _, c := range f.Cells[row][col : col+n] {
		switch {
		case c.Role == Muted && c.Attr&Faint != 0:
			s = append(s, 'f')
		case c.Role == Accent:
			s = append(s, 'A')
		case c.Role == Border:
			s = append(s, 'b')
		case c.Role == Muted:
			s = append(s, 'm')
		default:
			s = append(s, '.')
		}
	}
	return string(s)
}

// A HottyRangeSlider is a Slider's row with two knobs, the range between
// them ━ and the rest ⎯ in border; while a knob has the keyboard, it and
// the range are in the accent, and its number is bold. A Slider with fill
// "end" draws its active part from its knob to max. A disabled range is
// muted and faint.
func TestRangeSliderDraws(t *testing.T) {
	r, c := ranges(t)
	want := "  Price ⎯⎯⎯■━━━━━■⎯⎯⎯⎯ 20–70\n" +
		"  Rating ⎯⎯⎯⎯⎯⎯⎯⎯■━━━━━━━ 3\n" +
		"  Year ⎯⎯■━━■⎯⎯⎯⎯⎯⎯⎯ 2–4"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, tc := range []struct {
		name     string
		focus    string
		row, col int
		roles    string
	}{
		{"price", "", 0, 8, "bbb.......bbbb"},
		{"rating", "", 1, 9, "bbbbbbbb........"},
		{"rating focused", "rating", 1, 9, "bbbbbbbbAAAAAAAA"},
		{"price's start focused", "price/knob/0", 0, 8, "bbbAAAAAA.bbbb.AA..."},
		{"price's end focused", "price/knob/1", 0, 8, "bbb.AAAAAAbbbb....AA"},
		{"year, disabled", "", 2, 2, "ffffffffffffffffff.fff"},
	} {
		c.Focus(tc.focus)
		f = r.Draw(30)
		if got := rolesOf(f, tc.row, tc.col, len(tc.roles)); got != tc.roles {
			t.Errorf("%s: roles %s, want %s", tc.name, got, tc.roles)
		}
	}
	c.Focus("price/knob/1")
	f = r.Draw(30)
	if num, dash := f.Cells[0][26], f.Cells[0][25]; num.Attr&Bold == 0 || dash.Attr&Bold != 0 {
		t.Errorf("the end's focused, its number %q %v, the dash %q %v", num.Text, num.Attr, dash.Text, dash.Attr)
	}
	// Where the knobs meet, the end's stands a column right of the
	// start's, and at max, the start's a column left.
	for _, tc := range []struct {
		lo, hi float64
		want   string
	}{
		{50, 50, "⎯⎯⎯⎯⎯⎯⎯■■⎯⎯⎯⎯⎯"},
		{100, 100, "⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯■■"},
		{0, 0, "■■⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯"},
		{0, 100, "■━━━━━━━━━━━━■"},
	} {
		if err := c.S.Write("/lo", tc.lo); err != nil {
			t.Fatal(err)
		}
		if err := c.S.Write("/hi", tc.hi); err != nil {
			t.Fatal(err)
		}
		c.Rebuild()
		f = r.Draw(30)
		var got strings.Builder
		for _, cell := range f.Cells[0][8:22] {
			got.WriteString(cell.Text)
		}
		if got.String() != tc.want {
			t.Errorf("at %v–%v: %s, want %s", tc.lo, tc.hi, got.String(), tc.want)
		}
	}
}

// A click on the track off the knobs moves the nearer knob there, which
// takes the keyboard; a drag then moves it, stopped where it meets the
// other. A click on a knob gives it the keyboard where it is. A disabled
// range takes neither.
func TestRangeSliderClicks(t *testing.T) {
	r, c := ranges(t)
	r.Draw(30)
	data := func(p string) any { return c.S.Data.Value(p) }
	at := func(id string, v float64) (int, int) {
		t.Helper()
		col, row, ok := r.TrackCell(id, v)
		if !ok {
			t.Fatalf("no track for %s", id)
		}
		return col, row
	}
	click := func(id string, v float64) {
		t.Helper()
		if err := r.Click(at(id, v)); err != nil {
			t.Fatal(err)
		}
		r.Draw(30)
	}
	drag := func(id string, v float64) {
		t.Helper()
		if err := r.Drag(at(id, v)); err != nil {
			t.Fatal(err)
		}
		r.Draw(30)
	}
	click("price", 85)
	if data("/hi") != 85.0 || data("/lo") != 20.0 || c.St.Focus != "price/knob/1" || !c.St.Keyboard {
		t.Errorf("a click near the end: %v–%v, focus %q", data("/lo"), data("/hi"), c.St.Focus)
	}
	drag("price", 0)
	if data("/hi") != 20.0 {
		t.Errorf("the end dragged past the start: %v", data("/hi"))
	}
	r.Release()
	// They meet at 20: a click there moves neither, and the drag's first
	// move picks the knob it goes towards, here the start's, down.
	click("price", 20)
	if data("/lo") != 20.0 || data("/hi") != 20.0 {
		t.Errorf("a click where they meet moved one: %v–%v", data("/lo"), data("/hi"))
	}
	drag("price", 0)
	if data("/lo") != 0.0 || data("/hi") != 20.0 || c.St.Focus != "price/knob/0" {
		t.Errorf("a drag down from where they meet: %v–%v, focus %q", data("/lo"), data("/hi"), c.St.Focus)
	}
	r.Release()
	click("price", 20) // the end's knob
	if data("/hi") != 20.0 || c.St.Focus != "price/knob/1" {
		t.Errorf("a click on the end's knob: %v, focus %q", data("/hi"), c.St.Focus)
	}
	drag("price", 60)
	r.Release()
	if data("/lo") != 0.0 || data("/hi") != 60.0 {
		t.Errorf("the end's knob dragged to 60: %v–%v", data("/lo"), data("/hi"))
	}
	click("year", 8)
	if lo, hi := c.V.Find("year").Range(); lo != 2 || hi != 4 || c.St.Focus == "year/knob/1" {
		t.Errorf("a click on a disabled range: %v–%v, focus %q", lo, hi, c.St.Focus)
	}
}

// A Slider with fill "end" takes a click as any Slider does.
func TestSliderFillEndClick(t *testing.T) {
	r, c := ranges(t)
	r.Draw(30)
	col, row, ok := r.TrackCell("rating", 5)
	if !ok {
		t.Fatal("no track for rating")
	}
	if err := r.Click(col, row); err != nil {
		t.Fatal(err)
	}
	if v := c.V.Find("rating").Value; v != 5.0 {
		t.Errorf("a click at the track's end: %v", v)
	}
	if got := r.Draw(30).Plain(); !strings.Contains(got, "Rating ⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯■ 5") {
		t.Errorf("at 5:\n%s", got)
	}
}
