package vectors

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
)

type vector struct {
	Name     string           `yaml:"name"`
	Story    string           `yaml:"story"`
	Messages []any            `yaml:"messages"`
	Steps    []map[string]any `yaml:"steps"`
}

// player is one rendition as a vector plays it: the user's keys and
// clicks, the program's focus and blur, and who has the keyboard.
type player interface {
	key(k string)
	focus(id string)
	blur()
	click(id string)
	// focused is the element with the keyboard; "" for none.
	focused() string
}

func TestKeys(t *testing.T) {
	var vs []vector
	if err := yaml.Unmarshal(Keys, &vs); err != nil {
		t.Fatal(err)
	}
	for _, v := range vs {
		for _, rendition := range []string{"cells", "host"} {
			t.Run(rendition+"/"+v.Name, func(t *testing.T) { play(t, v, rendition) })
		}
	}
}

func play(t *testing.T, v vector, rendition string) {
	run := story.NewRun()
	var actions []*a2ui.ActionMessage
	run.Out = func(o a2ui.Outbound) {
		if o.Action != nil {
			actions = append(actions, o.Action)
		}
	}
	msgs := v.Messages
	if v.Story != "" {
		st := story.Find(v.Story)
		if st == nil {
			t.Fatalf("no story %s", v.Story)
		}
		for _, m := range st.Messages {
			msgs = append(msgs, m)
		}
	}
	for _, m := range msgs {
		b, err := json.Marshal(jsonable(m))
		if err != nil {
			t.Fatal(err)
		}
		if err := run.Feed(b); err != nil {
			t.Fatal(err)
		}
	}
	s := run.Surfaces()[0]
	var p player
	if rendition == "cells" {
		p = newCellsPlayer(t, s.C)
	} else {
		p = newHostPlayer(t, s.C)
	}
	for i, st := range v.Steps {
		where := fmt.Sprintf("step %d %v", i+1, st)
		switch {
		case st["key"] != nil:
			p.key(fmt.Sprint(st["key"]))
		case st["type"] != nil:
			for _, r := range fmt.Sprint(st["type"]) {
				p.key(string(r))
			}
		case st["focus"] != nil:
			p.focus(fmt.Sprint(st["focus"]))
		case st["blur"] != nil:
			p.blur()
		case st["click"] != nil:
			p.click(fmt.Sprint(st["click"]))
		case st["expect"] != nil:
			ex, _ := st["expect"].(map[string]any)
			if f, ok := ex["focused"]; ok {
				want, _ := f.(string)
				if got := p.focused(); got != want {
					t.Errorf("%s: focused %q, want %q", where, got, want)
				}
			}
			if l, ok := ex["actions"].([]any); ok {
				var got, want []string
				for _, a := range actions {
					got = append(got, a.Name)
				}
				for _, a := range l {
					want = append(want, fmt.Sprint(a))
				}
				if strings.Join(got, ",") != strings.Join(want, ",") {
					t.Errorf("%s: actions %v, want %v", where, got, want)
				}
			}
			if c, ok := ex["context"].(map[string]any); ok {
				if len(actions) == 0 {
					t.Errorf("%s: no action", where)
				} else {
					last := actions[len(actions)-1].Context
					for k, want := range c {
						if !same(last[k], want) {
							t.Errorf("%s: context %s = %v, want %v", where, k, last[k], want)
						}
					}
				}
			}
			if d, ok := ex["data"].(map[string]any); ok {
				for path, want := range d {
					if got := s.S.Data.Value(path); !same(got, want) {
						t.Errorf("%s: %s = %v, want %v", where, path, got, want)
					}
				}
			}
		default:
			t.Fatalf("%s: what step?", where)
		}
	}
}

// same compares a JSON value with a YAML one.
func same(got, want any) bool {
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(jsonable(want))
	var x, y any
	_ = json.Unmarshal(a, &x)
	_ = json.Unmarshal(b, &y)
	return reflect.DeepEqual(x, y)
}

// jsonable is a YAML value as JSON can have it.
func jsonable(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, e := range x {
			out[k] = jsonable(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = jsonable(e)
		}
		return out
	}
	return v
}

// cellsPlayer plays the cells rendition: its keys and clicks are its own.
type cellsPlayer struct {
	t *testing.T
	c *view.Controller
	r *cells.Rendition
}

func newCellsPlayer(t *testing.T, c *view.Controller) *cellsPlayer {
	p := &cellsPlayer{t: t, c: c, r: cells.New(c)}
	p.r.Draw(80)
	return p
}

func (p *cellsPlayer) key(k string) {
	if _, err := p.r.Key(k); err != nil {
		p.t.Fatal(err)
	}
	p.r.Draw(80)
}

func (p *cellsPlayer) focus(id string) { p.c.Focus(id); p.r.Draw(80) }
func (p *cellsPlayer) blur()           { p.c.Focus(""); p.r.Draw(80) }

func (p *cellsPlayer) click(id string) {
	col, row, _, _, ok := p.r.Box(id)
	if !ok {
		p.t.Fatalf("no box for %s", id)
	}
	if err := p.r.Click(col, row); err != nil {
		p.t.Fatal(err)
	}
	p.r.Draw(80)
}

func (p *cellsPlayer) focused() string {
	if !p.c.St.Keyboard {
		return ""
	}
	return p.c.St.Focus
}

// hostPlayer plays the HTML rendition on a host: the host takes the keys
// it uses (hottytest.Key, SPEC §10.2), and the rest reach the rendition
// as they reach a program.
type hostPlayer struct {
	t    *testing.T
	h    *hottytest.Host
	r    *html.Rendition
	seen int
}

const surface = "story"

func newHostPlayer(t *testing.T, c *view.Controller) *hostPlayer {
	p := &hostPlayer{t: t, h: hottytest.New(t), r: html.New(c, surface)}
	p.send(hotty.Doc(surface, p.r.Doc()), hotty.Place(surface, hotty.Placement{Cols: 80, Rows: 24}))
	p.settle()
	return p
}

func (p *hostPlayer) send(cmds ...string) {
	p.t.Helper()
	for _, c := range cmds {
		_, _ = io.WriteString(p.h, c)
	}
	if errs := p.h.Errors(); len(errs) > 0 {
		p.t.Fatalf("the host refused: %v", errs)
	}
}

// settle sends the rendition's deltas, and hands it the host's events,
// until neither has more.
func (p *hostPlayer) settle() {
	p.t.Helper()
	p.send(p.r.Update()...)
	for evs := p.h.Events(); p.seen < len(evs); evs = p.h.Events() {
		for _, ev := range evs[p.seen:] {
			p.seen++
			if err := p.r.Event(ev); err != nil {
				p.t.Fatal(err)
			}
			p.send(p.r.Update()...)
		}
	}
}

func (p *hostPlayer) key(k string) {
	if !p.h.Key(k) {
		cmds, _, err := p.r.Key(k)
		if err != nil {
			p.t.Fatal(err)
		}
		p.send(cmds...)
	}
	p.settle()
}

func (p *hostPlayer) focus(id string) { p.r.C.Focus(id); p.settle() }
func (p *hostPlayer) blur()           { p.r.C.Focus(""); p.settle() }

func (p *hostPlayer) click(id string) {
	if err := p.h.Click(surface, html.DOMID(id)); err != nil {
		p.t.Fatal(err)
	}
	p.settle()
}

func (p *hostPlayer) focused() string {
	dom := p.h.Surface(surface).Focused()
	if dom == "" {
		return ""
	}
	id, ok := html.ViewID(dom)
	if !ok {
		return "?" + dom
	}
	return id
}
