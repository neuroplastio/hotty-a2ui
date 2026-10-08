package view

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarkdown(t *testing.T) {
	src := "# Heading 1\n\nThis is **bold** text and *italic* text.\n\n- List item 1\n- List item 2\n  1. nested\n\n[Link to Google](https://google.com)\n\n> quoted `code`\n\n- [x] done"
	got := Markdown(src)
	b, _ := json.Marshal(got)
	want := `[{"kind":"h","level":1,"runs":[{"text":"Heading 1"}]},` +
		`{"kind":"p","runs":[{"text":"This is "},{"text":"bold","style":1},{"text":" text and "},{"text":"italic","style":2},{"text":" text."}]},` +
		`{"kind":"li","runs":[{"text":"List item 1"}]},` +
		`{"kind":"li","runs":[{"text":"List item 2"}]},` +
		`{"kind":"li","level":1,"number":1,"runs":[{"text":"nested"}]},` +
		`{"kind":"p","runs":[{"text":"Link to Google","href":"https://google.com"}]},` +
		`{"kind":"quote","runs":[{"text":"quoted "},{"text":"code","style":4}]},` +
		`{"kind":"li","task":2,"runs":[{"text":"done"}]}]`
	if string(b) != want {
		t.Errorf("got  %s\nwant %s", b, want)
	}
	if h := MarkdownHTML("[x](https://e.com) <b>no</b>"); !strings.Contains(h, `target="_blank"`) || strings.Contains(h, "<b>") {
		t.Errorf("html %s", h)
	}
}

func TestListNumbers(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"1. one", "<ol>\n<li><span class=\"k-n\">1.</span>one</li>\n</ol>"},
		{"3. three", "<ol start=\"3\">\n<li><span class=\"k-n\">3.</span>three</li>\n</ol>"},
		{"9. a\n10. b", "<li><span class=\"k-n\">10.</span>b</li>"},
		{"1. a\n\n2. b", "<li><span class=\"k-n\">2.</span>\n<p>b</p>\n</li>"},
		{"- a", "<ul>\n<li>a</li>\n</ul>"},
	} {
		if h := MarkdownHTML(c.src); !strings.Contains(h, c.want) {
			t.Errorf("%q: html %s, want %s", c.src, h, c.want)
		}
	}
}
