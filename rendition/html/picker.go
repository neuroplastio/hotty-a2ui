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
// edits, its caret the host's; a rule; its status, muted and smaller,
// with a rule under it, so that it reads as the head's and not as a row;
// its rows, a heading wherever their section changes, each an option whose id is the
// picker's, "~i" and the row's index, its keys at its right, its
// description on a second line; and its footer, the selected row's note at
// its left and what its keys do, flush right, as Zed's. While the actions
// panel is open, the
// selected row's actions take the rows' place. A picker that keeps keys
// (view.Element.KeepKeys) is k-pick-keep, whose rows don't wrap their keys
// out of sight (kit.css). The input gives the program the keys
// the picker takes (pickKeys).
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
	if e.KeepKeys {
		box.set("class", "k-pick k-pick-keep")
	}
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
				m.keycaps("k-pick-keys", m.keyText(key)))
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
		icons := slices.ContainsFunc(e.Rows, func(r view.PickRow) bool { return r.Icon != "" })
		if icons {
			list.set("class", "k-pick-rows k-pick-icons")
		}
		if m.page {
			top, height = 0, len(lines)+1
		}
		for k := top; k < len(lines) && k < top+height; k++ {
			ln := lines[k]
			switch {
			case ln.Row < 0:
				// A heading on the last line has none of its rows under it.
				if k < top+height-1 {
					list.add(el("div", "class", "k-pick-head").add(texts(ln.Heading)...))
				}
			case !ln.Desc:
				// A description shows on its line, as in cells: a row on
				// the last line shows without it.
				desc := k+1 < len(lines) && lines[k+1].Desc && k+1 < top+height
				list.add(m.pickItem(e, ln.Row, desc, icons))
			}
		}
	}
	box.add(list)
	box.add(m.pickFooter(e))
	return box
}

// pickFooter is a HottyPicker's footer (view.Element.PickFooter), in the
// order it shows, from the left: the note (k-pick-note), Actions and the
// panel key, the other keys in k-pick-more, the most significant first,
// which the CSS puts at the right, and Enter's at the right. Each is a
// k-pick-act, its title and its key in k-pick-key. k-pick-more has only
// the room the others leave, its least significant wrapping out of sight;
// short of room still, the note is cut (kit.css).
func (m *markup) pickFooter(e *view.Element) *node {
	f := e.PickFooter()
	act := func(h view.Hint) *node {
		return el("span", "class", "k-pick-act").add(texts(h.Desc)...).add(m.keycaps("k-pick-key", m.keyText(h.Key)))
	}
	foot := el("div", "class", "k-pick-keys-line")
	if f.Note != "" {
		foot.add(el("span", "class", "k-pick-note").add(texts(f.Note)...))
	}
	if f.Panel.Key != "" {
		foot.add(act(f.Panel))
	}
	if len(f.Keys) > 1 {
		more := el("span", "class", "k-pick-more")
		for _, h := range f.Keys[1:] {
			more.add(act(h))
		}
		foot.add(more)
	}
	if len(f.Keys) > 0 {
		foot.add(act(f.Keys[0]))
	}
	return foot
}

// pickItem is a row: its icon, where the picker's rows have icons (a
// blank one where it has none); its name (k-pick-name), its group and
// title, their matched characters in spans of class k-match; its detail;
// and its keys; with desc, those in a line of their own (k-pick-line), and
// its description under them.
func (m *markup) pickItem(e *view.Element, i int, desc, icons bool) *node {
	r := e.Rows[i]
	cls := "k-pick-row"
	if desc {
		cls += " k-pick-two"
	}
	if i == e.Selected {
		cls += " k-sel"
	}
	item := el("div", "id", partID(e.ID, partItem+strconv.Itoa(i)), "class", cls, "role", "option",
		"aria-selected", strconv.FormatBool(i == e.Selected), "data-on", "click")
	title, name := el("span", "class", "k-pick-title"), el("span", "class", "k-pick-name")
	if icons {
		title.add(pickIcon(r.Icon))
	}
	if r.Group != "" {
		name.add(el("span", "class", "k-pick-group").add(pickMarked(r.Group, r.GroupMatch)...).add(texts(": ")...))
	}
	title.add(name.add(pickMarked(r.Title, r.Match)...))
	if r.Detail != "" {
		title.add(el("span", "class", "k-pick-detail").add(texts(r.Detail)...))
	}
	first := item
	if desc {
		first = el("div", "class", "k-pick-line")
		item.add(first)
	}
	first.add(title)
	if r.Keys != "" {
		first.add(m.keycaps("k-pick-keys", view.KeysText(r.Keys, m.keyText)))
	}
	if desc {
		item.add(el("div", "class", "k-pick-desc").add(texts(r.Description)...))
	}
	return item
}

// pickIcon is a row's icon: a name's, as an Icon draws it beside a label,
// or a character of the agent's own, such as an emoji, as it is; a blank
// as wide for none. A screen reader skips it, as the title says the row.
func pickIcon(name string) *node {
	switch {
	case name == "":
		return el("span", "class", "k-icon k-icon-blank", "aria-hidden", "true")
	case view.PickIconNamed(name):
		return beside(name)
	}
	return el("span", "class", "k-icon", "aria-hidden", "true").add(texts(name)...)
}

// keyText is a key as the markup writes it (markup.key), as written where
// it has no reader.
func (m *markup) keyText(k string) string {
	if m.key != nil {
		return m.key(k)
	}
	return view.KeyText(k)
}

// caps is a hint's keys as caps, glyphs unless the user set text
// (view.Keycaps.Present).
func (r *Rendition) caps(keys string) [][]view.Chord {
	return r.C.Caps(keys, r.C.Keycaps.Glyphs(true))
}

// said is keys as a screen reader says them: the caps' text labels, for
// the user's system (Command-Shift-P on a Mac).
func (r *Rendition) said(keys string) string {
	return view.CapsText(r.C.Caps(keys, false))
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
