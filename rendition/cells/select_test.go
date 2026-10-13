package cells

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// sidePage is a docs page: a nav beside a HottyMarkdown, on the same
// rows.
func sidePage(t *testing.T, md string) (*Rendition, *view.Controller) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	text, _ := json.Marshal(md)
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Row","children":["nav","doc"]},
	 {"id":"nav","component":"Text","text":"Nav one\n\nNav two\n\nNav three"},
	 {"id":"doc","component":"HottyMarkdown",` + h + `,"weight":1,"text":` + string(text) + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c
}

const selPage = "First paragraph that is long enough to wrap twice here.\n\n" +
	"- one item\n- two item\n\n> quoted words\n\nLast [link](more.md) line."

// selectDrag presses at the first cell with from, drags to the last cell
// of to, and lets go; it returns the frame drawn during the drag.
func selectDrag(t *testing.T, r *Rendition, from, to string) *Frame {
	t.Helper()
	f := r.Draw(48)
	fy := rowOf(t, f, from)
	fx := colIn(t, f, fy, from)
	ty := rowOf(t, f, to)
	tx := colIn(t, f, ty, to) + len([]rune(to)) - 1
	if err := r.Click(fx, fy); err != nil {
		t.Fatal(err)
	}
	if err := r.Drag(tx, ty); err != nil {
		t.Fatal(err)
	}
	during := r.Draw(48)
	if err := r.Release(); err != nil {
		t.Fatal(err)
	}
	return during
}

// TestSelectCopy: a drag on a page's text selects it in reading order,
// its wrapped lines joined, its list's markers kept and the quote's bar
// left out, paragraphs a blank line apart; the release copies it once.
func TestSelectCopy(t *testing.T) {
	r, _ := sidePage(t, selPage)
	f := selectDrag(t, r, "paragraph", "quoted")
	want := "paragraph that is long enough to wrap twice here.\n\n• one item\n• two item\n\nquoted"
	got, ok := r.Copied()
	if !ok || got != want {
		t.Fatalf("copied %q, %v; want %q\n%s", got, ok, want, f.Plain())
	}
	if _, ok := r.Copied(); ok {
		t.Error("Copied twice")
	}
	if got := r.Selection(); got != want {
		t.Errorf("the selection after the release: %q", got)
	}
	y := rowOf(t, f, "one item")
	x := colIn(t, f, y, "one item")
	if c := f.Cells[y][x]; c.Back != Selection || c.BackMix != 255 {
		t.Errorf("a selected cell: Back %v, BackMix %d", c.Back, c.BackMix)
	}
	if c := f.Cells[y][0]; c.BackMix != 0 {
		t.Errorf("the nav's cell on a selected row is selected too: %+v", c)
	}
}

// TestSelectStaysInText: a drag from the page into the nav beside it
// selects to the start of the page's row there, never the nav's text;
// a click (no drag) selects nothing and copies nothing; a press on a link
// follows it rather than selecting; Escape drops the selection.
func TestSelectStaysInText(t *testing.T) {
	r, c := sidePage(t, selPage)
	f := r.Draw(48)
	y := rowOf(t, f, "two item")
	x := colIn(t, f, y, "item")
	if err := r.Click(x+3, y); err != nil {
		t.Fatal(err)
	}
	if err := r.Drag(1, y); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if got, _ := r.Copied(); got != "• two item" {
		t.Errorf("into the nav: %q, want the page's row from its start", got)
	}
	r.Draw(48)
	if err := r.Click(x, y); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if got, ok := r.Copied(); ok || r.Selection() != "" {
		t.Errorf("a click: copied %q, selected %q", got, r.Selection())
	}
	f = r.Draw(48)
	ly := rowOf(t, f, "link")
	lx := colIn(t, f, ly, "link")
	if got := r.Pointer(lx, ly); got != "pointer" {
		t.Errorf("over the link: %q", got)
	}
	if got := r.Pointer(lx-3, ly); got != "text" {
		t.Errorf("over the page's text: %q", got)
	}
	if err := r.Click(lx, ly); err != nil {
		t.Fatal(err)
	}
	r.Drag(lx+6, ly)
	r.Release()
	if _, ok := r.Copied(); ok || !strings.HasPrefix(c.St.Focus, "doc/link/") {
		t.Errorf("a press on the link: copied %v, the keyboard on %q", ok, c.St.Focus)
	}
	selectDrag(t, r, "Last", "line")
	if r.Selection() == "" {
		t.Fatal("no selection")
	}
	r.Key("Escape")
	if got := r.Selection(); got != "" {
		t.Errorf("after Escape: %q", got)
	}
}

func TestOSC52(t *testing.T) {
	if got, want := OSC52("hi"), "\x1b]52;c;aGk=\x1b\\"; got != want {
		t.Errorf("%q, want %q", got, want)
	}
}

// TestSelectWords: in a Text, a double click selects the word under it
// and a triple click its paragraph, the rows a wrap made of one joined,
// and each copies at once; a word on a wrapped row is selected as well,
// and a drag after it extends nothing.
func TestSelectWords(t *testing.T) {
	r, _ := sidePage(t, selPage)
	now := time.Unix(0, 0)
	r.Clock = func() time.Time { return now }
	f := r.Draw(48)
	run := func(word string, n int) string {
		t.Helper()
		now = now.Add(time.Second)
		y := rowOf(t, f, word)
		x := colIn(t, f, y, word) + 1
		for range n {
			now = now.Add(100 * time.Millisecond)
			if err := r.Click(x, y); err != nil {
				t.Fatal(err)
			}
			if err := r.Release(); err != nil {
				t.Fatal(err)
			}
		}
		got, ok := r.Copied()
		if !ok {
			t.Fatalf("%d clicks on %q copied nothing", n, word)
		}
		return got
	}
	if got := run("paragraph", 2); got != "paragraph" {
		t.Errorf("a double click on paragraph copied %q", got)
	}
	if got := r.Selection(); got != "paragraph" {
		t.Errorf("the selection after it: %q", got)
	}
	if got := run("paragraph", 3); got != "First paragraph that is long enough to wrap twice here." {
		t.Errorf("a triple click copied %q", got)
	}
	if got := run("twice", 2); got != "twice" {
		t.Errorf("a double click on a wrapped row's word copied %q", got)
	}
	now = now.Add(time.Second)
	y := rowOf(t, f, "quoted")
	x := colIn(t, f, y, "quoted")
	for range 2 {
		now = now.Add(100 * time.Millisecond)
		if err := r.Click(x, y); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Drag(x+10, y); err != nil {
		t.Fatal(err)
	}
	if err := r.Release(); err != nil {
		t.Fatal(err)
	}
	if got := r.Selection(); got != "quoted" {
		t.Errorf("a drag after a double click: %q", got)
	}
}
