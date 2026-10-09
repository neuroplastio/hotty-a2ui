package view

import "time"

// A SpinnerSet is a Spinner's frames and how long each shows.
type SpinnerSet struct {
	Frames   []string
	Interval time.Duration
}

// Spinners are the frame sets a HottySpinner names (its `spinner`), as
// bubbles' spinner package has them (MIT, charm.land/bubbles/v2/spinner):
// the same frames at the same rates, so that the kit's spinners look like
// Bubble Tea's. Dot's frames lose bubbles' trailing space; the renditions
// put a space before the label.
var Spinners = map[string]SpinnerSet{
	"line":      {[]string{"|", "/", "-", "\\"}, time.Second / 10},
	"dot":       {[]string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}, time.Second / 10},
	"miniDot":   {[]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, time.Second / 12},
	"jump":      {[]string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"}, time.Second / 10},
	"pulse":     {[]string{"█", "▓", "▒", "░"}, time.Second / 8},
	"points":    {[]string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●"}, time.Second / 7},
	"globe":     {[]string{"🌍", "🌎", "🌏"}, time.Second / 4},
	"moon":      {[]string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}, time.Second / 8},
	"monkey":    {[]string{"🙈", "🙉", "🙊"}, time.Second / 3},
	"meter":     {[]string{"▱▱▱", "▰▱▱", "▰▰▱", "▰▰▰", "▰▰▱", "▰▱▱", "▱▱▱"}, time.Second / 7},
	"hamburger": {[]string{"☱", "☲", "☴", "☲"}, time.Second / 3},
	"ellipsis":  {[]string{"", ".", "..", "..."}, time.Second / 3},
}

// SpinnerFrames is a Spinner's frame set: its Variant's, else dot's.
func (e *Element) SpinnerFrames() SpinnerSet {
	if s, ok := Spinners[e.Variant]; ok {
		return s
	}
	return Spinners["dot"]
}

// ProgressInterval is how often an indeterminate Progress moves.
const ProgressInterval = time.Second / 10
