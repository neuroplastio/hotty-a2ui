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
