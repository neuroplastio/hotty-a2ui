package storybook

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

// keyNames are the W3C UI Events key values of the keys that are not
// characters.
var keyNames = map[rune]string{
	tea.KeyEnter: "Enter", tea.KeyTab: "Tab", tea.KeyEscape: "Escape", tea.KeyBackspace: "Backspace",
	tea.KeyDelete: "Delete", tea.KeyInsert: "Insert", tea.KeySpace: " ",
	tea.KeyUp: "ArrowUp", tea.KeyDown: "ArrowDown", tea.KeyLeft: "ArrowLeft", tea.KeyRight: "ArrowRight",
	tea.KeyHome: "Home", tea.KeyEnd: "End", tea.KeyPgUp: "PageUp", tea.KeyPgDown: "PageDown",
	tea.KeyF1: "F1", tea.KeyF2: "F2", tea.KeyF3: "F3", tea.KeyF4: "F4", tea.KeyF5: "F5", tea.KeyF6: "F6",
	tea.KeyF7: "F7", tea.KeyF8: "F8", tea.KeyF9: "F9", tea.KeyF10: "F10", tea.KeyF11: "F11", tea.KeyF12: "F12",
}

// keyValue is a key as a Shortcut writes it (NEIO-11): its modifiers,
// then a W3C UI Events key value, joined by "+": "a", "A", "Control+s",
// "Shift+Tab", "Escape". Shift goes with a key only where the character
// does not already say it.
func keyValue(k tea.Key) string {
	var mods []string
	if k.Mod&tea.ModCtrl != 0 {
		mods = append(mods, "Control")
	}
	if k.Mod&tea.ModAlt != 0 {
		mods = append(mods, "Alt")
	}
	if k.Mod&(tea.ModMeta|tea.ModSuper) != 0 {
		mods = append(mods, "Meta")
	}
	name, named := keyNames[k.Code]
	switch {
	case named:
		if k.Mod&tea.ModShift != 0 && name != " " {
			mods = append(mods, "Shift")
		}
	case k.Text != "" && len(mods) == 0:
		return k.Text
	default:
		c := k.Code
		if k.Mod&tea.ModShift != 0 {
			if k.ShiftedCode != 0 {
				c = k.ShiftedCode
			} else {
				c = unicode.ToUpper(c)
			}
			if c == k.Code {
				mods = append(mods, "Shift")
			}
		}
		name = string(c)
	}
	return strings.Join(append(mods, name), "+")
}
