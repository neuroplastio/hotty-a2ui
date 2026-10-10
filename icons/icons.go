// Package icons is the shapes the kit draws its icons with on a HOTTY
// host, and the glyphs it draws them with in cells (gov R-4).
//
// The basic catalog's 59 names are Material Symbols, Sharp, filled
// (basic.go, made by `make icons` from the Iconify data REV pins), and an
// svgPath is checked by Path. HottyIcon's names are Material's own: the
// 59's (account_circle), and those a program registers, which `make icons`
// generates for it from a list (cmd/iconsgen -set). Either way a rendition
// gets an Icon, one path it writes into its own <svg>. The package imports
// nothing of the kit's.
package icons

// Icon is a shape: one path in a square viewBox from 0 0, filled with the
// text's colour by the nonzero rule, or stroked with it.
type Icon struct {
	// Box is the side of the viewBox.
	Box int
	// Path is the path data.
	Path string
	// Stroke, when it is more than 0, is the width the path is stroked
	// at, in the viewBox's units, and it is not filled: a stroke icon's.
	Stroke float64
}

// Basic is a basic catalog icon name's shape, if it is one of the 59.
func Basic(name string) (Icon, bool) {
	d, ok := basic[name]
	return Icon{Box: 24, Path: d}, ok
}

// registered are the names programs registered (Register).
var registered = map[string]Icon{}

// Register registers a Material Symbols name's shape for HottyIcon. Call it
// from init, as the files cmd/iconsgen -set writes do; a later
// registration of a name replaces the earlier one.
func Register(name string, ic Icon) { registered[name] = ic }

// Named is a Material Symbols name's shape: one a program registered, or
// one of the 59 by Material's name for it (account_circle).
func Named(name string) (Icon, bool) {
	if ic, ok := registered[name]; ok {
		return ic, true
	}
	if b, ok := material[name]; ok {
		return Basic(b)
	}
	return Icon{}, false
}

// MaxPath is the most path data an svgPath may hold.
const MaxPath = 8 << 10

// Path is an svgPath's shape: one path in a 24 box, as A2UI's reference
// renderers draw it. It is one only when d is path data and nothing else
// (commands, numbers, separators) and at most MaxPath bytes, so that
// whatever an agent sends can't be anything but a path in the markup.
func Path(d string) (Icon, bool) {
	if d == "" || len(d) > MaxPath {
		return Icon{}, false
	}
	for i := 0; i < len(d); i++ {
		if !pathByte[d[i]] {
			return Icon{}, false
		}
	}
	return Icon{Box: 24, Path: d}, true
}

// pathByte are the bytes path data is made of.
var pathByte = func() (t [256]bool) {
	for _, c := range "MmLlHhVvCcSsQqTtAaZz0123456789eE+-., \t\n\r\f" {
		t[c] = true
	}
	return t
}()

// Glyphs are the basic catalog's icon names as one character each, of
// width one in a terminal: no emoji, whose width terminals disagree on.
// Cells and text draw an Icon so, and so does a host when it has no shape
// for one.
var Glyphs = map[string]string{
	"accountCircle": "◉", "add": "+", "arrowBack": "←", "arrowForward": "→",
	"attachFile": "⌇", "calendarToday": "▦", "call": "☏", "camera": "◘",
	"check": "✓", "close": "✕", "delete": "⌫", "download": "⤓",
	"edit": "✎", "event": "▦", "error": "✗", "fastForward": "»",
	"favorite": "♥", "favoriteOff": "♡", "folder": "▭", "help": "?",
	"home": "⌂", "info": "ⓘ", "locationOn": "⌖", "lock": "▣",
	"lockOpen": "□", "mail": "✉", "menu": "≡", "moreVert": "⋮",
	"moreHoriz": "⋯", "notificationsOff": "◌", "notifications": "◔", "pause": "‖",
	"payment": "▭", "person": "☺", "phone": "☏", "photo": "▨",
	"play": "▶", "print": "⎙", "refresh": "↻", "rewind": "«",
	"search": "⌕", "send": "➤", "settings": "⚙", "share": "⇪",
	"shoppingCart": "⊔", "skipNext": "⇥", "skipPrevious": "⇤", "star": "★",
	"starHalf": "✫", "starOff": "☆", "stop": "■", "upload": "⤒",
	"visibility": "◎", "visibilityOff": "⊘", "volumeDown": "◂", "volumeMute": "◃",
	"volumeOff": "×", "volumeUp": "▸", "warning": "!",
}

// Unknown is the glyph of a name that isn't one of the 59, and of an
// svgPath in cells.
const Unknown = "◇"

// Glyph is an icon name's character: one of the 59's, by its basic name
// or its Material one.
func Glyph(name string) string {
	if g, ok := Glyphs[name]; ok {
		return g
	}
	if g, ok := Glyphs[material[name]]; ok {
		return g
	}
	return Unknown
}
