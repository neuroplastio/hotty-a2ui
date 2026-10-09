package view_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/highlight"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyCode is its code lexed as its language says, a line of tokens
// each; its line numbers count from startLine (1 without); a mark covers
// its line, or line to end, within the code, a later one over an earlier;
// it wraps unless wrap is false.
func TestCode(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"src":"a := 1\nb := 2\nc := 3\n"}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["go","plain"]},
	 {"id":"go","component":"HottyCode","catalogId":"` + hotty.ID + `","code":{"@path":"/src"},"language":"go","lineNumbers":true,"startLine":10,
	  "marks":[{"line":9,"end":10,"kind":"highlight"},{"line":11,"end":99,"kind":"added"},{"line":12,"kind":"error"}]},
	 {"id":"plain","component":"HottyCode","catalogId":"` + hotty.ID + `","code":"x\ty","wrap":false}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	e := c.V.Find("go")
	if e.Kind != view.Listing || e.Lang != "go" || !e.Numbers || e.FirstLine != 10 || !e.Wrap || len(e.Code) != 3 {
		t.Fatalf("go: %+v", e)
	}
	if got := e.Code[0]; !slices.Equal(got, []highlight.Token{{Text: "a := ", Kind: highlight.Plain}, {Text: "1", Kind: highlight.Literal}}) {
		t.Errorf("a := 1 lexed as %+v", got)
	}
	want := map[int]view.Mark{10: view.MarkHighlight, 11: view.MarkAdded, 12: view.MarkError}
	if !maps.Equal(e.Marks, want) {
		t.Errorf("marks %v, want %v", e.Marks, want)
	}
	pl := c.V.Find("plain")
	if pl.Wrap || pl.Numbers || pl.FirstLine != 1 || pl.Marks != nil || len(pl.Code) != 1 || pl.Code[0][0].Text != "x   y" {
		t.Errorf("plain: %+v", pl)
	}
}

// A fenced code block keeps its language; on a host it is goldmark's pre
// and code, with a span for each token that is not plain.
func TestFencedCode(t *testing.T) {
	src := "Run:\n\n```go\nfunc main() {}\n```\n\n    indented\n"
	var langs []string
	for _, b := range view.Markdown(src) {
		if b.Kind == view.CodeBlock {
			langs = append(langs, b.Lang)
		}
	}
	if strings.Join(langs, ",") != "go," {
		t.Errorf("code blocks' languages: %q", langs)
	}
	html := view.MarkdownHTML(src)
	for _, want := range []string{`<pre><code class="language-go"><span class="k-t-keyword">func</span> <span class="k-t-function">main</span>() {}` + "\n</code></pre>",
		"<pre><code>indented\n</code></pre>"} {
		if !strings.Contains(html, want) {
			t.Errorf("no %q in\n%s", want, html)
		}
	}
	if got := view.TokensHTML([]highlight.Token{{Text: "a<b", Kind: highlight.String}, {Text: " & ", Kind: highlight.Plain}}); got != `<span class="k-t-string">a&lt;b</span> &amp; ` {
		t.Errorf("escaping: %q", got)
	}
}
