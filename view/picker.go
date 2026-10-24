package view

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A HottyPicker (profile §6.32) in the view: a query the user types, the
// rows the agent ranked for it, in its order, the selected one, the lines
// the rows and their headings take, and the actions panel. The picker
// never filters or ranks: a change to the query runs onQuery, and the
// agent answers with rows.

// PickRow is one of a HottyPicker's rows, as the agent wrote it: what
// identifies it, its title, the group before it and the detail after it,
// its icon (a HottyIcon's name, or a character of its own, such as an
// emoji), the keys that run it where the program binds them (a sequence,
// key names apart by spaces), the description on a line under it, the
// heading it falls under, where the query matched its title and its group
// (character ranges, from and to), the note the footer shows while it is
// selected, its actions, the first Enter's, and what Tab fills the query
// with ("" for its title).
type PickRow struct {
	Value       string       `json:"value"`
	Title       string       `json:"title"`
	Complete    string       `json:"complete,omitempty"`
	Group       string       `json:"group,omitempty"`
	Detail      string       `json:"detail,omitempty"`
	Icon        string       `json:"icon,omitempty"`
	Keys        string       `json:"keys,omitempty"`
	Description string       `json:"description,omitempty"`
	Section     string       `json:"section,omitempty"`
	Match       [][2]int     `json:"match,omitempty"`
	GroupMatch  [][2]int     `json:"groupMatch,omitempty"`
	Footer      string       `json:"footer,omitempty"`
	Actions     []PickAction `json:"actions,omitempty"`
}

// PickIconNamed reports whether a row's icon is an icon's name, letters,
// digits and _ (Material's "terminal", the basic Icon's "settings"), which
// is drawn as an Icon draws it, rather than a character of the agent's
// own, such as an emoji, which is drawn as it is.
func PickIconNamed(icon string) bool {
	for _, c := range icon {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return icon != ""
}

// PickAction is one of a row's actions: its id, what it does, and the key
// that runs it ("" for none but Enter, on a row's first).
type PickAction struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Key   string `json:"key,omitempty"`
}

// Pick is a HottyPicker as the user worked it, by its id: the row the
// user selected, where selected is not bound (Moved: since the query last
// changed); the actions panel, open, and the action it highlights; the
// query as the user last typed it (Typed), and how many times another
// query was written since the picker was made (Written: the agent's, a
// completion or a staged command, which a host's focused field would not
// show, SPEC §6.2).
type Pick struct {
	Selected string
	Moved    bool
	Panel    bool
	At       int
	Typed    string
	Written  int
}

// PickLine is a line of a HottyPicker's rows: a heading (Row < 0), or a
// row, by its index, its title's line or its description's (Desc).
type PickLine struct {
	Row     int
	Heading string
	Desc    bool
}

// pickHeight is the lines a HottyPicker's rows take without a height.
const pickHeight = 10

// mapPicker makes a HottyPicker's element: its query (its Value), its
// rows, the selected one (the first where selected names none), its
// panel, and the lines that show, the selected row's among them.
func mapPicker(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Picker, Value: a2ui.ToString(b.Value(n, "query")), Placeholder: b.String(n, "placeholder"),
		Chip: b.String(n, "chip"), Status: b.String(n, "status"), StatusError: b.String(n, "statusKind") == "error",
		Busy: b.Bool(n, "loading"), PanelKey: b.String(n, "panelKey"), NoRows: b.String(n, "emptyText"), Selected: -1,
		KeepKeys: b.String(n, "keep") == "keys"}
	if _, set := n.Props["emptyText"]; !set {
		e.NoRows = "No matches."
	}
	for _, it := range b.List(n, "rows") {
		m, _ := it.(map[string]any)
		e.Rows = append(e.Rows, pickRow(m))
	}
	st := b.St.Pick[n.Key]
	if q, _ := e.Value.(string); q != st.Typed {
		st.Typed, st.Written = q, st.Written+1
		b.St.Pick[n.Key] = st
	}
	e.Written = st.Written
	sel := b.String(n, "selected")
	if bd, ok := n.Props["selected"].(a2ui.Bound); (!ok || !bd.Writable()) && st.Moved {
		sel = st.Selected
	}
	for i, r := range e.Rows {
		if r.Value == sel {
			e.Selected = i
			break
		}
	}
	if e.Selected < 0 && len(e.Rows) > 0 {
		e.Selected = 0
	}
	if r := e.SelectedPick(); st.Panel && r != nil && len(r.Actions) > 0 {
		e.Panel, e.PanelAt = true, min(max(st.At, 0), len(r.Actions)-1)
	}
	e.Height = pickHeight
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
	}
	// The lines follow the selection when it moves, or the rows move
	// under it, and stay where the wheel left them otherwise.
	lines, shown := e.PickLines(), ""
	if r := e.SelectedPick(); r != nil {
		shown = strconv.Itoa(slices.Index(lines, PickLine{Row: e.Selected})) + " " + r.Value
	}
	reveal := b.St.Revealed[n.Key] != shown
	b.St.Revealed[n.Key] = shown
	e.Top = pickTop(lines, e.Selected, b.St.Scroll[n.Key], e.Height, reveal)
	b.St.Scroll[n.Key] = e.Top
	return e
}

// pickRow reads a row as the agent wrote it.
func pickRow(m map[string]any) PickRow {
	r := PickRow{Value: a2ui.ToString(m["value"]), Title: a2ui.ToString(m["title"]), Complete: a2ui.ToString(m["complete"]), Group: a2ui.ToString(m["group"]),
		Detail: a2ui.ToString(m["detail"]), Icon: a2ui.ToString(m["icon"]), Keys: a2ui.ToString(m["keys"]), Description: a2ui.ToString(m["description"]),
		Section: a2ui.ToString(m["section"]), Match: pickRanges(m["match"]), GroupMatch: pickRanges(m["groupMatch"]), Footer: a2ui.ToString(m["footer"])}
	as, _ := m["actions"].([]any)
	for _, x := range as {
		a, _ := x.(map[string]any)
		r.Actions = append(r.Actions, PickAction{ID: a2ui.ToString(a["id"]), Title: a2ui.ToString(a["title"]), Key: a2ui.ToString(a["key"])})
	}
	return r
}

// pickRanges reads a row's [from, to) character ranges.
func pickRanges(v any) [][2]int {
	ms, _ := v.([]any)
	var out [][2]int
	for _, x := range ms {
		p, _ := x.([]any)
		if len(p) == 2 {
			out = append(out, [2]int{int(a2ui.ToNumber(p[0])), int(a2ui.ToNumber(p[1]))})
		}
	}
	return out
}

// pickTop is the first line that shows: where it was (where the wheel
// left it, Controller.ScrollRows), or, when the selection moved (reveal),
// moved as little as brings the selected row into view, its description
// too, with its heading when that is the line above it. It is never a
// description's line, whose row's title would be above it.
func pickTop(lines []PickLine, sel, top, height int, reveal bool) int {
	for at, l := range lines {
		if !reveal || l.Row != sel || sel < 0 || l.Desc {
			continue
		}
		first, last := at, at
		if at > 0 && lines[at-1].Row < 0 {
			first = at - 1
		}
		if at+1 < len(lines) && lines[at+1].Desc {
			last = at + 1
		}
		top = max(min(top, first), last+1-height)
		break
	}
	top = max(min(top, len(lines)-height), 0)
	if top < len(lines) && lines[top].Desc {
		if reveal && lines[top].Row == sel || !reveal && top+height >= len(lines) {
			return top - 1
		}
		return top + 1
	}
	return top
}

// PickLines are a HottyPicker's lines: each row, after a heading wherever
// its section differs from the row before's, and its description on the
// line under it.
func (e *Element) PickLines() []PickLine {
	lines := make([]PickLine, 0, len(e.Rows))
	prev := ""
	for i, r := range e.Rows {
		if r.Section != "" && r.Section != prev {
			lines = append(lines, PickLine{Row: -1, Heading: r.Section})
		}
		prev = r.Section
		lines = append(lines, PickLine{Row: i})
		if r.Description != "" {
			lines = append(lines, PickLine{Row: i, Desc: true})
		}
	}
	return lines
}

// pickPage is how many rows Page Up and Page Down move a HottyPicker's
// selection by: its height, less the descriptions' lines from Top, one at
// least.
func (e *Element) pickPage() int {
	n := e.Height
	for k, lines := e.Top, e.PickLines(); k < len(lines) && k < e.Top+e.Height; k++ {
		if lines[k].Desc {
			n--
		}
	}
	return max(n, 1)
}

// SelectedPick is a HottyPicker's selected row, nil for none.
func (e *Element) SelectedPick() *PickRow {
	if e.Selected < 0 || e.Selected >= len(e.Rows) {
		return nil
	}
	return &e.Rows[e.Selected]
}

// PickFoot is what a HottyPicker's footer shows (Element.PickFooter): the
// selected row's note, at its left; what the keys do, each a key and its
// title, the most significant first, for a footer that draws them from its
// right, as Zed's does; and the panel's key (Key "" for none), left of
// them. Where they don't all fit, the footer keeps Keys[0], Enter's, then
// the panel's, then the note, then the other keys in order: the last goes
// first.
type PickFoot struct {
	Note  string
	Keys  []Hint
	Panel Hint
}

// PickFooter is what a HottyPicker's footer shows: the selected row's
// footer note; in the actions panel, Enter's run and Escape's back; else
// the row's actions, Enter's first, then each of the others that has a
// key, and Actions on the panel key, where it lists them. Titles are as
// written.
func (e *Element) PickFooter() PickFoot {
	r := e.SelectedPick()
	if r == nil {
		return PickFoot{}
	}
	f := PickFoot{Note: r.Footer}
	if e.Panel {
		f.Keys = []Hint{{Key: "enter", Desc: "Run"}, {Key: "escape", Desc: "Back"}}
		return f
	}
	if len(r.Actions) == 0 {
		return f
	}
	f.Keys = []Hint{{Key: "enter", Desc: r.Actions[0].Title}}
	for _, a := range r.Actions[1:] {
		if a.Key != "" {
			f.Keys = append(f.Keys, Hint{Key: a.Key, Desc: a.Title})
		}
	}
	if e.PanelKey != "" {
		f.Panel = Hint{Key: e.PanelKey, Desc: "Actions"}
	}
	return f
}

// PickerKey works a HottyPicker by a key, named as SPEC §10.4 has it. The
// arrows up and down, Control+p and Control+n, move the selection a row,
// round from either end, and Page Up and Page Down by the lines it shows,
// a row's description counted with it. Tab completes the query to the
// selected row (its complete, else its title), as a query the agent wrote.
// Enter runs the selected row's first action, and a key one of its actions
// names runs that one. panelKey opens the selected row's actions, where
// the arrows move, Enter runs the one highlighted, and Escape or the key
// again closes them; another key closes them and is taken as if they were
// shut. ok reports whether the picker took the key: one it doesn't edits
// the query.
func (c *Controller) PickerKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != Picker {
		return false, nil
	}
	if e.Panel {
		acts := e.SelectedPick().Actions
		switch {
		case key == "arrow_up" || key == "ctrl+p":
			c.pickPanel(e, true, (e.PanelAt+len(acts)-1)%len(acts))
			return true, nil
		case key == "arrow_down" || key == "ctrl+n":
			c.pickPanel(e, true, (e.PanelAt+1)%len(acts))
			return true, nil
		case key == "enter":
			return true, c.pickRun(e, acts[e.PanelAt].ID)
		case key == "escape" || sameKey(c.Platform, e.PanelKey, key):
			c.pickPanel(e, false, 0)
			return true, nil
		}
		c.pickPanel(e, false, 0)
		e = c.V.Find(id)
	}
	n, page := len(e.Rows), e.pickPage()
	switch {
	case key == "arrow_up" || key == "ctrl+p":
		return true, c.pickMove(e, (e.Selected+n-1)%max(n, 1))
	case key == "arrow_down" || key == "ctrl+n":
		return true, c.pickMove(e, (e.Selected+1)%max(n, 1))
	case key == "page_up":
		return true, c.pickMove(e, max(e.Selected-page, 0))
	case key == "page_down":
		return true, c.pickMove(e, min(e.Selected+page, n-1))
	case key == "tab":
		if r := e.SelectedPick(); r != nil {
			return true, c.writePickQuery(e, cmp.Or(r.Complete, r.Title), false)
		}
		return true, nil
	case e.PanelKey != "" && sameKey(c.Platform, e.PanelKey, key):
		if r := e.SelectedPick(); r != nil && len(r.Actions) > 0 {
			c.pickPanel(e, true, 0)
		}
		return true, nil
	}
	r := e.SelectedPick()
	if key == "enter" {
		if r == nil {
			return true, nil
		}
		return true, c.pickRun(e, firstAction(r))
	}
	if r != nil {
		for _, a := range r.Actions {
			if a.Key != "" && sameKey(c.Platform, a.Key, key) {
				return true, c.pickRun(e, a.ID)
			}
		}
	}
	return false, nil
}

// firstAction is the id of the action Enter runs on r, "" for none.
func firstAction(r *PickRow) string {
	if len(r.Actions) == 0 {
		return ""
	}
	return r.Actions[0].ID
}

// pickMove selects row i: its value goes to where selected is bound, else
// to the picker's state.
func (c *Controller) pickMove(e *Element, i int) error {
	if i < 0 || i >= len(e.Rows) {
		return nil
	}
	return c.pickSelect(e, e.Rows[i].Value)
}

func (c *Controller) pickSelect(e *Element, value string) error {
	st := c.pick(e.ID)
	st.Selected, st.Moved = value, true
	c.St.Pick[e.ID] = st
	var err error
	if bd, ok := c.V.Node(e.ID).Props["selected"].(a2ui.Bound); ok && bd.Writable() {
		err = c.S.Write(bd.Path, value)
	}
	c.Rebuild()
	return err
}

// pickPanel opens the actions panel on action at, or closes it.
func (c *Controller) pickPanel(e *Element, open bool, at int) {
	st := c.pick(e.ID)
	st.Panel, st.At = open, at
	c.St.Pick[e.ID] = st
	c.Rebuild()
}

func (c *Controller) pick(id string) Pick {
	if c.St.Pick == nil {
		c.St.Pick = map[string]Pick{}
	}
	return c.St.Pick[id]
}

// pickRun runs action on the selected row: the row's value goes to
// selected and the action's id to action, where they are bound, the panel
// closes, and onAction runs, its context reading them there.
func (c *Controller) pickRun(e *Element, action string) error {
	r := e.SelectedPick()
	if r == nil {
		return nil
	}
	if err := c.pickSelect(e, r.Value); err != nil {
		return err
	}
	c.pickPanel(e, false, 0)
	n := c.V.Node(e.ID)
	if bd, ok := n.Props["action"].(a2ui.Bound); ok && bd.Writable() {
		if err := c.S.Write(bd.Path, action); err != nil {
			return err
		}
	}
	c.Rebuild()
	if n.Props["onAction"] == nil {
		return nil
	}
	err := c.invoke(n, "onAction", true)
	c.Rebuild()
	return err
}

// setPickQuery makes q a HottyPicker's query, as the user typed it.
func (c *Controller) setPickQuery(e *Element, q string) error { return c.writePickQuery(e, q, true) }

// writePickQuery makes q a HottyPicker's query: it is written where query
// is bound (else kept as the renderer's), the selection goes back to the
// first row, the panel closes, and onQuery runs, its context reading the
// query there. typed: the user typed it in the field; else it is written
// to the field (Tab's completion), as the agent's queries are
// (Pick.Written).
func (c *Controller) writePickQuery(e *Element, q string, typed bool) error {
	if v, _ := e.Value.(string); v == q {
		return nil
	}
	c.St.Touched[e.ID] = true
	n := c.V.Node(e.ID)
	var err error
	if bd, ok := n.Props["query"].(a2ui.Bound); ok && bd.Writable() {
		err = c.S.Write(bd.Path, q)
	} else {
		c.St.Local[e.ID] = q
	}
	was := c.pick(e.ID)
	c.St.Pick[e.ID] = Pick{Typed: was.Typed, Written: was.Written}
	if typed {
		c.St.Pick[e.ID] = Pick{Typed: q, Written: was.Written}
	}
	c.St.Scroll[e.ID] = 0
	if bd, ok := n.Props["selected"].(a2ui.Bound); ok && bd.Writable() && err == nil {
		err = c.S.Write(bd.Path, "")
	}
	c.Rebuild()
	if err != nil || n.Props["onQuery"] == nil {
		return err
	}
	err = c.invoke(n, "onQuery", true)
	c.Rebuild()
	return err
}

// clickPick is a click on a HottyPicker's row i: it selects the row, or,
// when it was selected already, runs the row's first action.
func (c *Controller) clickPick(e *Element, i int, acts bool) error {
	if acts {
		if r := e.SelectedPick(); r != nil {
			return c.pickRun(e, firstAction(r))
		}
		return nil
	}
	return c.pickMove(e, i)
}
