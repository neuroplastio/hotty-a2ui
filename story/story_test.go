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
	if groups[Basic] != 43 || groups[Hotty] == 0 || groups[Fallback] == 0 {
		t.Errorf("groups %v", groups)
	}
	if Find("00_simple-login-form") == nil || Find("hotty/form") == nil {
		t.Error("Find")
	}
}
