// Package thirdparty holds A2UI at one commit (a2ui/REV, written by
// scripts/a2ui.sh) and embeds what the kit reads from it at run time: the
// basic catalog and its examples. The conformance suites are read from disk
// by the tests.
package thirdparty

import "embed"

// BasicCatalog is A2UI's basic catalog, v1.
//
//go:embed a2ui/catalogs/basic/v1/catalog.json
var BasicCatalog []byte

// BasicExamples are the basic catalog's examples, as examples/*.json.
//
//go:embed a2ui/catalogs/basic/v1/examples/*.json
var BasicExamples embed.FS

// Rev is the A2UI commit vendored here, with its date and subject.
//
//go:embed a2ui/REV
var Rev string
