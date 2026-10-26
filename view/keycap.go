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
	// Case is how a cap's words are written: "title", each word's first
	// letter a capital (Ctrl+Shift+L, Esc); "lower" (ctrl+shift+l); or
	// "shifted", lower but for a letter with Shift, which is its capital
	// without the Shift (ctrl+L, as Shift types it). "" is the system's:
	// title, but for macOS's glyphs, whose Esc is "esc", as Apple's keys
	// print it.
	Case string
	// Sep joins a chord's labels. "" is the system's: "-" on macOS, as
	// Apple writes keys (Command-Shift-P), "+" elsewhere.
	Sep string
	// Then stands between a sequence's chords, dim: "" is ">" ([Ctrl+B] >
	// [N]), and a space a gap alone ([Ctrl+B] [N]; Keycaps.Between).
	Then string
	// Glue is whether a chord's modifiers drawn as glyphs sit tight against
	// its key, with no Sep (⌘⇧P), after any written as words
	// (Meta+Alt+⌃⇧R): "on" or "off". "" is the system's: on in macOS's
	// glyphs, off elsewhere.
	Glue string
	// Labels relabels keys by name (SPEC §10.4), over the system's: "super"
	// as Tux, say. A user may put a logo there; the kit's defaults show
	// none (HOST.md §2.4).
	Labels map[string]string
}

// The presentations (Keycaps.Present), cases (Keycaps.Case) and glue
// (Keycaps.Glue).
const (
	PresentGlyphs = "glyphs"
	PresentText   = "text"
	CaseTitle     = "title"
	CaseLower     = "lower"
	CaseShifted   = "shifted"
	GlueOn        = "on"
	GlueOff       = "off"
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

// Between is what a rendition draws between a sequence's chords
// (Keycaps.Then): ">" by default, a gap where it is blank.
func (k Keycaps) Between() string {
	if k.Then == "" {
		return ">"
	}
	return k.Then
}

// A Chord is a key with its modifiers, as a cap shows it: its labels,
// modifiers first in myelin's order, with Sep between them but for the
// last Glued, which sit tight (⌘⇧P, Meta+Alt+⌃⇧R; Keycaps.Glue).
type Chord struct {
	Labels []string
	Sep    string
	Glued  int
}

// Joined reports whether Sep comes before the chord's kth label.
func (ch Chord) Joined(k int) bool { return k > 0 && k <= len(ch.Labels)-ch.Glued }

// CapsText is caps as words, as a screen reader is to say them: each
// chord its labels joined as its cap joins them, a sequence's chords apart
// by spaces, alternatives by " or ".
func CapsText(alts [][]Chord) string {
	var b strings.Builder
	for i, seq := range alts {
		if i > 0 {
			b.WriteString(" or ")
		}
		for j, ch := range seq {
			if j > 0 {
				b.WriteByte(' ')
			}
			for k, l := range ch.Labels {
				if ch.Joined(k) {
					b.WriteString(ch.Sep)
				}
				b.WriteString(l)
			}
		}
	}
	return b.String()
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

// splitAlts is a hint's alternatives, split at each "/" that is not a key.
// A "/" is a key where it starts one (the alternative's start, after a
// space, after a modifier's "+") and ends one (the end, a space, the "/"
// that follows): "/", "ctrl+/", "ctrl+b /", "a//". Else it parts two
// alternatives: "↑/k", "+/-".
func splitAlts(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '/' {
			continue
		}
		starts := i == start || s[i-1] == ' ' || s[i-1] == '+'
		ends := i+1 == len(s) || s[i+1] == ' ' || s[i+1] == '/'
		if starts && ends {
			continue
		}
		if p := s[start:i]; strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
		start = i + 1
	}
	if p := s[start:]; strings.TrimSpace(p) != "" {
		out = append(out, p)
	}
	return out
}

// KeysText is keys a program wrote (a HottyKeys', a HottyPicker row's) as
// Caps reads them: HottyShortcut's names, a sequence's chords apart by
// spaces and alternatives by "/", each chord as text writes it (read with
// the Platform, then KeyText).
func KeysText(keys string, text func(string) string) string {
	alts := splitAlts(keys)
	for i, alt := range alts {
		ks := strings.Fields(alt)
		for j, k := range ks {
			ks[j] = text(k)
		}
		alts[i] = strings.Join(ks, " ")
	}
	return strings.Join(alts, "/")
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

// modOrder is the order caps put modifiers in on every system: myelin's
// (myelinkey SPEC §2.1).
var modOrder = []string{"super", "ctrl", "alt", "shift"}

// chord is a key's cap on os. With glue on, the modifiers drawn as glyphs
// (a label of one character) go last, tight against the key. In the
// shifted case a letter with Shift is its capital, and the Shift goes,
// unless the user labels the letter.
func (c *Controller) chord(os string, glyphs bool, mods []string, key string) Chord {
	mac := os == "mac"
	mods = slices.Clone(mods)
	slices.SortStableFunc(mods, func(a, b string) int {
		return slices.Index(modOrder, a) - slices.Index(modOrder, b)
	})
	k := c.Keycaps
	label := c.keyLabel(os, glyphs, key)
	_, own := k.Labels[key]
	if r, n := utf8.DecodeRuneInString(key); k.Case == CaseShifted && !own && n == len(key) && unicode.ToUpper(r) != r && slices.Contains(mods, "shift") {
		label = string(unicode.ToUpper(r))
		mods = slices.DeleteFunc(mods, func(m string) bool { return m == "shift" })
	}
	ch := Chord{Sep: k.Sep}
	switch {
	case ch.Sep != "":
	case mac:
		ch.Sep = "-"
	default:
		ch.Sep = "+"
	}
	glue := k.Glue == GlueOn || k.Glue == "" && mac && glyphs
	var marks []string
	for _, m := range mods {
		l := c.keyLabel(os, glyphs, m)
		if glue && utf8.RuneCountInString(l) == 1 {
			marks = append(marks, l)
		} else {
			ch.Labels = append(ch.Labels, l)
		}
	}
	ch.Labels = append(append(ch.Labels, marks...), label)
	if len(marks) > 0 {
		ch.Glued = len(marks) + 1
	}
	return ch
}

// keyLabel is a key's label on a cap: the user's (Keycaps.Labels), as
// it is, else the system's in the Keycaps' Case.
func (c *Controller) keyLabel(os string, glyphs bool, key string) string {
	if l, ok := c.Keycaps.Labels[key]; ok {
		return l
	}
	return cased(systemLabel(os, glyphs, key), c.Keycaps.Case)
}

// systemLabel is a key's label on os: its table's, else its name, a
// letter or an F-key in capitals and a name's words capitalised
// ("caps_lock" is "Caps Lock").
func systemLabel(os string, glyphs bool, key string) string {
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
	return cased(strings.ReplaceAll(key, "_", " "), CaseTitle)
}

// cased is a label in a Case (Keycaps.Case): its words' first letters
// capitals, or all of it lower case (the shifted case's capitals are
// chord's); "" leaves it as it is.
func cased(l, how string) string {
	switch how {
	case CaseLower, CaseShifted:
		return strings.ToLower(l)
	case CaseTitle:
		words := strings.Split(l, " ")
		for i, w := range words {
			if r, n := utf8.DecodeRuneInString(w); n > 0 {
				words[i] = string(unicode.ToUpper(r)) + w[n:]
			}
		}
		return strings.Join(words, " ")
	}
	return l
}

// The labels, by system and presentation; the first table to have a key
// gives it. The glyphs are Material Design Icons in the font hosts ship
// (HOST.md §2): Apple's key symbols drawn evenly, where the font's own
// ⌃⌥⇧⌘ are not; ⌃, ⇧ and Backspace's on every system, as Ctrl's caret
// and Shift's arrow read anywhere, while Alt and Meta are the system's.
// Esc is a word everywhere: ⎋ reads as reload. macOS's words are Apple's
// (Command, Control, Option; Backspace is Delete and Delete is Forward
// Delete).
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
	macWords = map[string]string{
		"super": "Command", "ctrl": "Control", "alt": "Option", "enter": "Return",
		"backspace": "Delete", "delete": "Forward Delete",
	}
	macIcons = map[string]string{
		"ctrl": "\U000F0634", "alt": "\U000F0635", "shift": "\U000F0636", "super": "\U000F0633",
		"enter": "\U000F0311", "backspace": "\U000F0B5C", "delete": "\U000F0E7F", "tab": "\U000F0312",
		"caps_lock": "\U000F0632", "escape": "esc",
	}
	pcIcons = map[string]string{"ctrl": "\U000F0634", "shift": "\U000F0636", "backspace": "\U000F0B5C"}

	pcLabels  = []map[string]string{textArrows, words}
	pcGlyphs  = []map[string]string{pcIcons, iconArrows, words}
	macText   = []map[string]string{macWords, textArrows, words}
	macGlyphs = []map[string]string{macIcons, iconArrows, words}
)
