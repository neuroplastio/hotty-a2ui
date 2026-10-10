package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// suggesting is a Command field with git's commands as suggestions and
// "git c" typed, with the keyboard, and a Note field after it.
func suggesting(t *testing.T) (*Rendition, *view.Controller) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"cmd":"git c","note":"",
	  "cmds":["git status","git commit","git checkout","git cherry-pick","git clone","git config","git diff"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["cmd","note"]},
	 {"id":"cmd","component":"TextField","label":"Command","placeholder":"git …","value":{"@path":"/cmd"},
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"autofocus":true,"suggestions":{"options":{"@path":"/cmds"}}}}}},
	 {"id":"note","component":"TextField","label":"Note","value":{"@path":"/note"}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c
}

func drawn(t *testing.T, r *Rendition, want string) *Frame {
	t.Helper()
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	return f
}

// A field with the keyboard shows the rest of its first suggestion faint
// after the value, under the caret, and the suggestions its value leaves
// under it, a row each in its gutter's bar: what was typed plain, the rest
// muted. More than five show five, and which they are. The highlighted
// one is reversed in the accent. Without the keyboard it is a plain field.
func TestSuggestDraws(t *testing.T) {
	r, c := suggesting(t)
	f := drawn(t, r, "┃ Command\n"+
		"┃ > git commit\n"+
		"┃   git commit\n"+
		"┃   git checkout\n"+
		"┃   git cherry-pick\n"+
		"┃   git clone\n"+
		"┃   git config\n"+
		"\n"+
		"  Note\n"+
		"  >")
	for _, tc := range []struct {
		row, col int
		role     Role
		attr     Attr
	}{
		{1, 8, Fg, 0}, {1, 9, Muted, Faint}, {1, 13, Muted, Faint},
		{2, 0, Accent, 0}, {2, 8, Fg, 0}, {2, 9, Muted, 0},
		{6, 0, Accent, 0},
	} {
		if cell := f.Cells[tc.row][tc.col]; cell.Role != tc.role || cell.Attr != tc.attr {
			t.Errorf("row %d col %d %q: %v %v, want %v %v", tc.row, tc.col, cell.Text, cell.Role, cell.Attr, tc.role, tc.attr)
		}
	}
	if col, row, ok := f.Cursor(); !ok || col != 9 || row != 1 {
		t.Errorf("the caret is at %d,%d %v, want after the value, on the ghost", col, row, ok)
	}

	r.Key("ArrowDown")
	r.Key("ArrowDown")
	f = drawn(t, r, "┃ Command\n"+
		"┃ > git checkout\n"+
		"┃   git commit\n"+
		"┃   git checkout\n"+
		"┃   git cherry-pick\n"+
		"┃   git clone\n"+
		"┃   git config\n"+
		"\n"+
		"  Note\n"+
		"  >")
	for col := 3; col <= 16; col++ {
		if cell := f.Cells[3][col]; cell.Role != Accent || cell.Attr != Reverse {
			t.Errorf("highlighted, col %d %q: %v %v", col, cell.Text, cell.Role, cell.Attr)
		}
	}
	if cell := f.Cells[3][2]; cell.Attr != 0 {
		t.Errorf("highlighted, col 2 is %v", cell.Attr)
	}

	r.Key("Backspace")
	r.Key("Backspace")
	drawn(t, r, "┃ Command\n"+
		"┃ > git status\n"+
		"┃   git status\n"+
		"┃   git commit\n"+
		"┃   git checkout\n"+
		"┃   git cherry-pick\n"+
		"┃   git clone\n"+
		"┃   1–5 of 7\n"+
		"\n"+
		"  Note\n"+
		"  >")

	r.Key("Tab")
	r.Key("Tab")
	drawn(t, r, "  Command\n"+
		"  > git status\n"+
		"\n"+
		"┃ Note\n"+
		"┃ >")
	if c.St.Focus != "note" {
		t.Errorf("focus is on %s", c.St.Focus)
	}
}

// With nothing typed, or nothing it leaves, a field shows no list and no
// ghost: its placeholder, or its value alone.
func TestSuggestEmpty(t *testing.T) {
	r, c := suggesting(t)
	c.SetValue("cmd", "")
	f := drawn(t, r, "┃ Command\n"+
		"┃ > git …\n"+
		"\n"+
		"  Note\n"+
		"  >")
	if cell := f.Cells[1][4]; cell.Role != Muted || cell.Attr != Faint {
		t.Errorf("the placeholder is %v %v", cell.Role, cell.Attr)
	}
	c.SetValue("cmd", "hg")
	drawn(t, r, "┃ Command\n"+
		"┃ > hg\n"+
		"\n"+
		"  Note\n"+
		"  >")
}

// A click on a suggestion's row picks it, and the caret goes after it;
// the list shuts.
func TestSuggestClick(t *testing.T) {
	r, c := suggesting(t)
	r.Draw(30)
	if err := r.Click(6, 4); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if got := c.S.Data.Value("/cmd"); got != "git cherry-pick" {
		t.Errorf("picked %v", got)
	}
	f := drawn(t, r, "┃ Command\n"+
		"┃ > git cherry-pick\n"+
		"\n"+
		"  Note\n"+
		"  >")
	if col, row, ok := f.Cursor(); !ok || col != 19 || row != 1 {
		t.Errorf("the caret is at %d,%d %v", col, row, ok)
	}
}
