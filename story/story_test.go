package story

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// TestStories runs every story: none fails, each makes a surface, and only
// the fallback stories show placeholders, or report errors to the agent.
func TestStories(t *testing.T) {
	all := All()
	groups := map[string]int{}
	for _, st := range all {
		groups[st.Group]++
		r, err := Start(st)
		if err != nil {
			t.Errorf("%s: %v", st.Name, err)
			continue
		}
		if len(r.Surfaces()) == 0 {
			t.Errorf("%s: no surface", st.Name)
		}
		for _, s := range r.Surfaces() {
			s.C.V.Walk(func(e *view.Element) bool {
				if e.Kind == view.Placeholder && st.Group != Fallback {
					t.Errorf("%s: %s is a placeholder (%s)", st.Name, e.ID, e.State)
				}
				return true
			})
		}
		for _, e := range r.Log {
			if e.Out && strings.Contains(string(e.JSON), `"error"`) && st.Group != Fallback {
				t.Errorf("%s: the renderer reported %s", st.Name, e.JSON)
			}
		}
	}
	// Basic: A2UI's 43 examples, then the kit's icons.
	if groups[Basic] != 43+1 || groups[Hotty] == 0 || groups[Fallback] == 0 {
		t.Errorf("groups %v", groups)
	}
	if Find("00_simple-login-form") == nil || Find("hotty/form") == nil {
		t.Error("Find")
	}
}

// TestTreeButtons: the tree story's buttons call hottyExpandAll and
// hottyCollapseAll, which the run implements.
func TestTreeButtons(t *testing.T) {
	run, err := Start(Find("hotty/tree"))
	if err != nil {
		t.Fatal(err)
	}
	c := run.Surfaces()[0].C
	if err := c.Activate("expand"); err != nil {
		t.Fatal(err)
	}
	open, _ := c.S.Data.Value("/expanded").([]any)
	if len(open) < 20 {
		t.Errorf("expand all opened %d branches", len(open))
	}
	if err := c.Activate("collapse"); err != nil {
		t.Fatal(err)
	}
	if open, _ := c.S.Data.Value("/expanded").([]any); len(open) != 0 || c.S.Data.Value("/selected") != "rendition" {
		t.Errorf("collapse all: open %v, selected %v", open, c.S.Data.Value("/selected"))
	}
}

// TestKitStoriesSayWhatTheyAre: every story of the kit's own has an icon,
// and every one of the basic and hotty groups a kind and the A2UI names
// it is about, so that the storybook can list it.
func TestKitStoriesSayWhatTheyAre(t *testing.T) {
	for _, st := range All() {
		own := st.Group != Basic || st.Kind != ""
		if !own {
			continue
		}
		if st.Icon == "" {
			t.Errorf("%s has no icon", st.Name)
		}
		if st.Group != Fallback && (st.Component == "" || st.Kind != KindComponent && st.Kind != KindBehaviour) {
			t.Errorf("%s: kind %q, component %q", st.Name, st.Kind, st.Component)
		}
	}
}
