package view

import (
	"slices"
	"unicode"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A text field's suggestions (profile §6.22): the extension
// io_neuroplast_hotty.suggestions on a basic TextField, as bubbles' text
// input has them. Its options are the agent's, a list literal or bound;
// the field shows those that start with its value, case aside, as a list
// under it, and the first, or the one the arrows highlight, after the
// caret, faint, for Tab or → to take. The agent rewrites a bound list as
// it hears what was typed (onInput), and the field narrows it meanwhile.

// SuggestRows is the most rows a text field's list of suggestions shows
// at once; past it the list scrolls, its highlight in view.
const SuggestRows = 5

// Suggest is the renderer's state of a text field's suggestions, as the
// user works them: the one the arrows highlighted, Hi ("" for none), and
// whether Escape, a pick or Tab closed the list, Closed. Both hold for the
// value At, at which they were set: once the user edits the value, the
// list opens again with nothing highlighted.
type Suggest struct {
	At     string
	Hi     string
	Closed bool
}

// suggestable reports whether a TextField takes suggestions: a one-line
// field that shows what it holds, so not a longText or an obscured one.
func suggestable(e *Element) bool {
	return e.Kind == TextField && e.Variant != "longText" && e.Variant != "obscured"
}

// suggestions reads a TextField's io_neuroplast_hotty.suggestions into
// its element: Suggestions, its options resolved in its scope; Shown,
// those its value leaves (by index, in order), unless the list is closed;
// Selected, the highlighted one's place in Shown (-1 for none); Height,
// the rows the list shows at most, and Top, the first of Shown it shows,
// moved as little as keeps the highlight in view.
func (b *Builder) suggestions(n *a2ui.Node, e *Element) {
	spec, ok := hottyExt(n)["suggestions"].(map[string]any)
	if !ok || !suggestable(e) {
		return
	}
	e.Selected = -1
	if l, ok := spec["list"]; ok {
		on, _ := b.S.Context(n.Scope).Resolve(l)
		e.GhostOnly = on == false
	}
	opts, _ := b.S.Context(n.Scope).Resolve(spec["options"])
	e.Suggestions = []string{}
	if l, ok := opts.([]any); ok {
		for _, x := range l {
			if s, ok := x.(string); ok && s != "" && !slices.Contains(e.Suggestions, s) {
				e.Suggestions = append(e.Suggestions, s)
			}
		}
	}
	v, _ := e.Value.(string)
	st := b.St.Suggest[n.Key]
	if st.At != v {
		st = Suggest{}
		delete(b.St.Suggest, n.Key)
		b.St.Scroll[n.Key] = 0
	}
	if st.Closed {
		return
	}
	e.Shown = matchSuggestions(e.Suggestions, v)
	for i, j := range e.Shown {
		if st.Hi != "" && e.Suggestions[j] == st.Hi {
			e.Selected = i
		}
	}
	e.Height = SuggestRows
	top := b.St.Scroll[n.Key]
	if e.Selected >= 0 {
		top = min(top, e.Selected)
		top = max(top, e.Selected-e.Height+1)
	}
	e.Top = max(min(top, len(e.Shown)-e.Height), 0)
	b.St.Scroll[n.Key] = e.Top
}

// matchSuggestions are the suggestions a value leaves, by index, in
// order: those that start with it, case aside, as bubbles matches them,
// but for the value itself; none while the value is empty.
func matchSuggestions(all []string, v string) []int {
	if v == "" {
		return nil
	}
	var out []int
	for i, s := range all {
		if _, ok := cutPrefixFold(s, v); ok && s != v {
			out = append(out, i)
		}
	}
	return out
}

// cutPrefixFold is s without prefix, which s starts with, case aside
// (simple case folding, rune by rune); ok is false when it does not.
func cutPrefixFold(s, prefix string) (rest string, ok bool) {
	for prefix != "" {
		if s == "" {
			return "", false
		}
		a, n := utf8.DecodeRuneInString(s)
		b, m := utf8.DecodeRuneInString(prefix)
		if !sameFold(a, b) {
			return "", false
		}
		s, prefix = s[n:], prefix[m:]
	}
	return s, true
}

// sameFold reports whether two runes are one under simple case folding.
func sameFold(a, b rune) bool {
	if a == b {
		return true
	}
	for r := unicode.SimpleFold(a); r != a; r = unicode.SimpleFold(r) {
		if r == b {
			return true
		}
	}
	return false
}

// Suggestion is a text field's suggestion that Tab takes: the highlighted
// one, else the first its value leaves; "" while none shows.
func (e *Element) Suggestion() string {
	switch {
	case len(e.Shown) == 0:
		return ""
	case e.Selected >= 0 && e.Selected < len(e.Shown):
		return e.Suggestions[e.Shown[e.Selected]]
	}
	return e.Suggestions[e.Shown[0]]
}

// Ghost is what a text field shows after its value, faint, for Tab or →
// to take: the rest of its Suggestion after what was typed; "" for none.
func (e *Element) Ghost() string {
	v, _ := e.Value.(string)
	rest, _ := cutPrefixFold(e.Suggestion(), v)
	return rest
}

// SuggestionParts splits the suggestion at index i of Shown into what the
// value typed of it and the rest, which the list shows apart.
func (e *Element) SuggestionParts(i int) (typed, rest string) {
	s := e.Suggestions[e.Shown[i]]
	v, _ := e.Value.(string)
	rest, _ = cutPrefixFold(s, v)
	return s[:len(s)-len(rest)], rest
}

// SuggestKey works a text field's suggestions by a key, named as SPEC
// §10.4 has it, before the field's keymap (profile §3.7): Tab takes the
// suggestion, and so does ArrowRight at the value's end (atEnd: the caret
// there, nothing selected); ArrowDown and Control+n highlight the next one
// and ArrowUp and Control+p the one before, round from either end, as
// bubbles' keys move through them; Enter picks the highlighted one; Escape
// closes the list. ArrowDown and Control+n open a closed list again. A
// taken suggestion is the field's value, as written, and closes the list
// until the next edit. ok reports whether the key was the suggestions'.
// Tab with nothing to take, Enter with nothing highlighted and Escape with
// no list are not: they move focus, submit and reach the program as ever.
func (c *Controller) SuggestKey(id, key string, atEnd bool) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Suggestions == nil {
		return false, nil
	}
	v, _ := e.Value.(string)
	n := len(e.Shown)
	switch key {
	case "Tab":
		if s := e.Suggestion(); s != "" && s != v {
			return true, c.takeSuggestion(e, s)
		}
	case "ArrowRight":
		if atEnd && e.Ghost() != "" {
			return true, c.takeSuggestion(e, e.Suggestion())
		}
	case "ArrowDown", "Control+n":
		if n == 0 {
			if m := matchSuggestions(e.Suggestions, v); len(m) > 0 {
				c.highlight(e, e.Suggestions[m[0]])
				return true, nil
			}
			return false, nil
		}
		c.highlight(e, e.Suggestions[e.Shown[(e.Selected+1)%n]])
		return true, nil
	case "ArrowUp", "Control+p":
		if n == 0 {
			return false, nil
		}
		i := e.Selected - 1
		if i < 0 {
			i = n - 1
		}
		c.highlight(e, e.Suggestions[e.Shown[i]])
		return true, nil
	case "Enter":
		if e.Selected >= 0 && e.Selected < n {
			return true, c.takeSuggestion(e, e.Suggestions[e.Shown[e.Selected]])
		}
	case "Escape":
		if n > 0 {
			c.St.Suggest[id] = Suggest{At: v, Closed: true}
			c.Rebuild()
			return true, nil
		}
	}
	return false, nil
}

// PickSuggestion takes the suggestion at index i of a text field's Shown,
// as a click on its row does.
func (c *Controller) PickSuggestion(id string, i int) error {
	e := c.V.Find(id)
	if e == nil || i < 0 || i >= len(e.Shown) {
		return nil
	}
	return c.takeSuggestion(e, e.Suggestions[e.Shown[i]])
}

// highlight highlights a text field's suggestion s, at its value now.
func (c *Controller) highlight(e *Element, s string) {
	v, _ := e.Value.(string)
	c.St.Suggest[e.ID] = Suggest{At: v, Hi: s}
	c.Rebuild()
}

// takeSuggestion makes s a text field's value, as the user's edit, and
// closes its list until the next.
func (c *Controller) takeSuggestion(e *Element, s string) error {
	c.St.Suggest[e.ID] = Suggest{At: s, Closed: true}
	return c.SetValue(e.ID, s)
}

// suggestInput runs a text field's suggestions' onInput, after the user
// changed its value: in its scope, so that its context reads the value
// where it is bound, and the agent may rewrite the options.
func (c *Controller) suggestInput(id string) error {
	n := c.V.Node(id)
	if n == nil {
		return nil
	}
	spec, _ := hottyExt(n)["suggestions"].(map[string]any)
	a, ok := spec["onInput"].(map[string]any)
	if !ok {
		return nil
	}
	return c.dispatch(a, n.Scope, n.ComponentID, true)
}
