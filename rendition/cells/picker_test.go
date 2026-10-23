package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// palette is a surface with a HottyPicker of four rows under two headings,
// three lines of rows at once, and the actions it sent, "name:row:action".
func palette(t *testing.T) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	return paletteWith(t, "")
}

// paletteWith is palette with extra, more of the picker's properties.
func paletteWith(t *testing.T, extra string) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name+":"+a2ui.ToString(o.Action.Context["row"])+":"+a2ui.ToString(o.Action.Context["action"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"q":"","sel":"","act":"","rows":[
	 {"value":"pane.split","group":"Pane","title":"Split right","keys":"ctrl+b %","section":"Used","match":[[0,2]],
	  "actions":[{"id":"run","title":"Run"},{"id":"again","title":"Run again","key":"mod+enter"}]},
	 {"value":"p-1","group":"Pane","title":"fish","detail":"~/code","section":"Panes",
	  "actions":[{"id":"focus","title":"Focus"},{"id":"close","title":"Close pane","key":"shift+delete"}]},
	 {"value":"p-2","group":"Pane","title":"htop","section":"Panes"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyPicker","catalogId":"` + hotty.ID + `","query":{"@path":"/q"},"rows":{"@path":"/rows"},
	  "selected":{"@path":"/sel"},"action":{"@path":"/act"},"panelKey":"mod+k","placeholder":"Type a command","height":3,` + extra + `
	  "onAction":{"event":{"name":"act","context":{"row":{"@path":"/sel"},"action":{"@path":"/act"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// trimmed is a frame's plain text, each row's trailing spaces gone.
func trimmed(f *Frame) string {
	rows := strings.Split(f.Plain(), "\n")
	for i, r := range rows {
		rows[i] = strings.TrimRight(r, " ")
	}
	return strings.Join(rows, "\n")
}

// A HottyPicker is laid out as a palette: the query (its placeholder while
// it is empty), a rule, the lines that show (a heading where the section
// changes, a row's group before its title, its keys at its right), then
// the footer, the selected row's actions flush right, Enter's at the right
// and those that don't fit at the left left out; a heading on the last
// line, with none of its rows under it, is left out. The selected row is
// reversed across the picker; the title's matched characters are drawn
// out.
func TestPickerDraws(t *testing.T) {
	r, c, _ := palette(t)
	want := " Type a command\n" +
		"──────────────────────────────────────────\n" +
		" Used\n" +
		" Pane: Split right            Ctrl+B > %\n" +
		"\n" +
		"    Run again  Ctrl+Enter     Run  Enter"
	f := r.Draw(42)
	if got := trimmed(f); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if cell := f.Cells[0][1]; cell.Role != Muted || cell.Attr&Faint == 0 {
		t.Errorf("the placeholder: %v %v, want muted and faint", cell.Role, cell.Attr)
	}
	if cell := f.Cells[3][0]; cell.Role != Muted || cell.Attr&Reverse == 0 {
		t.Errorf("the selected row, unfocused: %v %v, want muted reversed", cell.Role, cell.Attr)
	}
	if cell, key := f.Cells[5][31], f.Cells[5][35]; cell.Role != 0 || key.BackMix != capTint {
		t.Errorf("the footer: a title %v, a key tinted %d; want the text's and a cap", cell.Role, key.BackMix)
	}
	c.Focus("root")
	f = r.Draw(42)
	if cell := f.Cells[3][41]; cell.Role != Accent || cell.Attr&Reverse == 0 {
		t.Errorf("the selected row, focused: %v %v, want the accent reversed across it", cell.Role, cell.Attr)
	}
	if cell := f.Cells[3][7]; cell.Attr&Underline == 0 || cell.Attr&Bold == 0 {
		t.Errorf("a matched character in the selected row: %v, want bold and underlined", cell.Attr)
	}
	// Narrow, the detail wins: the selected row keeps its keys, and is
	// cut; the footer keeps Enter's.
	want = " Type a command\n" +
		"────────────────────────────\n" +
		" Used\n" +
		" Pane: Split…   Ctrl+B > %\n" +
		"\n" +
		"                Run  Enter"
	if got := trimmed(r.Draw(28)); got != want {
		t.Errorf("at 28 columns:\n%s\nwant\n%s", got, want)
	}
	keys(t, r, "arrow_down")
	want = " Type a command\n" +
		"──────────────────────────────────────────\n" +
		" Pane: Split right            Ctrl+B > %\n" +
		" Panes\n" +
		" Pane: fish  ~/code\n" +
		"  Close pane  Shift+Del     Focus  Enter"
	if got := trimmed(r.Draw(42)); got != want {
		t.Errorf("after arrow_down, the lines follow the selection, its heading with it:\n%s\nwant\n%s", got, want)
	}
	// A row whose keys don't fit beside its whole text shows none.
	want = " Type a command\n" +
		"────────────────────────────\n" +
		" Pane: Split right\n" +
		" Panes\n" +
		" Pane: fish  ~/code\n" +
		"              Focus  Enter"
	if got := trimmed(r.Draw(28)); got != want {
		t.Errorf("at 28 columns:\n%s\nwant\n%s", got, want)
	}
	keys(t, r, "ctrl+k")
	if got := strings.Split(trimmed(r.Draw(42)), "\n")[5]; got != "                Back  Esc     Run  Enter" {
		t.Errorf("the panel's footer %q", got)
	}
}

// A picker that keeps keys (keep: "keys", a picker of bindings) shows every
// row's keys, cutting its text for them, where a narrow one shows the
// selected row's alone.
func TestPickerKeepsKeys(t *testing.T) {
	r, c, _ := paletteWith(t, `"keep":"keys",`)
	c.Focus("root")
	keys(t, r, "arrow_down")
	want := " Type a command\n" +
		"────────────────────────────\n" +
		" Pane: Split…   Ctrl+B > %\n" +
		" Panes\n" +
		" Pane: fish  ~/code\n" +
		"              Focus  Enter"
	if got := trimmed(r.Draw(28)); got != want {
		t.Errorf("at 28 columns:\n%s\nwant\n%s", got, want)
	}
}

// A row's text is cut in its name first, so that its detail shows whole,
// while the name keeps pickName columns; then at its end.
func TestPickFit(t *testing.T) {
	for _, c := range []struct {
		name, detail string
		w            int
		want         string
	}{
		{"Switch plx's theme, seeing each", "  theme.pick", 24, "Switch plx'…  theme.pick"},
		{"Switch plx's theme, seeing each", "  theme.pick", 18, "Switch plx's them…"},
		{"fish", "  ~/code/plexos", 19, "fish  ~/code/plexos"},
		{"fish", "  ~/code/plexos", 17, "fish  ~/code/ple…"},
		{"Split right", "", 8, "Split r…"},
	} {
		gs := pickFit(line(c.name, style{}), line(c.detail, style{}), c.w)
		var b strings.Builder
		for _, g := range gs {
			b.WriteString(g.text)
		}
		if b.String() != c.want {
			t.Errorf("%q %q in %d: %q, want %q", c.name, c.detail, c.w, b.String(), c.want)
		}
	}
}

// A row's description is on the line under it, in muted, the selected
// row's reversed with it across the picker; a click on it is a click on
// its row. A row without one takes a line.
func TestPickerDescription(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"q":"","sel":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyPicker","catalogId":"` + hotty.ID + `","query":{"@path":"/q"},"selected":{"@path":"/sel"},"placeholder":"Type a command",
	  "rows":[{"value":"pane.new","title":"pane.new","description":"New pane (your shell)","detail":"ctrl+b n"},
	   {"value":"pane.split","title":"pane.split"},{"value":"tab.new","title":"tab.new","description":"New tab"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	r := New(c)
	c.Focus("root")
	want := " Type a command\n" +
		"──────────────────────────────\n" +
		" pane.new  ctrl+b n\n" +
		" New pane (your shell)\n" +
		" pane.split\n" +
		" tab.new\n" +
		" New tab\n"
	f := r.Draw(30)
	if got := trimmed(f); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, y := range []int{2, 3} {
		if cell := f.Cells[y][29]; cell.Role != Accent || cell.Attr&Reverse == 0 {
			t.Errorf("the selected row's line %d: %v %v, want the accent reversed across it", y, cell.Role, cell.Attr)
		}
	}
	if cell := f.Cells[6][1]; cell.Role != Muted || cell.Attr&Reverse != 0 {
		t.Errorf("a description: %v %v, want muted", cell.Role, cell.Attr)
	}
	if cell := f.Cells[5][1]; cell.Role == Muted {
		t.Errorf("a title: %v, want the text's", cell.Role)
	}
	if err := r.Click(3, 6); err != nil || c.S.Data.Value("/sel") != "tab.new" {
		t.Errorf("a click on a description: %v, /sel %v", err, c.S.Data.Value("/sel"))
	}
}

// Typed keys edit the query, as a field's, with the caret after them; the
// picker's keys (the arrows, Enter, an action's key) run on the selected
// row; the panel key lists its actions in the rows' place.
func TestPickerKeys(t *testing.T) {
	r, c, actions := palette(t)
	c.Focus("root")
	keys(t, r, "s", "p", "backspace", "f")
	if q := c.V.Find("root").Value; q != "sf" {
		t.Errorf("query %q, want sf", q)
	}
	keys(t, r, "enter", "mod+enter", "arrow_down", "shift+delete")
	if got := strings.Join(*actions, " "); got != "act:pane.split:run act:pane.split:again act:p-1:close" {
		t.Errorf("actions %s", got)
	}
	keys(t, r, "ctrl+k")
	f := r.Draw(42)
	if got := strings.Split(trimmed(f), "\n")[2:5]; strings.Join(got, "|") != " Actions: fish| Focus                             Enter| Close pane                    Shift+Del" {
		t.Errorf("the panel: %q", got)
	}
	keys(t, r, "arrow_down", "enter")
	if n := len(*actions); n != 4 || (*actions)[3] != "act:p-1:close" {
		t.Errorf("enter in the panel: %q", *actions)
	}
}

// A HottyPicker's status reads as its head's, not as a row: muted, with a
// rule under it, and the picker two rows taller for it. An error status
// keeps its colour and its ✗.
func TestPickerStatus(t *testing.T) {
	plain, _, _ := palette(t)
	r, _, _ := paletteWith(t, `"status":"theme.use · theme","statusKind":"info",`)
	f := r.Draw(42)
	want := " Type a command\n" +
		"──────────────────────────────────────────\n" +
		" theme.use · theme\n" +
		"──────────────────────────────────────────\n" +
		" Used\n" +
		" Pane: Split right            Ctrl+B > %\n" +
		"\n" +
		"    Run again  Ctrl+Enter     Run  Enter"
	if got := trimmed(f); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if got, want := len(f.Cells), len(plain.Draw(42).Cells)+2; got != want {
		t.Errorf("%d rows with a status, want %d", got, want)
	}
	if cell := f.Cells[2][1]; cell.Role != Muted {
		t.Errorf("the status: %v, want muted", cell.Role)
	}
	if cell := f.Cells[3][0]; cell.Role != Border {
		t.Errorf("the rule under the status: %v, want the border's", cell.Role)
	}
	r, _, _ = paletteWith(t, `"status":"theme.use · theme","statusKind":"error",`)
	f = r.Draw(42)
	if got := strings.Split(trimmed(f), "\n")[2]; got != " ✗ theme.use · theme" {
		t.Errorf("an error status: %q", got)
	}
	if cell := f.Cells[2][1]; cell.Role != Error {
		t.Errorf("an error status: %v, want the error's", cell.Role)
	}
}
