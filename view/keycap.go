package view

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Keycaps is how the renditions draw keys (KIT-08h): each chord in a cap
// of its own, a sequence's caps in order, labelled for the user's system.
// The zero Keycaps is the Platform's system, drawn as the rendition draws
// by default.
type Keycaps struct {
	// OS is the system the labels are for: "mac", "linux" or "windows".
	// "" guesses from the Platform: mac where Mod is super, linux elsewhere
	// (no host tells Windows from Linux).
	OS string
	// Present is how a cap draws its keys: "glyphs", as symbols where the
	// system has them (⌘, Return's arrow), the Nerd Font icons hosts ship
	// (hotty HOST.md §2), or "text", as words, in any font. "" is the
	// rendition's default: glyphs on a host, text in cells, whose font is
	// the terminal's.
	Present string
	// Labels relabels keys by name (SPEC §10.4), over the system's: "super"
	// as "◆", say. A user may put a logo there; the kit's defaults show
	// none (HOST.md §2.4).
	Labels map[string]string
}

// The presentations (Keycaps.Present).
const (
	PresentGlyphs = "glyphs"
	PresentText   = "text"
)

// Glyphs reports whether caps draw as glyphs, where the rendition's
// default is host's (Keycaps.Present).
func (k Keycaps) Glyphs(host bool) bool {
	switch k.Present {
	case PresentGlyphs:
		return true
	case PresentText:
		return false
	}
	return host
}

// A Chord is a key with its modifiers, as a cap shows it: its labels,
// modifiers first in the system's order; Tight where they sit together
// (macOS's glyphs, ⇧⌘P), else joined by a "+".
type Chord struct {
	Labels []string
	Tight  bool
}

// Caps is a hint's keys (KeyText's form: alternatives joined by "/", a
// sequence's keys by spaces) as caps: the alternatives, each a sequence of
// chords. An arrow alone after an arrow with modifiers takes them, as in
// "alt+↑/↓". glyphs is the presentation (Keycaps.Glyphs).
func (c *Controller) Caps(keys string, glyphs bool) [][]Chord {
	os := c.Keycaps.OS
	if os == "" {
		os = "linux"
		if c.Platform.Mod == "super" {
			os = "mac"
		}
	}
	var out [][]Chord
	var last []string
	for _, alt := range splitAlts(keys) {
		var seq []Chord
		for _, k := range strings.Fields(alt) {
			mods, key := readChord(k)
			if len(mods) == 0 && len(seq) == 0 && arrow(key) && last != nil {
				mods = last
			}
			last = nil
			if arrow(key) && len(mods) > 0 {
				last = mods
			}
			seq = append(seq, c.chord(os, glyphs, mods, key))
		}
		if len(seq) > 0 {
			out = append(out, seq)
		}
	}
	return out
}

// splitAlts is a hint's alternatives, split at each "/" that is not a key:
// "/" alone, or after a modifier's "+" ("ctrl+/").
func splitAlts(s string) []string {
	ps := strings.Split(s, "/")
	var out []string
	for i := 0; i < len(ps); i++ {
		p := ps[i]
		if (p == "" || strings.HasSuffix(p, "+")) && i+1 < len(ps) && ps[i+1] == "" {
			p += "/"
			i++
		}
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// readChord is a key as KeyText writes it, as its modifiers and its name
// (SPEC §10.4): the arrows and abbreviations read back, a capital alone as
// the letter with shift.
func readChord(k string) (mods []string, key string) {
	switch {
	case k == "+":
		key = "+"
	case strings.HasSuffix(k, "++"):
		mods, key = strings.Split(strings.TrimSuffix(k, "++"), "+"), "+"
	default:
		parts := strings.Split(k, "+")
		mods, key = parts[:len(parts)-1], parts[len(parts)-1]
	}
	for name, text := range keyNames {
		if key == text {
			key = name
		}
	}
	if r, n := utf8.DecodeRuneInString(key); n == len(key) && unicode.IsUpper(r) && len(mods) == 0 {
		mods, key = []string{"shift"}, string(unicode.ToLower(r))
	}
	return mods, key
}

func arrow(key string) bool { return strings.HasPrefix(key, "arrow_") }

// modOrder is the order caps put modifiers in: Apple's on macOS (⌃⌥⇧⌘),
// the PC's elsewhere (Win+Ctrl+Alt+Shift).
var modOrder = map[bool][]string{
	true:  {"ctrl", "alt", "shift", "super"},
	false: {"super", "ctrl", "alt", "shift"},
}

// chord is a key's cap on os.
func (c *Controller) chord(os string, glyphs bool, mods []string, key string) Chord {
	mac := os == "mac"
	order := modOrder[mac]
	mods = slices.Clone(mods)
	slices.SortStableFunc(mods, func(a, b string) int {
		return slices.Index(order, a) - slices.Index(order, b)
	})
	ch := Chord{Tight: mac && glyphs}
	for _, m := range append(mods, key) {
		ch.Labels = append(ch.Labels, c.keyLabel(os, glyphs, m))
	}
	return ch
}

// keyLabel is a key's label on a cap: the user's (Keycaps.Labels), else
// the system's, else its name, a letter or an F-key in capitals and a
// name's words capitalised ("caps_lock" is "Caps Lock").
func (c *Controller) keyLabel(os string, glyphs bool, key string) string {
	if l, ok := c.Keycaps.Labels[key]; ok {
		return l
	}
	if os == "windows" && key == "super" {
		if glyphs {
			return "⊞"
		}
		return "Win"
	}
	tables := pcLabels
	switch {
	case os == "mac" && glyphs:
		tables = macGlyphs
	case os == "mac":
		tables = macText
	case glyphs:
		tables = pcGlyphs
	}
	for _, t := range tables {
		if l, ok := t[key]; ok {
			return l
		}
	}
	if utf8.RuneCountInString(key) == 1 || len(key) <= 3 && key[0] == 'f' && strings.Trim(key[1:], "0123456789") == "" {
		return strings.ToUpper(key)
	}
	words := strings.Split(key, "_")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// The labels, by system and presentation; the first table to have a key
// gives it. The glyphs are Material Design Icons in the font hosts ship
// (HOST.md §2): Apple's key symbols drawn evenly, where the font's own
// ⌃⌥⇧⌘ are not; Backspace's on every system. Esc is a word everywhere:
// ⎋ reads as reload.
var (
	words = map[string]string{
		"super": "Meta", "ctrl": "Ctrl", "alt": "Alt", "shift": "Shift",
		"enter": "Enter", "escape": "Esc", "backspace": "Backspace", "delete": "Del",
		"tab": "Tab", "space": "Space", "page_up": "PgUp", "page_down": "PgDn",
		"insert": "Ins", "context_menu": "Menu",
	}
	textArrows = map[string]string{"arrow_up": "↑", "arrow_down": "↓", "arrow_left": "←", "arrow_right": "→"}
	iconArrows = map[string]string{
		"arrow_up": "\U000F005D", "arrow_down": "\U000F0045", "arrow_left": "\U000F004D", "arrow_right": "\U000F0054",
	}
	macWords = map[string]string{"super": "Cmd", "alt": "Option", "enter": "Return", "escape": "esc"}
	macIcons = map[string]string{
		"ctrl": "\U000F0634", "alt": "\U000F0635", "shift": "\U000F0636", "super": "\U000F0633",
		"enter": "\U000F0311", "backspace": "\U000F0B5C", "tab": "\U000F0312", "caps_lock": "\U000F0632",
		"escape": "esc",
	}
	pcIcons = map[string]string{"backspace": "\U000F0B5C"}

	pcLabels  = []map[string]string{textArrows, words}
	pcGlyphs  = []map[string]string{pcIcons, iconArrows, words}
	macText   = []map[string]string{macWords, textArrows, words}
	macGlyphs = []map[string]string{macIcons, iconArrows, words}
)
