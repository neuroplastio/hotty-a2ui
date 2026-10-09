package cells

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// diffing is a surface with a HottyDiff of the props given, its selection
// bound to /h, and the actions it sent, each its name and the hunk.
func diffing(t *testing.T, props string) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name+":"+a2ui.ToString(o.Action.Context["hunk"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyDiff","catalogId":"` + hotty.ID + `","selected":{"@path":"/h"},
	  "onActivate":{"event":{"name":"stage","context":{"hunk":{"@path":"/h"}}}},` + props + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// texts is the props of a HottyDiff of two texts.
func texts(file, old, new string, more string) string {
	b, _ := json.Marshal(map[string]string{"file": file, "old": old, "new": new})
	s := strings.TrimSuffix(strings.TrimPrefix(string(b), "{"), "}")
	if more != "" {
		s += "," + more
	}
	return s
}

// A unified diff is the file's name and what it adds and removes, then
// each hunk's header and its lines: the old and the new number, the sign
// and the code, a column left for the rail.
func TestDiffDraws(t *testing.T) {
	r, _, _ := diffing(t, texts("a.txt", "one\ntwo\nthree\n", "one\n2\nthree\n", ""))
	want := " a.txt  +1 -1\n" +
		" @@ -1,3 +1,3 @@\n" +
		" 1 1   one\n" +
		" 2   - two\n" +
		"   2 + 2\n" +
		" 3 3   three"
	if got := r.Draw(30).Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	f := r.Draw(30)
	for _, c := range []struct {
		x, y int
		role Role
		attr Attr
	}{{1, 0, Fg, Bold}, {8, 0, Success, 0}, {11, 0, Error, 0}, {1, 1, Info, 0}, {1, 2, Muted, 0}, {5, 3, Error, 0}, {5, 4, Success, 0}} {
		if got := f.Cells[c.y][c.x]; got.Role != c.role || got.Attr != c.attr {
			t.Errorf("(%d,%d) %q: %v %v, want %v %v", c.x, c.y, got.Text, got.Role, got.Attr, c.role, c.attr)
		}
	}

	// No numbers, no wrap: a long line is cut.
	r, _, _ = diffing(t, texts("", "a\n", "abcdefghijklmnopqrstuvwxyz\n", `"lineNumbers":false,"wrap":false`))
	if got := r.Draw(16).Plain(); got != " @@ -1 +1 @@\n - a\n + abcdefghijkl…" {
		t.Errorf("cut:\n%s", got)
	}
	// Wrapped, the further rows leave the gutter blank.
	r, _, _ = diffing(t, texts("", "a\n", "abcdefghijklmnopqrstuvwxyz\n", ""))
	if got := r.Draw(16).Plain(); got != " @@ -1 +1 @@\n 1   - a\n   1 + abcdefghi\n       jklmnopqr\n       stuvwxyz" {
		t.Errorf("wrapped:\n%s", got)
	}
	r, _, _ = diffing(t, texts("same.txt", "a\n", "a\n", ""))
	if got := r.Draw(20).Plain(); got != " same.txt" {
		t.Errorf("no change: %q", got)
	}
	r, _, _ = diffing(t, texts("", "a\n", "a\n", ""))
	if got := r.Draw(20).Plain(); got != " No changes" {
		t.Errorf("no change, no name: %q", got)
	}
}

// A changed line's row is tinted toward its role to its end, where the
// theme can tint; its changed words more so. With the terminal's own
// colours, the line's text takes its role's colour instead, and its words
// are reversed, as git's diff-highlight does.
func TestDiffTint(t *testing.T) {
	r, _, _ := diffing(t, texts("", "x := errors.New(e)\n", "x := fmt.Errorf(e)\n", `"lineNumbers":false`))
	f := r.Draw(30)
	if c := f.Cells[1][29]; c.Back != Error || c.BackMix != lineTint {
		t.Errorf("the removed row's last cell: %v %d", c.Back, c.BackMix)
	}
	if c := f.Cells[1][3]; c.Back != Error || c.BackMix != lineTint || !c.BackFg || c.BackAttr != 0 {
		t.Errorf("an unchanged word of the removed line (%q): %+v", c.Text, c)
	}
	if c := f.Cells[1][8]; c.Text != "e" || c.BackMix != wordTint || c.BackAttr != Reverse {
		t.Errorf("errors, changed: %+v", c)
	}
	if c := f.Cells[2][8]; c.Text != "f" || c.Back != Success || c.BackMix != wordTint {
		t.Errorf("fmt, changed: %+v", c)
	}
	if c := f.Cells[0][1]; c.BackMix != 0 {
		t.Errorf("the header is tinted: %+v", c)
	}
	th := theme.Theme{Name: "t", Bg: "#000000", Fg: "#ffffff", Error: "#ff0000", Success: "#00ff00"}
	if out := f.Themed(th); !strings.Contains(out, "48;2;42;0;0") || strings.Contains(out, "[7m") {
		t.Errorf("tints, themed: %q", out)
	}
	if out := f.Themed(theme.Default); strings.Contains(out, "48;") {
		t.Errorf("the terminal's background is tinted: %q", out)
	}
	if st := f.Cells[1][3].style(&theme.Default); st != ansi16[Error] {
		t.Errorf("an unchanged word, the terminal's: %q, want %q", st, ansi16[Error])
	}
	if st := f.Cells[1][8].style(&theme.Default); st != "7;"+ansi16[Error] {
		t.Errorf("a changed word, the terminal's: %q", st)
	}
}

// Split, the old side is left and the new right, a changed line beside the
// line that replaces it; where a side has no room for 16 columns of code,
// the diff is unified.
func TestDiffSplit(t *testing.T) {
	props := texts("", "a\nb\nc\n", "a\nB\nnew\nc\n", `"view":"split"`)
	r, _, _ := diffing(t, props)
	want := " @@ -1,3 +1,4 @@\n" +
		" 1   a                 │ 1   a\n" +
		" 2 - b                 │ 2 + B\n" +
		"                       │ 3 + new\n" +
		" 3   c                 │ 4   c"
	if got := r.Draw(46).Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if got := r.Draw(30).Plain(); !strings.Contains(got, "2   - b") {
		t.Errorf("narrow, not unified:\n%s", got)
	}
}

// Hunks are selected with the arrows (or k and j), Home and End (g and
// G), or a click; Enter, or a click on the selected hunk, acts on it. The
// selected one's rows have a rail, and its header is reversed.
func TestDiffKeysAndClicks(t *testing.T) {
	var old, new []string
	for i := range 20 {
		old = append(old, string(rune('a'+i)))
	}
	new = append(new, old...)
	new[1], new[17] = "B", "R"
	r, c, actions := diffing(t, texts("x", strings.Join(old, "\n")+"\n", strings.Join(new, "\n")+"\n", `"context":1`))
	f := r.Draw(30)
	if !strings.Contains(f.Plain(), "@@ -1,3 +1,3 @@") || !strings.Contains(f.Plain(), "@@ -17,3 +17,3 @@") {
		t.Fatalf("two hunks:\n%s", f.Plain())
	}
	// The unchanged runs between and after the hunks are folded.
	rows := strings.Split(f.Plain(), "\n")
	if rows[6] != " ⋯ 13 unchanged lines" || rows[7] != " @@ -17,3 +17,3 @@" || rows[len(rows)-1] != " ⋯ 1 unchanged line" {
		t.Fatalf("rows 6 and 7, and the last:\n%s", f.Plain())
	}
	if err := r.Click(3, 6); err != nil || c.S.Data.Value("/h") != nil {
		t.Fatalf("a click on a fold: %v %v", err, c.S.Data.Value("/h"))
	}
	if err := r.Click(3, 8); err != nil {
		t.Fatal(err)
	}
	if !r.focused("root") || c.S.Data.Value("/h") != "x:17" || len(*actions) != 0 {
		t.Fatalf("a click on the second hunk: focused %v, /h %v, actions %v", r.focused("root"), c.S.Data.Value("/h"), *actions)
	}
	f = r.Draw(30)
	if c := f.Cells[7][0]; c.Text != "▎" || c.Role != Accent {
		t.Errorf("the rail: %+v", c)
	}
	if c := f.Cells[6][0]; c.Text == "▎" {
		t.Error("the fold has the rail")
	}
	if c := f.Cells[7][1]; c.Attr != Reverse || c.Role != Accent {
		t.Errorf("the selected header: %+v", c)
	}
	if c := f.Cells[2][0]; c.Text == "▎" {
		t.Error("the first hunk has the rail")
	}
	if err := r.Click(3, 8); err != nil || len(*actions) != 1 || (*actions)[0] != "stage:x:17" {
		t.Fatalf("a second click: %v %v", err, *actions)
	}
	keys(t, r, "k")
	if got := c.S.Data.Value("/h"); got != "x:1" {
		t.Errorf("k: %v", got)
	}
	keys(t, r, "G", "Enter")
	if got := c.S.Data.Value("/h"); got != "x:17" || len(*actions) != 2 || (*actions)[1] != "stage:x:17" {
		t.Errorf("G, Enter: %v %v", got, *actions)
	}
	if err := r.Click(3, 0); err != nil || c.S.Data.Value("/h") != "x:17" {
		t.Errorf("a click on the file's name changed the selection: %v %v", err, c.S.Data.Value("/h"))
	}
}

// A HottyDiff takes its gutter and its widest line, and a blank row
// follows it in a Column, as one follows a HottyCode.
func TestDiffLayout(t *testing.T) {
	r := coding(t, `{"id":"d","component":"HottyDiff","catalogId":"`+hotty.ID+`","old":"a\n","new":"abc\n"}`,
		`{"id":"after","component":"Text","text":"|"}`)
	if got := r.Draw(40).Plain(); got != " @@ -1 +1 @@\n 1   - a\n   1 + abc\n\n|" {
		t.Errorf("got %q", got)
	}
	if w := diffWidth(r.c.V.Find("d")); w != 1+len("@@ -1 +1 @@") {
		t.Errorf("natural width %d", w)
	}
}

// In a HottyScrollView, the selected hunk is kept in sight with the rows
// that lead to it, its start where it is taller than the box.
func TestDiffInScrollView(t *testing.T) {
	var old, new []string
	for i := range 30 {
		old = append(old, string(rune('a'+i%26)))
	}
	new = slices.Clone(old)
	for i := 2; i < 10; i++ {
		new[i] = "X"
	}
	new[25] = "Y"
	b, _ := json.Marshal(map[string]string{"old": strings.Join(old, "\n") + "\n", "new": strings.Join(new, "\n") + "\n"})
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyScrollView","catalogId":"` + hotty.ID + `","height":6,"child":"d"},
	 {"id":"d","component":"HottyDiff","catalogId":"` + hotty.ID + `","file":"x",` + string(b)[1:] + `]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	r := New(c)
	r.Draw(40)
	c.Focus("d")
	keys(t, r, "G")
	keys(t, r, "G")
	// Taller than the box: its start, the fold before it first.
	if got := r.Draw(40).Plain(); !strings.HasPrefix(got, " ⋯ 9 unchanged lines") || !strings.Contains(got, "@@ -23,7 +23,7 @@") {
		t.Fatalf("G: the last hunk's start not at the top:\n%s", got)
	}
	keys(t, r, "g")
	got := r.Draw(40).Plain()
	if lines := strings.Split(got, "\n"); !strings.HasPrefix(strings.TrimSpace(lines[0]), "x") || !strings.Contains(lines[1], "@@ -1,") {
		t.Errorf("g: want the file's name, then the first hunk's header:\n%s", got)
	}
}
