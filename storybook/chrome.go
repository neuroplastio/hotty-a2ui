package storybook

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// The storybook's own surfaces are A2UI too, rendered by the kit like the
// stories (NEIO-11's Q6, leaning to the kit): the storybook is their
// agent. pick picks the rendition, the keymap and the theme, and filters
// the stories; it stays put above nav, the stories in a HottyTree, which
// scrolls; together they are the sidebar. panel is the right column: what
// the story shown is, over its tabs: its preview, which the Book draws
// under the tab bar, then all it says of itself, and its actions, data
// model and messages, as JSON in HottyCode.
const (
	pickID  = "pick"
	navID   = "nav"
	panelID = "panel"
	// streamName is the stream's entry among the stories.
	streamName = "stream"
	// treeID is nav's HottyTree.
	treeID = "tree"
)

// entry is a story in nav: its handle, its title, the branch it is
// listed in ("" for the stream, a root of its own), its icon, and the
// topic it is under in the branch, if any (story.Story.Topic).
type entry struct{ name, label, branch, icon, topic string }

// topicIcons are the topics' icons in nav.
var topicIcons = map[string]string{"Drag and drop": "drag_pan"}

// branch is one of nav's: its id, which is its node's value, its title
// and its icon.
type branch struct{ id, title, icon string }

// branches are nav's, in order: the kit's stories by what they show, A2UI's
// examples, and what the kit falls back to. All but the examples, which
// are many, start open, and so do the topics in them.
var branches = []branch{
	{"components", "Components", "widgets"},
	{"behaviours", "Behaviours", "touch_app"},
	{"examples", "A2UI examples", "dashboard"},
	{"fallbacks", "Fallbacks", "warning"},
}

// branchOf is the branch a story is listed in: a fallback's group says;
// else its kind, and none is A2UI's.
func branchOf(st *story.Story) string {
	switch {
	case st.Group == story.Fallback:
		return "fallbacks"
	case st.Kind == story.KindComponent:
		return "components"
	case st.Kind == story.KindBehaviour:
		return "behaviours"
	}
	return "examples"
}

// chrome is the storybook as the agent of its own surfaces.
type chrome struct {
	run *story.Run
	// acts are the actions its surfaces sent, until the storybook takes
	// them: the processor is still dispatching when they come.
	acts []*a2ui.ActionMessage
	// sent are the values last set by path, so only changes go out.
	sent map[string]string
	// cur is the story shown.
	cur string
	// logged is what report made of logRun's log so far, an entry each:
	// a log only grows.
	logRun *story.Run
	logged []logged
}

// logged is a log entry as panel shows it: the message, pretty, under a
// comment that says which way it went and what it is, and its error after
// it. An action is also shown alone.
type logged struct {
	message string
	failed  bool
	action  string
}

type renditionOption struct{ value, label string }

// The keymaps the stories' text fields edit by (Book.keymap).
const (
	keysTerminal = "terminal" // hotty.TerminalKeys over SPEC §10.2's default
	keysDefault  = "default"  // SPEC §10.2's default alone
)

// newChrome makes pick, nav and panel. pick is three selects, one over
// the other, and the filter; their options have icons on a host (native). nav is a
// HottyTree: the stream, then the branches, the stories in them; the
// story shown is the node selected, and selecting one shows it (Book.settle).
// panel heads its tabs with the story's title, the A2UI names it is
// about, and its description's first sentence (showing).
func newChrome(entries []entry, rends []renditionOption, rend, th, keys string, native bool) *chrome {
	ch := &chrome{run: story.NewRun(), sent: map[string]string{}}
	ch.run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			ch.acts = append(ch.acts, o.Action)
		}
	}
	// A select's title: on a host, a caption before it, the selects in a
	// column, as wide as each other; in cells, where a title takes a row of
	// its own, its options' labels say it instead ("Theme: Nord"), so that
	// the three take a row each.
	var pick []map[string]any
	picker := func(id, title, path string, options [][3]string) map[string]any {
		var opts []any
		icons := map[string]any{}
		for _, o := range options {
			label := o[1]
			if !native {
				label = title + ": " + label
			}
			opts = append(opts, map[string]any{"label": label, "value": o[0]})
			icons[o[0]] = o[2]
		}
		p := withIcons(obj("id", id, "component", "ChoicePicker", "value", obj("@path", path), "options", opts, "weight", 2,
			"accessibility", obj("label", title)), icons)
		if native {
			pick = append(pick, obj("id", id+"_r", "component", "Row", "children", []any{id + "_l", id}, "align", "center"),
				obj("id", id+"_l", "component", "Text", "variant", "caption", "text", title, "weight", 1))
		}
		return p
	}
	var rendOpts, themeOpts [][3]string
	for _, r := range rends {
		rendOpts = append(rendOpts, [3]string{r.value, r.label, renditionIcons[r.value]})
	}
	for _, t := range theme.All {
		themeOpts = append(themeOpts, [3]string{t.Name, t.Name, themeIcon(t)})
	}
	pick = append(pick,
		obj("id", "root", "component", "Column", "children", []any{"view"}),
		obj("id", "view", "component", "Column", "children", map[bool][]any{true: {"rend_r", "keys_p_r", "theme_p_r", "find"}, false: {"rend", "keys_p", "theme_p", "find"}}[native]),
		picker("rend", "Rendition", "/rendition", rendOpts),
		picker("theme_p", "Theme", "/theme", themeOpts),
		picker("keys_p", "Keys", "/keys", [][3]string{{keysDefault, "Default", "keyboard"}, {keysTerminal, "Terminal", "terminal"}}),
		// The filter is nav's (Book.settle hands it over): here it stays
		// in sight while nav scrolls.
		// On a host its caption lines it up with the selects; in cells its
		// icon alone says what it is, as a search box's does. Its label is a
		// screen reader's.
		obj("id", "find", "component", "Row", "children", map[bool][]any{true: {"find_l", "find_f"}, false: {"find_i", "find_f"}}[native], "align", "center"),
		obj("id", "find_l", "component", "Row", "children", []any{"find_i", "find_t"}, "align", "center", "weight", 1),
		obj("id", "find_i", "component", "Icon", "name", "search"),
		obj("id", "find_t", "component", "Text", "variant", "caption", "text", "Filter"),
		obj("id", "find_f", "component", "TextField", "label", "", "value", obj("@path", "/q"), "weight", 2, "accessibility", obj("label", "Filter the stories")),
	)
	var open []any
	for _, br := range branches {
		if br.id != "examples" {
			open = append(open, br.id)
		}
	}
	for _, e := range entries {
		if t := e.branch + "/" + e.topic; e.topic != "" && e.branch != "examples" && !slices.Contains(open, any(t)) {
			open = append(open, t)
		}
	}
	ch.feed(
		create(pickID, map[string]any{"rendition": []any{rend}, "theme": []any{th}, "keys": []any{keys}, "q": ""}),
		components(pickID, pick...),
		create(navID, map[string]any{"cur": "", "open": open, "q": ""}),
		components(navID,
			obj("id", "root", "component", "Column", "children", []any{treeID}),
			obj("id", treeID, "component", "HottyTree", "catalogId", hottycat.ID, "items", navItems(entries),
				"selected", obj("@path", "/cur"), "expanded", obj("@path", "/open"), "filter", obj("@path", "/q"),
				"emptyText", "No story matches.", "accessibility", obj("label", "Stories"),
				"onActivate", obj("event", obj("name", "enter"))),
		),
		create(panelID, map[string]any{"title": "", "names": "", "summary": "", "about": "", "actions": "", "data": "", "messages": ""}),
		components(panelID,
			obj("id", "root", "component", "Column", "children", []any{"header", tabsID}),
			obj("id", "header", "component", "Column", "children", []any{"heading", "summary"}),
			// The story's icon is a host's, as nav's are.
			obj("id", "heading", "component", "Row", "children", map[bool][]any{true: {"icon", "title", "names"}, false: {"title", "names"}}[native], "align", "center"),
			storyIcon(""),
			obj("id", "title", "component", "Text", "text", obj("@path", "/title")),
			obj("id", "names", "component", "Text", "variant", "caption", "text", obj("@path", "/names")),
			obj("id", "summary", "component", "Text", "variant", "caption", "text", obj("@path", "/summary")),
			withIcons(obj("id", tabsID, "component", "Tabs", "tabs", []any{
				obj("title", "Preview", "child", "preview"),
				obj("title", "About", "child", "about"),
				obj("title", "Actions", "child", "actions"),
				obj("title", "Data model", "child", "data"),
				obj("title", "Messages", "child", "messages"),
			}), []any{"visibility", "info", "send", "data_object", "forum"}),
			// The preview is the Book's: the tab holds nothing, and the
			// story's surfaces go under its bar.
			obj("id", "preview", "component", "Column", "children", []any{}),
			obj("id", "about", "component", "Text", "text", obj("@path", "/about")),
			jsonCode("actions", nil),
			jsonCode("data", nil),
			jsonCode("messages", nil),
		),
	)
	return ch
}

// renditionIcons are the renditions' icons in pick.
var renditionIcons = map[string]string{rendSurfaces: "web", rendCells: "terminal", rendText: "notes", rendSide: "vertical_split"}

// themeIcon is a theme's icon in pick: light or dark, as its background
// is; the terminal's own colours, which the kit does not know, are
// neither.
func themeIcon(t theme.Theme) string {
	var r, g, b int
	if _, err := fmt.Sscanf(t.Bg, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return "contrast"
	}
	if 299*r+587*g+114*b > 128*1000 {
		return "light_mode"
	}
	return "dark_mode"
}

// withIcons gives a Tabs' titles or a ChoicePicker's options their icons
// (the kit's io_neuroplast_hotty.icons): a name a tab, in order, or a name
// an option's value.
func withIcons(c map[string]any, icons any) map[string]any {
	c["metadata"] = obj("extensions", obj("io_neuroplast_hotty", obj("icons", icons)))
	return c
}

// storyIcon is panel's icon of the story shown: a HottyIcon, whose name
// is no binding, so it goes again with each story.
func storyIcon(name string) map[string]any {
	return obj("id", "icon", "component", "HottyIcon", "catalogId", hottycat.ID, "name", cmp.Or(name, "article"))
}

// navItems are nav's nodes: the stream, then a node a branch with its
// stories in it, a value each: its handle, or the branch's id. The
// components and the behaviours go by their titles; A2UI's examples and
// the fallbacks as they are numbered. A topic's stories are a node in
// their branch, where the topic's name goes, its value the branch's id
// and the topic's.
func navItems(entries []entry) []any {
	var items []any
	in := map[string][]entry{}
	for _, e := range entries {
		if e.branch == "" {
			items = append(items, node(e))
			continue
		}
		in[e.branch] = append(in[e.branch], e)
	}
	for _, br := range branches {
		var kids []any
		topics := map[string]map[string]any{}
		for _, e := range in[br.id] {
			if e.topic == "" {
				kids = append(kids, node(e))
				continue
			}
			t, ok := topics[e.topic]
			if !ok {
				t = obj("label", e.topic, "value", br.id+"/"+e.topic, "icon", cmp.Or(topicIcons[e.topic], br.icon), "children", []any{})
				topics[e.topic] = t
				kids = append(kids, t)
			}
			t["children"] = append(t["children"].([]any), node(e))
		}
		if len(kids) > 0 {
			items = append(items, obj("label", br.title, "value", br.id, "icon", br.icon, "children", kids))
		}
	}
	return items
}

func node(e entry) map[string]any {
	n := obj("label", e.label, "value", e.name)
	if e.icon != "" {
		n["icon"] = e.icon
	}
	return n
}

// tabsID is panel's Tabs; its first tab, the preview, is 0.
const tabsID = "tabs"

// panelTabs is how many tabs panel has.
const panelTabs = 5

// tab is the tab panel shows: 0 for the preview.
func (ch *chrome) tab() int { return ch.surface(panelID).C.St.Tabs[tabsID] }

// setTab shows panel's tab i, as a click on its title would.
func (ch *chrome) setTab(i int) {
	c := ch.surface(panelID).C
	c.St.Tabs[tabsID] = (i + panelTabs) % panelTabs
	c.Rebuild()
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

// selected is nav's selected node: a story's handle, or a branch's id.
func (ch *chrome) selected() string {
	s, _ := ch.surface(navID).S.Data.Value("/cur").(string)
	return s
}

// filter hands what is typed in pick's filter to nav's tree.
func (ch *chrome) filter() {
	q, _ := ch.surface(pickID).S.Data.Value("/q").(string)
	ch.set(navID, "/q", q)
}

// about is what panel says of a story: its title and icon, the A2UI names
// it is about, and its description.
type about struct{ title, icon, names, description string }

// showing selects the story shown in nav and says in panel what it is:
// its header the first sentence of its description, and About all of it.
func (ch *chrome) showing(name string, a about) {
	if name != ch.cur {
		ch.cur = name
		ch.feed(components(panelID, storyIcon(a.icon)))
	}
	// The user moves the selection too, so it goes whatever was sent last.
	delete(ch.sent, navID+"/cur")
	ch.set(navID, "/cur", name)
	ch.set(panelID, "/title", "**"+a.title+"**")
	ch.set(panelID, "/names", a.names)
	ch.set(panelID, "/summary", firstSentence(a.description))
	head := "**" + a.title + "**"
	if a.names != "" {
		head += " · " + a.names
	}
	ch.set(panelID, "/about", head+"\n\n"+a.description)
}

// firstSentence is a description's first sentence: to its first full stop
// or colon outside brackets, which a full stop ends.
func firstSentence(s string) string {
	depth := 0
	for i, r := range s {
		switch r {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case '.', ':':
			if depth == 0 && (i+1 == len(s) || s[i+1] == ' ') {
				return s[:i] + "."
			}
		}
	}
	return s
}

// report fills panel from a run: newest first, so what fits is what
// just happened.
func (ch *chrome) report(run *story.Run) {
	if run != ch.logRun || len(run.Log) < len(ch.logged) {
		ch.logRun, ch.logged = run, nil
	}
	for _, e := range run.Log[len(ch.logged):] {
		ch.logged = append(ch.logged, logEntry(e))
	}
	var acts, msgs []string
	var marks []any
	line := 1
	for i := len(ch.logged) - 1; i >= 0 && len(msgs) < 200; i-- {
		l := ch.logged[i]
		if l.action != "" {
			acts = append(acts, l.action)
		}
		n := strings.Count(l.message, "\n") + 1
		if l.failed {
			// The message and its error, under the comment.
			marks = append(marks, obj("line", line+1, "end", line+n-1, "kind", "error"))
		}
		msgs = append(msgs, l.message)
		line += n
	}
	var data []string
	for _, s := range run.Surfaces() {
		b, _ := json.Marshal(s.S.Data.Root())
		data = append(data, "// "+s.S.ID+"\n"+pretty(b))
	}
	ch.set(panelID, "/actions", orNone(strings.Join(acts, "\n\n"), len(acts), "No action yet: the user's go here, as the agent gets them."))
	ch.set(panelID, "/data", orNone(strings.Join(data, "\n\n"), len(data), "No surface."))
	ch.set(panelID, "/messages", orNone(strings.Join(msgs, "\n"), len(msgs), "No message."))
	// Marks are the component's, not its data's: it goes again when they
	// change.
	b, _ := json.Marshal(marks)
	if key := panelID + "#marks"; ch.sent[key] != string(b) {
		ch.sent[key] = string(b)
		ch.feed(components(panelID, jsonCode("messages", marks)))
	}
}

// logEntry is how panel shows a log entry.
func logEntry(e story.Entry) logged {
	var kind string
	var top map[string]json.RawMessage
	if json.Unmarshal(e.JSON, &top) == nil {
		for k := range top {
			if k != "version" {
				kind = k
			}
		}
	}
	arrow := "→"
	if e.Out {
		arrow = "←"
	}
	l := logged{message: "// " + arrow + " " + kind + "\n" + pretty(e.JSON)}
	if e.Err != nil {
		l.failed = true
		l.message += "\n// ✗ " + oneLine(strings.TrimPrefix(e.Err.Error(), "a2ui: "))
	}
	if e.Out && top["action"] != nil {
		l.action = pretty(e.JSON)
	}
	return l
}

// oneLine keeps a comment on its line.
func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// jsonCode is panel's tab id: its data at /id, as JSON, with marks. It
// is the hotty catalog's in panel's basic one, as a story mixes them.
func jsonCode(id string, marks []any) map[string]any {
	c := obj("id", id, "component", "HottyCode", "catalogId", hottycat.ID, "code", obj("@path", "/"+id), "language", "json")
	if len(marks) > 0 {
		c["marks"] = marks
	}
	return c
}

// orNone is s, or a comment that says there is nothing yet.
func orNone(s string, n int, none string) string {
	if n == 0 {
		return "// " + none
	}
	return s
}

// entries are the storybook's stories as nav lists them: the stream first
// when there is one, then each branch's, the components and the behaviours
// by their titles, a topic's stories together where its name goes. A2UI's examples' titles are in title case, and go in
// sentence case, as the kit's are.
func entries(stream bool) []entry {
	var out []entry
	if stream {
		out = append(out, entry{streamName, "The stream", "", "stream", ""})
	}
	in := map[string][]entry{}
	for _, st := range story.All() {
		e := entry{st.Name, st.Title, branchOf(st), st.Icon, st.Topic}
		if e.branch == "examples" {
			e.label = sentenceCase(e.label)
		}
		in[e.branch] = append(in[e.branch], e)
	}
	for _, br := range branches {
		es := in[br.id]
		if br.id == "components" || br.id == "behaviours" {
			slices.SortStableFunc(es, func(a, b entry) int {
				return cmp.Or(cmp.Compare(cmp.Or(a.topic, a.label), cmp.Or(b.topic, b.label)), cmp.Compare(a.label, b.label))
			})
		}
		out = append(out, es...)
	}
	return out
}

// sentenceCase lowers the capitals title case gives words after the first:
// a word that is a capital and lower case letters ("Login Form" is "Login
// form"); a name ("ChildList") or an acronym stays as it is.
func sentenceCase(s string) string {
	words := strings.Split(s, " ")
	for i, w := range words[1:] {
		rs := []rune(w)
		if len(rs) > 1 && unicode.IsUpper(rs[0]) && strings.ToLower(w[len(string(rs[0])):]) == w[len(string(rs[0])):] {
			words[i+1] = strings.ToLower(w)
		}
	}
	return strings.Join(words, " ")
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
