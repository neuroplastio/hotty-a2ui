package story

import (
	"encoding/json"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// cheatsheetProgram is the program behind hotty/cheatsheet, as plx is
// behind its onboarding's keys: it holds the keymap, and when the user
// picks a leader (a leader action's context's leader) it sends the sheet
// its keys again (/keys), as an agent's message, logged as one. The
// HottyKeys are bound to them.
func cheatsheetProgram(r *Run) {
	r.hears = func(a *a2ui.ActionMessage) {
		leader, _ := a.Context["leader"].(string)
		if a.Name != "leader" || leader == "" {
			return
		}
		keys := map[string]any{}
		for id, seq := range cheatsheetKeys {
			keys[id] = leader + " " + seq
		}
		msg, _ := json.Marshal(map[string]any{"version": "v1.0", "updateDataModel": map[string]any{
			"surfaceId": a.SurfaceID, "path": "/keys", "value": keys}})
		r.Feed(msg)
	}
}

// cheatsheetKeys are hotty/cheatsheet's keys after the leader, by the
// line's id.
var cheatsheetKeys = map[string]string{"v": "v", "s": "s", "t": "w t", "x": "x", "f": "/"}
