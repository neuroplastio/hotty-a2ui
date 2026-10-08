package html

import "github.com/neuroplastio/hotty-a2ui/view"

// iconEmoji are the icons a host draws as colour emoji, where the emoji
// says the name plainly. The rest keep the one-character glyph every
// rendition draws (view.Icons): a name not here is the glyph too.
var iconEmoji = map[string]string{
	"accountCircle": "👤", "add": "➕", "arrowBack": "⬅️", "arrowForward": "➡️",
	"attachFile": "📎", "calendarToday": "📅", "call": "📞", "camera": "📷",
	"check": "✅", "close": "❌", "delete": "🗑️", "download": "⬇️",
	"edit": "✏️", "error": "❗", "event": "📅", "fastForward": "⏩",
	"favorite": "❤️", "favoriteOff": "🤍", "folder": "📁", "help": "❓",
	"home": "🏠", "info": "ℹ️", "locationOn": "📍", "lock": "🔒",
	"lockOpen": "🔓", "mail": "✉️", "notifications": "🔔", "notificationsOff": "🔕",
	"pause": "⏸️", "payment": "💳", "person": "🙂", "phone": "📱",
	"photo": "🖼️", "play": "▶️", "print": "🖨️", "refresh": "🔄",
	"rewind": "⏪", "search": "🔍", "send": "📤", "settings": "⚙️",
	"share": "🔗", "shoppingCart": "🛒", "skipNext": "⏭️", "skipPrevious": "⏮️",
	"star": "⭐", "stop": "⏹️", "upload": "⬆️", "visibility": "👁️",
	"visibilityOff": "🙈", "volumeDown": "🔉", "volumeMute": "🔇", "volumeOff": "🔇",
	"volumeUp": "🔊", "warning": "⚠️",
}

// iconText is an Icon as the host shows it: its emoji, else the glyph;
// emoji says which.
func iconText(name string) (g string, emoji bool) {
	if g, ok := iconEmoji[name]; ok {
		return g, true
	}
	return view.IconGlyph(name), false
}
