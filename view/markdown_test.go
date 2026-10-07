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
