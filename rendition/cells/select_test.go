package cells

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	hottygo "github.com/neuroplastio/hotty-go"

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
// left out, paragraphs a blank line apart; the release puts it in the
// primary selection, once.
func TestSelectCopy(t *testing.T) {
	r, _ := sidePage(t, selPage)
	f := selectDrag(t, r, "paragraph", "quoted")
	want := "paragraph that is long enough to wrap twice here.\n\n• one item\n• two item\n\nquoted"
	got, ok := r.Copied()
	if !ok || got != (Clip{Text: want, Primary: true}) {
		t.Fatalf("copied %+v, %v; want %q to the primary selection\n%s", got, ok, want, f.Plain())
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
	if got, _ := r.Copied(); got.Text != "• two item" {
		t.Errorf("into the nav: %q, want the page's row from its start", got.Text)
	}
	r.Draw(48)
	if err := r.Click(x, y); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if got, ok := r.Copied(); ok || r.Selection() != "" {
		t.Errorf("a click: copied %q, selected %q", got.Text, r.Selection())
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
	r.Key("escape")
	if got := r.Selection(); got != "" {
		t.Errorf("after Escape: %q", got)
	}
}

func TestOSC52(t *testing.T) {
	if got, want := (Clip{Text: "hi"}).OSC52(), "\x1b]52;c;aGk=\x1b\\"; got != want {
		t.Errorf("%q, want %q", got, want)
	}
	if got, want := (Clip{Text: "hi", Primary: true}).OSC52(), "\x1b]52;p;aGk=\x1b\\"; got != want {
		t.Errorf("%q, want %q", got, want)
	}
}

// TestCopyKeys: the keymap's copy, mod+c, copies a Text's selection to
// the clipboard and leaves it selected: ctrl+c where Mod is ctrl,
// the default, and super+c (Command+c) where the host says super, the other
// then not bound. With nothing selected the key is the program's, as it
// is once the keymap binds it to program, and another key bound to copy
// copies instead (SPEC §10.2, §10.4).
func TestCopyKeys(t *testing.T) {
	r, c := sidePage(t, selPage)
	if ok, _ := r.Key("ctrl+c"); ok {
		t.Fatal("ctrl+c with nothing selected was taken")
	}
	selectDrag(t, r, "one item", "one item")
	r.Copied()
	copies := func(k string) bool {
		t.Helper()
		ok, err := r.Key(k)
		got, copied := r.Copied()
		if ok != copied || err != nil || copied && got != (Clip{Text: "one item"}) {
			t.Errorf("%s: handled %v, %v; copied %+v, %v", k, ok, err, got, copied)
		}
		return copied
	}
	if !copies("ctrl+c") || copies("super+c") {
		t.Error("where Mod is ctrl, ctrl+c copies and super+c does not")
	}
	if got := r.Selection(); got != "one item" {
		t.Errorf("the selection after a copy: %q", got)
	}
	c.Platform = hottygo.Platform{Mod: "super"}
	if !copies("super+c") || copies("ctrl+c") {
		t.Error("where Mod is super, super+c copies and ctrl+c does not")
	}
	c.Platform = hottygo.Platform{}
	r.SetKeys("mod+c=program")
	if copies("ctrl+c") {
		t.Error("ctrl+c bound to program copied")
	}
	r.SetKeys("ctrl+shift+c=copy")
	if !copies("ctrl+shift+c") {
		t.Error("ctrl+shift+c, bound to copy, did not copy")
	}
	r.Deselect()
	if ok, _ := r.Key("ctrl+shift+c"); ok || r.Selection() != "" {
		t.Error("Deselect left a selection")
	}
}

// TestCopyField: in the text control that has the keyboard, a copy key
// copies what is selected there; a double click there puts the word in
// the primary selection; a password field copies nothing.
func TestCopyField(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	keys(t, r, "shift+a", "d", "a", "space", "shift+l", "shift+arrow_left", "shift+arrow_left")
	if ok, _ := r.Key("ctrl+c"); !ok {
		t.Fatal("ctrl+c on a selection in the field was not taken")
	}
	if got, _ := r.Copied(); got != (Clip{Text: " L"}) {
		t.Errorf("copied %+v, want \" L\"", got)
	}
	keys(t, r, "arrow_right")
	if ok, _ := r.Key("ctrl+c"); ok {
		t.Error("ctrl+c with nothing selected in the field was taken")
	}
	now := time.Unix(0, 0)
	r.Clock = func() time.Time { return now }
	for range 2 {
		now = now.Add(100 * time.Millisecond)
		if err := r.Click(nameAt+1, 0); err != nil {
			t.Fatal(err)
		}
	}
	r.Release()
	if got, _ := r.Copied(); got != (Clip{Text: "Ada", Primary: true}) {
		t.Errorf("a double click on Ada: %+v", got)
	}
	p := a2ui.NewProcessor(basic.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"TextField","label":"Password","variant":"obscured","value":"secret"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c = view.NewController(p.Surface("s"))
	r = New(c)
	c.Focus("root")
	r.Draw(40)
	keys(t, r, "ctrl+a")
	if ok, _ := r.Key("ctrl+c"); ok {
		t.Error("ctrl+c copied from a password field")
	}
}

// TestCutPaste: in a text control with the keyboard, its keymap's cut cuts what
// is selected to the clipboard, and with nothing selected is the
// program's; a paste key asks for the clipboard, once, and Paste types it
// in place of the selection, a one-line field without its line breaks.
// Elsewhere both keys are the program's, and so is a paste.
func TestCutPaste(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	keys(t, r, "shift+a", "d", "a", "space", "shift+l", "shift+arrow_left", "shift+arrow_left")
	if ok, err := r.Key("ctrl+x"); !ok || err != nil {
		t.Fatalf("ctrl+x on a selection: %v, %v", ok, err)
	}
	if got, _ := r.Copied(); got != (Clip{Text: " L"}) {
		t.Errorf("cut %+v, want \" L\" on the clipboard", got)
	}
	if got := data()["name"]; got != "Ada" {
		t.Errorf("after the cut the name is %q", got)
	}
	if ok, _ := r.Key("ctrl+x"); ok {
		t.Error("ctrl+x with nothing selected was taken")
	}
	if r.PasteWanted() {
		t.Error("a paste was wanted before a paste key")
	}
	if ok, _ := r.Key("ctrl+v"); !ok || !r.PasteWanted() || r.PasteWanted() {
		t.Error("ctrl+v did not ask for the clipboard once")
	}
	keys(t, r, "shift+arrow_left", "shift+arrow_left")
	if ok, err := r.Paste("ny\nthing"); !ok || err != nil {
		t.Fatalf("Paste: %v, %v", ok, err)
	}
	if got := data()["name"]; got != "Anything" {
		t.Errorf("pasting over \"da\" made %q", got)
	}
	c.Focus("note")
	r.Draw(40)
	if _, err := r.Paste("one\r\ntwo"); err != nil {
		t.Fatal(err)
	}
	if got := data()["note"]; got != "one\ntwo" {
		t.Errorf("pasting in a textarea made %q", got)
	}
	r.SetKeys("mod+v=program ctrl+shift+x=cut")
	if ok, _ := r.Key("ctrl+v"); ok {
		t.Error("ctrl+v bound to program was taken")
	}
	keys(t, r, "shift+arrow_left")
	if ok, _ := r.Key("ctrl+shift+x"); !ok {
		t.Error("ctrl+shift+x, bound to cut, did not cut")
	}
	c.Focus("go")
	r.Draw(40)
	r.SetKeys("")
	if ok, _ := r.Key("ctrl+v"); ok || r.PasteWanted() {
		t.Error("ctrl+v on a button was taken")
	}
	if ok, _ := r.Paste("x"); ok {
		t.Error("a paste with no text control was taken")
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
		if !ok || !got.Primary {
			t.Fatalf("%d clicks on %q: %+v, %v", n, word, got, ok)
		}
		return got.Text
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
