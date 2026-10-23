package story

import (
	"encoding/json"
)

// cheatsheetProgram is the program behind hotty/cheatsheet, as plx is
// behind its onboarding's keys: it holds the keymaps, and the sheet's
// HottyKeys are bound to the keys it sends (/keys). A ChoicePicker sends
// no action, so it reads the keymap picked (/keymap) before each draw, as
// a program reads its own surface's data, and when the keys it sent are
// another keymap's, it sends that one's, as an agent's message, logged as
// one.
func cheatsheetProgram(r *Run) {
	r.config = func(s *Surface) {
		l, _ := s.S.Data.Value("/keymap").([]any)
		if len(l) == 0 {
			return
		}
		name, _ := l[0].(string)
		leader := cheatsheetLeaders[name]
		if sent, _ := s.S.Data.Value("/keys/v").(string); leader == "" || sent == leader+" "+cheatsheetKeys["v"] {
			return
		}
		keys := map[string]any{}
		for id, seq := range cheatsheetKeys {
			keys[id] = leader + " " + seq
		}
		msg, _ := json.Marshal(map[string]any{"version": "v1.0", "updateDataModel": map[string]any{
			"surfaceId": s.S.ID, "path": "/keys", "value": keys}})
		r.Feed(msg)
	}
}

// cheatsheetLeaders are the leaders of hotty/cheatsheet's keymaps, by the
// Keymap's value; cheatsheetKeys its keys after the leader, by the line's
// id.
var (
	cheatsheetLeaders = map[string]string{"tmux": "ctrl+b", "screen": "ctrl+a", "emacs": "ctrl+x"}
	cheatsheetKeys    = map[string]string{"v": "v", "s": "s", "t": "w t", "x": "x", "f": "/"}
)
