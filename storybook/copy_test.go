package storybook

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
)

// TestCopyOrQuit: in cells, Control+c with a story's text selected copies
// it to the clipboard and leaves the storybook running; a press in
// another pane drops the selection, and Control+c then quits. With no
// copy keys, it quits with the text selected.
func TestCopyOrQuit(t *testing.T) {
	for _, o := range []Options{{First: "basic/00_simple-text"}, {First: "basic/00_simple-text", CopyKeys: []string{}}} {
		s := hottytea.New()
		s.Mode = hottytea.Text
		b := New(o)
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
		quit := b.key("Control+c", s)
		if o.CopyKeys != nil {
			if !quit {
				t.Error("no copy keys: Control+c did not quit")
			}
			continue
		}
		if c, ok := b.Copied(); quit || !ok || c != (cells.Clip{Text: "Minimal"}) {
			t.Fatalf("Control+c: quit %v, copied %+v", quit, c)
		}
		nav := b.panes[navID]
		b.Update(tea.MouseClickMsg{X: nav.rect.X + 1, Y: nav.rect.Y + nav.rect.H - 1, Button: tea.MouseLeft}, s)
		b.Update(tea.MouseReleaseMsg{X: nav.rect.X + 1, Y: nav.rect.Y + nav.rect.H - 1, Button: tea.MouseLeft}, s)
		if got := b.Selection(); got != "" {
			t.Errorf("a press in nav left %q selected", got)
		}
		if !b.key("Control+c", s) {
			t.Error("Control+c with nothing selected did not quit")
		}
	}
}
