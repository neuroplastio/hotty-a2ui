package view

import (
	"strings"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-go"
)

// A select's pick, and a cycle's steps (io_neuroplast_hotty.cycle,
// profile §6.16): a ChoicePicker that picks one, drawn "‹ staging ›", whose
// arrows and keys step through its options in place, round from either
// end, where a select opens a list.

// Picked is the index of a Choice's picked option, the first of several;
// -1 if none is.
func (e *Element) Picked() int {
	vs, _ := e.Value.([]string)
	for i, o := range e.Options {
		for _, v := range vs {
			if v == o.Value {
				return i
			}
		}
	}
	return -1
}

// StepChoice picks the option n on from a select's picked one (back when
// n < 0), round from either end, as a cycle's arrows do; with none picked,
// going on starts at the first and going back at the last. It does
// nothing to any other element.
func (c *Controller) StepChoice(id string, n int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Choice || len(e.Children) > 0 || len(e.Options) == 0 {
		return nil
	}
	k, i := len(e.Options), e.Picked()
	if i < 0 && n < 0 {
		i = 0
	}
	return c.SetValue(id, e.Options[((i+n)%k+k)%k].Value)
}

// CycleKey works a cycle by a key, named as SPEC §10.4 has it: arrow_left
// and arrow_up pick the option before, arrow_right, arrow_down, space and
// enter the one after, round from either end, as its arrows and a click
// on it do (on a host it is a button, which Space and Enter click); home
// and end the first and the last; a character the next option whose label
// starts with it, case aside, wrapping. ok reports whether it took the
// key: a character too, matched or not, as a select takes it.
func (c *Controller) CycleKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || !e.Cycle || len(e.Options) == 0 {
		return false, nil
	}
	switch key {
	case "arrow_left", "arrow_up":
		return true, c.StepChoice(id, -1)
	case "arrow_right", "arrow_down", "space", "enter":
		return true, c.StepChoice(id, 1)
	case "home":
		return true, c.SetValue(id, e.Options[0].Value)
	case "end":
		return true, c.SetValue(id, e.Options[len(e.Options)-1].Value)
	}
	text, _ := hotty.KeyText(key)
	if i := e.NextByLetter(e.Picked(), text); text != "" && i >= 0 {
		return true, c.SetValue(id, e.Options[i].Value)
	}
	return utf8.RuneCountInString(text) == 1, nil
}

// NextByLetter is a Choice's next option after from, wrapping, whose label
// starts with key, the text a key types (hotty.KeyText; case aside); -1
// if none does, or key is not one character.
func (e *Element) NextByLetter(from int, key string) int {
	n := len(e.Options)
	if utf8.RuneCountInString(key) != 1 {
		return -1
	}
	for k := 1; k <= n; k++ {
		j := (max(from, -1) + k + n) % n
		if strings.HasPrefix(strings.ToLower(e.Options[j].Label), strings.ToLower(key)) {
			return j
		}
	}
	return -1
}
