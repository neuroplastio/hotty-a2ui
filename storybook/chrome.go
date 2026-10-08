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

// newChrome makes pick, nav and panel. pick is two selects, the rendition
// and the theme, side by side on a host (native), one over the other in
// cells, whose Row does not wrap. nav is the stories in their groups: a
// List of Buttons, which the kit draws as a menu's rows. The story shown
// is the row whose Button is not borderless (showing).
func newChrome(entries []entry, rends []renditionOption, rend string, th string, native bool) *chrome {
	ch := &chrome{run: story.NewRun(), sent: map[string]string{}, entries: entries}
	ch.run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			ch.acts = append(ch.acts, o.Action)
		}
	}
	var opts []any
	for _, r := range rends {
		opts = append(opts, map[string]any{"label": r.label, "value": r.value})
	}
	var themes []any
	for _, t := range theme.All {
		themes = append(themes, map[string]any{"label": t.Name, "value": t.Name})
	}
	pick := []map[string]any{
		obj("id", "root", "component", "Column", "children", []any{"view", "rule"}),
		obj("id", "view", "component", map[bool]string{true: "Row", false: "Column"}[native], "children", []any{"rend", "theme_p"}),
		obj("id", "rend", "component", "ChoicePicker", "label", "Rendition", "value", obj("@path", "/rendition"), "options", opts, "weight", 1),
		obj("id", "theme_p", "component", "ChoicePicker", "label", "Theme", "value", obj("@path", "/theme"), "options", themes, "weight", 1),
		obj("id", "rule", "component", "Divider"),
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
		create(pickID, map[string]any{"rendition": []any{rend}, "theme": []any{th}}),
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
func (ch *chrome) theme() string {
	v := ch.surface(pickID).S.Data.Value("/theme")
	if l, ok := v.([]any); ok && len(l) > 0 {
		s, _ := l[0].(string)
		return s
	}
	return ""
}

// rendition is the rendition picked.
func (ch *chrome) rendition() string {
	v := ch.surface(pickID).S.Data.Value("/rendition")
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
