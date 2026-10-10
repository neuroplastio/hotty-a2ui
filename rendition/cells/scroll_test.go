package cells

import (
	"fmt"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// scrolling is a surface whose root is a HottyScrollView with the props
// given (JSON, after its id), lines /log bound to log; and the actions it
// sent.
func scrolling(t *testing.T, props, log string, more ...string) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name)
		}
	}
	comps := `{"id":"root","component":"HottyScrollView","catalogId":"` + hotty.ID + `",` + props + `}`
	for _, m := range more {
		comps += "," + m
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"log":` + log + `}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[` + comps + `]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// sixLines are "line 1" to "line 6".
const sixLines = `["line 1","line 2","line 3","line 4","line 5","line 6"]`

// first is a frame's first row as text.
func first(f *Frame) string { return strings.SplitN(f.Plain(), "\n", 2)[0] }

// A scroll view shows height rows of its lines a column in, a blank
// column, then the scrollbar: a thumb of ┃ as long as the share that shows, where it shows,
// on a track of │. The thumb is muted, and accent while the view has the
// keyboard; the track is border, faint.
func TestScrollDraws(t *testing.T) {
	r, c, _ := scrolling(t, `"lines":{"@path":"/log"},"height":3`, sixLines)
	f := r.Draw(20)
	want := fmt.Sprintf(" %-18s┃\n %-18s┃\n %-18s│", "line 1", "line 2", "line 3")
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if thumb, track := f.Cells[0][19], f.Cells[2][19]; thumb.Role != Muted || track.Role != Border || track.Attr&Faint == 0 {
		t.Errorf("thumb %v, track %v %v", thumb.Role, track.Role, track.Attr)
	}
	c.Focus("root")
	if thumb := r.Draw(20).Cells[0][19]; thumb.Role != Accent {
		t.Errorf("focused, the thumb is %v", thumb.Role)
	}
	r2, _, _ := scrolling(t, `"lines":["short"],"height":3`, `[]`)
	if got := r2.Draw(20).Plain(); got != " short\n\n" {
		t.Errorf("all of it shows, so no bar: %q", got)
	}
}

// Keys scroll it as bubbles' viewport takes them, clamped to its lines;
// so does the wheel, three rows a notch, over the box only.
func TestScrollKeys(t *testing.T) {
	r, c, _ := scrolling(t, `"lines":{"@path":"/log"},"height":3`, sixLines)
	c.Focus("root")
	r.Draw(20)
	for _, step := range []struct {
		key string
		top int
	}{{"j", 1}, {"ArrowDown", 2}, {"G", 3}, {"ArrowDown", 3}, {"k", 2}, {"g", 0}, {"f", 3}, {"b", 0}, {"Space", 3},
		{"PageUp", 0}, {"d", 1}, {"Control+d", 2}, {"u", 1}, {"Control+u", 0}, {"End", 3}, {"Home", 0}} {
		if ok, err := r.Key(step.key); !ok || err != nil {
			t.Fatalf("%s: %v %v", step.key, ok, err)
		}
		r.Draw(20)
		if got := c.V.Find("root").Top; got != step.top {
			t.Fatalf("after %s: top %d, want %d", step.key, got, step.top)
		}
	}
	if !r.Wheel(5, 1, 0, 1) || c.V.Find("root").Top != 3 {
		t.Fatalf("the wheel down: top %d", c.V.Find("root").Top)
	}
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " line 4") {
		t.Errorf("after the wheel: %q", got)
	}
	if r.Wheel(5, 1, 0, 1) {
		t.Error("the wheel past the end was taken")
	}
	if r.Wheel(5, 3, 0, -1) {
		t.Error("the wheel below the box was taken")
	}
	if ok, _ := r.Key("ArrowRight"); ok {
		t.Error("→ scrolls lines that all fit")
	}
}

// With follow it starts at the end and stays there as lines arrive, until
// the user scrolls up; back at the end, it follows again.
func TestScrollFollows(t *testing.T) {
	r, c, _ := scrolling(t, `"lines":{"@path":"/log"},"height":3,"follow":true`, sixLines)
	add := func(i int) {
		t.Helper()
		if err := c.S.Write(fmt.Sprintf("/log/%d", i-1), fmt.Sprintf("line %d", i)); err != nil {
			t.Fatal(err)
		}
		c.Rebuild()
	}
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " line 4") {
		t.Fatalf("it starts at %q", got)
	}
	add(7)
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " line 5") {
		t.Fatalf("a line arrived: %q", got)
	}
	c.Focus("root")
	keys(t, r, "k")
	add(8)
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " line 4") {
		t.Fatalf("scrolled up, a line arrived: %q", got)
	}
	keys(t, r, "End")
	add(9)
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " line 7") {
		t.Errorf("back at the end, a line arrived: %q", got)
	}
	// Without follow, hottyScrollTo's end is where it goes, once.
	r, c, _ = scrolling(t, `"lines":{"@path":"/log"},"height":3`, sixLines)
	r.Draw(20)
	c.ScrollTo("root", "end")
	r.Draw(20)
	add(7)
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " line 4") {
		t.Errorf("scrolled to the end, then a line: %q", got)
	}
}

// Lines that do not wrap are cut at the box, and ← and → scroll them six
// columns; lines that wrap take the rows they need, and do not scroll
// sideways.
func TestScrollSideways(t *testing.T) {
	long := `["0123456789abcdefghijklmnopqrstuvwxyz","x"]`
	r, c, _ := scrolling(t, `"lines":{"@path":"/log"},"height":2`, long)
	c.Focus("root")
	if got := first(r.Draw(21)); got != " 0123456789abcdefgh" {
		t.Fatalf("cut: %q", got)
	}
	// At 21 columns, 18 of lines: keys() draws at 40, where the lines fit.
	at21 := func(ks ...string) string {
		t.Helper()
		for _, k := range ks {
			if ok, err := r.Key(k); !ok || err != nil {
				t.Fatalf("%s: %v %v", k, ok, err)
			}
			r.Draw(21)
		}
		return first(r.Draw(21))
	}
	if got := at21("l", "ArrowRight"); got != " cdefghijklmnopqrst" {
		t.Errorf("12 columns right: %q", got)
	}
	if got := at21("l", "l", "l"); got != " ijklmnopqrstuvwxyz" {
		t.Errorf("right, clamped to the widest line: %q", got)
	}
	if got := at21("h", "h", "h", "h", "h"); got != " 0123456789abcdefgh" {
		t.Errorf("back left, clamped: %q", got)
	}
	r, c, _ = scrolling(t, `"lines":{"@path":"/log"},"height":3,"wrap":true`, long)
	c.Focus("root")
	if got := r.Draw(21).Plain(); got != " 0123456789abcdefgh\n ijklmnopqrstuvwxyz\n x" {
		t.Errorf("wrapped:\n%s", got)
	}
	if ok, _ := r.Key("l"); ok {
		t.Error("l scrolls lines that wrap")
	}
}

// A scroll view's box is filled with the surface's colour, so that it
// stands apart, where the theme colours the surface; what its content
// tints keeps its own, and the terminal's background is left alone.
func TestScrollSurface(t *testing.T) {
	r, _, _ := scrolling(t, `"lines":{"@path":"/log"},"height":2`, `["a"]`)
	f := r.Draw(10)
	for _, at := range [][2]int{{0, 0}, {9, 0}, {9, 1}} {
		if c := f.Cells[at[1]][at[0]]; c.Back != Surface || c.BackMix != 255 {
			t.Errorf("(%d,%d): %v %d", at[0], at[1], c.Back, c.BackMix)
		}
	}
	th := theme.Theme{Name: "t", Bg: "#000000", Fg: "#ffffff", Surface: "#112233"}
	if !strings.Contains(f.Themed(th), "48;2;17;34;51") {
		t.Error("no surface in a theme that colours it")
	}
	if strings.Contains(f.Themed(theme.Default), "48;") {
		t.Error("the terminal's background is painted")
	}
	r = coding(t, `{"id":"v","component":"HottyScrollView","catalogId":"`+hotty.ID+`","height":1,"child":"c"}`,
		codeComp("c", `"code":"a","marks":[{"line":1,"kind":"error"}]`))
	if c := r.Draw(10).Cells[0][2]; c.Back != Error || c.BackMix != 42 {
		t.Errorf("a marked line in the box: %v %d", c.Back, c.BackMix)
	}
}

// A child is laid out at the box's width and shows through it: what it
// hides takes no click, what shows takes it where it shows, and the
// element with the keyboard is scrolled into sight.
func TestScrollChild(t *testing.T) {
	var buttons, ids []string
	for i := 1; i <= 6; i++ {
		buttons = append(buttons, fmt.Sprintf(`{"id":"b%d","component":"Button","child":"t%d","action":{"event":{"name":"press%d"}}},{"id":"t%d","component":"Text","text":"B%d"}`, i, i, i, i, i))
		ids = append(ids, fmt.Sprintf(`"b%d"`, i))
	}
	col := `{"id":"col","component":"Column","children":[` + strings.Join(ids, ",") + `]}`
	r, c, actions := scrolling(t, `"child":"col","height":3`, `[]`, append([]string{col}, buttons...)...)
	if got := r.Draw(20).Plain(); got != fmt.Sprintf(" %-18s┃\n %-18s┃\n %-18s│", "[ B1 ]", "[ B2 ]", "[ B3 ]") {
		t.Fatalf("got\n%s", r.Draw(20).Plain())
	}
	if _, _, _, _, ok := r.Box("b4"); ok {
		t.Error("b4 is hidden, yet has a box")
	}
	must(t, r.Click(2, 1))
	c.Focus("root")
	keys(t, r, "G")
	if got := first(r.Draw(20)); !strings.HasPrefix(got, " [ B4 ]") {
		t.Fatalf("at the end: %q", got)
	}
	must(t, r.Click(2, 2))
	if got := strings.Join(*actions, " "); got != "press2 press6" {
		t.Errorf("clicks pressed %q", got)
	}
	c.Focus("b1")
	r.Draw(20)
	if top := c.V.Find("root").Top; top != 0 {
		t.Errorf("b1 has the keyboard, top %d", top)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// A Column sets a blank row after a scroll view, whose box draws no edge
// but its scrollbar, so that what follows does not read as its content;
// its title above it stays on the row before it.
func TestScrollSetApart(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["title","log","after"]},
	 {"id":"title","component":"Text","text":"Log"},
	 {"id":"log","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["one"],"height":1},
	 {"id":"after","component":"Text","text":"After"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	if got := New(view.NewController(p.Surface("s"))).Draw(20).Plain(); got != "Log\n one\n\nAfter" {
		t.Errorf("got %q", got)
	}
}

// A scroll view's natural width is its widest line, its padding and its
// scrollbar, whether it wraps them or not: in a Row it takes that, not
// less.
func TestScrollWidth(t *testing.T) {
	for _, wrap := range []string{"false", "true"} {
		p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
		msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
		{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
		 {"id":"root","component":"Row","children":["log","after"]},
		 {"id":"log","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["a line"],"height":1,"wrap":` + wrap + `},
		 {"id":"after","component":"Text","text":"|"}]}}]`
		if err := p.ProcessJSON([]byte(msgs)); err != nil {
			t.Fatal(err)
		}
		if got := New(view.NewController(p.Surface("s"))).Draw(40).Plain(); got != " a line   |" {
			t.Errorf("wrap %s: %q", wrap, got)
		}
	}
}
