package storybook

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
)

// TestCopyOrQuit: in cells, ctrl+c with a story's text selected copies
// it to the clipboard and leaves the storybook running; a press in
// another pane drops the selection, and ctrl+c twice then quits.
func TestCopyOrQuit(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Text
	b := New(Options{First: "basic/00_simple-text"})
	var p *pane
	b.View(hottytea.Rect{W: 120, H: 24}, s)
	for _, q := range b.order {
		if q.kind == asCells && !q.chrome(b) {
			p = q
		}
	}
	if p == nil {
		t.Fatal("no pane for the story")
	}
	x, y := -1, -1
	for i, l := range strings.Split(p.frame.Plain(), "\n") {
		if c := strings.Index(l, "Minimal"); c >= 0 {
			x, y = p.rect.X+len([]rune(l[:c])), p.rect.Y+i
		}
	}
	if y < 0 {
		t.Fatalf("no Minimal in the story:\n%s", p.frame.Plain())
	}
	b.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
	b.Update(tea.MouseMotionMsg{X: x + 6, Y: y, Button: tea.MouseLeft}, s)
	b.Update(tea.MouseReleaseMsg{X: x + 6, Y: y, Button: tea.MouseLeft}, s)
	if c, ok := b.Copied(); !ok || c != (cells.Clip{Text: "Minimal", Primary: true}) {
		t.Fatalf("the selection let go: %+v, %v", c, ok)
	}
	if got := b.Selection(); got != "Minimal" {
		t.Errorf("the Book's selection: %q", got)
	}
	quit := b.key("ctrl+c", s)
	if c, ok := b.Copied(); quit || !ok || c != (cells.Clip{Text: "Minimal"}) {
		t.Fatalf("ctrl+c: quit %v, copied %+v", quit, c)
	}
	nav := b.panes[navID]
	b.Update(tea.MouseClickMsg{X: nav.rect.X + 1, Y: nav.rect.Y + nav.rect.H - 1, Button: tea.MouseLeft}, s)
	b.Update(tea.MouseReleaseMsg{X: nav.rect.X + 1, Y: nav.rect.Y + nav.rect.H - 1, Button: tea.MouseLeft}, s)
	if got := b.Selection(); got != "" {
		t.Errorf("a press in nav left %q selected", got)
	}
	if b.key("ctrl+c", s) || !b.key("ctrl+c", s) {
		t.Error("ctrl+c twice with nothing selected did not quit, or once did")
	}
}

// TestQuitTwice: a ctrl+c that nothing copied with asks on the bar for
// another, and a quick second one quits: in a password field, which never
// copies, one press does not. Another key, a click, or QuitWindow passing
// ends the ask; ctrl+q quits at once.
func TestQuitTwice(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Text
	b := New(Options{First: "basic/00_simple-login-form", Bars: true})
	b.View(hottytea.Rect{W: 120, H: 24}, s)
	var p *pane
	for _, q := range b.order {
		if q.kind == asCells && !q.chrome(b) {
			p = q
		}
	}
	x, y := -1, -1
	for i, l := range strings.Split(p.frame.Plain(), "\n") {
		if c := strings.Index(l, "Password"); c >= 0 {
			x, y = p.rect.X+len([]rune(l[:c]))+len("Password: "), p.rect.Y+i
		}
	}
	if y < 0 {
		t.Fatalf("no Password in the story:\n%s", p.frame.Plain())
	}
	click := func() {
		b.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
		b.Update(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
	}
	click()
	if bar, _ := b.View(hottytea.Rect{W: 120, H: 24}, s); !strings.Contains(bar, "Ctrl+C copy, twice to quit") {
		t.Errorf("the bar's keys:\n%s", bar)
	}
	for _, k := range []string{"p", "w", "ctrl+a"} {
		b.key(k, s)
	}
	asks := func(when string) {
		t.Helper()
		if b.key("ctrl+c", s) {
			t.Fatalf("%s: one ctrl+c quit", when)
		}
		if bar, _ := b.View(hottytea.Rect{W: 120, H: 24}, s); !strings.Contains(bar, "Ctrl+C again to quit") {
			t.Fatalf("%s: the bar does not ask for another:\n%s", when, bar)
		}
	}
	asks("in the password field")
	if e := p.s.C.V.Find(p.s.C.St.Focus); e == nil || e.Value != "pw" {
		t.Fatalf("the password field: %+v", e)
	}
	b.key("x", s)
	asks("after another key")
	click()
	if bar, _ := b.View(hottytea.Rect{W: 120, H: 24}, s); strings.Contains(bar, "again to quit") {
		t.Errorf("a click left the ask on the bar:\n%s", bar)
	}
	if b.Tick() == nil {
		t.Fatal("no timer for the ask")
	}
	b.Update(quitEnd{b.quitAt}, s)
	if bar, _ := b.View(hottytea.Rect{W: 120, H: 24}, s); strings.Contains(bar, "again to quit") {
		t.Errorf("the window's end left the ask on the bar:\n%s", bar)
	}
	asks("after the window")
	old := b.quitAt
	b.key("x", s)
	asks("after another ask")
	b.Update(quitEnd{old}, s)
	if bar, _ := b.View(hottytea.Rect{W: 120, H: 24}, s); !strings.Contains(bar, "again to quit") {
		t.Errorf("an earlier ask's end took a later one off the bar:\n%s", bar)
	}
	b.quitAt = b.quitAt.Add(-2 * QuitWindow)
	asks("a second press after the window")
	if !b.key("ctrl+c", s) {
		t.Error("a quick second ctrl+c did not quit")
	}
	if !b.key("ctrl+q", s) {
		t.Error("ctrl+q did not quit")
	}
}

// TestPaste: in cells, a paste key in a story's text field asks for the
// terminal's clipboard (PasteWanted), and Update pastes what comes back
// there, as it does the terminal's own paste.
func TestPaste(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Text
	b := New(Options{First: "basic/00_simple-login-form"})
	b.View(hottytea.Rect{W: 120, H: 24}, s)
	var p *pane
	for _, q := range b.order {
		if q.kind == asCells && !q.chrome(b) {
			p = q
		}
	}
	x, y := -1, -1
	for i, l := range strings.Split(p.frame.Plain(), "\n") {
		if c := strings.Index(l, "Username"); c >= 0 {
			x, y = p.rect.X+len([]rune(l[:c]))+len("Username: "), p.rect.Y+i
		}
	}
	if y < 0 {
		t.Fatalf("no Username in the story:\n%s", p.frame.Plain())
	}
	b.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
	b.Update(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
	if b.key("ctrl+v", s) || !b.PasteWanted() || b.PasteWanted() {
		t.Fatal("ctrl+v in the field did not ask for the clipboard once")
	}
	b.Update(tea.ClipboardMsg{Content: "ada", Selection: 'c'}, s)
	b.Update(tea.PasteMsg{Content: "\nlovelace"}, s)
	if e := p.s.C.V.Find(p.s.C.St.Focus); e == nil || e.Value != "adalovelace" {
		t.Errorf("the field after the pastes: %+v", e)
	}
}

// TestEditCommands: on a host that sends edit commands (caps edit), the
// storybook takes them all (SPEC §10.6). A paste into a story's text field
// in cells brings its text, typed with no clipboard read; a copy with
// nothing selected is its key, ctrl+c, which quits the second time.
func TestEditCommands(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Native
	b := New(Options{First: "basic/00_simple-login-form", Rendition: rendCells})
	caps := hotty.Caps{Edit: hotty.DefaultEdit("ctrl"), Mod: "ctrl"}
	b.Update(hottytea.ReadyMsg{Mode: hottytea.Native, Caps: caps}, s)
	if cmd := s.Flush(); cmd == nil {
		t.Fatal("nothing sent on ready")
	} else if raw, _ := cmd().(tea.RawMsg); !strings.Contains(raw.Msg.(string), "a=edit:take=select-all,copy,cut,paste,undo,redo") {
		t.Errorf("the edit commands were not taken: %q", raw.Msg)
	}
	b.View(hottytea.Rect{W: 120, H: 24}, s)
	var p *pane
	for _, q := range b.order {
		if q.kind == asCells && !q.chrome(b) {
			p = q
		}
	}
	if p == nil {
		t.Fatal("no cells pane for the story")
	}
	x, y := -1, -1
	for i, l := range strings.Split(p.frame.Plain(), "\n") {
		if c := strings.Index(l, "Username"); c >= 0 {
			x, y = p.rect.X+len([]rune(l[:c]))+len("Username: "), p.rect.Y+i
		}
	}
	if y < 0 {
		t.Fatalf("no Username in the story:\n%s", p.frame.Plain())
	}
	b.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
	b.Update(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft}, s)
	command := func(c hotty.Command) bool {
		t.Helper()
		return b.Update(hottytea.EventMsg{Event: hotty.Event{Kind: hotty.EventCommand, Command: &c}}, s)
	}
	if command(hotty.Command{Name: "paste", Key: "ctrl+v", Text: "ada", HasText: true}) || b.PasteWanted() {
		t.Fatal("a paste quit, or read the clipboard")
	}
	if e := p.s.C.V.Find(p.s.C.St.Focus); e == nil || e.Value != "ada" {
		t.Errorf("the field after the paste: %+v", e)
	}
	if command(hotty.Command{Name: "copy", Key: "ctrl+c"}) || !command(hotty.Command{Name: "copy", Key: "ctrl+c"}) {
		t.Error("copy twice with nothing selected did not quit, or once did")
	}
}
