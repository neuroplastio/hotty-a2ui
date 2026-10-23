// Package story is the storybook's model: the stories, and a story as it
// runs. A story is A2UI messages, as A2UI's examples are written. Running
// one feeds them to a processor, keeps a view controller for each surface
// it makes, and logs every message both ways: what the agent sent, and
// what the renderer sends back (the user's actions, errors, function
// calls and answers).
package story

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/a2ui/schema"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// The groups of stories.
const (
	Basic    = "basic"    // the basic catalog: A2UI's examples as A2UI has them, then the kit's own
	Hotty    = "hotty"    // the hotty catalog: Shortcut, Form, focus, blur
	Fallback = "fallback" // what has no terminal meaning, and what fails
)

// Story is one story: A2UI messages, and what they show.
type Story struct {
	// Name is the story's handle: its group and its file's name without
	// .json ("basic/00_simple-login-form").
	Name        string            `json:"-"`
	Group       string            `json:"-"`
	Title       string            `json:"name"`
	Description string            `json:"description"`
	Messages    []json.RawMessage `json:"messages"`
	// Kind, Component and Icon are the kit's own stories' (A2UI's examples
	// have none): what the storybook lists it under, "component" or
	// "behaviour" (a fallback's group says); the A2UI names it is about
	// ("HottyTree", "hottyFocus, hottyBlur"); and its icon, a Material
	// Symbols name the storybook registers (storybook/icons.txt).
	Kind      string `json:"kind,omitempty"`
	Component string `json:"component,omitempty"`
	Icon      string `json:"icon,omitempty"`
	// Topic gathers stories of a kind under a node of their own in the
	// storybook's list ("Drag and drop"); "" for none.
	Topic string `json:"topic,omitempty"`
}

// The kinds of the kit's own stories (Story.Kind).
const (
	KindComponent = "component"
	KindBehaviour = "behaviour"
)

//go:embed stories/*/*.json
var stories embed.FS

// All are the stories, by group (basic, hotty, fallback) and name. They
// are read once, and shared: a story is not changed once read.
func All() []*Story { return slices.Clone(all()) }

var all = sync.OnceValue(func() []*Story {
	var out []*Story
	add := func(fsys fs.FS, glob, group string) {
		names, _ := fs.Glob(fsys, glob)
		for _, n := range names {
			b, _ := fs.ReadFile(fsys, n)
			st, err := Parse(b)
			if err != nil {
				panic(fmt.Sprintf("story %s: %v", n, err))
			}
			st.Group = group
			st.Name = group + "/" + strings.TrimSuffix(path.Base(n), ".json")
			out = append(out, st)
		}
	}
	add(thirdparty.BasicExamples, "a2ui/catalogs/basic/v1/examples/*.json", Basic)
	add(stories, "stories/basic/*.json", Basic)
	add(stories, "stories/hotty/*.json", Hotty)
	add(stories, "stories/fallback/*.json", Fallback)
	return out
})

// Find is the story with a name, or with a name that ends in it
// ("00_simple-login-form"); nil if none, or more than one.
func Find(name string) *Story {
	var found *Story
	for _, st := range all() {
		if st.Name == name {
			return st
		}
		if strings.HasSuffix(st.Name, "/"+name) {
			if found != nil {
				return nil
			}
			found = st
		}
	}
	return found
}

// Parse reads a story, as A2UI's examples are written: {"name",
// "description", "messages"}.
func Parse(b []byte) (*Story, error) {
	var st Story
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	if len(st.Messages) == 0 {
		return nil, errors.New("no messages")
	}
	return &st, nil
}

// Entry is one message of a run's log.
type Entry struct {
	// Out is set for what the renderer sent; else the agent sent it.
	Out  bool
	At   time.Time
	JSON json.RawMessage
	// Err is the processor's error for an agent's message, if it failed.
	Err error
}

// Surface is one surface of a running story, with its view.
type Surface struct {
	S *a2ui.Surface
	C *view.Controller
}

// Run is a story running.
type Run struct {
	P   *a2ui.Processor
	Log []Entry
	// Out, if set, gets what the renderer sends the agent, as it goes.
	Out func(a2ui.Outbound)
	// Changed, if set, is called when the agent's messages change a
	// surface, after its view is built again; deleted when it went.
	Changed func(s *Surface, deleted bool)
	// roots are the filesystems granted to each surface's controller, by
	// name (Grant); hears hears the user's actions, and config sets a
	// view up before it is drawn, as the program behind the story does
	// (Program, Configure).
	roots  map[string]fs.FS
	hears  func(a *a2ui.ActionMessage)
	config func(s *Surface)

	surfaces []*Surface
}

// validator is every run's: it compiles the catalogs' schemas once.
var validator schema.Validator

// NewRun is a run with nothing fed yet. Its processor has the catalogs
// the kit renders, basic and hotty, and hotty's hottyFocus and hottyBlur
// move the keyboard in the run's views, hottyScrollTo scrolls them,
// hottyExpandAll and hottyCollapseAll fold their trees, hottyToast and
// hottyDismissToast show and take away their toasts, and hottyStartTimer,
// hottyStopTimer, hottyToggleTimer and hottyResetTimer work their timers.
func NewRun() *Run {
	r := &Run{}
	h := hotty.Catalog()
	hotty.Implement(h, hotty.Renderer{Focus: r.focus, Blur: r.blur, ScrollTo: r.scrollTo, FoldAll: r.foldAll,
		Toast: r.toast, DismissToast: r.dismissToast, Timer: r.timer})
	// The stories and -stream are A2UI another program wrote: checked,
	// against schemas compiled once for every run.
	r.P = schema.NewProcessor(basic.Catalog(), h)
	r.P.Validator = &validator
	r.P.Send = func(o a2ui.Outbound) {
		b, _ := json.Marshal(o)
		r.Log = append(r.Log, Entry{Out: true, At: time.Now(), JSON: b})
		if r.Out != nil {
			r.Out(o)
		}
		if o.Action != nil && r.hears != nil {
			r.hears(o.Action)
		}
	}
	r.P.Changed = func(s *a2ui.Surface, deleted bool) {
		i := slices.IndexFunc(r.surfaces, func(x *Surface) bool { return x.S == s })
		switch {
		case deleted && i >= 0:
			gone := r.surfaces[i]
			r.surfaces = slices.Delete(r.surfaces, i, i+1)
			if r.Changed != nil {
				r.Changed(gone, true)
			}
			return
		case deleted:
			return
		case i < 0:
			c := view.NewController(s)
			if len(r.roots) > 0 {
				for name, fsys := range r.roots {
					c.Grant(name, fsys)
				}
				c.Rebuild()
			}
			r.surfaces = append(r.surfaces, &Surface{S: s, C: c})
			i = len(r.surfaces) - 1
		default:
			r.surfaces[i].C.Rebuild()
		}
		if r.Changed != nil {
			r.Changed(r.surfaces[i], false)
		}
	}
	return r
}

// Grant gives the run's views a root to list, by name (view.Controller.Grant):
// those it has, and those it makes.
func (r *Run) Grant(name string, fsys fs.FS) {
	if r.roots == nil {
		r.roots = map[string]fs.FS{}
	}
	r.roots[name] = fsys
	for _, x := range r.surfaces {
		x.C.Grant(name, fsys)
		x.C.Rebuild()
	}
}

// programs are the programs behind the stories that need one, by the
// story's name: what the program around the agent's surface does, which
// no A2UI message says (granting a root, reading a file).
var programs = map[string]func(r *Run){
	"hotty/filepicker": filePickerProgram,
	"hotty/keycaps":    keycapsProgram,
	"hotty/picker":     pickerProgram,
}

// filePickerProgram is the program behind hotty/filepicker: it grants
// FixtureFS as the root "project" and, since it granted it, it (not the
// agent) reads the file a filePicked action's context names, and sends
// its text to the surface's /code, as an agent's message, logged as one.
func filePickerProgram(r *Run) {
	r.Grant("project", FixtureFS)
	r.hears = func(a *a2ui.ActionMessage) {
		file, _ := a.Context["file"].(string)
		if a.Name != "filePicked" || file == "" {
			return
		}
		text, err := fs.ReadFile(FixtureFS, file)
		if err != nil {
			return
		}
		msg, _ := json.Marshal(map[string]any{"version": "v1.0", "updateDataModel": map[string]any{
			"surfaceId": a.SurfaceID, "path": "/code", "value": string(text)}})
		r.Feed(msg)
	}
}

// Configure sets a surface's view up as the program behind the story
// has it, before each time it is drawn: hotty/keycaps's keycaps, as its
// controls say. A view of a story with no program draws as the kit does
// by default.
func (r *Run) Configure(s *Surface) {
	if r.config != nil {
		r.config(s)
	}
}

// Program sets the run up as the program behind a story, if it has one
// (programs), before its messages are fed.
func (r *Run) Program(st *Story) {
	if p := programs[st.Name]; p != nil {
		p(r)
	}
}

// Play plays a story in the run: the program behind it (Program), then
// its messages, until one fails.
func (r *Run) Play(st *Story) error {
	r.Program(st)
	for _, m := range st.Messages {
		if err := r.Feed(m); err != nil {
			return err
		}
	}
	return nil
}

// Start runs a story: a new run, playing it.
func Start(st *Story) (*Run, error) {
	r := NewRun()
	return r, r.Play(st)
}

// Surfaces are the run's surfaces, in the order they were made.
func (r *Run) Surfaces() []*Surface { return slices.Clone(r.surfaces) }

// Feed hands the processor what the agent sent: a message, a list of
// them, or a story ({"messages": [...]}). It is logged, and fails as the
// processor does; the processor has then sent the agent its error.
func (r *Run) Feed(raw json.RawMessage) error {
	raw = bytes.TrimSpace(raw)
	if bytes.HasPrefix(raw, []byte("{")) {
		var st struct {
			Messages []json.RawMessage `json:"messages"`
		}
		if json.Unmarshal(raw, &st) == nil && st.Messages != nil {
			for _, m := range st.Messages {
				if err := r.Feed(m); err != nil {
					return err
				}
			}
			return nil
		}
	}
	at := len(r.Log)
	r.Log = append(r.Log, Entry{At: time.Now(), JSON: slices.Clone(raw)})
	err := r.P.ProcessJSON(raw)
	r.Log[at].Err = err
	return err
}

// Stream feeds every JSON value a reader has, as they come (JSON lines,
// or values one after another), calling each after one is fed. A value
// the processor refuses does not stop it; one that is not JSON does.
func (r *Run) Stream(in io.Reader, each func(err error)) error {
	dec := json.NewDecoder(in)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err := r.Feed(raw)
		if each != nil {
			each(err)
		}
	}
}

// focus gives the keyboard to a component: in the caller's surface, or
// for the agent's call, the first that has it. Only one surface has the
// keyboard at a time.
func (r *Run) focus(s *a2ui.Surface, scope a2ui.Scope, id string) error {
	for _, x := range r.surfaces {
		if s != nil && x.S != s {
			continue
		}
		if el := x.C.FindComponent(id, scope); el != "" {
			for _, o := range r.surfaces {
				o.C.St.Keyboard = false
			}
			x.C.Focus(el)
			return nil
		}
	}
	return fmt.Errorf("no component '%s' to focus", id)
}

// scrollTo scrolls a HottyScrollView of the caller's surface, or for the
// agent's call, of any surface: the instance in the caller's scope.
func (r *Run) scrollTo(s *a2ui.Surface, scope a2ui.Scope, id, to string) error {
	for _, x := range r.surfaces {
		if s != nil && x.S != s {
			continue
		}
		if el := x.C.FindComponent(id, scope); el != "" {
			x.C.ScrollTo(el, to)
			return nil
		}
	}
	return fmt.Errorf("no component '%s' to scroll", id)
}

// foldAll opens or closes every branch of a HottyTree of the caller's
// surface, or for the agent's call, of any surface: the instance in the
// caller's scope.
func (r *Run) foldAll(s *a2ui.Surface, scope a2ui.Scope, id string, open bool) error {
	for _, x := range r.surfaces {
		if s != nil && x.S != s {
			continue
		}
		if el := x.C.FindComponent(id, scope); el != "" {
			return x.C.FoldAll(el, open)
		}
	}
	return fmt.Errorf("no component '%s' to fold", id)
}

// timer starts, stops, toggles or resets a HottyTimer or a HottyStopwatch
// of the caller's surface, or for the agent's call, of any surface: the
// instance in the caller's scope, shown or not (view.Controller.TimerDo).
func (r *Run) timer(s *a2ui.Surface, scope a2ui.Scope, id, do string) error {
	for _, x := range r.surfaces {
		if s != nil && x.S != s {
			continue
		}
		if key := x.C.FindTimer(id, scope); key != "" {
			return x.C.TimerDo(key, do)
		}
	}
	return fmt.Errorf("no timer '%s'", id)
}

// toast shows a toast (hottyToast): on the surface its surfaceId names,
// else the caller's, or for the agent's call the surface that has the
// keyboard, else the first. The agent's arguments come as it sent them,
// and are read in that surface's data model now, as a component's were
// read in its scope.
func (r *Run) toast(s *a2ui.Surface, args map[string]any) error {
	var to *Surface
	id, named := args["surfaceId"].(string)
	for _, x := range r.surfaces {
		switch {
		case named && x.S.ID == id, !named && x.S == s:
			to = x
		case !named && s == nil && to == nil && x.C.St.Keyboard:
			to = x
		}
	}
	if to == nil && !named && s == nil && len(r.surfaces) > 0 {
		to = r.surfaces[0]
	}
	if to == nil {
		return fmt.Errorf("no surface '%s' to show a toast on", id)
	}
	if s == nil {
		v, err := to.S.Context(a2ui.RootScope).Resolve(args)
		if err != nil {
			return err
		}
		args, _ = v.(map[string]any)
	}
	_, err := to.C.ShowToast(args)
	return err
}

// dismissToast takes a toast away (hottyDismissToast), from the caller's
// surface, or for the agent's call, from whichever has it.
func (r *Run) dismissToast(s *a2ui.Surface, id string) error {
	for _, x := range r.surfaces {
		if s == nil || x.S == s {
			x.C.DismissToast(id)
		}
	}
	return nil
}

// blur takes the keyboard from the caller's surface, or for the agent's
// call, from every surface.
func (r *Run) blur(s *a2ui.Surface) error {
	for _, x := range r.surfaces {
		if s == nil || x.S == s {
			x.C.Focus("")
		}
	}
	return nil
}

// Actions are the user's actions so far: the log's action messages.
func (r *Run) Actions() []Entry {
	var out []Entry
	for _, e := range r.Log {
		if e.Out && bytes.Contains(e.JSON, []byte(`"action":`)) {
			out = append(out, e)
		}
	}
	return out
}
