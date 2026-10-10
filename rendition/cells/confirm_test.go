package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/story"
)

// The Confirm stories (vault KIT-16c) are a pattern, not a component: two
// Buttons in a Row, the one with the keyboard reversed in the accent and
// the other filled grey, as huh shows the answer it would give, and the
// key hints of the HottyShortcuts that answer and move. ArrowLeft and
// ArrowRight move the keyboard between them (hottyFocus), y and n press
// them.
func TestConfirm(t *testing.T) {
	run, err := story.Start(story.Find("hotty/confirm"))
	if err != nil {
		t.Fatal(err)
	}
	r := New(run.Surfaces()[0].C)
	want := "Delete 3 files?\n" +
		"notes.md, todo.md and draft.md go to the trash.\n" +
		" Yes   No\n" +
		"\n" +
		"enter press • ←/→ move • y yes • n no • ? more"
	f := r.Draw(64)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	// picked reports whether a Button is reversed in the accent, the other
	// filled.
	picked := func(f *Frame, id string) bool {
		col, row, _, _, ok := r.Box(id)
		if !ok {
			t.Fatalf("no box for %s", id)
		}
		g := f.Cells[row][col+1]
		switch {
		case g.Role == Accent && g.Attr == Reverse && g.BackMix == 0:
			return true
		case g.Attr == 0 && g.Back == Fg && g.BackMix == buttonTint:
			return false
		}
		t.Fatalf("%s is neither picked nor filled: %+v", id, g)
		return false
	}
	if picked(f, "yes") || !picked(f, "no") {
		t.Error("at the start, No is not the one picked")
	}
	for _, step := range []struct {
		key, picked string
		actions     int
	}{
		{"ArrowLeft", "yes", 0},
		{"ArrowRight", "no", 0},
		{"y", "no", 1},
		{"ArrowLeft", "yes", 1},
		{"n", "yes", 2},
	} {
		if _, err := r.Key(step.key); err != nil {
			t.Fatal(err)
		}
		f = r.Draw(64)
		if picked(f, "yes") != (step.picked == "yes") || picked(f, "no") != (step.picked == "no") {
			t.Errorf("after %s, %s is not the one picked", step.key, step.picked)
		}
		if n := len(run.Actions()); n != step.actions {
			t.Errorf("after %s, %d actions, want %d", step.key, n, step.actions)
		}
	}

	// In the form, the field's keys are its own: the hints leave out what
	// it types and the arrows it moves by, until Yes has the keyboard.
	run, err = story.Start(story.Find("hotty/confirm-form"))
	if err != nil {
		t.Fatal(err)
	}
	r = New(run.Surfaces()[0].C)
	want = "Create the repository?\n" +
		"\n" +
		"┃ Name         hotty-kit\n" +
		"  Description\n" +
		"\n" +
		" Yes   No\n" +
		"\n" +
		"enter submit • ? more"
	if got := r.Draw(64).Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, k := range []string{"Tab", "Tab"} {
		if _, err := r.Key(k); err != nil {
			t.Fatal(err)
		}
	}
	f = r.Draw(64)
	if !picked(f, "yes") || picked(f, "no") {
		t.Error("Tab past the fields does not pick Yes")
	}
	if got := f.Plain(); !strings.HasSuffix(got, "\nenter press • ←/→ move • y yes • n no • ? more") {
		t.Errorf("the hints on Yes:\n%s", got)
	}
}
