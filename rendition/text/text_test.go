package text_test

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
	"github.com/neuroplastio/hotty-a2ui/view"
)

func TestLoginForm(t *testing.T) {
	b, _ := fs.ReadFile(thirdparty.BasicExamples, "a2ui/catalogs/basic/v1/examples/00_simple-login-form.json")
	var ex struct{ Messages []any }
	_ = json.Unmarshal(b, &ex)
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	if err := p.Process(ex.Messages); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surfaces()[0])
	_ = c.SetValue(c.V.Root.Children[1].ID, "ada")
	_ = c.SetValue(c.V.Root.Children[2].ID, "secret")
	got := text.Render(c.V)
	want := "Login\nUsername: ada\nPassword: ••••••\n[ Sign In ]\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if strings.Contains(got, "secret") {
		t.Error("the password shows")
	}
}

// TestProgressAndSpinner: a bar reads as its percentage, or "…" while it
// has none; a spinner reads as "…" while it spins.
func TestProgressAndSpinner(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["p","i","s","x"]},
	 {"id":"p","component":"HottyProgress","catalogId":"` + hotty.ID + `","label":"Download","value":0.25},
	 {"id":"i","component":"HottyProgress","catalogId":"` + hotty.ID + `","label":"Indexing"},
	 {"id":"s","component":"HottySpinner","catalogId":"` + hotty.ID + `","label":"Working"},
	 {"id":"x","component":"HottySpinner","catalogId":"` + hotty.ID + `","label":"Idle","active":false}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	if want := "Download: 25%\nIndexing: …\nWorking: …\nIdle\n"; got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestTable: a table reads as its header and every row, aligned, "> "
// on the selected one, whatever its height.
func TestTable(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyTable","catalogId":"` + hotty.ID + `","columns":[{"key":"c","header":"City"},{"key":"p","header":"Pop","align":"end"}],
	  "rows":[{"c":"Tokyo","p":37},{"c":"São Paulo","p":22},{"c":"東京","p":1}],"rowKey":"c","selected":"São Paulo","height":1}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	want := "  City       Pop\n  Tokyo       37\n> São Paulo   22\n  東京         1\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestTree: a tree reads as every node it shows, whatever its height, as
// cells draws them, "> " on the selected one.
func TestTree(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyTree","catalogId":"` + hotty.ID + `","selected":"b","height":1,"items":[
	  {"label":"src","children":[{"label":"a.go"},{"label":"pkg","children":[{"label":"b.go","value":"b"}]}]},
	  {"label":"docs","children":[{"label":"x.md"}]},{"label":"README.md"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	want := "  ▼ src\n  ├── a.go\n  └── ▼ pkg\n>     └── b.go\n  ▶ docs 1\n    README.md\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestSwitch: a switch reads as a field does, "Label: on" or "off",
// "(disabled)" after a disabled one, its error on the next line.
func TestSwitch(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"tfa":false}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["wifi","lock","bare","tfa"]},
	 {"id":"wifi","component":"HottySwitch",` + h + `,"label":"Wi-Fi","value":true},
	 {"id":"lock","component":"HottySwitch",` + h + `,"label":"Locked","value":false,"disabled":true},
	 {"id":"bare","component":"HottySwitch",` + h + `,"value":true},
	 {"id":"tfa","component":"HottySwitch",` + h + `,"label":"Two-factor","value":{"@path":"/tfa"},
	  "checks":[{"condition":{"@path":"/tfa"},"message":"Required"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	c.St.Touched["tfa"] = true
	c.Rebuild()
	got := text.Render(c.V)
	want := "Wi-Fi: on\nLocked: off (disabled)\non\nTwo-factor: off\n✗ Required\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestScrollView: a scroll view reads as all of its content, whatever its
// height: its lines, or its child.
func TestScrollView(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["log","doc"]},
	 {"id":"log","component":"HottyScrollView","catalogId":"` + hotty.ID + `","lines":["one","two","three"],"height":1,"follow":true},
	 {"id":"doc","component":"HottyScrollView","catalogId":"` + hotty.ID + `","child":"t","height":1},
	 {"id":"t","component":"Text","text":"# Title\n\nA paragraph."}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	if want := "one\ntwo\nthree\nTitle\nA paragraph.\n"; got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestCode: code reads as it is, after its line numbers when they show
// and its marks' signs when it has any.
func TestCode(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["a","b"]},
	 {"id":"a","component":"HottyCode","catalogId":"` + hotty.ID + `","code":"one\n\ttwo\nthree","lineNumbers":true,"startLine":9,
	  "marks":[{"line":10,"kind":"added"},{"line":11,"kind":"highlight"}]},
	 {"id":"b","component":"HottyCode","catalogId":"` + hotty.ID + `","code":"plain"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	if want := " 9   one\n10 +     two\n11 > three\nplain\n"; got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestDiff: a HottyDiff reads as a unified diff, which patch reads: each
// named file's --- and +++ lines, its hunks' headers and their lines,
// signed, and git's line for a binary file; a split one too.
func TestDiff(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["a","b"]},
	 {"id":"a","component":"HottyDiff","catalogId":"` + hotty.ID + `","view":"split",
	  "patch":"--- a/x.go\n+++ b/x.go\n@@ -1,2 +1,2 @@ func f() {\n-\treturn a\n+\treturn b\n }\nBinary files a/i.png and b/i.png differ\n"},
	 {"id":"b","component":"HottyDiff","catalogId":"` + hotty.ID + `","file":"n.txt","old":"","new":"hi\n"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	want := "--- a/x.go\n+++ b/x.go\n@@ -1,2 +1,2 @@ func f() {\n-    return a\n+    return b\n }\n" +
		"Binary files a/i.png and b/i.png differ\n" +
		"--- a/n.txt\n+++ b/n.txt\n@@ -0,0 +1 @@\n+hi\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// TestChart: a HottyChart reads as a table of its values, a row a point,
// its label first (else its number), a column a series under its label,
// the numbers aligned to the end and a missing one blank; "No data" with
// none. A HottySparkline reads as its values, "–" where one is missing.
func TestChart(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"mem":[40,null,1250.5]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["a","b","c","d"]},
	 {"id":"a","component":"HottyChart","catalogId":"` + hotty.ID + `","labels":["10:00","10:01","10:02"],
	  "series":[{"label":"cpu","values":[12,-8,61]},{"label":"memory","values":{"@path":"/mem"}}]},
	 {"id":"b","component":"HottyChart","catalogId":"` + hotty.ID + `","kind":"bar","values":[3,4]},
	 {"id":"c","component":"HottyChart","catalogId":"` + hotty.ID + `","values":{"@path":"/none"}},
	 {"id":"d","component":"HottySparkline","catalogId":"` + hotty.ID + `","values":[1,null,2.5,-3]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	got := text.Render(view.NewController(p.Surface("s")).V)
	want := "       cpu  memory\n" +
		"10:00   12      40\n" +
		"10:01   -8\n" +
		"10:02   61  1250.5\n" +
		"   value\n" +
		"1      3\n" +
		"2      4\n" +
		"No data\n" +
		"1 – 2.5 -3\n"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
