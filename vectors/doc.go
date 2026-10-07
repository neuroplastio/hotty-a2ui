// Package vectors holds the vectors that check the renditions against one
// another: keys.yaml, keys and focus, run against the cells rendition and
// against the HTML rendition on a host (NEIO-11).
package vectors

import _ "embed"

// Keys is keys.yaml.
//
//go:embed keys.yaml
var Keys []byte
