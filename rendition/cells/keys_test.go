package cells

import (
	"slices"
	"strings"
	"testing"
	"time"

	hottygo "github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// form is a surface with a Form, its fields, a select, a Button and two
// Shortcuts; actions are the actions it has sent.
func form(t *testing.T) (c *view.Controller, data func() map[string]any, actions *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions = new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"name":"","note":"","agree":false,"size":["m"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyForm","catalogId":"` + hotty.ID + `","child":"col","onSubmit":{"event":{"name":"send"}}},
	 {"id":"col","component":"Column","children":["name","note","agree","size","go","save","quit"]},
	 {"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"}},
	 {"id":"note","component":"TextField","label":"Note","variant":"longText","value":{"@path":"/note"}},
	 {"id":"agree","component":"CheckBox","label":"Agree","value":{"@path":"/agree"}},
	 {"id":"size","component":"ChoicePicker","label":"Size","value":{"@path":"/size"},"options":[
	  {"label":"Small","value":"s"},{"label":"Medium","value":"m"},{"label":"Large","value":"l"}]},
	 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go"}}},
	 {"id":"go_t","component":"Text","text":"Go"},
	 {"id":"save","component":"HottyShortcut","catalogId":"` + hotty.ID + `","key":"Control+s","press":"go"},
	 {"id":"quit","component":"HottyShortcut","catalogId":"` + hotty.ID + `","key":"q","action":{"event":{"name":"quit"}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	s := p.Surface("s")
	return view.NewController(s), func() map[string]any { return s.Data.Root().(map[string]any) }, actions
}

// nameAt is the column of the form's Name field's value, on its label's
// row, the first: "┃ Name  abc", past the gutter, the label and a space,
// and the inset.
const nameAt = gutter + len("Name") + 1 + inset

func keys(t *testing.T, r *Rendition, ks ...string) {
	t.Helper()
	for _, k := range ks {
		if ok, err := r.Key(k); !ok || err != nil {
			t.Fatalf("key %q: handled %v, %v", k, ok, err)
		}
		r.Draw(40)
	}
}

func TestTyping(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	keys(t, r, "A", "d", "a", "Space", "L")
	if got := data()["name"]; got != "Ada L" {
		t.Fatalf("name %q", got)
	}
	keys(t, r, "Backspace", "Backspace")
	if got := data()["name"]; got != "Ada" {
		t.Fatalf("after Backspace: %q", got)
	}
	keys(t, r, "ArrowLeft", "ArrowLeft", "Shift+L", "Home", "Delete", "End", "!")
	if got := data()["name"]; got != "Lda!" {
		t.Fatalf("after editing: %q", got)
	}
	f := r.Draw(40)
	// The gutter's bar, the label and the value's inset come first:
	// "┃ Name  Lda!".
	if col, row, ok := f.Cursor(); !ok || col != nameAt+4 || row != 0 {
		t.Errorf("cursor %d,%d %v", col, row, ok)
	}
	if !strings.HasPrefix(strings.Split(f.Plain(), "\n")[0], "┃ Name  Lda!") {
		t.Errorf("field shows\n%s", f.Plain())
	}
}

// TestFieldSelection: Shift with a move and Control+a select (SPEC §10.2),
// and the field shows what is selected on the selection colour, reversed
// where the theme can't tint. Under the default keymap the cursor is a
// line, the terminal's; under a terminal's it is a block, the cell
// reversed, but not beside a selection.
func TestFieldSelection(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	shows := func(when string, sel []bool, block int) {
		t.Helper()
		f := r.Draw(40)
		// "┃ Name  abc": the value from nameAt, on the label's row.
		row := f.Cells[0]
		for i, want := range sel {
			cell := row[nameAt+i]
			if got := cell.Back == Selection && cell.BackMix == 255 && cell.BackAttr == Reverse; got != want {
				t.Errorf("%s: %q selected %v", when, cell.Text, got)
			}
			if got := cell.Attr&Reverse != 0; got != (i == block) {
				t.Errorf("%s: %q a block cursor %v", when, cell.Text, got)
			}
		}
		if _, _, ok := f.Cursor(); !ok || f.BlockCursor() != (block >= 0) {
			t.Errorf("%s: a cursor %v, a block %v", when, ok, f.BlockCursor())
		}
	}
	keys(t, r, "a", "b", "c", "Shift+ArrowLeft", "Shift+ArrowLeft")
	shows("Shift+ArrowLeft twice", []bool{false, true, true, false}, -1)
	if col, _, ok := r.Draw(40).Cursor(); !ok || col != nameAt+1 {
		t.Errorf("the cursor at %d %v, want %d", col, ok, nameAt+1)
	}
	keys(t, r, "End")
	shows("End", []bool{false, false, false, false}, -1)
	keys(t, r, "Control+a")
	shows("Control+a", []bool{true, true, true, false}, -1)

	r.SetKeys(hottygo.TerminalKeys)
	keys(t, r, "End")
	shows("End, in the terminal keymap", []bool{false, false, false, false}, 3)
	keys(t, r, "Shift+Home")
	shows("Shift+Home, in the terminal keymap", []bool{true, true, true, false}, -1)
}

// TestFieldMouseSelection: a press in a field's value puts the caret
// there, and a drag selects from it to the pointer, past the value's ends
// to them; typing replaces the selection.
func TestFieldMouseSelection(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	keys(t, r, "a", "b", "c", "d")
	// "┃ Name  abcd" on row 0: b is at column nameAt+1.
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(r.Click(nameAt+1, 0))
	f := r.Draw(40)
	if col, _, _ := f.Cursor(); col != nameAt+1 {
		t.Errorf("a press at b: the cursor at %d", col)
	}
	must(r.Drag(nameAt+2, 0))
	must(r.Drag(30, 0))
	r.Release()
	f = r.Draw(40)
	for i, want := range []bool{false, true, true, true} {
		if got := f.Cells[0][nameAt+i].Back == Selection; got != want {
			t.Errorf("dragged past the end: %q selected %v", f.Cells[0][nameAt+i].Text, got)
		}
	}
	keys(t, r, "x")
	if got := data()["name"]; got != "ax" {
		t.Errorf("typing over the selection: %q", got)
	}
	must(r.Click(nameAt+2, 0))
	must(r.Drag(0, 0))
	r.Release()
	keys(t, r, "Delete")
	if got := data()["name"]; got != "" {
		t.Errorf("a drag from the end to left of the value, and Delete: %q", got)
	}
}

// TestCaretBlinks: the caret shows when the field takes the keyboard and
// after a key, then hides and shows by turns, caretBlink each, the frame
// saying when to draw it again.
func TestCaretBlinks(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	now := time.Unix(0, 0)
	r.Clock = func() time.Time { return now }
	c.Focus("name")
	shows := func(when string, want bool) {
		t.Helper()
		f := r.Draw(40)
		if _, _, ok := f.Cursor(); ok != want {
			t.Errorf("%s: the caret shows %v", when, ok)
		}
		if d := r.Animating(); d <= 0 || d > caretBlink {
			t.Errorf("%s: drawn again in %v", when, d)
		}
	}
	shows("focused", true)
	now = now.Add(caretBlink)
	shows("a blink on", false)
	now = now.Add(caretBlink)
	shows("two on", true)
	now = now.Add(caretBlink + 100*time.Millisecond)
	shows("three on and a bit", false)
	keys(t, r, "a")
	shows("after a key", true)
}

// TestFieldShiftClick: with Shift, a press in the field that has the
// keyboard extends the selection from its anchor, or from the caret; in a
// field without it, it is a click.
func TestFieldShiftClick(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	keys(t, r, "a", "b", "c", "d", "Home")
	// "┃ Name  abcd" on row 0: d is at column nameAt+3.
	if err := r.ShiftClick(nameAt+3, 0); err != nil {
		t.Fatal(err)
	}
	r.Release()
	keys(t, r, "x")
	if got := data()["name"]; got != "xd" {
		t.Fatalf("Shift with a press from the caret, then x: %q", got)
	}
	keys(t, r, "End")
	for _, col := range []int{nameAt, nameAt + 1} {
		if err := r.ShiftClick(col, 0); err != nil {
			t.Fatal(err)
		}
		r.Release()
	}
	keys(t, r, "!")
	if got := data()["name"]; got != "x!" {
		t.Errorf("a second Shift press moves the selection's end, not its anchor: %q", got)
	}
	c.Focus("note")
	r.Draw(40)
	if err := r.ShiftClick(nameAt+1, 0); err != nil {
		t.Fatal(err)
	}
	if c.St.Focus != "name" {
		t.Errorf("Shift with a press in another field: the keyboard on %q", c.St.Focus)
	}
}

// TestPointer: the pointer is an I-beam over a field's input, and the
// terminal's own elsewhere: on its label, its inset, and past its end.
func TestPointer(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	r.Draw(60)
	for _, at := range []struct {
		col, row int
		want     string
	}{{nameAt + 1, 0, "text"}, {nameAt + textInput - 1, 0, "text"}, {nameAt + textInput, 0, ""}, {5, 0, ""}, {nameAt - 1, 0, ""}} {
		if got := r.Pointer(at.col, at.row); got != at.want {
			t.Errorf("at %d,%d: %q, want %q", at.col, at.row, got, at.want)
		}
	}
}

// TestFieldInput: a one-line field's input is underlined, and nothing
// else on its row: not its label, the space after it or the inset, nor
// what is past the input's width, however wide the field's box; a select's
// input is its widest option, a space and "▾".
func TestFieldInput(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	f := r.Draw(60)
	under := func(row, from, to int) {
		t.Helper()
		for col := 0; col < f.Cols; col++ {
			if got, want := f.Cells[row][col].Attr&Underline != 0, col >= from && col < to; got != want {
				t.Errorf("row %d col %d %q: underlined %v", row, col, f.Cells[row][col].Text, got)
			}
		}
	}
	under(0, nameAt, nameAt+textInput)
	size := strings.Split(f.Plain(), "\n")
	for y, l := range size {
		if strings.HasPrefix(l, "  Size") {
			if l != "  Size  Medium ▾" {
				t.Errorf("the select's row %q", l)
			}
			under(y, nameAt, nameAt+len("Medium")+2)
		}
	}
}

// TestFieldUnderline: a one-line field's underline is border at half its
// contrast with the background, the accent at half while the field has
// the keyboard; where the terminal has not said its colours, each its own.
// Its label is the text's colour faded toward the background.
func TestFieldUnderline(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	for _, focus := range []bool{false, true} {
		want := Border
		if focus {
			c.Focus(c.V.Focusables()[0])
			want = Accent
		}
		f := r.Draw(60)
		cell := f.Cells[0][nameAt]
		if cell.Attr&Underline == 0 || !cell.LineSet || cell.Line != want || cell.LineMix != fieldLineMix {
			t.Fatalf("focused %v: %+v, want %v toned by %d", focus, cell, want, fieldLineMix)
		}
		if !focus {
			if l := f.Cells[0][gutter]; l.Role != Fg || l.To != Bg || l.Mix != labelFade {
				t.Errorf("the label: %+v, want the text faded by %d", l, labelFade)
			}
		}
	}
	term := theme.Default
	term.Term = theme.Terminal{Bg: "#000000"}
	term.Term.ANSI[8], term.Term.ANSI[12] = "#808080", "#ffffff"
	line := func(r Role) Cell {
		return Cell{Text: "a", Width: 1, Attr: Underline, Line: r, LineSet: true, LineMix: fieldLineMix}
	}
	for _, at := range []struct {
		role       Role
		said, none string
	}{
		{Border, "4;58;2;64;64;64", "4;58;5;8"},
		{Accent, "4;58;2;128;128;128", "4;58;5;12"},
	} {
		if got := line(at.role).style(&term); got != at.said {
			t.Errorf("%v, the terminal's colours said: %q, want %q", at.role, got, at.said)
		}
		if got := line(at.role).style(&theme.Default); got != at.none {
			t.Errorf("%v, the terminal's colours unknown: %q, want %q", at.role, got, at.none)
		}
	}
}

// TestFieldError: a one-line field's error starts in its input's column,
// under the value it is about; a narrow field's, at the field's edge.
func TestFieldError(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"email":"ada@exa"}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","align":"start","children":["a","b"]},
	 {"id":"a","component":"TextField","label":"Name","value":"Ada"},
	 {"id":"b","component":"TextField","label":"Email","value":{"@path":"/email"},
	  "checks":[{"condition":{"@call":"email","args":{"value":{"@path":"/email"}}},"message":"An email address, please"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.St.Touched["b"] = true // the user typed it
	c.Rebuild()
	r := New(c)
	want := "  Name   Ada\n" +
		"  Email  ada@exa\n" +
		"         ✗ An email address, please"
	if got := r.Draw(60).Plain(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	want = "  Name   Ada\n" +
		"  Email  ada@exa\n" +
		"  ✗ An email address,\n" +
		"    please"
	if got := r.Draw(22).Plain(); got != want {
		t.Errorf("narrow, got\n%s\nwant\n%s", got, want)
	}
}

// TestLabelRun: the labels of a run of one-line fields are padded to the
// widest, so that their inputs start in one column, in a Column that
// does not stretch them too, none of them cut for it.
func TestLabelRun(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","align":"start","children":["a","b"]},
	 {"id":"a","component":"TextField","label":"To","value":"ada@example.com"},
	 {"id":"b","component":"TextField","label":"Subject","value":"Hello"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	f := New(view.NewController(p.Surface("s"))).Draw(80)
	want := "  To       ada@example.com\n" +
		"  Subject  Hello"
	if got := f.Plain(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if end := gutter + len("Subject") + 1 + inset + textInput; f.Cells[0][end-1].Attr&Underline == 0 {
		t.Errorf("To's input is cut before column %d", end)
	}
}

func TestLongText(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("note")
	r.Draw(40)
	keys(t, r, "a", "b", "Enter", "c", "ArrowUp", "x", "PageDown", "y")
	if got := data()["note"]; got != "axb\ncy" {
		t.Fatalf("note %q", got)
	}
}

func TestTabOrder(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	want := c.V.Focusables()
	ids := []string{want[0]}
	c.Focus(want[0])
	for {
		r.Draw(40)
		if ok, err := r.Key("Tab"); !ok || err != nil {
			t.Fatalf("Tab: %v %v", ok, err)
		}
		if !c.St.Keyboard {
			break
		}
		ids = append(ids, c.St.Focus)
	}
	if !slices.Equal(ids, want) {
		t.Errorf("Tab order %v, focusables %v", ids, want)
	}
	c.Focus(want[len(want)-1])
	if _, err := r.Key("Shift+Tab"); err != nil || c.St.Focus != want[len(want)-2] {
		t.Errorf("Shift+Tab went to %s", c.St.Focus)
	}
}

func TestEnterSubmits(t *testing.T) {
	c, _, actions := form(t)
	r := New(c)
	c.Focus("name")
	keys(t, r, "x", "Enter")
	if !slices.Equal(*actions, []string{"send"}) {
		t.Errorf("actions %v", *actions)
	}
}

func TestSpaceToggles(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("agree")
	keys(t, r, " ")
	if data()["agree"] != true {
		t.Fatal("Space did not tick the CheckBox")
	}
	keys(t, r, "Space")
	if data()["agree"] != false {
		t.Fatal("Space did not untick it")
	}
	if f := r.Draw(40).Plain(); !strings.Contains(f, "[ ] Agree") {
		t.Errorf("box:\n%s", f)
	}
}

func TestShortcutPresses(t *testing.T) {
	c, _, actions := form(t)
	r := New(c)
	c.Focus("name")
	keys(t, r, "Control+s")
	c.Focus("")
	keys(t, r, "Control+s")
	if !slices.Equal(*actions, []string{"go", "go"}) {
		t.Errorf("actions %v", *actions)
	}
}

func TestPrintableNeverReachesShortcut(t *testing.T) {
	c, data, actions := form(t)
	r := New(c)
	c.Focus("name")
	keys(t, r, "q", "Shift+Q")
	if data()["name"] != "qQ" || len(*actions) != 0 {
		t.Fatalf("name %q, actions %v", data()["name"], *actions)
	}
	c.Focus("go")
	keys(t, r, "q")
	if !slices.Equal(*actions, []string{"quit"}) {
		t.Errorf("q on a Button: actions %v", *actions)
	}
	if ok, _ := r.Key("Escape"); ok {
		t.Error("Escape was taken with no Modal open")
	}
	if ok, _ := r.Key("F5"); ok {
		t.Error("F5 was taken")
	}
}

func TestSelect(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("size")
	r.Draw(40)
	keys(t, r, "ArrowDown")
	if got := data()["size"]; !slices.Equal(got.([]any), []any{"l"}) {
		t.Fatalf("ArrowDown: size %v", got)
	}
	keys(t, r, "s")
	if got := data()["size"]; !slices.Equal(got.([]any), []any{"s"}) {
		t.Fatalf("s: size %v", got)
	}
	keys(t, r, "Enter", "ArrowDown")
	f := r.Draw(40).Plain()
	// The input is a text field's, after the label and the inset, as wide
	// as the widest option and the chevron; the list's labels under the
	// value's.
	if !strings.Contains(f, "┃ Size  Small  ▾\n") || !strings.Contains(f, "┃     ● Small\n┃   > ○ Medium") {
		t.Fatalf("open list:\n%s", f)
	}
	keys(t, r, " ")
	if got := data()["size"]; !slices.Equal(got.([]any), []any{"m"}) {
		t.Fatalf("pick: size %v", got)
	}
	if strings.Contains(r.Draw(40).Plain(), "○") {
		t.Error("the list stayed open")
	}
}

func TestClick(t *testing.T) {
	c, data, actions := form(t)
	r := New(c)
	f := r.Draw(40)
	row := func(prefix string) int {
		for y, l := range strings.Split(f.Plain(), "\n") {
			if strings.HasPrefix(l, prefix) {
				return y
			}
		}
		t.Fatalf("no row %q in\n%s", prefix, f.Plain())
		return -1
	}
	if err := r.Click(3, row("  [ ] Agree")); err != nil || data()["agree"] != true || c.St.Focus != "agree" {
		t.Fatalf("click on the box: %v %v %s", err, data()["agree"], c.St.Focus)
	}
	f = r.Draw(40)
	if err := r.Click(2, row(" Go")); err != nil || !slices.Equal(*actions, []string{"go"}) {
		t.Fatalf("click on the Button: %v %v", err, *actions)
	}
	f = r.Draw(40)
	if err := r.Click(30, row(" Go")); err != nil || c.St.Keyboard {
		t.Fatal("a click on nothing kept the keyboard")
	}
	_ = c.SetValue("name", "Ada")
	f = r.Draw(40)
	if err := r.Click(nameAt+1, row("  Name")); err != nil || !c.St.Keyboard || c.St.Focus != "name" {
		t.Fatal("a click on the field did not focus it")
	}
	keys(t, r, "x")
	if data()["name"] != "Axda" {
		t.Errorf("the click put the cursor elsewhere: %q", data()["name"])
	}
}

func TestModalPanel(t *testing.T) {
	c := example(t, "36_modal")
	r := New(c)
	openModal(t, c)
	f := r.Draw(60)
	if !strings.Contains(f.Plain(), "╭") || f.Cells[0][0].Attr&Faint == 0 {
		t.Fatalf("no panel over a faint surface:\n%s", f.Plain())
	}
	if err := r.Click(0, f.Rows-1); err != nil || c.V.Overlay != nil {
		t.Error("a click outside the panel left the Modal open")
	}
	openModal(t, c)
	if ok, err := r.Key("Escape"); !ok || err != nil || c.V.Overlay != nil {
		t.Error("Escape left the Modal open")
	}
}

func TestBox(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	if _, _, _, _, ok := r.Box("go"); ok {
		t.Error("a box before any Draw")
	}
	f := r.Draw(40)
	col, row, w, h, ok := r.Box("go")
	if !ok || w != 4 || h != 1 || strings.Split(f.Plain(), "\n")[row][col:] != " Go" {
		t.Errorf("go: %d,%d %dx%d %v", col, row, w, h, ok)
	}
	if col, row, w, h, ok := r.Box("name"); !ok || col != 0 || row != 0 || w != 40 || h != 1 {
		t.Errorf("name: %d,%d %dx%d %v", col, row, w, h, ok)
	}
	if _, _, _, _, ok := r.Box("nothing"); ok {
		t.Error("a box for an id not drawn")
	}
}

// A Button is " label " on a fill a grey off the background, a tint of it
// toward the text; the terminal's bright black where nothing tints;
// underlined under NO_COLOR. With the keyboard it is reversed in the
// accent, unfilled; disabled, its label is the text's, faint, on half the
// fill; borderless, it is its label underlined, unfilled.
func TestButtonFill(t *testing.T) {
	c, _, _ := form(t)
	r := New(c)
	f := r.Draw(40)
	col, row, _, _, _ := r.Box("go")
	for _, x := range []int{col, col + 1, col + 3} {
		if g := f.Cells[row][x]; g.Back != Fg || g.BackMix != buttonTint || !g.BackShade || g.MonoAttr != Underline || g.Attr != 0 {
			t.Fatalf("unfocused, column %d: %+v", x-col, g)
		}
	}
	g := f.Cells[row][col+1]
	th := theme.Theme{Name: "t", Bg: "#000000", Fg: "#ffffff"}
	if st := g.style(&th); !strings.Contains(st, "48;2;48;48;48") {
		t.Errorf("themed: %q", st)
	}
	term := theme.Default
	term.Term.Bg, term.Term.Fg = "#000000", "#ffffff"
	if st := g.style(&term); st != "48;2;48;48;48" {
		t.Errorf("the terminal's colours: %q", st)
	}
	if st := g.style(&theme.Default); st != "100" {
		t.Errorf("the floor: %q, want the bright black background", st)
	}
	if st := g.style(nil); st != "4" {
		t.Errorf("NO_COLOR: %q, want underlined", st)
	}

	c.Focus("go")
	if g := r.Draw(40).Cells[row][col+1]; g.Role != Accent || g.Attr != Reverse || g.BackMix != 0 {
		t.Errorf("focused: %+v", g)
	}

	r = coding(t, `{"id":"b","component":"Button","child":"bt","action":{"event":{"name":"b"}},"checks":[{"condition":{"@path":"/off"},"message":"No"}]}`,
		`{"id":"bt","component":"Text","text":"Off"}`,
		`{"id":"l","component":"Button","variant":"borderless","child":"lt","action":{"event":{"name":"l"}}}`,
		`{"id":"lt","component":"Text","text":"Link"}`)
	f = r.Draw(20)
	col, row, _, _, _ = r.Box("b")
	if g := f.Cells[row][col+1]; g.Role != Fg || g.Attr != Faint || g.BackMix != buttonTint/2 || !g.BackShade {
		t.Errorf("disabled: %+v", g)
	}
	if st := f.Cells[row][col+1].style(&theme.Default); st != "2;100" {
		t.Errorf("disabled at the floor: %q, want faint on the bright black", st)
	}
	col, row, w, _, _ := r.Box("l")
	if g := f.Cells[row][col]; w != 4 || g.Attr != Underline || g.BackMix != 0 {
		t.Errorf("borderless: %d wide, %+v", w, g)
	}
}

// TestSliderDrag: a click on a Slider's track sets it, and the pointer
// moved with the button down drags it, clamped to the track's ends; after
// the release, moves do nothing. Arrows step it.
func TestSliderDrag(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"v":0}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[{"id":"root","component":"Slider","min":0,"max":10,"value":{"@path":"/v"}}]}}]`)); err != nil {
		t.Fatal(err)
	}
	s := p.Surface("s")
	c := view.NewController(s)
	r := New(c)
	r.Draw(20)
	b, ok := r.boxes["root"]
	if !ok {
		t.Fatal("no box for the slider")
	}
	var tr *trackArea
	for _, h := range r.hits {
		if h.id == "root" && h.track != nil {
			tr = h.track
		}
	}
	if tr == nil {
		t.Fatal("no track drawn")
	}
	v := func() any { return s.Data.Value("/v") }
	if err := r.Click(tr.x, b.y); err != nil || v() != 0.0 {
		t.Fatalf("a click on the track's start: %v %v", err, v())
	}
	r.Draw(20)
	if err := r.Drag(tr.x+tr.n-1, b.y); err != nil || v() != 10.0 {
		t.Fatalf("dragged to the track's end: %v %v", err, v())
	}
	r.Draw(20)
	if err := r.Drag(-5, b.y+3); err != nil || v() != 0.0 {
		t.Fatalf("dragged past the start: %v %v", err, v())
	}
	must(t, r.Release())
	if err := r.Drag(tr.x+tr.n-1, b.y); err != nil || v() != 0.0 {
		t.Fatalf("a move after the release: %v %v", err, v())
	}
	keys(t, r, "ArrowRight", "ArrowRight")
	if v() != 1.0 {
		t.Errorf("two steps right of 0: %v", v())
	}
	// The knob is a square, as a HottySwitch's, in the accent with the
	// track up to it while the slider has the keyboard.
	f := r.Draw(20)
	knob, done := f.Cells[b.y][tr.x+1], f.Cells[b.y][tr.x]
	if knob.Text != "■" || knob.Role != Accent || knob.Attr != 0 || done.Text != "━" || done.Role != Accent {
		t.Errorf("focused at 1: track %q %v, knob %q %v %v", done.Text, done.Role, knob.Text, knob.Role, knob.Attr)
	}
}

// TestSliderSteady: the track keeps its length and place whatever the
// value, 0.45 or a whole 1.
func TestSliderSteady(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"v":0.45}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[{"id":"root","component":"Slider","min":0,"max":1,"value":{"@path":"/v"}}]}}]`)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	r := New(c)
	var first *trackArea
	for _, v := range []float64{0.45, 1, 0.5, 0} {
		if err := c.SetValue("root", v); err != nil {
			t.Fatal(err)
		}
		r.Draw(20)
		for _, h := range r.hits {
			if h.id != "root" || h.track == nil {
				continue
			}
			if first == nil {
				first = h.track
			} else if *h.track != *first {
				t.Errorf("at %v the track is %+v, at 0.45 %+v", v, *h.track, *first)
			}
		}
	}
}

// TestListRows: a List's Buttons are its rows, as wide as it, whatever
// their variant; a click anywhere on a row presses its Button.
func TestListRows(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	var acts []string
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			acts = append(acts, o.Action.Name)
		}
	}
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"List","children":["a","b"]},
	 {"id":"a","component":"Button","variant":"borderless","child":"at","action":{"event":{"name":"a"}}},
	 {"id":"at","component":"Text","text":"One"},
	 {"id":"b","component":"Button","child":"bt","action":{"event":{"name":"b"}}},
	 {"id":"bt","component":"Text","text":"Two"}]}}]`)); err != nil {
		t.Fatal(err)
	}
	r := New(view.NewController(p.Surface("s")))
	f := r.Draw(12)
	var rows []string
	for l := range strings.Lines(f.Plain()) {
		rows = append(rows, strings.TrimRight(l, " \n"))
	}
	if want := []string{" One", " Two"}; !slices.Equal(rows, want) {
		t.Errorf("rows %q, want %q", rows, want)
	}
	for _, id := range []string{"a", "b"} {
		if b := r.boxes[id]; b.w != 12 {
			t.Errorf("%s's row is %d wide, want 12", id, b.w)
		}
	}
	if err := r.Click(10, r.boxes["b"].y); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(acts, []string{"b"}) {
		t.Errorf("a click at the row's end sent %v, want [b]", acts)
	}
}
