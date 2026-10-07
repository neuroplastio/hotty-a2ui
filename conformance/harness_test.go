package conformance

// Package conformance runs A2UI's conformance suites
// (third_party/a2ui/conformance, as its README.md describes them) against
// the core and the basic catalog. Each suite has its own test function;
// this file holds what they share.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"go.yaml.in/yaml/v3"
)

const conformanceDir = "../third_party/a2ui/conformance"

// suite is one conformance file's cases, each a generic JSON value: YAML
// read as JSON reads it, numbers as float64.
func suite(t *testing.T, name string) []map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(conformanceDir, "core", name))
	if err != nil {
		t.Fatal(err)
	}
	var raw any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var cases []map[string]any
	if err := json.Unmarshal(mustJSON(t, raw), &cases); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return cases
}

// asJSON makes v a generic JSON value: maps, slices, float64s.
func asJSON(t *testing.T, v any) any {
	t.Helper()
	var out any
	if err := json.Unmarshal(mustJSON(t, v), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func jsonEqual(t *testing.T, got, want any) bool {
	t.Helper()
	return reflect.DeepEqual(asJSON(t, got), asJSON(t, want))
}

// expectError checks err against a case's expect_error: {category,
// message}, the message a regular expression on the error's text.
func expectError(t *testing.T, err error, want map[string]any) {
	t.Helper()
	if err == nil {
		t.Fatalf("no error; want %v", want)
	}
	if m, ok := want["message"].(string); ok {
		if !regexp.MustCompile(m).MatchString(err.Error()) {
			t.Errorf("error %q does not match %q", err, m)
		}
	}
}

func str(v any) string { s, _ := v.(string); return s }
