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
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name+":"+a2ui.ToString(o.Action.Context["row"])+":"+a2ui.ToString(o.Action.Context["action"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"q":"","sel":"","act":"","rows":[
	 {"value":"pane.split","group":"Pane","title":"Split right","section":"Used","match":[[0,2]],
	  "actions":[{"id":"run","title":"Run"},{"id":"again","title":"Run again","key":"mod+enter"}]},
	 {"value":"p-1","group":"Pane","title":"fish","detail":"~/code","section":"Panes",
	  "actions":[{"id":"focus","title":"Focus"},{"id":"close","title":"Close pane","key":"shift+delete"}]},
	 {"value":"p-2","group":"Pane","title":"htop","section":"Panes"}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyPicker","catalogId":"` + hotty.ID + `","query":{"@path":"/q"},"rows":{"@path":"/rows"},
	  "selected":{"@path":"/sel"},"action":{"@path":"/act"},"panelKey":"mod+k","placeholder":"Type a command","height":3,
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
// changes, a row's group before its title, its other actions' keys at its
// right), then its line of keys; a heading on the last line, with none of
// its rows under it, is left out. The selected row is reversed across the
// picker; the title's matched characters are drawn out.
func TestPickerDraws(t *testing.T) {
	r, c, _ := palette(t)
	want := " Type a command\n" +
		"──────────────────────────────────────────\n" +
		" Used\n" +
		" Pane: Split right             ctrl+enter\n" +
		"\n" +
		" ↑/↓ move • enter run • ctrl+k actions"
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
	c.Focus("root")
	f = r.Draw(42)
	if cell := f.Cells[3][41]; cell.Role != Accent || cell.Attr&Reverse == 0 {
		t.Errorf("the selected row, focused: %v %v, want the accent reversed across it", cell.Role, cell.Attr)
	}
	if cell := f.Cells[3][7]; cell.Attr&Underline == 0 || cell.Attr&Bold == 0 {
		t.Errorf("a matched character in the selected row: %v, want bold and underlined", cell.Attr)
	}
	keys(t, r, "arrow_down")
	want = " Type a command\n" +
		"──────────────────────────────────────────\n" +
		" Pane: Split right             ctrl+enter\n" +
		" Panes\n" +
		" Pane: fish  ~/code          shift+delete\n" +
		" ↑/↓ move • enter focus • ctrl+k actions"
	if got := trimmed(r.Draw(42)); got != want {
		t.Errorf("after arrow_down, the lines follow the selection, its heading with it:\n%s\nwant\n%s", got, want)
	}
	// Narrow, the detail wins: a row whose keys don't fit beside its whole
	// text shows none; the selected row keeps its keys, and is cut.
	want = " Type a command\n" +
		"────────────────────────────\n" +
		" Pane: Split right\n" +
		" Panes\n" +
		" Pane: fish …  shift+delete\n" +
		" ↑/↓ move • enter focus …"
	if got := trimmed(r.Draw(28)); got != want {
		t.Errorf("at 28 columns:\n%s\nwant\n%s", got, want)
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
	if got := strings.Split(trimmed(f), "\n")[2:5]; strings.Join(got, "|") != " Actions: fish| Focus                              enter| Close pane                  shift+delete" {
		t.Errorf("the panel: %q", got)
	}
	keys(t, r, "arrow_down", "enter")
	if n := len(*actions); n != 4 || (*actions)[3] != "act:p-1:close" {
		t.Errorf("enter in the panel: %q", *actions)
	}
}
