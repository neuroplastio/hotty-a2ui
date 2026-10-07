// Package thirdparty holds A2UI at one commit (a2ui/REV, written by
// scripts/a2ui.sh) and embeds what the kit reads from it at run time: the
// v1.0 schemas the renderer validates against, the basic catalog and its
// examples. The conformance suites are read from disk by the tests.
package thirdparty

import "embed"

// CommonTypes is A2UI v1.0's common_types.json: the types catalogs and
// messages refer to.
//
//go:embed a2ui/specification/v1_0/json/common_types.json
var CommonTypes []byte

// AgentToRenderer is A2UI v1.0's agent_to_renderer.json: the messages an
// agent sends.
//
//go:embed a2ui/specification/v1_0/json/agent_to_renderer.json
var AgentToRenderer []byte

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
