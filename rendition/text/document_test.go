package text_test

import (
	"encoding/json"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// TestDocument: a HottyMarkdown reads as its blocks, a blank line apart
// but a tight list's items, each line after its frames; an alert as its
// mark and title, then its lines; a table's columns aligned; HTML and
// comments nothing; an image its alt text.
func TestDocument(t *testing.T) {
	md, _ := json.Marshal("# Guide\n\nSee [usage](#usage).\n\n> [!TIP]\n> Use **Tab**.\n\n- one\n  - nested\n- [x] done\n\n" +
		"| Name | Size |\n| :-- | --: |\n| a.go | 10 |\n| main.go | 200 |\n\n<!-- docs:figure -->\n![Logo](logo.png)\n\n```sh\nmake\nmake check\n```\n\n---\n\n## Usage")
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyMarkdown","catalogId":"` + hotty.ID + `","text":` + string(md) + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surfaces()[0]).V)
	want := `Guide

See usage.

▎ ✓ Tip
▎ Use Tab.

• one
  • nested
✓ done

Name     Size
a.go       10
main.go   200

[image: Logo]

make
make check

───

Usage
`
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
