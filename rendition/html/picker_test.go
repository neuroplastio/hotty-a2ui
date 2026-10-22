package html

import (
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// TestPickerOnHost: a HottyPicker is a box holding its query as an input
// of its own, which the host edits and gives the program the picker's keys
// from, its rows (the selected one marked) and its line of keys. A click on
// a row selects it, and a second runs its first action. A query the program
// writes goes to the input with a=field, which keeps its id and the
// keyboard, and puts the caret after it (SPEC §6.4).
func TestPickerOnHost(t *testing.T) {
	x, r := pickerHarness(t)
	s := x.h.Surface(r.name)
	input := pickInput(r.C.V.Find("root"))
	for _, want := range [][3]string{{input, "role", "combobox"}, {input, "placeholder", "Type a command"},
		{input, "data-keys", pickKeys(r.C.V.Find("root"))}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	row := func(i int) string { return partID("root", partItem+strconv.Itoa(i)) }
	if got := s.TextOf(row(0)); got != "Pane: Splitctrl+b %" {
		t.Errorf("the first row reads %q", got)
	}
	if v, _ := s.Attr(row(0), "aria-selected"); v != "true" {
		t.Errorf("the first row is not selected")
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if got := r.C.S.Data.Value("/sel"); got != "p-1" || r.C.St.Focus != "root" || s.Focused() != input {
		t.Fatalf("a click on fish: selected %v, focus %q, the host's on %q", got, r.C.St.Focus, s.Focused())
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if len(x.actions) != 1 || x.actions[0].Context["row"] != "p-1" || x.actions[0].Context["action"] != "" {
		t.Fatalf("a second click: %+v", x.actions)
	}
	typeQuery(t, x, r, "sp")
	if got := r.C.S.Data.Value("/q"); got != "sp" {
		t.Fatalf("after s, p: /q %v", got)
	}
	if len(fields(x)) != 0 {
		t.Fatalf("typing sent a=field: %+v", fields(x))
	}
	// The program writes another query: the same input, with the keyboard,
	// its caret after the query, and the rev the user's two keys made.
	must(t, r.C.S.Write("/q", "split"))
	r.C.Rebuild()
	x.update(r)
	if v, _ := s.Value(input); v != "split" || s.Focused() != input {
		t.Fatalf("the program's query: value %q, the host's keyboard on %q", v, s.Focused())
	}
	if a, c, _ := s.Selection(input); a != 5 || c != 5 {
		t.Errorf("the program's query: selection %d, %d, want the caret after it", a, c)
	}
	if rev, _ := s.Rev(input); rev != 2 {
		t.Errorf("the program's query made the rev %d", rev)
	}
	// The user types on after it.
	typeQuery(t, x, r, "s")
	if got := r.C.S.Data.Value("/q"); got != "splits" {
		t.Errorf("typed after the program's query: /q %v", got)
	}
	leave(t, x, r)
}

// TestPickerCompletes: Tab completes the query to the selected row's
// title, in the input the user typed in: the caret goes after the
// completed text, and what they type next follows it.
func TestPickerCompletes(t *testing.T) {
	x, r := pickerHarness(t)
	s := x.h.Surface(r.name)
	input := pickInput(r.C.V.Find("root"))
	focusPicker(x, r)
	typeQuery(t, x, r, "sp")
	x.h.Key("tab")
	x.pump()
	if v := r.C.S.Data.Value("/q"); v != "Split" {
		t.Fatalf("Tab after sp: /q %v", v)
	}
	if got := fields(x); len(got) != 1 || !splice(got[0], 2, 0, 2, "Split") {
		t.Fatalf("Tab's completion: %+v, want a=field from 0 to 2, Split, against rev 2", got)
	}
	if v, _ := s.Value(input); v != "Split" || s.Focused() != input {
		t.Fatalf("completed: value %q, the host's keyboard on %q", v, s.Focused())
	}
	if a, c, _ := s.Selection(input); a != 5 || c != 5 {
		t.Errorf("completed: selection %d, %d, want the caret after Split", a, c)
	}
	typeQuery(t, x, r, " ")
	if v, _ := s.Value(input); v != "Split " || r.C.S.Data.Value("/q") != "Split " {
		t.Errorf("typed after the completion: %q, /q %v", v, r.C.S.Data.Value("/q"))
	}
	leave(t, x, r)
}

// TestPickerStale: a query the program writes while the user types on is
// stale (ESTALE), and is dropped: what they typed stays, and becomes the
// query once their input event comes (SPEC §6.4).
func TestPickerStale(t *testing.T) {
	x, r := pickerHarness(t)
	s := x.h.Surface(r.name)
	input := pickInput(r.C.V.Find("root"))
	focusPicker(x, r)
	typeQuery(t, x, r, "sp")
	// l, its input event not heard yet.
	x.h.Key("l")
	must(t, r.C.S.Write("/q", "split"))
	r.C.Rebuild()
	x.update(r)
	if v, _ := s.Value(input); v != "spl" {
		t.Fatalf("the stale query changed the field: %q", v)
	}
	var stale []hotty.Reply
	for _, m := range x.h.Replies() {
		if rep, ok := hotty.ReplyOf(m); ok && rep.Re == "field" {
			stale = append(stale, rep)
		}
	}
	if len(stale) != 1 || stale[0].Code != hotty.ESTALE || stale[0].Detail != "3" || !r.Reply(stale[0]) {
		t.Fatalf("the replies to a=field: %+v", stale)
	}
	if r.Reply(hotty.Reply{Re: "field", Surface: r.name, Code: hotty.EINVAL}) || r.Reply(hotty.Reply{Re: "field", Surface: "other", Code: hotty.ESTALE}) {
		t.Error("Reply took an error not its own")
	}
	x.pump()
	x.update(r)
	if v := r.C.S.Data.Value("/q"); v != "spl" {
		t.Errorf("after the user's input event: /q %v", v)
	}
	// The next query the program writes is made against the rev it heard.
	must(t, r.C.S.Write("/q", "split"))
	r.C.Rebuild()
	x.update(r)
	if v, _ := s.Value(input); v != "split" {
		t.Errorf("the program's next query: %q", v)
	}
	leave(t, x, r)
}

// TestPickerDescription: a row with a description is two lines, its
// title's (with its keys) and the description's under it, in a
// class of its own (k-pick-desc, smaller and muted); a row without one is
// one line, as it was.
func TestPickerDescription(t *testing.T) {
	x, r := pickerHarnessOf(t, 10, `[{"value":"pane.new","title":"pane.new","description":"New pane (your shell)","keys":"ctrl+b n",
  "actions":[{"id":"run","title":"Run"},{"id":"keep","title":"Keep","key":"mod+enter"}]},
 {"value":"pane.split","title":"pane.split","detail":"right"}]`)
	s := x.h.Surface(r.name)
	row := func(i int) string { return partID("root", partItem+strconv.Itoa(i)) }
	html := s.HTML()
	for _, want := range []string{
		`<div class="k-pick-line"><span class="k-pick-title">pane.new</span><span class="k-pick-keys">ctrl+b n</span></div><div class="k-pick-desc">New pane (your shell)</div>`,
		`<span class="k-pick-title">pane.split<span class="k-pick-detail">right</span></span></div>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("no %s in\n%s", want, html)
		}
	}
	for i, want := range []string{"k-pick-row k-pick-two k-sel", "k-pick-row"} {
		if got, _ := s.Attr(row(i), "class"); got != want {
			t.Errorf("row %d's class is %q, want %q", i, got, want)
		}
	}
	// Two lines high, the first row shows without its description.
	x, r = pickerHarnessOf(t, 2, `[{"value":"a","title":"a"},{"value":"b","title":"b","description":"Bee"}]`)
	if html := x.h.Surface(r.name).HTML(); strings.Contains(html, "Bee") || !strings.Contains(html, ">b<") {
		t.Errorf("the description past the last line shows:\n%s", html)
	}
}

// TestPickerStatus: a picker's status reads as its head's, not as a row:
// under the query, in a class of its own (k-pick-status, muted and smaller,
// with a rule under it), and an error's keeps its colour and its ✗.
func TestPickerStatus(t *testing.T) {
	if !regexp.MustCompile(`\.k-pick-status \{[^}]*color: var\(--k-muted\);[^}]*font-size: 0\.875rem;[^}]*border-bottom: `).MatchString(kitCSS) {
		t.Error("kit.css sets the status off neither muted, smaller nor ruled")
	}
	for _, c := range []struct{ kind, class, text string }{
		{"info", "k-pick-status", "theme.use · theme"},
		{"error", "k-pick-status k-error", "✗ theme.use · theme"},
	} {
		x := newHarness(t)
		var msgs []any
		must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"q":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"HottyPicker","catalogId":"`+hottycat.ID+`","query":{"@path":"/q"},"status":"theme.use · theme","statusKind":"`+c.kind+`",
  "rows":[{"value":"a","title":"a"}]}]}}]`), &msgs))
		x.process(msgs...)
		s := x.h.Surface(x.rs["s"].name)
		want := `</div><div id="` + partID("root", partStatus) + `" class="` + c.class + `" aria-live="polite">` + c.text + `</div><div class="k-pick-rows"`
		if html := s.HTML(); !strings.Contains(html, want) {
			t.Errorf("%s: no %s in\n%s", c.kind, want, html)
		}
	}
}

// TestPickerFooter: a row's actions show in the footer alone, not on the
// row: each its title and its key, the most significant first, which the
// CSS puts at the right (Enter's, the others with a key, the panel's); in
// the panel, Enter's and Escape's.
func TestPickerFooter(t *testing.T) {
	x, r := pickerHarnessOf(t, 10, `[{"value":"p-1","title":"Pane: gov","detail":"main · p-3",
  "actions":[{"id":"focus","title":"Go to pane"},{"id":"close","title":"Close pane","key":"shift+delete"}]}]`)
	s := x.h.Surface(r.name)
	if got := s.TextOf(partID("root", partItem+"0")); got != "Pane: govmain · p-3" {
		t.Errorf("the row reads %q", got)
	}
	act := func(title, key string) string {
		return `<span class="k-pick-act">` + title + `<span class="k-pick-key">` + key + `</span></span>`
	}
	want := `<div class="k-pick-keys-line">` + act("Go to pane", "enter") + act("Close pane", "shift+delete") + act("Actions", "ctrl+k") + `</div>`
	if html := s.HTML(); !strings.Contains(html, want) {
		t.Errorf("no %s in\n%s", want, html)
	}
	must(t, func() error { _, err := r.C.PickerKey("root", "ctrl+k"); return err }())
	x.update(r)
	want = `<div class="k-pick-keys-line">` + act("Run", "enter") + act("Back", "esc") + `</div>`
	if html := s.HTML(); !strings.Contains(html, want) {
		t.Errorf("in the panel, no %s in\n%s", want, html)
	}
}

// pickerHarness shows a HottyPicker, root, as surface s's, whose query is
// /q.
func pickerHarness(t *testing.T) (*harness, *Rendition) {
	t.Helper()
	return pickerHarnessOf(t, 10, `[{"value":"split","group":"Pane","title":"Split","keys":"ctrl+b %","section":"Used","match":[[0,2]],"actions":[{"id":"run","title":"Run"},{"id":"again","title":"Again","key":"mod+enter"}]},
   {"value":"p-1","group":"Pane","title":"fish","detail":"~/code","section":"Panes"}]`)
}

// pickerHarnessOf is pickerHarness on rows, height lines high.
func pickerHarnessOf(t *testing.T, height int, rows string) (*harness, *Rendition) {
	t.Helper()
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"q":"","sel":"","act":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"HottyPicker","catalogId":"`+hottycat.ID+`","query":{"@path":"/q"},"selected":{"@path":"/sel"},"action":{"@path":"/act"},
  "chip":">","placeholder":"Type a command","panelKey":"mod+k","height":`+strconv.Itoa(height)+`,
  "rows":`+rows+`,
  "onAction":{"event":{"name":"act","context":{"row":{"@path":"/sel"},"action":{"@path":"/act"}}}}}]}}]`), &msgs))
	x.process(msgs...)
	return x, x.rs["s"]
}

// focusPicker gives the picker the keyboard, which Update gives its input.
func focusPicker(x *harness, r *Rendition) {
	r.C.Focus("root")
	x.update(r)
	x.pump()
}

// leave: the user leaves the picker. Its input reads the query, and the
// host's document is the view's, but for the input's value attribute,
// which stays the query it entered with (complete).
func leave(t *testing.T, x *harness, r *Rendition) {
	t.Helper()
	must(t, x.h.Blur(r.name))
	x.pump()
	s := x.h.Surface(r.name)
	input := pickInput(r.C.V.Find("root"))
	if v, _ := s.Value(input); v != r.C.S.Data.Value("/q") {
		t.Errorf("left: the input reads %q, the query is %v", v, r.C.S.Data.Value("/q"))
	}
	entered, _ := s.Attr(input, "value")
	h2 := hottytest.New(t)
	fresh := &Rendition{C: r.C, name: r.name, keys: r.keys, fit: r.fit, Clock: r.Clock}
	for _, c := range []string{hotty.Doc(r.name, fresh.Doc()), hotty.Delta(r.name, []hotty.Op{hotty.OpAttr(input, "value", entered)})} {
		_, err := io.WriteString(h2, c)
		must(t, err)
	}
	if got, want := s.HTML(), h2.Surface(r.name).HTML(); got != want {
		t.Errorf("the deltas made\n%s\nthe document is\n%s", got, want)
	}
}

// splice reports whether f replaces from..to with text, against rev, and
// does nothing else.
func splice(f hotty.FieldChange, rev, from, to int, text string) bool {
	return f.Rev == rev && f.From != nil && *f.From == from && f.To != nil && *f.To == to && f.Text != nil && *f.Text == text &&
		f.Value == nil && f.Sel == nil && f.Checked == nil
}

// fields are the a=field commands the host took.
func fields(x *harness) []hotty.FieldChange {
	var out []hotty.FieldChange
	for _, m := range x.h.Commands() {
		if hotty.Get(m.Control, "a") == "field" {
			f, err := hotty.FieldOf(m)
			must(x.t, err)
			out = append(out, f)
		}
	}
	return out
}

// typeQuery types keys into the focused picker's input, the rendition
// hearing each.
func typeQuery(t *testing.T, x *harness, r *Rendition, keys string) {
	t.Helper()
	for _, k := range keys {
		key := string(k)
		if k == ' ' {
			key = "space"
		}
		if !x.h.Key(key) {
			t.Fatalf("the host did not type %q into the query", key)
		}
		x.pump()
		x.update(r)
	}
}
