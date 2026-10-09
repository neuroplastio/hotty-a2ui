package storybook

import "testing"

// TestPretty: what fits on its line stays on it, spaced; what doesn't is
// a member a line, two spaces deeper. Keys keep their order, strings
// their colons, commas and escapes.
func TestPretty(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{`{"b":1,"a":[1,2]}`, `{"b": 1, "a": [1, 2]}`},
		{`"a,b:c"`, `"a,b:c"`},
		{`{"s":"x\"y,z:\u00e9 <&>"}`, `{"s": "x\"y,z:\u00e9 <&>"}`},
		{`{}`, `{}`},
		{` [ ] `, `[]`},
		{`not json`, `not json`},
		{
			`{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[{"id":"root","component":"Column","children":["title","note"]},{"id":"title","component":"Text","text":"## A note"}]}}`,
			`{
  "version": "v1.0",
  "updateComponents": {
    "surfaceId": "s",
    "components": [
      {"id": "root", "component": "Column", "children": ["title", "note"]},
      {"id": "title", "component": "Text", "text": "## A note"}
    ]
  }
}`,
		},
		// The key counts toward the line: this one pushes its object over.
		{
			`{"aVeryLongKeyThatTakesUpMostOfTheLineAlready":{"x":"0123456789","y":"0123456789"}}`,
			`{
  "aVeryLongKeyThatTakesUpMostOfTheLineAlready": {
    "x": "0123456789",
    "y": "0123456789"
  }
}`,
		},
	} {
		if got := pretty([]byte(c.in)); got != c.want {
			t.Errorf("pretty(%s):\n%s\nwant\n%s", c.in, got, c.want)
		}
	}
}
