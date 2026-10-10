package view

import (
	"cmp"
	"slices"
	"strings"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// HottyKeyHints (profile §6.10) in the view: the keys a surface takes now,
// as bubbles' help shows them. The renditions ask the controller for them
// as they draw, since they follow the keyboard, which moves without a
// rebuild.

// Hint is a key, as bubbles' help writes it ("↑/k", "ctrl+s"), and what
// it does.
type Hint struct {
	Key  string `json:"key"`
	Desc string `json:"desc"`
}

// mapKeyHints makes a HottyKeyHints' element: Open while the full view
// shows, Active when ? switches between the views.
func mapKeyHints(b *Builder, n *a2ui.Node) *Element {
	_, set := n.Props["toggle"]
	return &Element{Kind: KeyHints, Active: !set || b.Bool(n, "toggle"), Open: b.St.FullHints}
}

// KeyHints are the keys the surface takes now: short, one line, as
// bubbles' short help has them; full, the columns of its full help. The
// element with the keyboard comes first, then the surface's
// HottyShortcuts that have a label, then, in the full view, Tab, Shift+Tab
// and ?. Short ends with "? more" when ? shows the full view. keys is the
// surface's keymap (a rendition's SetKeys), which a text field's
// components' keymaps override.
//
// host is true for a rendition on a host, which moves focus among the
// elements it works itself without telling the renderer (profile §2, the
// keyboard): their keys are left out, as a guess would be wrong once Tab
// moved on. Those of an element whose keys the program takes (a Table, a
// HottyList, a Slider, a select) show, as the program's keys go there.
func (c *Controller) KeyHints(keys string, host bool) (short []Hint, full [][]Hint) {
	if c.St.Keyboard {
		if e := c.V.Find(c.St.Focus); e != nil && (!host || programKeys(e)) {
			short, full = c.elementHints(e, keys)
		}
	}
	if c.St.Modal != "" {
		esc := Hint{"esc", "close"}
		short = append(short, esc)
		if len(full) == 0 {
			full = [][]Hint{nil}
		}
		full[len(full)-1] = append(full[len(full)-1], esc)
	}
	var shortcuts []Hint
	for _, sc := range c.V.Shortcuts {
		if sc.Label != "" {
			shortcuts = append(shortcuts, Hint{KeyText(sc.Key), sc.Label})
		}
	}
	short = append(short, shortcuts...)
	general := []Hint{{"tab", "next"}, {"shift+tab", "back"}}
	toggle := false
	c.V.Walk(func(e *Element) bool {
		toggle = toggle || e.Kind == KeyHints && e.Active
		return !toggle
	})
	if toggle {
		short = append(short, Hint{"?", "more"})
		general = append(general, Hint{"?", "close help"})
	}
	if len(shortcuts) > 0 {
		full = append(full, shortcuts)
	}
	return short, append(full, general)
}

// programKeys reports whether a host leaves an element's keys to the
// program, which works it: a box (a Table, a HottyList), a Slider, a
// HottyRangeSlider's knob, a select (rendition/html's Key).
func programKeys(e *Element) bool {
	switch e.Kind {
	case Table, RichList, Slider, Knob, DiffView, Tree:
		return true
	case Choice:
		return len(e.Children) == 0
	}
	return false
}

// elementHints are an element's keys, keys the surface's keymap: those
// its short line shows, and all of them in groups, as bubbles splits a
// component's full help into columns (a list's moves, then its filter).
// The keys that move its item (MoveKey) follow, in a group of their own.
func (c *Controller) elementHints(e *Element, keys string) (short []Hint, groups [][]Hint) {
	short, groups = c.ownHints(e, keys)
	if mv := c.moveHints(e); len(mv) > 0 {
		short = append(short, mv[0])
		groups = append(groups, mv)
	}
	return short, groups
}

// moveHints are the keys that move the item the keyboard is on
// (MoveKey): a reorderable list's, table's or tree's selected item, while
// no filter applies, or the item of a List with reorder that holds it.
func (c *Controller) moveHints(e *Element) []Hint {
	if moves(e) && isRows(e) {
		if e.filtering() || e.Query.Editing {
			return nil
		}
		h := []Hint{{"alt+↑/↓", "move"}}
		if e.Kind == Tree {
			h = append(h, Hint{"alt+←/→", "move out/in"})
		}
		return h
	}
	for _, a := range c.Ancestors(e.ID) {
		if a.Kind == Stack && moves(a) {
			return []Hint{{"alt+↑/↓", "move"}}
		}
	}
	return nil
}

// ownHints are an element's own keys, as elementHints has them.
func (c *Controller) ownHints(e *Element, keys string) (short []Hint, groups [][]Hint) {
	enter := c.FormOf(e.ID) != nil
	act := c.V.Node(e.ID) != nil && c.V.Node(e.ID).Props["onActivate"] != nil
	switch e.Kind {
	case RichList:
		if e.Query.Editing {
			short = []Hint{{"enter", "apply filter"}, {"esc", "cancel"}}
			return short, [][]Hint{short}
		}
		short = []Hint{{"↑/k", "up"}, {"↓/j", "down"}}
		moves := slices.Clone(short)
		if n, _ := e.Pages(); n > 1 {
			moves = append(moves, Hint{"→/l/pgdn", "next page"}, Hint{"←/h/pgup", "prev page"})
		}
		moves = append(moves, Hint{"g/home", "go to start"}, Hint{"G/end", "go to end"})
		var more []Hint
		if e.Filter {
			more = append(more, Hint{"/", "filter"})
		}
		if e.Query.Text != "" {
			more = append(more, Hint{"esc", "clear filter"})
		}
		if act {
			more = append(more, Hint{"enter", "choose"})
		}
		return append(short, more...), nonEmpty(moves, more)
	case Table:
		short = []Hint{{"↑/k", "up"}, {"↓/j", "down"}}
		moves := append(slices.Clone(short), Hint{"g/home", "go to start"}, Hint{"G/end", "go to end"})
		pages := []Hint{{"pgup", "page up"}, {"pgdn", "page down"}}
		if act {
			short = append(short, Hint{"enter", "choose"})
			pages = append(pages, Hint{"enter", "choose"})
		}
		return short, [][]Hint{moves, pages}
	case DiffView:
		short = []Hint{{"↑/k", "prev hunk"}, {"↓/j", "next hunk"}}
		moves := append(slices.Clone(short), Hint{"g/home", "first hunk"}, Hint{"G/end", "last hunk"})
		if act {
			short = append(short, Hint{"enter", "choose"})
			moves = append(moves, Hint{"enter", "choose"})
		}
		return short, [][]Hint{moves}
	case Tree:
		short = []Hint{{"↑/k", "up"}, {"↓/j", "down"}, {"→/l", "open"}, {"←/h", "close"}}
		moves := append(slices.Clone(short), Hint{"g/home", "go to start"}, Hint{"G/end", "go to end"})
		more := []Hint{{"enter", "open/close"}}
		if act {
			more = []Hint{{"enter", "choose"}}
		}
		return append(short, more...), [][]Hint{moves, more}
	case ScrollView:
		// bubbles' viewport's words. A host scrolls the box itself (SPEC
		// §5.3), so on one these never show (programKeys).
		short = []Hint{{"↑/k", "up"}, {"↓/j", "down"}, {"f/pgdn", "page down"}, {"b/pgup", "page up"}}
		groups = [][]Hint{
			{{"↑/k", "up"}, {"↓/j", "down"}, {"g/home", "go to start"}, {"G/end", "go to end"}},
			{{"f/pgdn", "page down"}, {"b/pgup", "page up"}, {"d", "½ page down"}, {"u", "½ page up"}},
		}
		if len(e.Children) == 0 && !e.Wrap {
			groups = append(groups, []Hint{{"←/h", "move left"}, {"→/l", "move right"}})
		}
		return short, groups
	case TextField, DateTime:
		long := e.Kind == TextField && e.Variant == "longText"
		if enter && !long {
			short = []Hint{{"enter", "submit"}}
		}
		return short, fieldHints(hotty.Resolve(long, append([]string{keys}, c.V.KeyChain(e.ID)...)...), long, enter)
	case CheckBox:
		short = []Hint{{"space", "toggle"}}
	case Switch:
		// A switch is a button on a host, which Enter clicks too (SPEC
		// §10.2): it flips, in a HottyForm as anywhere.
		short = []Hint{{"space/enter", "toggle"}}
	case Option:
		short = []Hint{{"space", "pick"}}
	case Choice:
		short = []Hint{{"enter", "open"}}
	case Slider, Knob:
		short = []Hint{{"←/→", "adjust"}}
		return short, [][]Hint{append(slices.Clone(short), Hint{"home/end", "min/max"})}
	case Button:
		short = []Hint{{"enter", "press"}}
		return short, [][]Hint{short}
	case Tab:
		short = []Hint{{"enter", "show"}}
	case Media, Modal:
		short = []Hint{{"enter", "open"}}
	}
	// A box, as a host's checkbox, leaves Enter to its HottyForm; a chip
	// takes it.
	box := e.Kind == CheckBox
	if anc := c.Ancestors(e.ID); e.Kind == Option && len(anc) > 0 {
		box = anc[len(anc)-1].Variant != "chips"
	}
	if enter && box {
		short = append(short, Hint{"enter", "submit"})
	}
	return short, nonEmpty(short)
}

// nonEmpty are the groups that hold a hint.
func nonEmpty(groups ...[]Hint) [][]Hint {
	var out [][]Hint
	for _, g := range groups {
		if len(g) > 0 {
			out = append(out, g)
		}
	}
	return out
}

// editNames are what a text field's actions do, in the words of
// bubbles' text input's help, in the order the full view lists them: the
// moves in its first column, the edits in its second.
var editNames = []struct {
	action hotty.Action
	desc   string
	edit   bool
}{
	{hotty.CharBackward, "character backward", false}, {hotty.CharForward, "character forward", false},
	{hotty.WordBackward, "word backward", false}, {hotty.WordForward, "word forward", false},
	{hotty.LineStart, "line start", false}, {hotty.LineEnd, "line end", false},
	{hotty.LinePrevious, "line up", false}, {hotty.LineNext, "line down", false},
	{hotty.InputStart, "input begin", false}, {hotty.InputEnd, "input end", false},
	{hotty.SelectAll, "select all", false},
	{hotty.DeleteCharBackward, "delete character backward", true}, {hotty.DeleteCharForward, "delete character forward", true},
	{hotty.DeleteWordBackward, "delete word backward", true}, {hotty.DeleteWordForward, "delete word forward", true},
	{hotty.DeleteToLineStart, "delete before cursor", true}, {hotty.DeleteToLineEnd, "delete after cursor", true},
	{hotty.Newline, "insert newline", true}, {hotty.Submit, "submit", true},
}

// fieldHints are a text field's keys, from its keymap, in two groups, the
// moves and the edits: each action once, its first two keys, those with
// Meta (macOS's Command) after the others, so that a terminal keymap's
// Control+e shows before the default's Meta+ArrowRight; Submit only in a
// HottyForm, where Enter submits; the lines' only in a long text, as
// bubbles' text input has none.
func fieldHints(km *hotty.Keymap, long, inForm bool) [][]Hint {
	keys := map[hotty.Action][]string{}
	for _, b := range strings.Fields(km.Format()) {
		i := strings.LastIndexByte(b, '=')
		a := hotty.Action(b[i+1:])
		keys[a] = append(keys[a], b[:i])
	}
	for a, ks := range keys {
		meta := func(k string) int {
			if strings.Contains(k, "Meta+") {
				return 1
			}
			return 0
		}
		slices.SortStableFunc(ks, func(x, y string) int { return cmp.Compare(meta(x), meta(y)) })
		ks = ks[:min(len(ks), 2)]
		for i, k := range ks {
			ks[i] = KeyText(k)
		}
		keys[a] = ks
	}
	var moves, edits []Hint
	for _, n := range editNames {
		lines := n.action == hotty.LinePrevious || n.action == hotty.LineNext || n.action == hotty.Newline
		if len(keys[n.action]) == 0 || n.action == hotty.Submit && !inForm || lines && !long {
			continue
		}
		h := Hint{strings.Join(keys[n.action], "/"), n.desc}
		if n.edit {
			edits = append(edits, h)
		} else {
			moves = append(moves, h)
		}
	}
	return nonEmpty(moves, edits)
}

// keyNames are the W3C key values bubbles' help writes otherwise.
var keyNames = map[string]string{
	"ArrowUp": "↑", "ArrowDown": "↓", "ArrowLeft": "←", "ArrowRight": "→",
	"Escape": "esc", "PageUp": "pgup", "PageDown": "pgdn", "Control": "ctrl",
}

// KeyText is a key value (SPEC §10.4) as bubbles' help writes keys:
// modifiers and names in lower case, "ctrl" for Control, arrows as
// arrows; a character as it is ("G").
func KeyText(key string) string {
	if k, ok := hotty.ParseKey(key); ok {
		key = k
	}
	parts := strings.Split(key, "+")
	if strings.HasSuffix(key, "++") {
		parts = append(strings.Split(strings.TrimSuffix(key, "++"), "+"), "+")
	}
	for i, p := range parts {
		if n, ok := keyNames[p]; ok {
			parts[i] = n
		} else if len([]rune(p)) > 1 {
			parts[i] = strings.ToLower(p)
		}
	}
	return strings.Join(parts, "+")
}

// ToggleHints switches the surface's key hints between their short view
// and their full one, if a HottyKeyHints on it takes ?; ok reports
// whether one did.
func (c *Controller) ToggleHints() bool {
	toggle := false
	c.V.Walk(func(e *Element) bool {
		toggle = toggle || e.Kind == KeyHints && e.Active
		return !toggle
	})
	if !toggle {
		return false
	}
	c.St.FullHints = !c.St.FullHints
	c.Rebuild()
	return true
}
