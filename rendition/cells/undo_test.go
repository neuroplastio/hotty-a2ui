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
	keys(t, r, "A", "d", "a", "Space", "L", "o", "Backspace", "Backspace")
	want("Ada ")
	keys(t, r, "Control+z")
	want("Ada Lo")
	keys(t, r, "Control+z")
	want("")
	keys(t, r, "Control+z") // nothing to undo: the field takes it still
	want("")
	keys(t, r, "Control+y")
	want("Ada Lo")
	keys(t, r, "Shift+Control+z")
	want("Ada ")
	keys(t, r, "Control+z", "x")
	want("Ada Lox")
	keys(t, r, "Control+y")
	want("Ada Lox")
	keys(t, r, "ArrowLeft", "y", "z")
	want("Ada Loyzx")
	keys(t, r, "Control+z")
	want("Ada Lox")

	keys(t, r, "End", "Shift+ArrowLeft", "Shift+ArrowLeft")
	keys(t, r, "Control+x")
	want("Ada L")
	keys(t, r, "Control+z")
	want("Ada Lox")
	if got := r.Selection(); got != "ox" {
		t.Errorf("undoing the cut selected %q, want \"ox\"", got)
	}
	if _, err := r.Paste("!"); err != nil {
		t.Fatal(err)
	}
	want("Ada L!")
	keys(t, r, "Control+z")
	want("Ada Lox")

	if err := c.SetValue("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	r.Draw(40)
	keys(t, r, "Control+z")
	want("Grace")

	c.Focus("note")
	r.Draw(40)
	keys(t, r, "h", "i", "Control+z")
	if got := data()["note"]; got != "" {
		t.Errorf("the note after an undo: %q", got)
	}
	want("Grace")

	r.SetKeys("Mod+z=program")
	if ok, _ := r.Key("Control+z"); ok {
		t.Error("Control+z bound to program was taken")
	}
	c.Focus("go")
	r.Draw(40)
	if ok, _ := r.Key("Control+y"); ok {
		t.Error("Control+y on a button was taken")
	}
}
