package cells

import "testing"

// TestUndo: a text control's history, as a page keeps one. A run of
// typing undoes as one step, and so does a run of deleting; redo brings
// them back, and an edit after an undo drops what could be redone. A
// caret moved between keys starts a new step; a cut undoes with its text
// selected again, and a paste is a step of its own. A value set from
// outside drops the history, and each control has its own.
func TestUndo(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	want := func(v string) {
		t.Helper()
		if got := data()["name"]; got != v {
			t.Fatalf("the name is %q, want %q", got, v)
		}
	}
	keys(t, r, "shift+a", "d", "a", "space", "shift+l", "o", "backspace", "backspace")
	want("Ada ")
	keys(t, r, "ctrl+z")
	want("Ada Lo")
	keys(t, r, "ctrl+z")
	want("")
	keys(t, r, "ctrl+z") // nothing to undo: the field takes it still
	want("")
	keys(t, r, "ctrl+y")
	want("Ada Lo")
	keys(t, r, "ctrl+shift+z")
	want("Ada ")
	keys(t, r, "ctrl+z", "x")
	want("Ada Lox")
	keys(t, r, "ctrl+y")
	want("Ada Lox")
	keys(t, r, "arrow_left", "y", "z")
	want("Ada Loyzx")
	keys(t, r, "ctrl+z")
	want("Ada Lox")

	keys(t, r, "end", "shift+arrow_left", "shift+arrow_left")
	keys(t, r, "ctrl+x")
	want("Ada L")
	keys(t, r, "ctrl+z")
	want("Ada Lox")
	if got := r.Selection(); got != "ox" {
		t.Errorf("undoing the cut selected %q, want \"ox\"", got)
	}
	if _, err := r.Paste("!"); err != nil {
		t.Fatal(err)
	}
	want("Ada L!")
	keys(t, r, "ctrl+z")
	want("Ada Lox")

	if err := c.SetValue("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	r.Draw(40)
	keys(t, r, "ctrl+z")
	want("Grace")

	c.Focus("note")
	r.Draw(40)
	keys(t, r, "h", "i", "ctrl+z")
	if got := data()["note"]; got != "" {
		t.Errorf("the note after an undo: %q", got)
	}
	want("Grace")

	r.SetKeys("mod+z=program")
	if ok, _ := r.Key("ctrl+z"); ok {
		t.Error("ctrl+z bound to program was taken")
	}
	c.Focus("go")
	r.Draw(40)
	if ok, _ := r.Key("ctrl+y"); ok {
		t.Error("ctrl+y on a button was taken")
	}
}
