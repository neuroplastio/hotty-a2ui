package html

import (
	"slices"
	"strconv"
	"strings"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottyedit"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// picker is a HottyPicker on a host (profile §2, §6.32): a focusable box,
// as a HottyList is, holding its query after its chip, with a caret where
// the program's caret is; a rule; its status; its rows, a heading wherever
// their section changes, each an option whose id is the picker's, "~i" and
// the row's index; and its line of keys. While the actions panel is open,
// the selected row's actions take the rows' place. The box's keys are the
// program's (listKeys): the query is edited by the program, as cells
// edits it (Rendition.pickEdit), so that what the program writes to it
// shows, which a host's own input would not while it is focused (SPEC
// §6.2).
func (m *markup) picker(e *view.Element, caret int) *node {
	box := el("div", "id", domID(e.ID), "class", "k-pick", "tabindex", "0", "role", "combobox",
		"aria-expanded", "true", "data-keys", listKeys)
	q, _ := e.Value.(string)
	query := el("div", "id", partID(e.ID, partTitle), "class", "k-pick-query")
	if e.Chip != "" {
		query.add(el("span", "class", "k-pick-chip").add(texts(e.Chip)...))
	}
	cl := clustersOf(q)
	caret = min(max(caret, 0), len(cl))
	before, after := strings.Join(cl[:caret], ""), strings.Join(cl[caret:], "")
	query.add(el("span", "class", "k-pick-text").add(texts(before)...),
		el("span", "class", "k-caret", "aria-hidden", "true"),
		el("span", "class", "k-pick-text").add(texts(after)...))
	if q == "" && e.Placeholder != "" {
		query.add(el("span", "class", "k-pick-hint").add(texts(e.Placeholder)...))
	}
	box.add(query)
	if e.Status != "" {
		cls, text := "k-pick-status", e.Status
		if e.StatusError {
			cls, text = "k-pick-status k-error", "✗ "+text
		}
		box.add(el("div", "id", partID(e.ID, partStatus), "class", cls, "aria-live", "polite").add(texts(text)...))
	}
	list := el("div", "class", "k-pick-rows", "role", "listbox")
	switch {
	case e.Panel:
		row := e.SelectedPick()
		list.add(el("div", "class", "k-pick-head").add(texts("Actions: " + row.Title)...))
		for i, a := range row.Actions {
			key := a.Key
			if i == 0 {
				key = "enter"
			}
			item := el("div", "class", "k-pick-row", "role", "option", "aria-selected", strconv.FormatBool(i == e.PanelAt))
			if i == e.PanelAt {
				item.set("class", "k-pick-row k-sel")
			}
			item.add(el("span", "class", "k-pick-title").add(texts(a.Title)...),
				el("span", "class", "k-pick-keys").add(texts(m.keyText(key))...))
			list.add(item)
		}
	case len(e.Rows) == 0:
		text := e.NoRows
		if e.Busy {
			text = "Loading…"
		}
		list.add(el("div", "class", "k-pick-empty").add(texts(text)...))
	default:
		// A page shows every line, not a window of them (§2.1).
		lines, top, height := e.PickLines(), e.Top, e.Height
		if m.page {
			top, height = 0, len(lines)+1
		}
		for k := top; k < len(lines) && k < top+height; k++ {
			ln := lines[k]
			if ln.Row < 0 {
				// A heading on the last line has none of its rows under it.
				if k < top+height-1 {
					list.add(el("div", "class", "k-pick-head").add(texts(ln.Heading)...))
				}
				continue
			}
			list.add(m.pickItem(e, ln.Row))
		}
	}
	box.add(list)
	var hints []string
	if e.Panel {
		hints = []string{"↑/↓ move", "enter run", "esc back"}
	} else if enter, panel := e.PickFooter(); enter != "" {
		hints = []string{"↑/↓ move", "enter " + strings.ToLower(enter)}
		if panel {
			hints = append(hints, m.keyText(e.PanelKey)+" actions")
		}
	}
	box.add(el("div", "class", "k-pick-keys-line").add(texts(strings.Join(hints, " • "))...))
	return box
}

// pickItem is a row: its group and title, the title's matched characters
// in spans of class k-match, its detail, and its other actions' keys.
func (m *markup) pickItem(e *view.Element, i int) *node {
	r := e.Rows[i]
	item := el("div", "id", partID(e.ID, partItem+strconv.Itoa(i)), "class", "k-pick-row", "role", "option",
		"aria-selected", strconv.FormatBool(i == e.Selected), "data-on", "click")
	if i == e.Selected {
		item.set("class", "k-pick-row k-sel")
	}
	title := el("span", "class", "k-pick-title")
	if r.Group != "" {
		title.add(el("span", "class", "k-pick-group").add(texts(r.Group + ": ")...))
	}
	title.add(pickMarked(r.Title, r.Match)...)
	if r.Detail != "" {
		title.add(el("span", "class", "k-pick-detail").add(texts(r.Detail)...))
	}
	item.add(title)
	var keys []string
	for k, a := range r.Actions {
		if k > 0 && a.Key != "" {
			keys = append(keys, m.keyText(a.Key))
		}
	}
	if len(keys) > 0 {
		item.add(el("span", "class", "k-pick-keys").add(texts(strings.Join(keys, "  "))...))
	}
	return item
}

// keyText is a key as the markup writes it (markup.key), as written where
// it has no reader.
func (m *markup) keyText(k string) string {
	if m.key != nil {
		return m.key(k)
	}
	return view.KeyText(k)
}

// keyText is a key a program wrote, read with the Platform, as the key
// hints write it.
func (r *Rendition) keyText(k string) string {
	if p, ok := r.C.Platform.ReadKey(k); ok {
		k = p
	}
	return view.KeyText(k)
}

// pickMarked is a title as nodes, the clusters that hold a matched
// character (view.PickRow.Match, character ranges) in spans of class
// k-match.
func pickMarked(s string, match [][2]int) []*node {
	if len(match) == 0 {
		return texts(s)
	}
	var out []*node
	var run strings.Builder
	on, at := false, 0
	flush := func() {
		if run.Len() == 0 {
			return
		}
		if on {
			out = append(out, el("span", "class", "k-match").add(texts(run.String())...))
		} else {
			out = append(out, texts(run.String())...)
		}
		run.Reset()
	}
	for _, cl := range clustersOf(s) {
		n := len([]rune(cl))
		hit := slices.ContainsFunc(match, func(m [2]int) bool { return at < m[1] && at+n > m[0] })
		if hit != on {
			flush()
			on = hit
		}
		run.WriteString(cl)
		at += n
	}
	flush()
	return out
}

// clustersOf is a string's grapheme clusters, as hottyedit counts a
// field's characters.
func clustersOf(s string) []string {
	var out []string
	for g := graphemes.FromString(s); g.Next(); {
		out = append(out, g.Value())
	}
	return out
}

// pickEdit edits a focused HottyPicker's query by a key, as cells does
// (rendition/cells editKey): by the field's keymap, the surface's (SetKeys)
// and then those of the components it is in, typing the characters it does
// not bind; the caret is the rendition's. ok is false for a key the keymap
// leaves to the program.
func (r *Rendition) pickEdit(e *view.Element, key string) (ok bool, err error) {
	km := hotty.Resolve(r.C.Platform, false, append([]string{r.keys}, r.C.V.KeyChain(e.ID)...)...)
	f := r.pickField(e)
	switch km.Lookup(key) {
	case "", hotty.Copy, hotty.Cut, hotty.Paste, hotty.Undo, hotty.Redo, hotty.Submit:
		return false, nil
	}
	_, changed := f.Key(km, key)
	if changed {
		return true, r.C.SetValue(e.ID, f.Value)
	}
	return true, nil
}

// pickField is the hottyedit.Field a HottyPicker's query is edited by: the
// one kept while its value is the query, else a new one, its caret at the
// end.
func (r *Rendition) pickField(e *view.Element) *hottyedit.Field {
	q, _ := e.Value.(string)
	if r.picks == nil {
		r.picks = map[string]*hottyedit.Field{}
	}
	f := r.picks[e.ID]
	if f == nil || f.Value != q {
		f = &hottyedit.Field{Value: q, Caret: len(clustersOf(q))}
		r.picks[e.ID] = f
	}
	return f
}

// pickCaret is where a HottyPicker's caret is drawn: the field's while it
// edits the query, else its end.
func (r *Rendition) pickCaret(e *view.Element) int {
	q, _ := e.Value.(string)
	if f := r.picks[e.ID]; f != nil && f.Value == q {
		return f.Caret
	}
	return len(clustersOf(q))
}
