package a2ui

import (
	"sort"
	"testing"
)

// TestConformanceDataModel runs data_model.yaml: each case builds a model
// from initial, watches each path in watch, and runs its steps.
func TestConformanceDataModel(t *testing.T) {
	for _, c := range suite(t, "data_model.yaml") {
		t.Run(str(c["name"]), func(t *testing.T) {
			m := NewDataModel(c["initial"])
			var fired []string
			watch, _ := c["watch"].([]any)
			for _, w := range watch {
				path := str(w)
				m.Watch(path, func() { fired = append(fired, path) })
			}
			for i, s := range c["steps"].([]any) {
				step := s.(map[string]any)
				fired = nil
				path := str(step["path"])
				var err error
				switch step["op"] {
				case "get":
					var v any
					var ok bool
					v, ok, err = m.Get(path)
					if err != nil {
						break
					}
					if step["expect_absent"] == true && ok {
						t.Errorf("step %d: get %s = %v, want absent", i, path, v)
					}
					if want, has := step["expect"]; has && (!ok || !jsonEqual(t, v, want)) {
						t.Errorf("step %d: get %s = %v (%v), want %v", i, path, v, ok, want)
					}
					switch step["expect_type"] {
					case "list":
						if _, is := v.([]any); !is {
							t.Errorf("step %d: %s is %T, want a list", i, path, v)
						}
					case "object":
						if _, is := v.(map[string]any); !is {
							t.Errorf("step %d: %s is %T, want an object", i, path, v)
						}
					}
				case "set":
					err = m.Set(path, step["value"])
				case "delete":
					err = m.Delete(path)
				case "dispose":
					m.Dispose()
				default:
					t.Fatalf("step %d: op %v", i, step["op"])
				}
				if want, has := step["expect_error"].(map[string]any); has {
					expectError(t, err, want)
					continue
				}
				if err != nil {
					t.Fatalf("step %d: %v", i, err)
				}
				if want, has := step["expect_notified"].([]any); has {
					w := make([]string, len(want))
					for j, x := range want {
						w[j] = str(x)
					}
					sort.Strings(w)
					sort.Strings(fired)
					if len(fired)+len(w) > 0 && !jsonEqual(t, fired, w) {
						t.Errorf("step %d: notified %v, want %v", i, fired, w)
					}
				}
				if want, has := step["expect_values"].(map[string]any); has {
					for p, v := range want {
						if got := m.Value(p); !jsonEqual(t, got, v) {
							t.Errorf("step %d: %s = %v, want %v", i, p, got, v)
						}
					}
				}
			}
		})
	}
}
