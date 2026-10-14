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

// A DateTimeInput's min and max are a date, a time or both, in ISO 8601's
// forms, which the schema tells apart by format (plexos F146): each is
// one of them, and what is none is refused.
func TestDateBounds(t *testing.T) {
	for _, tc := range []struct {
		min string
		ok  bool
	}{
		{"2026-10-14", true},
		{"09:00", true},
		{"09:00:30.5", true},
		{"2026-10-14T09:00", true},
		{"2026-10-14T09:00:00Z", true},
		{"2026-10-14T09:00+02:00", true},
		{"14/10/2026", false},
		{"2026-02-30", false},
		{"24:00", false},
		{"9:00", false},
		{"2026-10-14 09:00", false},
	} {
		msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
		{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
		 {"id":"root","component":"DateTimeInput","label":"When","value":"","enableDate":true,"enableTime":true,"min":"` + tc.min + `"}]}}]`
		err := schema.NewProcessor(basic.Catalog()).ProcessJSON([]byte(msgs))
		if (err == nil) != tc.ok {
			t.Errorf("min %q: %v, want ok %v", tc.min, err, tc.ok)
		}
	}
}
