package view

// Icons are the basic catalog's icon names as one character each, of
// width one in a terminal: no emoji, whose width terminals disagree on.
// Every rendition draws an Icon so, and an unknown name as ◇.
var Icons = map[string]string{
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

// IconGlyph is an icon name's character.
func IconGlyph(name string) string {
	if g, ok := Icons[name]; ok {
		return g
	}
	return "◇"
}
