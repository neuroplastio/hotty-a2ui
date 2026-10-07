package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// The storybook's own surfaces are A2UI too, rendered by the kit like the
// stories (NEIO-11's Q6, leaning to the kit): the storybook is their
// agent. nav lists the stories and picks the rendition; panel shows the
// story's actions, data model and messages.
const (
	navID   = "nav"
	panelID = "panel"
	// streamName is the stream's entry among the stories.
	streamName = "stream"
)

// chrome is the storybook as the agent of its own surfaces.
type chrome struct {
	run *story.Run
	// acts are the actions its surfaces sent, until the storybook takes
	// them: the processor is still dispatching when they come.
	acts []*a2ui.ActionMessage
	// sent are the values last set by path, so only changes go out.
	sent map[string]string
}

type renditionOption struct{ value, label string }

func newChrome(entries [][2]string, rends []renditionOption, rend string) *chrome {
	ch := &chrome{run: story.NewRun(), sent: map[string]string{}}
	ch.run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			ch.acts = append(ch.acts, o.Action)
		}
	}
	var items []any
	for _, e := range entries {
		items = append(items, map[string]any{"name": e[0], "label": "  " + e[1]})
	}
	var opts []any
	for _, r := range rends {
		opts = append(opts, map[string]any{"label": r.label, "value": r.value})
	}
	ch.feed(
		create(navID, map[string]any{"rendition": []any{rend}, "stories": items}),
		components(navID,
			obj("id", "root", "component", "Column", "children", []any{"title", "rend", "list"}),
			obj("id", "title", "component", "Text", "text", "**HOTTY kit** · stories"),
			obj("id", "rend", "component", "ChoicePicker", "label", "Rendition", "displayStyle", "chips",
				"value", obj("@path", "/rendition"), "options", opts),
			obj("id", "list", "component", "List", "children", obj("componentId", "item", "path", "/stories")),
			obj("id", "item", "component", "Button", "variant", "borderless", "child", "item_t",
				"action", obj("event", obj("name", "open", "context", obj("name", obj("@path", "name"))))),
			obj("id", "item_t", "component", "Text", "text", obj("@path", "label")),
		),
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

// rendition is the rendition picked in nav.
func (ch *chrome) rendition() string {
	v := ch.surface(navID).S.Data.Value("/rendition")
	if l, ok := v.([]any); ok && len(l) > 0 {
		s, _ := l[0].(string)
		return s
	}
	return ""
}

// showing marks the story shown in nav, and says what it is in panel.
func (ch *chrome) showing(entries [][2]string, name, head string) {
	for i, e := range entries {
		mark := "  "
		if e[0] == name {
			mark = "▸ "
		}
		ch.set(navID, fmt.Sprintf("/stories/%d/label", i), mark+e[1])
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

// entries are the storybook's list: its stories, and the stream first
// when there is one.
func entries(stream bool) [][2]string {
	var out [][2]string
	if stream {
		out = append(out, [2]string{streamName, "The stream"})
	}
	for _, st := range story.All() {
		label := st.Title
		if st.Group != story.Basic {
			label = st.Group + " · " + label
		}
		out = append(out, [2]string{st.Name, label})
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
