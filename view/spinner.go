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

// Frame is the frame a set shows at a time: the same in every rendition
// drawn then (profile §3.4, the clock).
func (s SpinnerSet) Frame(at time.Time) string {
	if len(s.Frames) == 0 || s.Interval <= 0 {
		return ""
	}
	return s.Frames[step(at, s.Interval, int64(len(s.Frames)))]
}

// ProgressInterval is how often an indeterminate Progress moves, and
// ProgressSteps how many moves its sweep takes: its segment, a quarter of
// its bar, goes from just before the bar's start to just past its end in
// five seconds, in every rendition.
const (
	ProgressInterval = time.Second / 10
	ProgressSteps    = 50
)

// ProgressStep is which step of its sweep an indeterminate Progress is at,
// at a time: from 0 to ProgressSteps − 1.
func ProgressStep(at time.Time) int {
	return int(step(at, ProgressInterval, ProgressSteps))
}

// step is which of n steps, each d long, a time is at, counting from the
// Unix epoch round and round.
func step(at time.Time, d time.Duration, n int64) int64 {
	k := at.UnixNano() / int64(d) % n
	if k < 0 {
		k += n
	}
	return k
}
