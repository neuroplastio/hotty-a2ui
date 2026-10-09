package storybook

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// The storybook's own surfaces are A2UI too, rendered by the kit like the
// stories (NEIO-11's Q6, leaning to the kit): the storybook is their
// agent. pick picks the rendition and the theme, and stays put above nav,
// which lists the stories and scrolls; panel shows the story's actions,
// data model and messages.
const (
	pickID  = "pick"
	navID   = "nav"
	panelID = "panel"
	// streamName is the stream's entry among the stories.
	streamName = "stream"
)

// entry is a story in nav: its handle, its title, and the group it is
// listed under.
type entry struct{ name, label, group string }

// groupTitles head nav's groups.
var groupTitles = map[string]string{
	"":             "Live",
	story.Basic:    "A2UI basic catalog",
	story.Hotty:    "HOTTY catalog",
	story.Fallback: "Fallbacks",
}

// chrome is the storybook as the agent of its own surfaces.
type chrome struct {
	run *story.Run
	// acts are the actions its surfaces sent, until the storybook takes
	// them: the processor is still dispatching when they come.
	acts []*a2ui.ActionMessage
	// sent are the values last set by path, so only changes go out.
	sent map[string]string
	// entries are nav's stories; cur is the one shown.
	entries []entry
	cur     string
}

type renditionOption struct{ value, label string }

// The keymaps the stories' text fields edit by (Book.keymap).
const (
	keysTerminal = "terminal" // hotty.TerminalKeys over SPEC §10.2's default
	keysDefault  = "default"  // SPEC §10.2's default alone
)

// newChrome makes pick, nav and panel. pick is three selects: on a host
// (native), the rendition and the keymap side by side, their values
// short, over the theme; in cells, whose Row does not wrap, one over the
// other. nav is the stories in their groups: a List of Buttons, which the
// kit draws as a menu's rows. The story shown is the row whose Button is
// not borderless (showing).
func newChrome(entries []entry, rends []renditionOption, rend, th, keys string, native bool) *chrome {
	ch := &chrome{run: story.NewRun(), sent: map[string]string{}, entries: entries}
	ch.run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			ch.acts = append(ch.acts, o.Action)
		}
	}
	// A select's title: on a host its label; in cells, where a title takes a
	// row of its own, its options' labels say it instead ("Theme: Nord"),
	// so that the three take a row each.
	picker := func(id, title, path string, options [][2]string) map[string]any {
		var opts []any
		for _, o := range options {
			label := o[1]
			if !native {
				label = title + ": " + label
			}
			opts = append(opts, map[string]any{"label": label, "value": o[0]})
		}
		p := obj("id", id, "component", "ChoicePicker", "value", obj("@path", path), "options", opts, "weight", 1)
		if native {
			p["label"] = title
		} else {
			p["accessibility"] = obj("label", title)
		}
		return p
	}
	var rendOpts, themeOpts [][2]string
	for _, r := range rends {
		rendOpts = append(rendOpts, [2]string{r.value, r.label})
	}
	for _, t := range theme.All {
		themeOpts = append(themeOpts, [2]string{t.Name, t.Name})
	}
	pick := []map[string]any{
		obj("id", "root", "component", "Column", "children", []any{"view"}),
		obj("id", "view", "component", "Column", "children", map[bool][]any{true: {"top", "theme_p"}, false: {"rend", "keys_p", "theme_p"}}[native]),
		obj("id", "top", "component", "Row", "children", []any{"rend", "keys_p"}),
		picker("rend", "Rendition", "/rendition", rendOpts),
		picker("theme_p", "Theme", "/theme", themeOpts),
		picker("keys_p", "Keys", "/keys", [][2]string{{keysTerminal, "Terminal"}, {keysDefault, "Default"}}),
	}
	nav := []map[string]any{obj("id", "root", "component", "Column", "children", []any{"list"})}
	var list []any
	group := "-"
	for i, e := range entries {
		if e.group != group {
			group = e.group
			id := "group_" + cmp.Or(group, "live")
			list = append(list, id)
			nav = append(nav, obj("id", id, "component", "Text", "variant", "caption", "text", groupTitles[group]))
		}
		list = append(list, itemID(i))
		nav = append(nav, ch.item(i, false), obj("id", itemID(i)+"_t", "component", "Text", "text", e.label))
	}
	nav = append(nav, obj("id", "list", "component", "List", "children", list))
	ch.feed(
		create(pickID, map[string]any{"rendition": []any{rend}, "theme": []any{th}, "keys": []any{keys}}),
		components(pickID, pick...),
		create(navID, map[string]any{}),
		components(navID, nav...),
		create(panelID, map[string]any{"head": "", "actions": "", "data": "", "messages": ""}),
		components(panelID,
			obj("id", "root", "component", "Column", "children", []any{"head", "tabs"}),
			obj("id", "head", "component", "Text", "text", obj("@path", "/head")),
			obj("id", "tabs", "component", "Tabs", "tabs", []any{
				obj("title", "Actions", "child", "actions"),
				obj("title", "Data model", "child", "data"),
				obj("title", "Messages", "child", "messages"),
			}),
			obj("id", "actions", "component", "Text", "text", obj("@path", "/actions")),
			obj("id", "data", "component", "Text", "text", obj("@path", "/data")),
			obj("id", "messages", "component", "Text", "text", obj("@path", "/messages")),
		),
	)
	return ch
}

func itemID(i int) string { return "story_" + strconv.Itoa(i) }

// item is a story's row in nav: borderless, but for the story shown.
func (ch *chrome) item(i int, shown bool) map[string]any {
	variant := "borderless"
	if shown {
		variant = "default"
	}
	return obj("id", itemID(i), "component", "Button", "variant", variant, "child", itemID(i)+"_t",
		"action", obj("event", obj("name", "open", "context", obj("name", ch.entries[i].name))))
}

func (ch *chrome) feed(msgs ...any) {
	for _, m := range msgs {
		b, _ := json.Marshal(m)
		if err := ch.run.Feed(b); err != nil {
			panic(fmt.Sprintf("the storybook's own surfaces: %v", err))
		}
	}
}

// set sets a value of a surface's data model, if it changed.
func (ch *chrome) set(surface, path string, v any) {
	b, _ := json.Marshal(v)
	key := surface + path
	if ch.sent[key] == string(b) {
		return
	}
	ch.sent[key] = string(b)
	ch.feed(obj("version", a2ui.Version, "updateDataModel", obj("surfaceId", surface, "path", path, "value", v)))
}

func (ch *chrome) surface(id string) *story.Surface {
	for _, s := range ch.run.Surfaces() {
		if s.S.ID == id {
			return s
		}
	}
	return nil
}

// theme is the theme picked.
func (ch *chrome) theme() string { return ch.picked("/theme") }

// rendition is the rendition picked.
func (ch *chrome) rendition() string { return ch.picked("/rendition") }

// keys is the keymap picked: keysTerminal or keysDefault.
func (ch *chrome) keys() string { return ch.picked("/keys") }

// picked is the value a select of pick has, at its path.
func (ch *chrome) picked(path string) string {
	v := ch.surface(pickID).S.Data.Value(path)
	if l, ok := v.([]any); ok && len(l) > 0 {
		s, _ := l[0].(string)
		return s
	}
	return ""
}

// showing marks the story shown in nav, its row's Button not borderless,
// and says what it is in panel.
func (ch *chrome) showing(name, head string) {
	if name != ch.cur {
		var rows []map[string]any
		for i, e := range ch.entries {
			if e.name == ch.cur || e.name == name {
				rows = append(rows, ch.item(i, e.name == name))
			}
		}
		ch.cur = name
		if len(rows) > 0 {
			ch.feed(components(navID, rows...))
		}
	}
	ch.set(panelID, "/head", head)
}

// report fills panel from a run: newest first, so what fits is what
// just happened.
func (ch *chrome) report(run *story.Run) {
	var acts, msgs []string
	for i := len(run.Log) - 1; i >= 0 && len(msgs) < 200; i-- {
		e := run.Log[i]
		line := string(e.JSON)
		if e.Out && bytes.Contains(e.JSON, []byte(`"action":`)) {
			acts = append(acts, line)
		}
		arrow := "→ "
		if e.Out {
			arrow = "← "
		}
		if e.Err != nil {
			line += "  ✗ " + strings.TrimPrefix(e.Err.Error(), "a2ui: ")
		}
		msgs = append(msgs, arrow+line)
	}
	var data []string
	for _, s := range run.Surfaces() {
		b, _ := json.MarshalIndent(s.S.Data.Root(), "", "  ")
		data = append(data, "**"+s.S.ID+"**\n\n"+code(string(b)))
	}
	ch.set(panelID, "/actions", orNone(code(strings.Join(acts, "\n")), len(acts), "No action yet: the user's go here, as the agent gets them."))
	ch.set(panelID, "/data", orNone(strings.Join(data, "\n\n"), len(data), "No surface."))
	ch.set(panelID, "/messages", orNone(code(strings.Join(msgs, "\n")), len(msgs), "No message."))
}

// code is a Markdown code block that no line of s can close.
func code(s string) string {
	fence := "```"
	for strings.Contains(s, fence) {
		fence += "`"
	}
	return fence + "\n" + s + "\n" + fence
}

func orNone(s string, n int, none string) string {
	if n == 0 {
		return "*" + none + "*"
	}
	return s
}

// entries are the storybook's list: its stories by group, and the stream
// first when there is one.
func entries(stream bool) []entry {
	var out []entry
	if stream {
		out = append(out, entry{streamName, "The stream", ""})
	}
	for _, st := range story.All() {
		out = append(out, entry{st.Name, st.Title, st.Group})
	}
	return out
}

func create(id string, data map[string]any) map[string]any {
	return obj("version", a2ui.Version, "createSurface", obj("surfaceId", id, "catalogId", basic.ID, "dataModel", data))
}

func components(id string, cs ...map[string]any) map[string]any {
	return obj("version", a2ui.Version, "updateComponents", obj("surfaceId", id, "components", toAny(cs)))
}

func toAny(cs []map[string]any) []any {
	out := make([]any, len(cs))
	for i, c := range cs {
		out[i] = c
	}
	return out
}

// obj is a JSON object of keys and values.
func obj(kv ...any) map[string]any {
	m := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}
