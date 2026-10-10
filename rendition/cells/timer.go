package cells

import "github.com/neuroplastio/hotty-a2ui/view"

// timerFace is a HottyTimer or a HottyStopwatch as drawn, on one row, as
// bubbles' timer and stopwatch examples draw theirs ("Exiting in 4s",
// "Elapsed: 1.5s"): its label and a space, then its time as the view
// writes it (view.FormatTimer), in the text's colour while it counts and in
// muted while it stands still (stopped, or a timer run out), so that one
// that does not move does not look as if it were stuck.
func timerFace(e *view.Element) []glyph {
	t, _ := e.Value.(string)
	st := style{}
	if !e.Active {
		st.role = Muted
	}
	if e.Label == "" {
		return line(t, st)
	}
	return concat(line(e.Label+" ", style{}), line(t, st))
}
