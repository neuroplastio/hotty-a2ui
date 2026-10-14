package cells

import (
	"testing"

	hottygo "github.com/neuroplastio/hotty-go"
)

// TestCommands: an edit command a host sends in place of its key (SPEC
// §10.6). A paste into a text field brings its text, which is typed where
// the key would have asked for the clipboard; one with no text does
// nothing. Any other command with a key is that key, which the program
// gives to Key. One from the terminal's menu, with no key, does what it
// names in the field: select-all, copy, cut, undo.
func TestCommands(t *testing.T) {
	c, data, _ := form(t)
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	command := func(name, key, text string, hasText bool) bool {
		t.Helper()
		ok, err := r.Command(hottygo.Command{Name: name, Key: key, Text: text, HasText: hasText})
		if err != nil {
			t.Fatal(err)
		}
		r.Draw(40)
		return ok
	}
	if !command("paste", "Control+v", "Ada", true) || data()["name"] != "Ada" || r.PasteWanted() {
		t.Fatalf("a paste with its text: %v, asked again %v", data()["name"], r.PasteWanted())
	}
	if !command("paste", "Control+v", "", false) || data()["name"] != "Ada" {
		t.Errorf("a paste with no text changed the field: %v", data()["name"])
	}
	if command("copy", "Control+c", "", false) {
		t.Error("copy with a key was not left to Key")
	}
	if !command("select-all", "", "", false) || r.Selection() != "Ada" {
		t.Errorf("select-all from the menu selected %q", r.Selection())
	}
	if !command("copy", "", "", false) {
		t.Error("copy from the menu did not copy")
	}
	if got, _ := r.Copied(); got != (Clip{Text: "Ada"}) {
		t.Errorf("copy from the menu put %+v", got)
	}
	if !command("cut", "", "", false) || data()["name"] != "" {
		t.Errorf("cut from the menu left %v", data()["name"])
	}
	if !command("undo", "", "", false) || data()["name"] != "Ada" {
		t.Errorf("undo from the menu left %v", data()["name"])
	}
	r.SetKeys("Mod+v=program")
	if command("paste", "Control+v", "x", true) || data()["name"] != "Ada" {
		t.Error("a paste whose key is bound to program pasted")
	}
}

// TestEditPlatform: where Mod is Meta (macOS), a field's edit keys are
// Command's, and Control's are not edits: Control+z reaches the program.
func TestEditPlatform(t *testing.T) {
	c, data, _ := form(t)
	c.Platform = hottygo.Platform{Mod: "Meta"}
	r := New(c)
	c.Focus("name")
	r.Draw(40)
	keys(t, r, "A", "d", "a", "Meta+z")
	if got := data()["name"]; got != "" {
		t.Errorf("Meta+z left %q", got)
	}
	keys(t, r, "Meta+Z")
	if got := data()["name"]; got != "Ada" {
		t.Errorf("Meta+Shift+z left %q", got)
	}
	if ok, _ := r.Key("Control+z"); ok {
		t.Error("Control+z was taken where Mod is Meta")
	}
}
