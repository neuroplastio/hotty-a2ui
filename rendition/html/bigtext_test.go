package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// A HottyBigText on a host is its text as it is, in display type: k-big,
// its size's class and its align's, its lines apart at a br. Bound, a new
// text reaches the host as a delta.
func TestBigText(t *testing.T) {
	x := newHarness(t)
	h := `"catalogId":"` + hottycat.ID + `"`
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"score":"3 : 1"}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["title","score"]},
 {"id":"title","component":"HottyBigText",`+h+`,"text":"Hello\nworld","size":"large","align":"center"},
 {"id":"score","component":"HottyBigText",`+h+`,"text":{"@path":"/score"}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	for _, want := range [][3]string{{"title", "class", "k-big k-big-large k-big-center"}, {"score", "class", "k-big k-big-medium"}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	if got := s.TextOf("title"); got != "Helloworld" {
		t.Errorf("the title reads %q", got)
	}
	if !strings.Contains(s.HTML(), `class="k-big k-big-large k-big-center">Hello<br/>world</div>`) {
		t.Error("the title's lines are not apart at a br")
	}
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/score","value":"4 : 1"}}]`), &msgs))
	x.process(msgs...)
	if got := s.TextOf("score"); got != "4 : 1" {
		t.Errorf("the score reads %q once written", got)
	}
}
