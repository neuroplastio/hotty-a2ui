package cells

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// toasting is the toast story's surface in cells, its clock stopped at
// the time it returns: the agent's four toasts, one of each kind, the
// newest (info) at the top.
func toasting(t *testing.T) (*Rendition, *view.Controller, *story.Run, *time.Time) {
	t.Helper()
	st := story.Find("hotty/toast")
	if st == nil {
		t.Fatal("no story hotty/toast")
	}
	run, err := story.Start(st)
	if err != nil {
		t.Fatal(err)
	}
	c := run.Surfaces()[0].C
	r := New(c)
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	r.Clock = func() time.Time { return now }
	return r, c, run, &now
}

// span is the columns of a frame's row from col to end.
func span(f *Frame, row, col, end int) string {
	var b strings.Builder
	for _, c := range f.Cells[row][col:end] {
		b.WriteString(c.Text)
	}
	return b.String()
}

// A toast is a rounded box in its kind's colour, its kind's mark (the
// glyph of its Icon), its message, and its action at the end of the row;
// the toasts stack from the top right corner down, a column from the
// edge, the newest first, all as wide as the widest needs, over what is
// under them.
func TestToastRows(t *testing.T) {
	r, _, _, _ := toasting(t)
	f := r.Draw(80)
	rule := strings.Repeat("─", 37)
	want := []string{
		"╭" + rule + "╮",
		"│ ⓘ Indexing 1,204 files              │",
		"╰" + rule + "╯",
		"╭" + rule + "╮",
		"│ ✓ Draft saved                       │",
		"╰" + rule + "╯",
		"╭" + rule + "╮",
		"│ ! Disk almost full: 2 GB left       │",
		"╰" + rule + "╯",
		"╭" + rule + "╮",
		"│ ✗ Could not reach the server  Retry │",
		"╰" + rule + "╯",
	}
	for i, w := range want {
		if got := span(f, i, 40, 79); got != w {
			t.Errorf("row %d: %q, want %q", i, got, w)
		}
	}
	// What is under them is cut, not moved.
	if got := strings.Split(f.Plain(), "\n")[1]; !strings.HasPrefix(got, "The buttons show toasts. A click on a to│") {
		t.Errorf("row 1: %q", got)
	}
	// Each kind's colour on its border and its mark, which is bold; the
	// message is the text's; the action is bold and underlined.
	for row, role := range map[int]Role{1: Info, 4: Success, 7: Warning, 10: Error} {
		if c := f.Cells[row][40]; c.Role != role {
			t.Errorf("row %d: border %v, want %v", row, c.Role, role)
		}
		if c := f.Cells[row][42]; c.Role != role || c.Attr&Bold == 0 {
			t.Errorf("row %d: mark %v %v, want %v bold", row, c.Role, c.Attr, role)
		}
		if c := f.Cells[row][44]; c.Role != Fg && c.Role != 0 {
			t.Errorf("row %d: message %v", row, c.Role)
		}
	}
	if c := f.Cells[10][72]; c.Text != "R" || c.Attr&(Bold|Underline) != Bold|Underline {
		t.Errorf("Retry: %q %v", c.Text, c.Attr)
	}
	// Without colour the kinds still read, by their marks.
	plain := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(f.ANSI(false), "")
	for _, m := range []string{"ⓘ Indexing", "✓ Draft saved", "! Disk almost", "✗ Could not"} {
		if !strings.Contains(plain, m) {
			t.Errorf("no %q without colour", m)
		}
	}
	// Narrow, a toast takes the frame's width at most, past the gutter,
	// which is blank, and its message wraps under itself, past its mark.
	f = r.Draw(24)
	got := strings.Split(f.Plain(), "\n")
	want = []string{
		"╭─────────────────────╮",
		"│ ⓘ Indexing 1,204    │",
		"│   files             │",
		"╰─────────────────────╯",
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("24 columns, row %d: %q, want %q", i, got[i], w)
		}
	}
	if !strings.Contains(f.Plain(), "│ ✗ Could not reach   │\n│   the server  Retry │") {
		t.Errorf("24 columns, the error:\n%s", f.Plain())
	}
}

// A toast counts down on the rendition's clock while it is drawn, a step
// at a time (Animating), and goes when its time is up; the pointer on it
// (Hover) holds its time, as the keyboard on its action does. A toast
// without a timeout stays.
func TestToastClock(t *testing.T) {
	r, c, _, now := toasting(t)
	r.Draw(80)
	if d := r.Animating(); d != view.ToastStep {
		t.Fatalf("Animating %v, want %v", d, view.ToastStep)
	}
	names := func() string {
		var out []string
		for _, e := range c.V.Toasts {
			out = append(out, e.Name)
		}
		return strings.Join(out, " ")
	}
	*now = now.Add(5 * time.Second)
	r.Draw(80)
	if got := names(); got != "toast-3 toast-1 offline" {
		t.Errorf("after 5s: %s, want the success gone", got)
	}
	// The pointer on the info toast (row 1) holds it, from the frame that
	// follows (a program draws one when Hover says so); the warning goes
	// on.
	if !r.Hover(60, 1) {
		t.Error("the pointer onto a toast changes nothing")
	}
	r.Draw(80)
	*now = now.Add(5 * time.Second)
	r.Draw(80)
	if got := names(); got != "toast-3 offline" {
		t.Errorf("after 10s: %s, want the warning gone and the held info kept", got)
	}
	if !r.Hover(10, 10) {
		t.Error("the pointer off a toast changes nothing")
	}
	r.Draw(80)
	*now = now.Add(3 * time.Second)
	f := r.Draw(80)
	if got := names(); got != "offline" {
		t.Errorf("after 13s: %s, want the error alone", got)
	}
	// The error has no timeout: nothing moves.
	if d := r.Animating(); d != 0 {
		t.Errorf("Animating %v with a toast that stays", d)
	}
	if !strings.Contains(f.Plain(), "Could not reach the server") {
		t.Errorf("the error went:\n%s", f.Plain())
	}
}

// The tooltip is a row of the HottyKeyHints, over its keys: the
// description of the element the pointer is on, else of the one with the
// keyboard, in italics; a key or a click hides the pointer's until it
// leaves the element.
func TestTooltipRow(t *testing.T) {
	r, c, _, _ := toasting(t)
	// The story's four toasts reach down over its tip's row, which is
	// what they do to anything under them; three leave it clear.
	if ts := c.St.Toasts; len(ts) == 0 || !c.DismissToast(ts[len(ts)-1].ID) {
		t.Fatal("no toast to dismiss")
	}
	last2 := func(f *Frame) [2]string {
		rows := strings.Split(f.Plain(), "\n")
		return [2]string{rows[len(rows)-2], rows[len(rows)-1]}
	}
	f := r.Draw(80)
	if got := last2(f); got != [2]string{"", "? more"} {
		t.Errorf("nothing focused: %q", got)
	}
	c.Focus("mute")
	f = r.Draw(80)
	if got := last2(f); got != [2]string{"Silences notifications about this conversation", "enter press • ? more"} {
		t.Errorf("mute focused: %q", got)
	}
	if c := f.Cells[f.Rows-2][0]; c.Attr&Italic == 0 || c.Role != Fg && c.Role != 0 {
		t.Errorf("the tip: %v %v, want the text's, italic", c.Role, c.Attr)
	}
	// The pointer on the star: its description.
	col, row, _, _, ok := r.Box("star")
	if !ok {
		t.Fatal("no star")
	}
	if !r.Hover(col, row) {
		t.Error("the pointer onto the star changes nothing")
	}
	f = r.Draw(80)
	if got := last2(f)[0]; got != "Keeps the conversation at the top of the list" {
		t.Errorf("star hovered: %q", got)
	}
	// On the email field's label: the field's, the nearest description.
	col, row, _, _, _ = r.Box("email")
	r.Hover(col, row)
	f = r.Draw(80)
	if got := last2(f)[0]; got != "Where receipts go; nobody else sees it" {
		t.Errorf("email hovered: %q", got)
	}
	// A key hides it: the keyboard's shows, until the pointer leaves.
	if _, err := r.Key("Tab"); err != nil {
		t.Fatal(err)
	}
	f = r.Draw(80)
	if got := last2(f)[0]; got != "Keeps the conversation at the top of the list" {
		t.Errorf("after Tab: %q, want the star's, focused", got)
	}
	if r.Hover(col+1, row) {
		t.Error("a move within the element shows its tip again")
	}
	col, row, _, _, _ = r.Box("notify")
	r.Hover(col, row)
	f = r.Draw(80)
	if got := last2(f)[0]; got != "Toasts for builds and deploys while you work" {
		t.Errorf("notify hovered: %q", got)
	}
	// Off anything described, the keyboard's again; and nothing on a
	// toast.
	r.Hover(60, 1)
	f = r.Draw(80)
	if got := last2(f)[0]; got != "Keeps the conversation at the top of the list" {
		t.Errorf("on a toast: %q", got)
	}
}

// A field's cursor under a toast is not shown, as under any box drawn
// over it.
func TestToastCursor(t *testing.T) {
	r, c, _, _ := toasting(t)
	c.Focus("email")
	if _, _, ok := r.Draw(80).Cursor(); !ok {
		t.Error("no cursor at 80 columns")
	}
	if col, row, ok := r.Draw(56).Cursor(); ok {
		t.Errorf("a cursor at %d,%d, under the error toast", col, row)
	}
}
