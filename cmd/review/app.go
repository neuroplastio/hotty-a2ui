package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// The surface, by its A2UI id and its name on the host.
const surfaceID = "review"

const basicCatalog = "https://a2ui.org/specification/v1_0/catalogs/basic/catalog.json"

// shot is a picture to review.
type shot struct {
	// name is what the list and the feedback call it: its path from the
	// directory the feedback is in.
	name, path, mime string
	// id is its resource's on the host, for its cid: URL.
	id string
}

// app is review on the whole screen. It is the A2UI agent and the
// renderer in one: it sends the surface's messages to the run, the kit
// renders them as a HOTTY surface, and what the user does comes back as
// the run's actions and data model.
type app struct {
	s     *hottytea.Session
	run   *story.Run
	sf    *story.Surface
	r     *html.Rendition
	shots []shot
	notes *notes
	// brief is the file that says what to look for in each picture, read
	// again whenever one shows, so that it can be written as the review
	// goes.
	brief string
	// cur is the picture shown.
	cur int
	// sent is the pictures the host has.
	sent  map[string]bool
	acts  []*a2ui.ActionMessage
	w, h  int
	frame string
}

func newApp(shots []shot, n *notes, brief string, th theme.Theme) (*app, error) {
	a := &app{s: hottytea.New(), run: story.NewRun(), shots: shots, notes: n, brief: brief, sent: map[string]bool{}}
	a.run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			a.acts = append(a.acts, o.Action)
		}
	}
	if err := a.feed(a.surface()...); err != nil {
		return nil, err
	}
	a.sf = a.run.Surfaces()[0]
	a.r = html.New(a.sf.C, surfaceID)
	a.r.SetTheme(th)
	a.sf.C.Focus(a.sf.C.FindComponent("files", a2ui.Scope{Path: "/"}))
	return a, nil
}

// surface is the messages that make the surface: the picture's name, what
// to look for in it, the list of pictures, the feedback box and Save down
// the side; the picture beside them.
func (a *app) surface() []map[string]any {
	sh := a.shots[0]
	data := map[string]any{
		"files": a.files(), "cur": sh.name, "img": "cid:" + sh.id,
		"title": a.title(0), "brief": a.lookFor(sh.name), "note": a.notes.text[sh.name], "status": "",
	}
	bind := func(path string) map[string]any { return map[string]any{"@path": path} }
	cat := hottycat.ID
	return []map[string]any{
		msg("createSurface", obj("surfaceId", surfaceID, "catalogId", basicCatalog, "dataModel", data)),
		msg("updateComponents", obj("surfaceId", surfaceID, "components", []any{
			obj("id", "root", "component", "Row", "children", []any{"side", "shot"}),
			obj("id", "side", "component", "Column", "children", []any{"title", "brief", "files", "note", "act", "help", "save_key"}, "weight", 1),
			obj("id", "files", "component", "HottyList", "catalogId", cat, "title", "Pictures",
				"items", bind("/files"), "selected", bind("/cur"), "filterable", true, "height", 8,
				"onActivate", obj("functionCall", obj("@call", "hottyFocus", "catalogId", cat, "args", obj("id", "note")))),
			obj("id", "note", "component", "TextField", "label", "Feedback", "value", bind("/note"), "variant", "longText"),
			obj("id", "act", "component", "Row", "children", []any{"save", "status"}, "align", "center"),
			obj("id", "save", "component", "Button", "child", "save_t", "variant", "primary",
				"action", obj("event", obj("name", "save", "context", obj("file", bind("/cur"), "note", bind("/note"))))),
			obj("id", "save_t", "component", "Text", "text", "Save"),
			obj("id", "status", "component", "Text", "text", bind("/status"), "variant", "caption", "weight", 1),
			obj("id", "help", "component", "HottyKeyHints", "catalogId", cat),
			obj("id", "save_key", "component", "HottyShortcut", "catalogId", cat, "key", "Control+s", "press", "save", "label", "Save"),
			obj("id", "title", "component", "Text", "text", bind("/title")),
			obj("id", "brief", "component", "Text", "text", bind("/brief")),
			// Weighted in a Row: as wide as its share, as tall as the picture
			// is at that width, no taller than the surface (profile §2).
			obj("id", "shot", "component", "Image", "url", bind("/img"), "description", bind("/cur"), "fit", "contain", "weight", 3),
		})),
	}
}

// files is the list's items: each picture, under it its note's first
// line.
func (a *app) files() []any {
	var items []any
	for _, sh := range a.shots {
		item := obj("label", sh.name, "value", sh.name)
		if t := a.notes.text[sh.name]; t != "" {
			item["description"] = "✎ " + first(t)
		}
		items = append(items, item)
	}
	return items
}

// lookFor is what to look for in a picture, from the brief: its section,
// under a heading; nothing when it has none.
func (a *app) lookFor(name string) string {
	b, err := loadNotes(a.brief)
	if err != nil || b.text[name] == "" {
		return ""
	}
	return "**What to look for**\n\n" + b.text[name]
}

// title heads the picture: its name, and where it is in the list.
func (a *app) title(i int) string {
	return fmt.Sprintf("#### %s · %d of %d", escape(a.shots[i].name), i+1, len(a.shots))
}

// escape keeps a name's characters from being Markdown's: Text is
// Markdown.
func escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(`\`+"`"+`*_[]#<>!|~`, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (a *app) Init() tea.Cmd { return a.s.Detect() }

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	msg, cmd := a.s.Update(msg) // the Session's first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
	case hottytea.ReadyMsg:
		a.r.SetSteps(msg.Caps.Steps)
	case hottytea.EventMsg:
		if msg.Surface == surfaceID {
			a.fail(a.r.Event(msg.Event))
		}
	case hottytea.ErrorMsg:
		a.fail(fmt.Errorf("the host: %s %s", msg.Code, msg.Detail))
	case tea.KeyPressMsg:
		k := hottytea.KeyName(msg.Key())
		if k == "Control+c" || k == "q" && a.s.Mode != hottytea.Native {
			return a, a.quit()
		}
		if k == "" || a.s.Mode != hottytea.Native {
			break
		}
		// On a host, the keys the surface left; the kit's run its
		// shortcuts. Escape it leaves gives the keyboard back, Tab takes it
		// again, and q quits while nothing has it.
		cmds, ok, err := a.r.Key(k)
		a.s.Send(cmds...)
		a.fail(err)
		if !ok {
			switch k {
			case "Tab", "Shift+Tab":
				a.sf.C.FocusNext(k == "Shift+Tab")
			case "Escape":
				a.sf.C.Focus("")
			case "q":
				if !a.sf.C.St.Keyboard {
					return a, a.quit()
				}
			}
		}
	case endSignal:
		return a, a.quit()
	}
	a.settle()
	return a, tea.Batch(cmd, a.draw())
}

// settle answers what the user did: Save's action, and another picture
// picked in the list.
func (a *app) settle() {
	acts := a.acts
	a.acts = nil
	for _, act := range acts {
		if act.Name == "save" {
			file, _ := act.Context["file"].(string)
			note, _ := act.Context["note"].(string)
			a.save(file, note, true)
		}
	}
	name, _ := a.sf.S.Data.Value("/cur").(string)
	for i, sh := range a.shots {
		if sh.name == name && i != a.cur {
			a.show(i)
			break
		}
	}
}

// show shows another picture, its note in the box; what was typed about
// the one before is saved.
func (a *app) show(i int) {
	a.keep()
	a.cur = i
	sh := a.shots[i]
	a.fail(a.feed(
		data("/img", "cid:"+sh.id),
		data("/title", a.title(i)),
		data("/brief", a.lookFor(sh.name)),
		data("/note", a.notes.text[sh.name]),
	))
}

// keep saves the note in the box on the picture shown, if it changed.
func (a *app) keep() {
	note, _ := a.sf.S.Data.Value("/note").(string)
	a.save(a.shots[a.cur].name, note, false)
}

// save writes a picture's note to the file: when it changed, or when the
// user asked.
func (a *app) save(name, note string, asked bool) {
	if strings.TrimSpace(note) == a.notes.text[name] && !asked {
		return
	}
	a.notes.set(name, note)
	status := "Saved in " + filepath.Base(a.notes.path)
	if err := a.notes.write(); err != nil {
		status = "✗ " + err.Error()
	}
	a.fail(a.feed(data("/files", a.files()), data("/status", status)))
}

func (a *app) quit() tea.Cmd {
	a.keep()
	return tea.Sequence(a.s.Close(), tea.Quit)
}

// fail shows an error in the status line.
func (a *app) fail(err error) {
	if err != nil {
		_ = a.feed(data("/status", "✗ "+strings.TrimPrefix(err.Error(), "a2ui: ")))
	}
}

// draw lays the surface out over the whole screen, after the picture it
// shows, if the host does not have it yet.
func (a *app) draw() tea.Cmd {
	switch {
	case a.s.Mode == hottytea.Detecting || a.w == 0:
		a.frame = "Finding out whether the terminal is a HOTTY host…"
		return nil
	case a.s.Mode == hottytea.Text:
		a.frame = "review shows pictures on a HOTTY host, such as hottyterm, and this terminal is not one.\n\nq quits."
		return nil
	}
	// A space, not nothing: Bubble Tea erases the screen (ED 2) for an
	// empty frame, every time it draws one, and the Session places the
	// surface again after an erase, which draws another.
	a.frame = " "
	if sh := a.shots[a.cur]; !a.sent[sh.id] {
		b, err := os.ReadFile(sh.path)
		if err != nil {
			a.fail(err)
		} else {
			a.s.Send(hotty.Res(sh.id, sh.mime, b))
			a.sent[sh.id] = true
		}
	}
	a.s.Send(a.r.Update()...)
	a.s.Layout([]hottytea.Surface{{
		Name: surfaceID, Rect: hottytea.Rect{W: a.w, H: a.h},
		Doc: a.r.Doc, Scroll: hotty.ScrollVertical,
	}})
	return a.s.Flush()
}

func (a *app) View() tea.View {
	v := tea.NewView(a.frame)
	v.AltScreen = true
	v.WindowTitle = "review · " + a.shots[a.cur].name
	return v
}

// feed hands the run the agent's messages.
func (a *app) feed(msgs ...map[string]any) error {
	for _, m := range msgs {
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if err := a.run.Feed(b); err != nil {
			return err
		}
	}
	return nil
}

func msg(kind string, body map[string]any) map[string]any {
	return map[string]any{"version": "v1.0", kind: body}
}

// data is an updateDataModel message for the surface.
func data(path string, v any) map[string]any {
	return msg("updateDataModel", obj("surfaceId", surfaceID, "path", path, "value", v))
}

func obj(kv ...any) map[string]any {
	m := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}
