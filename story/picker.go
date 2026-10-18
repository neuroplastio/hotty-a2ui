package story

import (
	"encoding/json"
	"strings"

	"github.com/sahilm/fuzzy"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// The agent behind hotty/picker: a command palette's, which ranks its
// rows for each query the picker sends and says what each action did.

// pickerEntry is one of the palette's rows before ranking: what it is, the
// section it shows under with no query, and its actions.
type pickerEntry struct {
	value, group, title, detail, section string
	actions                              []map[string]any
}

// pickerRun, pickerPane and pickerAgent are a command's, a pane's and an
// agent's actions, as a terminal multiplexer's palette has them.
var (
	pickerRun = []map[string]any{
		{"id": "run", "title": "Run", "key": "enter"},
		{"id": "again", "title": "Run and keep open", "key": "mod+enter"},
		{"id": "copy", "title": "Copy the command", "key": "alt+enter"},
	}
	pickerPane = []map[string]any{
		{"id": "focus", "title": "Focus", "key": "enter"},
		{"id": "close", "title": "Close pane", "key": "shift+delete"},
		{"id": "copy", "title": "Copy its id", "key": "alt+enter"},
	}
)

// pickerEntries are the palette's rows with no query, in order.
var pickerEntries = []pickerEntry{
	{"pane.split", "Pane", "Split right", "", "Recently used", pickerRun},
	{"tab.new", "Tab", "New tab", "", "Recently used", pickerRun},
	{"p-1", "Pane", "fish", "~/code/plexos", "Panes", pickerPane},
	{"p-2", "Pane", "claude", "agent · plexos", "Panes", pickerPane},
	{"p-3", "Pane", "htop", "bastion", "Panes", pickerPane},
	{"pane.close", "Pane", "Close", "", "Commands", pickerRun},
	{"pane.zoom", "Pane", "Zoom", "", "Commands", pickerRun},
	{"workspace.new", "Workspace", "New workspace", "", "Commands", pickerRun},
	{"workspace.rename", "Workspace", "Rename…", "", "Commands", pickerRun},
	{"config.get", "Config", "Get a setting", "", "Commands", nil},
	{"session.quit", "Session", "Quit", "", "Commands", pickerRun},
}

// PickerRows are the palette's rows for a query: with none, every row
// under its section; else those whose "Group: Title" holds the query's
// characters in order, best first (sahilm/fuzzy, as HottyList's filter
// ranks), with no sections, the title's matched characters as match.
func PickerRows(query string) []any {
	row := func(p pickerEntry, section string, match [][2]int) map[string]any {
		m := map[string]any{"value": p.value, "title": p.title, "group": p.group}
		if p.detail != "" {
			m["detail"] = p.detail
		}
		if section != "" {
			m["section"] = section
		}
		if len(match) > 0 {
			ms := make([]any, len(match))
			for i, r := range match {
				ms[i] = []any{r[0], r[1]}
			}
			m["match"] = ms
		}
		if p.actions != nil {
			acts := make([]any, len(p.actions))
			for i, a := range p.actions {
				acts[i] = a
			}
			m["actions"] = acts
		}
		return m
	}
	out := []any{}
	if query == "" {
		for _, p := range pickerEntries {
			out = append(out, row(p, p.section, nil))
		}
		return out
	}
	targets := make([]string, len(pickerEntries))
	for i, p := range pickerEntries {
		targets[i] = p.group + ": " + p.title
	}
	for _, m := range fuzzy.Find(query, targets) {
		p := pickerEntries[m.Index]
		out = append(out, row(p, "", titleRuns(p.group, p.title, m.MatchedIndexes)))
	}
	return out
}

// titleRuns are the runs of a title's characters among a match of
// "Group: Title"'s bytes, as character ranges in the title.
func titleRuns(group, title string, at []int) [][2]int {
	start := len(group) + len(": ")
	var runs [][2]int
	for _, b := range at {
		if b < start {
			continue
		}
		c := len([]rune(title[:b-start]))
		if n := len(runs); n > 0 && runs[n-1][1] == c {
			runs[n-1][1]++
			continue
		}
		runs = append(runs, [2]int{c, c + 1})
	}
	return runs
}

// pickerProgram is the program behind hotty/picker: the palette's agent.
// A query action's context names the query, and it answers with the rows
// for it (PickerRows); an act action's, the row and the action, and it
// says in the status what it did, as an agent's message, logged as one: an
// error for an action on a pane of bastion, which doesn't answer.
func pickerProgram(r *Run) {
	r.hears = func(a *a2ui.ActionMessage) {
		switch a.Name {
		case "query":
			q, _ := a.Context["query"].(string)
			r.feedModel(a.SurfaceID, "/rows", PickerRows(q))
		case "act":
			row, _ := a.Context["row"].(string)
			act, _ := a.Context["action"].(string)
			status, kind := strings.TrimSpace(act+" "+row), "info"
			if row == "p-3" && act != "copy" {
				// htop runs on bastion, which doesn't answer: an error.
				status, kind = "bastion is not reachable", "error"
			}
			r.feedModel(a.SurfaceID, "/statusKind", kind)
			r.feedModel(a.SurfaceID, "/status", status)
		}
	}
}

// feedModel feeds the run an updateDataModel of path, as the agent's.
func (r *Run) feedModel(surface, path string, value any) {
	msg, _ := json.Marshal(map[string]any{"version": "v1.0", "updateDataModel": map[string]any{
		"surfaceId": surface, "path": path, "value": value}})
	r.Feed(msg)
}
