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

// TestFilePickerPickLoadsCode: picking a file in the filepicker story
// loads its content into /code.
func TestFilePickerPickLoadsCode(t *testing.T) {
	run, err := Start(Find("hotty/filepicker"))
	if err != nil {
		t.Fatal(err)
	}
	c := run.Surfaces()[0].C
	// Entry 2 is go.mod
	if err := c.ClickFile("fp", 2); err != nil {
		t.Fatal(err)
	}
	// Second click picks it
	if err := c.ClickFile("fp", 2); err != nil {
		t.Fatal(err)
	}
	code, _ := c.S.Data.Value("/code").(string)
	if !strings.Contains(code, "module example.com/project") {
		t.Errorf("got code %q, want module example.com/project", code)
	}
}

// TestCheatSheetKeymap: picking a keymap in hotty/cheatsheet sends the
// sheet that keymap's keys, before the next draw (Run.Configure), which
// its HottyKeys show; drawing it again sends nothing more.
func TestCheatSheetKeymap(t *testing.T) {
	run, err := Start(Find("hotty/cheatsheet"))
	if err != nil {
		t.Fatal(err)
	}
	s := run.Surfaces()[0]
	run.Configure(s)
	if e := s.C.V.Find("k_t"); e == nil || e.Label != "ctrl+b w t" {
		t.Fatalf("at first: %+v", e)
	}
	sent := len(run.Log)
	if err := s.C.SetValue("keymap", "screen"); err != nil {
		t.Fatal(err)
	}
	run.Configure(s)
	for id, want := range map[string]string{"k_t": "ctrl+a w t", "k_f": "ctrl+a /", "k_p": "mod+shift+p"} {
		if e := s.C.V.Find(id); e == nil || e.Label != want {
			t.Errorf("after screen, %s: %+v, want %q", id, e, want)
		}
	}
	if n := len(run.Log) - sent; n != 1 {
		t.Errorf("screen sent %d messages, want 1", n)
	}
	run.Configure(s)
	if n := len(run.Log) - sent; n != 1 {
		t.Errorf("drawn again, %d messages, want 1", n)
	}
}

// TestKeycapsControls: hotty/keycaps's controls set how its surfaces draw
// keys, as a program's config would (Run.Configure).
func TestKeycapsControls(t *testing.T) {
	run, err := Start(Find("hotty/keycaps"))
	if err != nil {
		t.Fatal(err)
	}
	s := run.Surfaces()[0]
	run.Configure(s)
	if k := s.C.Keycaps; k.OS != "" || k.Present != "" || k.Labels != nil {
		t.Errorf("at first %+v, want the zero Keycaps", k)
	}
	for _, m := range []string{
		`{"version":"v1.0","updateDataModel":{"surfaceId":"keys","path":"/system","value":["windows"]}}`,
		`{"version":"v1.0","updateDataModel":{"surfaceId":"keys","path":"/present","value":["text"]}}`,
		`{"version":"v1.0","updateDataModel":{"surfaceId":"keys","path":"/meta","value":["◆"]}}`,
		`{"version":"v1.0","updateDataModel":{"surfaceId":"keys","path":"/case","value":["lower"]}}`,
		`{"version":"v1.0","updateDataModel":{"surfaceId":"keys","path":"/sep","value":["-"]}}`,
		`{"version":"v1.0","updateDataModel":{"surfaceId":"keys","path":"/glue","value":["on"]}}`,
	} {
		if err := run.Feed([]byte(m)); err != nil {
			t.Fatal(err)
		}
	}
	run.Configure(s)
	if k := s.C.Keycaps; k.OS != "windows" || k.Present != view.PresentText || k.Labels["super"] != "◆" ||
		k.Case != view.CaseLower || k.Sep != "-" || k.Glue != view.GlueOn {
		t.Errorf("after the controls %+v", k)
	}
}
