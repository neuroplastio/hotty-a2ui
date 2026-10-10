package conformance

import (
	"encoding/json"
	"io/fs"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/a2ui/schema"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
)

// TestBasicExamples processes each of the basic catalog's examples as a
// renderer does, not strictly (00_incremental streams its components
// before their parent has them; 31_incremental-dashboard's ids have
// hyphens, which strict mode's identifier rule refuses): none may be
// refused, and none may report an error.
func TestBasicExamples(t *testing.T) {
	names, err := fs.Glob(thirdparty.BasicExamples, "a2ui/catalogs/basic/v1/examples/*.json")
	if err != nil || len(names) == 0 {
		t.Fatalf("examples: %v", err)
	}
	for _, name := range names {
		t.Run(name[len("a2ui/catalogs/basic/v1/examples/"):], func(t *testing.T) {
			b, _ := fs.ReadFile(thirdparty.BasicExamples, name)
			var ex struct{ Messages []any }
			if err := json.Unmarshal(b, &ex); err != nil {
				t.Fatal(err)
			}
			p := schema.NewProcessor(basic.Catalog())
			var reported []string
			p.Send = func(o a2ui.Outbound) {
				if o.Error != nil {
					reported = append(reported, o.Error.Message)
				}
			}
			if err := p.Process(ex.Messages); err != nil {
				t.Fatal(err)
			}
			if len(p.Surfaces()) == 0 {
				t.Fatal("no surface")
			}
			for _, m := range reported {
				t.Errorf("reported: %s", m)
			}
		})
	}
}
