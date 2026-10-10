package cells

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// suggesting is a Command field with git's commands as suggestions and
// "git c" typed, with the keyboard, and a Note field after it; ext adds
// to the suggestions' extension.
func suggesting(t *testing.T, ext ...string) (*Rendition, *view.Controller) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	more := ""
	for _, x := range ext {
		more += "," + x
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"cmd":"git c","note":"",
	  "cmds":["git status","git commit","git checkout","git cherry-pick","git clone","git config","git diff"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["cmd","note"]},
	 {"id":"cmd","component":"TextField","label":"Command","placeholder":"git …","value":{"@path":"/cmd"},
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"autofocus":true,"suggestions":{"options":{"@path":"/cmds"}` + more + `}}}}},
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
// in a box over what is under it, a row each, its text under the value's:
// what was typed plain, the rest muted. It moves no row: the Note field
// is under the box, and the frame grows to hold it. More than five show
// five, and which they are. The highlighted one is reversed in the accent
// across the box. Without the keyboard it is a plain field.
func TestSuggestDraws(t *testing.T) {
	r, c := suggesting(t)
	// "┃ Command  git c": the value from column 11, past the labels' run
	// and the inset, which is not underlined; the box's text under it.
	f := drawn(t, r, "┃ Command  git commit\n"+
		"  Note   ╭─────────────────╮\n"+
		"         │ git commit      │\n"+
		"         │ git checkout    │\n"+
		"         │ git cherry-pick │\n"+
		"         │ git clone       │\n"+
		"         │ git config      │\n"+
		"         ╰─────────────────╯")
	for _, tc := range []struct {
		row, col int
		role     Role
		attr     Attr
	}{
		{0, 0, Accent, 0}, {0, 10, Fg, 0}, {0, 15, Fg, Underline}, {0, 16, Muted, Faint | Underline}, {0, 20, Muted, Faint | Underline},
		{1, 0, Fg, 0}, {1, 9, Border, 0}, {2, 11, Fg, 0}, {2, 16, Muted, 0},
	} {
		if cell := f.Cells[tc.row][tc.col]; cell.Role != tc.role || cell.Attr != tc.attr {
			t.Errorf("row %d col %d %q: %v %v, want %v %v", tc.row, tc.col, cell.Text, cell.Role, cell.Attr, tc.role, tc.attr)
		}
	}
	if col, row, ok := f.Cursor(); !ok || col != 16 || row != 0 {
		t.Errorf("the caret is at %d,%d %v, want after the value, on the ghost", col, row, ok)
	}

	r.Key("ArrowDown")
	r.Key("ArrowDown")
	f = drawn(t, r, "┃ Command  git checkout\n"+
		"  Note   ╭─────────────────╮\n"+
		"         │ git commit      │\n"+
		"         │ git checkout    │\n"+
		"         │ git cherry-pick │\n"+
		"         │ git clone       │\n"+
		"         │ git config      │\n"+
		"         ╰─────────────────╯")
	for col := 10; col <= 26; col++ {
		if cell := f.Cells[3][col]; cell.Role != Accent || cell.Attr != Reverse {
			t.Errorf("highlighted, col %d %q: %v %v", col, cell.Text, cell.Role, cell.Attr)
		}
	}
	if cell := f.Cells[3][9]; cell.Attr != 0 || cell.Role != Border {
		t.Errorf("highlighted, col 9 is %v %v", cell.Role, cell.Attr)
	}

	r.Key("Backspace")
	r.Key("Backspace")
	drawn(t, r, "┃ Command  git status\n"+
		"  Note   ╭─────────────────╮\n"+
		"         │ git status      │\n"+
		"         │ git commit      │\n"+
		"         │ git checkout    │\n"+
		"         │ git cherry-pick │\n"+
		"         │ git clone       │\n"+
		"         │ 1–5 of 7        │\n"+
		"         ╰─────────────────╯")

	r.Key("Tab")
	r.Key("Tab")
	drawn(t, r, "  Command  git status\n"+
		"┃ Note")
	if c.St.Focus != "note" {
		t.Errorf("focus is on %s", c.St.Focus)
	}
}

// With nothing typed, or nothing it leaves, a field shows no list and no
// ghost: its placeholder, or its value alone.
func TestSuggestEmpty(t *testing.T) {
	r, c := suggesting(t)
	c.SetValue("cmd", "")
	f := drawn(t, r, "┃ Command  git …\n"+
		"  Note")
	if cell := f.Cells[0][11]; cell.Role != Muted || cell.Attr != Faint|Underline {
		t.Errorf("the placeholder is %v %v", cell.Role, cell.Attr)
	}
	c.SetValue("cmd", "hg")
	drawn(t, r, "┃ Command  hg\n"+
		"  Note")
}

// A click on a suggestion's row picks it, and the caret goes after it;
// the list shuts. A click on the box's border lands on the box, not on the
// field under it.
func TestSuggestClick(t *testing.T) {
	r, c := suggesting(t)
	r.Draw(30)
	if err := r.Click(9, 2); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if c.St.Focus != "cmd" || c.S.Data.Value("/cmd") != "git c" {
		t.Errorf("a click on the border: the keyboard on %s, the value %v", c.St.Focus, c.S.Data.Value("/cmd"))
	}
	r.Draw(30)
	if err := r.Click(14, 4); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if got := c.S.Data.Value("/cmd"); got != "git cherry-pick" {
		t.Errorf("picked %v", got)
	}
	f := drawn(t, r, "┃ Command  git cherry-pick\n"+
		"  Note")
	if col, row, ok := f.Cursor(); !ok || col != 26 || row != 0 {
		t.Errorf("the caret is at %d,%d %v", col, row, ok)
	}
}

// With list false, the field shows the ghost alone, as bubbles' text input
// does: the arrows still move through the suggestions, and the ghost
// follows them.
func TestSuggestGhostOnly(t *testing.T) {
	r, _ := suggesting(t, `"list":false`)
	drawn(t, r, "┃ Command  git commit\n"+
		"  Note")
	r.Key("ArrowDown")
	r.Key("ArrowDown")
	drawn(t, r, "┃ Command  git checkout\n"+
		"  Note")
}
