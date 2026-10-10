package text_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// TestConfirm: the Confirm pattern (vault KIT-16c) reads as its question
// and its two Buttons on one line, in a form after the fields; its
// HottyShortcuts say nothing, as a pipe takes no keys.
func TestConfirm(t *testing.T) {
	for name, want := range map[string]string{
		"hotty/confirm":      "Delete 3 files?\nnotes.md, todo.md and draft.md go to the trash.\n[ Yes ]  [ No ]\n",
		"hotty/confirm-form": "Create the repository?\nName: hotty-kit\nDescription: \n[ Yes ]  [ No ]\n",
	} {
		run, err := story.Start(story.Find(name))
		if err != nil {
			t.Fatal(err)
		}
		if got := text.Render(run.Surfaces()[0].C.V); got != want {
			t.Errorf("%s: got\n%s\nwant\n%s", name, got, want)
		}
	}
}
