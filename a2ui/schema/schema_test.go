package schema_test

import (
	"errors"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/a2ui/schema"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
)

// A processor with the schemas refuses a component its catalog's schema
// does not allow; one without them, a program's own agent, takes it.
func TestValidatorOptional(t *testing.T) {
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Text","text":"Hi","nonsense":1}]}}]`
	var ve *a2ui.ValidationError
	if err := schema.NewProcessor(basic.Catalog()).ProcessJSON([]byte(msgs)); !errors.As(err, &ve) {
		t.Errorf("with the schemas: %v, want a ValidationError", err)
	}
	if err := a2ui.NewProcessor(basic.Catalog()).ProcessJSON([]byte(msgs)); err != nil {
		t.Errorf("without them: %v", err)
	}
}
