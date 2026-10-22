package html

import (
	"slices"
	"strconv"
	"strings"

	"github.com/clipperhouse/uax29/v2/graphemes"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// picker is a HottyPicker on a host (profile §2, §6.32): a box holding
// its query, after its chip, as a text input of its own, which the host
// edits, its caret the host's; a rule; its status; its rows, a heading
// wherever their section changes, each an option whose id is the
// picker's, "~i" and the row's index; and its line of keys. While the
// actions panel is open, the selected row's actions take the rows' place.
// The input gives the program the keys the picker takes (pickKeys).
//
// A query the program writes (a completion, a staged command), other than
// as the user typed it (view.Element.Written), goes to the input with
// a=field (Rendition.complete), which changes it focused or not, if the
// user has not typed since, and puts the caret after it (SPEC §6.4). The
// input keeps its id, its history and its caret.
//
// A host keeps Tab and Shift+Tab, which move the focus (SPEC §10.2), so
// the input sits between two stops of the picker's own (pickTab after it,
// pickBack before it), focusable and unseen: Tab onto the one after
// completes the query (Rendition.Event), and either gives the input the
// keyboard back.
func (m *markup) picker(e *view.Element) *node {
	box := el("div", "id", domID(e.ID), "class", "k-pick")
	q, _ := e.Value.(string)
	query := el("div", "class", "k-pick-query")
	if e.Chip != "" {
		query.add(el("span", "class", "k-pick-chip").add(texts(e.Chip)...))
	}
	m.pickers = append(m.pickers, e)
	in := el("input", "id", pickInput(e), "class", "k-pick-input", "type", "text", "value", q,
		"data-on", "input", "data-keys", pickKeys(e), "role", "combobox", "aria-expanded", "true",
		"autocomplete", "off", "spellcheck", "false")
	if e.Placeholder != "" {
		in.set("placeholder", e.Placeholder)
	}
	stop := func(part string) *node {
		return el("span", "id", partID(e.ID, part), "class", "k-pick-stop", "tabindex", "0", "aria-hidden", "true")
	}
	box.add(query.add(stop(pickBack), in, stop(pickTab)))
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

// pickTab and pickBack are a HottyPicker's stops after and before its
// input, parts of the picker (a cycle's › and ‹): Tab and Shift+Tab land
// there.
const (
	pickTab  = partMore
	pickBack = partLess
)

// pickInput is the DOM id of a HottyPicker's input, a part of the picker.
func pickInput(e *view.Element) string { return partID(e.ID, partTitle) }

// pickKeys are the keys a HottyPicker's input gives the program, which
// works the picker with them (view.Controller.PickerKey): the moves, its
// actions' keys, and the panel's.
func pickKeys(e *view.Element) string {
	keys := "arrow_up=program arrow_down=program ctrl+p=program ctrl+n=program page_up=program page_down=program " +
		"enter=program mod+enter=program alt+enter=program shift+enter=program shift+delete=program"
	if e.PanelKey != "" {
		keys += " " + e.PanelKey + "=program"
	}
	return keys
}
