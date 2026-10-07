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
