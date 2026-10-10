package html

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// pageRendition is a surface's rendition named name, of the components given
// after a basic surface's createSurface with data, for its page.
func pageRendition(t *testing.T, name, data, components string) *Rendition {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hottycat.Catalog())
	must(t, p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":`+data+`}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":`+strings.ReplaceAll(components, "HOTTY", hottycat.ID)+`}}]`)))
	r := New(view.NewController(p.Surface("s")), name)
	r.Clock = func() time.Time { return time.Unix(0, 0) }
	return r
}

// count is how many times re matches s.
func count(s, re string) int { return len(regexp.MustCompile(re).FindAllString(s, -1)) }

// The kit's stylesheet goes in a page once, however many surfaces it
// shows: PageCSS is it, and each surface's Page is a fragment with none,
// whose ids its name prefixes, so that two surfaces of the same
// components share no id.
func TestPageStylesheetOnce(t *testing.T) {
	const comps = `[{"id":"root","component":"Column","children":["name"]},{"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"}}]`
	a := pageRendition(t, "a", `{"name":"Ada"}`, comps).Page()
	b := pageRendition(t, "b", `{"name":"Bo"}`, comps).Page()
	page := "<style>" + PageCSS() + "</style>" + a + b
	if n := strings.Count(page, ".k-surface {"); n != 1 {
		t.Errorf("the kit's sheet is in the page %d times", n)
	}
	if !strings.Contains(PageCSS(), kitCSS) || !strings.Contains(PageCSS(), ".k-surface.k-page") {
		t.Error("PageCSS is not kit.css and page mode's rules")
	}
	for _, f := range []string{a, b} {
		if strings.Contains(f, "<style") || strings.Contains(f, "<meta") {
			t.Errorf("a fragment brings a sheet or a head:\n%s", f)
		}
	}
	if !strings.HasPrefix(a, `<div id="a" class="k-surface k-page">`) {
		t.Errorf("a's top element: %.60s", a)
	}
	for _, want := range []string{`id="a~~name"`, `for="a~~name"`, `value="Ada" readonly`} {
		if !strings.Contains(a, want) {
			t.Errorf("a has no %s:\n%s", want, a)
		}
	}
	if !strings.Contains(b, `id="b~~name"`) || strings.Contains(b, `"a~~`) {
		t.Errorf("b's ids are not its own:\n%s", b)
	}
}

// A page shows every tab of a Tabs, each a section under its title, its
// content in it, as GitHub shows the same Markdown.
func TestPageTabs(t *testing.T) {
	r := pageRendition(t, "p", `{}`, `[{"id":"root","component":"Tabs","tabs":[{"title":"macOS","child":"mac"},{"title":"Linux","child":"linux"},{"title":"Source","child":"src"}]},
 {"id":"mac","component":"Text","text":"brew install x"},{"id":"linux","component":"Text","text":"apt install x"},{"id":"src","component":"Text","text":"go install x"}]`)
	h := r.Page()
	if n := count(h, `<section class="k-tab-section`); n != 3 {
		t.Fatalf("%d sections, want 3:\n%s", n, h)
	}
	last := 0
	for i, want := range [][2]string{{"macOS", "brew install x"}, {"Linux", "apt install x"}, {"Source", "go install x"}} {
		title := strings.Index(h, `class="k-tab k-tab-title">`+want[0]+`</div>`)
		body := strings.Index(h, want[1])
		if title < last || body < title {
			t.Errorf("tab %d: its title at %d, its content at %d, after %d", i, title, body, last)
		}
		last = body
	}
	if !strings.Contains(h, `aria-labelledby="p~~root/tab/1"`) || !strings.Contains(h, `id="p~~root/tab/1"`) {
		t.Errorf("the Linux section is not named by its title:\n%s", h)
	}
	for _, gone := range []string{`role="tab"`, `role="tablist"`, `data-on`, `<button`} {
		if strings.Contains(h, gone) {
			t.Errorf("a page's tabs have %s:\n%s", gone, h)
		}
	}
	// The controller's view still shows one tab, and its rendition on a
	// host is unchanged.
	if doc := r.Doc(); !strings.Contains(doc, `role="tab"`) || strings.Contains(doc, "apt install x") {
		t.Error("Page changed the host's document")
	}
}

// A HottyTree on a page has every node as a row, in lists that nest as
// the nodes do; a branch is a details element, open or closed as given,
// which the browser opens and closes.
func TestPageTree(t *testing.T) {
	r := pageRendition(t, "nav", `{"page":"install","open":["guides"]}`, `[{"id":"root","component":"HottyTree","catalogId":"HOTTY",
 "selected":{"@path":"/page"},"expanded":{"@path":"/open"},"items":[
  {"label":"Overview","value":"readme"},
  {"label":"Start","value":"start","children":[{"label":"Install","value":"install"},{"label":"Quick start","value":"quick"}]},
  {"label":"Guides","value":"guides","children":[{"label":"Configuration","value":"config"}]},
  {"label":"Reference","value":"ref","children":[{"label":"Commands","value":"commands","children":[{"label":"get","value":"get"}]}]}]}]`)
	h := r.Page()
	if n := count(h, `<li>`); n != 9 {
		t.Errorf("%d rows, want every node's 9:\n%s", n, h)
	}
	// Start holds the selection, Guides is expanded; Reference and
	// Commands are closed, their nodes still there.
	for _, want := range []string{
		`<details open><summary id="nav~~root~q1" class="k-node" style="--k-level: 0">`,
		`<details open><summary id="nav~~root~q4"`,
		`<details><summary id="nav~~root~q6"`,
		`<details><summary id="nav~~root~q7" class="k-node" style="--k-level: 1">`,
		`<span class="k-node-label">get</span>`,
		`<div id="nav~~root~q2" class="k-node k-sel" style="--k-level: 1"><span class="k-node-fold" aria-hidden="true"></span><span class="k-node-label" aria-current="true">Install</span></div>`,
		`<span class="k-node-count">1</span></summary>`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("no %s in:\n%s", want, h)
		}
	}
	if count(h, `<ul class="k-tree-list" role="list">`) != 5 {
		t.Errorf("the lists do not nest as the nodes do:\n%s", h)
	}
	for _, gone := range []string{`data-on`, `data-keys`, `tabindex`, `role="tree"`, `role="treeitem"`, `aria-expanded`, `▸`, `▾`} {
		if strings.Contains(h, gone) {
			t.Errorf("a page's tree has %s:\n%s", gone, h)
		}
	}
}

// A page's links are links: one that leaves the site opens in a new tab,
// any other in the same one, wherever it is (a Text's Markdown, a Video,
// a tree's node once it has an href); one to javascript: goes nowhere.
func TestPageLinks(t *testing.T) {
	r := pageRendition(t, "p", `{}`, `[{"id":"root","component":"Column","children":["t","v"]},
 {"id":"t","component":"Text","text":"[next](../quick/) [files](/docs/config/#files) [here](#top) [src](https://github.com/x/y?a=1&b=2) [host](//cdn.example.com/x) [mail](mailto:a@b.c) [bad](javascript:alert(1))"},
 {"id":"v","component":"Video","url":"https://example.com/talk.mp4"}]`)
	h := r.Page()
	for _, want := range []string{
		`<a href="../quick/">next</a>`,
		`<a href="/docs/config/#files">files</a>`,
		`<a href="#top">here</a>`,
		`<a href="https://github.com/x/y?a=1&amp;b=2" target="_blank" rel="noopener">src</a>`,
		`<a href="//cdn.example.com/x" target="_blank" rel="noopener">host</a>`,
		`<a href="mailto:a@b.c">mail</a>`,
		`<a>bad</a>`,
		`<a id="p~~v" class="k-media" href="https://example.com/talk.mp4" target="_blank" rel="noopener">`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("no %s in:\n%s", want, h)
		}
	}
	// A tree node's label with an href (KIT-26 wires view.TreeNode's in).
	for _, tc := range []struct {
		href     string
		selected bool
		want     string
	}{
		{"", false, `<span class="k-node-label">Install</span>`},
		{"", true, `<span class="k-node-label" aria-current="true">Install</span>`},
		{"../install/", true, `<a href="../install/" class="k-node-label" aria-current="page">Install</a>`},
		{"https://x.io", false, `<a href="https://x.io" target="_blank" rel="noopener" class="k-node-label">Install</a>`},
		{"javascript:x()", false, `<span class="k-node-label">Install</span>`},
	} {
		if got := treeLabel("Install", tc.href, tc.selected).html(); got != tc.want {
			t.Errorf("treeLabel(%q, %v) = %s, want %s", tc.href, tc.selected, got, tc.want)
		}
	}
}

// What a host's program makes work is left out of a page, and what it
// would show a piece at a time shows whole: controls show their state and
// take no input, a HottyForm submits nothing, a Modal is closed, a
// HottyKeyHints is not there, and a Table, a HottyList and a
// HottyPaginator show every row.
func TestPageStatic(t *testing.T) {
	r := pageRendition(t, "p", `{"name":"Ada","agree":true,"plan":["pro"],"cmd":"git"}`, `[{"id":"root","component":"HottyForm","catalogId":"HOTTY","child":"col"},
 {"id":"col","component":"Column","children":["name","agree","plan","cmd","go","modal","table","list","pager","hints"]},
 {"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"}},
 {"id":"agree","component":"CheckBox","label":"Agree","value":{"@path":"/agree"}},
 {"id":"plan","component":"ChoicePicker","label":"Plan","value":{"@path":"/plan"},"options":[{"label":"Free","value":"free"},{"label":"Pro","value":"pro"}]},
 {"id":"cmd","component":"TextField","label":"Command","value":{"@path":"/cmd"},"metadata":{"extensions":{"io_neuroplast_hotty":{"suggestions":["git commit","git checkout"]}}}},
 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go"}}},{"id":"go_t","component":"Text","text":"Go"},
 {"id":"modal","component":"Modal","trigger":"open","content":"inside"},{"id":"open","component":"Button","child":"open_t","action":{"event":{"name":"x"}}},
 {"id":"open_t","component":"Text","text":"Open"},{"id":"inside","component":"Text","text":"In the modal"},
 {"id":"table","component":"HottyTable","catalogId":"HOTTY","height":2,"columns":[{"key":"n","header":"N"}],"rows":[{"n":"one"},{"n":"two"},{"n":"three"}]},
 {"id":"list","component":"HottyList","catalogId":"HOTTY","height":1,"items":[{"label":"apple"},{"label":"pear"}]},
 {"id":"pager","component":"HottyPaginator","catalogId":"HOTTY","perPage":1,"child":"fruit"},
 {"id":"fruit","component":"Column","children":["f1","f2"]},{"id":"f1","component":"Text","text":"fig"},{"id":"f2","component":"Text","text":"kiwi"},
 {"id":"hints","component":"HottyKeyHints","catalogId":"HOTTY"}]`)
	r.C.St.Modal = "modal"
	r.C.Rebuild()
	h := r.Page()
	for _, want := range []string{
		`<div id="p~~root" class="k-form" role="form">`,
		`<input id="p~~name" class="k-input" type="text" value="Ada" readonly>`,
		`<input id="p~~agree" type="checkbox" checked tabindex="-1">`,
		`<span class="k-value">Pro</span>`,
		`<button id="p~~go" type="button" class="k-btn k-default" tabindex="-1">`,
		`Open`, `three`, `pear`, `fig`, `kiwi`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("no %s in:\n%s", want, h)
		}
	}
	for _, gone := range []string{`<form`, `type="submit"`, `datalist`, ` list=`, `In the modal`, `k-hints`, `k-dot`, `1–2 of 3`, `aria-haspopup`,
		`data-on`, `data-keys`, `data-steps`, `tabindex="0"`, `role="grid"`, `role="listbox"`, `role="option"`, `aria-selected`, `inert`} {
		if strings.Contains(h, gone) {
			t.Errorf("a page has %s:\n%s", gone, h)
		}
	}
}

func TestPageHref(t *testing.T) {
	for _, tc := range []struct {
		href     string
		h        string
		away, ok bool
	}{
		{"install.md", "install.md", false, true},
		{" ../quick/ ", "../quick/", false, true},
		{"?q=1", "?q=1", false, true},
		{"HTTPS://x.io", "HTTPS://x.io", true, true},
		{"//x.io/a", "//x.io/a", true, true},
		{"tel:+1", "tel:+1", false, true},
		{"JavaScript:x", "", false, false},
		{"data:text/html,x", "", false, false},
		{"", "", false, false},
		{"%zz", "", false, false},
	} {
		if h, away, ok := pageHref(tc.href); h != tc.h || away != tc.away || ok != tc.ok {
			t.Errorf("pageHref(%q) = %q %v %v, want %q %v %v", tc.href, h, away, ok, tc.h, tc.away, tc.ok)
		}
	}
}
