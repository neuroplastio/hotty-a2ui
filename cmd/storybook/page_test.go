package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// -page prints one web page: the kit's stylesheet once, then each story's
// surfaces under its title, or a stream's.
func TestPrintPage(t *testing.T) {
	var b strings.Builder
	if err := printPage(&b, []string{"24_recipe-card", "hotty/tree"}, "", theme.Default); err != nil {
		t.Fatal(err)
	}
	h := b.String()
	if n := strings.Count(h, ".k-surface.k-page {"); n != 1 {
		t.Errorf("the sheet is in the page %d times", n)
	}
	for _, want := range []string{"<!doctype html>", "Recipe Card <code>basic/24_recipe-card</code>", `<div id="s0-gallery-recipe-card" class="k-surface k-page">`,
		"Ingredients", "Instructions", `<div id="s1-files" class="k-surface k-page">`, "<details"} {
		if !strings.Contains(h, want) {
			t.Errorf("no %s in the page", want)
		}
	}
	stream := filepath.Join(t.TempDir(), "a.jsonl")
	if err := os.WriteFile(stream, []byte(`{"version":"v1.0","createSurface":{"surfaceId":"x","catalogId":"https://a2ui.org/specification/v1_0/catalogs/basic/catalog.json"}}
{"version":"v1.0","updateComponents":{"surfaceId":"x","components":[{"id":"root","component":"Text","text":"[next](next/)"}]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	b.Reset()
	if err := printPage(&b, nil, stream, theme.Default); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), `<h2>a.jsonl</h2>`) || !strings.Contains(b.String(), `<a href="next/">next</a>`) {
		t.Errorf("the stream's page:\n%s", b.String())
	}
	if err := printPage(&b, []string{"hotty/tree"}, stream, theme.Default); err == nil {
		t.Error("-page took stories and a stream")
	}
}
