package story

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// keycapsProgram is the program behind hotty/keycaps. What a program's
// config sets (plx's), the story's controls set here: every surface draws
// keys (view.Controller.Keycaps) for the system the keys surface's
// /system names ("auto", the Platform's guess), as its /present says
// ("default", the rendition's), with the label its /meta picks, if not
// "system", as super's.
func keycapsProgram(r *Run) {
	r.config = func(s *Surface) {
		k := view.Keycaps{}
		ctl := r.P.Surface("keys")
		if ctl == nil {
			s.C.Keycaps = k
			return
		}
		picked := func(path string) string {
			l, _ := ctl.Data.Value(path).([]any)
			if len(l) == 0 {
				return ""
			}
			v, _ := l[0].(string)
			return v
		}
		if os := picked("/system"); os != "auto" {
			k.OS = os
		}
		if p := picked("/present"); p != "default" {
			k.Present = p
		}
		if meta := picked("/meta"); meta != "system" && meta != "" {
			k.Labels = map[string]string{"super": meta}
		}
		s.C.Keycaps = k
	}
}
