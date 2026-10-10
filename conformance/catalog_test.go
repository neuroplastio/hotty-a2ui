package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/a2ui/schema"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
)

// TestConformanceCommonTypes runs common_types.yaml's v1.0
// validate_common_type cases: each value against a common_types.json
// definition, with the schema as the spec publishes it.
func TestConformanceCommonTypes(t *testing.T) {
	for _, c := range suite(t, "common_types.yaml") {
		if c["action"] != "validate_common_type" || c["protocolVersion"] != a2ui.Version {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			steps, _ := c["steps"].([]any)
			for i, st := range steps {
				step := st.(map[string]any)
				err := schema.CheckCommonType(str(c["definition"]), step["value"])
				if want, ok := step["expectError"].(map[string]any); ok {
					checkError(t, "value "+string(mustJSON(t, step["value"])), err, want)
				} else if err != nil {
					t.Errorf("step %d: %s: %v", i, mustJSON(t, step["value"]), err)
				}
			}
		})
	}
}

// catalogSkips are the v1.0 from_json cases about an SDK's catalog API,
// not the catalog's meaning.
var catalogSkips = map[string]string{
	"test_v10_catalog_from_json_no_theme":                "the Python SDK's catalog model (a theme that v1.0 has not)",
	"test_v10_published_basic_catalog_is_self_contained": "an SDK inlining common types into the catalog; this core refers to them",
}

// TestConformanceCatalog runs catalog.yaml's v1.0 from_json cases: the
// catalog reads (its id, version, components and functions as written),
// its entities are UAX #31 identifiers, and the published basic catalog
// accepts and refuses the components the cases list.
func TestConformanceCatalog(t *testing.T) {
	for _, c := range suite(t, "catalog.yaml") {
		doc, _ := c["catalog"].(map[string]any)
		version := caseVersion(c)
		if c["action"] != "from_json" || version != a2ui.Version {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			if why, ok := catalogSkips[str(c["name"])]; ok {
				t.Skip(why)
			}
			var cat *a2ui.Catalog
			var err error
			v := &schema.Validator{}
			if strings.HasSuffix(str(c["catalogPath"]), "catalogs/basic/v1/catalog.json") {
				b, rerr := os.ReadFile(filepath.Join(conformanceDir, "..", str(c["catalogPath"])))
				if rerr != nil {
					t.Fatal(rerr)
				}
				cat, err = a2ui.ParseCatalog(b)
				if err == nil {
					cat = basic.Catalog()
				}
			} else {
				d := asJSON(t, doc).(map[string]any)
				if id := str(c["catalogId"]); id != "" {
					d["catalogId"] = id
				}
				if str(d["protocolVersion"]) == "" {
					d["protocolVersion"] = version
				}
				cat, err = a2ui.ParseCatalog(mustJSON(t, d))
			}
			if want, ok := c["expectError"].(map[string]any); ok {
				checkError(t, "parse", err, want)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			exp, _ := c["expect"].(map[string]any)
			for _, k := range []string{"catalogId", "protocolVersion", "components", "functions"} {
				if w, ok := exp[k]; ok && !jsonEqual(t, cat.Doc[k], w) {
					t.Errorf("%s %v, want %v", k, cat.Doc[k], w)
				}
			}
			valid, _ := exp["validComponents"].([]any)
			for _, d := range valid {
				if err := v.Component(cat, d.(map[string]any)); err != nil {
					t.Errorf("%s: %v", mustJSON(t, d), err)
				}
			}
			invalid, _ := exp["invalidComponents"].([]any)
			for _, d := range invalid {
				if err := v.Component(cat, d.(map[string]any)); err == nil {
					t.Errorf("%s is accepted", mustJSON(t, d))
				}
			}
		})
	}
}

// outOfScope are the suites and actions this core does not run, and why.
// TestConformanceCoverage fails on any other, so that a new A2UI pin
// shows what it adds.
var outOfScope = map[string]string{
	"message_processor_v0_8.yaml":                     "A2UI v0.8; this core speaks v1.0",
	"message_processor_v0_9.yaml":                     "A2UI v0.9",
	"validator_v0_8.yaml":                             "A2UI v0.8",
	"validator_v0_9.yaml":                             "A2UI v0.9",
	"agent_to_renderer.yaml agent_to_renderer_schema": "an SDK's schema generated from its models; this core embeds the spec's",
	"common_types.yaml common_types_schema":           "an SDK's schema generated from its models; this core embeds the spec's",
	"catalog.yaml catalog_schema":                     "an SDK serialising its catalog model",
	"accessibility.yaml accessibility_check":          "a renderer's accessibility tree; the HOTTY renditions' business, not the core's",
}

// covered are the suites and actions the tests here run.
var covered = map[string]bool{
	"data_model.yaml data_model":                            true,
	"expressions.yaml parse_expression_template":            true,
	"functions.yaml evaluate_function":                      true,
	"node_resolution.yaml resolve_nodes":                    true,
	"actions.yaml dispatch_action":                          true,
	"data_context.yaml resolve_path":                        true,
	"rpc_functions.yaml handle_rpc":                         true,
	"multi_catalog.yaml select_catalog":                     true,
	"common_types.yaml validate_common_type":                true,
	"catalog.yaml from_json":                                true,
	"message_processor_v1_0.yaml get_renderer_capabilities": true,
}

func TestConformanceCoverage(t *testing.T) {
	for _, s := range processorSuites {
		for _, a := range []string{"process_messages", "validate"} {
			covered[s+" "+a] = true
		}
	}
	files, err := filepath.Glob(filepath.Join(conformanceDir, "core", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		name := filepath.Base(f)
		if _, ok := outOfScope[name]; ok {
			continue
		}
		for _, c := range suite(t, name) {
			key := name + " " + str(c["action"])
			if _, ok := outOfScope[key]; !ok && !covered[key] {
				t.Errorf("%s (%s): no test runs it", key, c["name"])
			}
		}
	}
}
