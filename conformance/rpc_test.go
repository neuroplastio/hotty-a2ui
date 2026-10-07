package conformance

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// TestConformanceRPC runs rpc_functions.yaml (handle_rpc in the brief):
// the agent's callRendererFunction against a catalog of the case's
// functions, with canned bodies; agentFunctionResponse; and the
// renderer's callAgentFunction, answered, timed out or duplicated.
func TestConformanceRPC(t *testing.T) {
	for _, c := range suite(t, "rpc_functions.yaml") {
		if c["action"] != "handle_rpc" {
			continue
		}
		t.Run(str(c["name"]), func(t *testing.T) {
			args, _ := c["args"].(map[string]any)
			exp, _ := c["expect"].(map[string]any)
			p := a2ui.NewProcessor(rpcCatalog(c, args, exp))
			var out []a2ui.Outbound
			p.Send = func(o a2ui.Outbound) { out = append(out, o) }
			activated := args["userActivationPresent"] == true
			p.Activated = func() bool { return activated }
			expErr, _ := c["expectError"].(map[string]any)

			switch {
			case args["message"] != nil && exp["error"] != nil:
				err := p.Process(args["message"])
				var v *a2ui.ValidationError
				if !errors.As(err, &v) {
					t.Fatalf("error %v, want a ValidationError", err)
				}
				want, _ := exp["error"].(map[string]any)
				if !strings.Contains(err.Error(), str(want["message"])) {
					t.Errorf("error %q does not say %q", err, want["message"])
				}
				for _, o := range out {
					if o.RendererFunctionResponse != nil {
						t.Errorf("answered %v", o.RendererFunctionResponse)
					}
				}

			case args["message"] != nil:
				if err := p.Process(args["message"]); err != nil {
					t.Fatal(err)
				}
				var responses []a2ui.Outbound
				for _, o := range out {
					if o.RendererFunctionResponse != nil {
						responses = append(responses, o)
					}
				}
				if exp["response"] == nil {
					if len(responses) > 0 {
						t.Errorf("answered %s", mustJSON(t, responses))
					}
					return
				}
				if len(responses) != 1 {
					t.Fatalf("%d responses, want 1", len(responses))
				}
				got := asJSON(t, responses[0]).(map[string]any)
				want := exp["response"].(map[string]any)
				if got["version"] != want["version"] {
					t.Errorf("version %v, want %v", got["version"], want["version"])
				}
				g := got["rendererFunctionResponse"].(map[string]any)
				w := want["rendererFunctionResponse"].(map[string]any)
				if g["functionCallId"] != w["functionCallId"] {
					t.Errorf("functionCallId %v, want %v", g["functionCallId"], w["functionCallId"])
				}
				if wv, ok := w["value"]; ok {
					if gv, ok := g["value"]; !ok || !jsonEqual(t, gv, wv) {
						t.Errorf("value %v (present %v), want %v", gv, ok, wv)
					}
				}
				if we, ok := w["error"].(map[string]any); ok {
					ge, _ := g["error"].(map[string]any)
					if ge == nil || ge["code"] != we["code"] {
						t.Errorf("error %v, want %v", ge, we)
					} else if m := str(we["message"]); !strings.Contains(str(ge["message"]), m) {
						t.Errorf("error message %q does not say %q", ge["message"], m)
					}
				}

			case args["outboundCall"] != nil && args["inboundResponse"] != nil:
				oc := args["outboundCall"].(map[string]any)
				call, err := p.CallAgentFunction(str(oc["surfaceId"]), callOf(oc), str(oc["functionCallId"]), timeoutOf(oc))
				if err != nil {
					t.Fatal(err)
				}
				if len(out) != 1 || out[0].CallAgentFunction == nil {
					t.Fatalf("sent %s, want one callAgentFunction", mustJSON(t, out))
				}
				sent := out[0].CallAgentFunction
				if sent.FunctionCallID != exp["correlatedCallId"] || sent.CallFunction.Call != callOf(oc).Call {
					t.Errorf("sent %s", mustJSON(t, sent))
				}
				in := args["inboundResponse"].(map[string]any)
				if id := in["agentFunctionResponse"].(map[string]any)["functionCallId"]; id != exp["correlatedCallId"] {
					t.Errorf("the response answers %v, want %v", id, exp["correlatedCallId"])
				}
				if err := p.Process(in); err != nil {
					t.Fatal(err)
				}
				v, err := call.Result()
				if err != nil {
					t.Fatal(err)
				}
				if !jsonEqual(t, v, exp["result"]) {
					t.Errorf("result %v, want %v", v, exp["result"])
				}

			case args["outboundCall"] != nil && expErr != nil:
				oc := args["outboundCall"].(map[string]any)
				call, err := p.CallAgentFunction(str(oc["surfaceId"]), callOf(oc), str(oc["functionCallId"]), timeoutOf(oc))
				if err != nil {
					t.Fatal(err)
				}
				if sc, ok := args["secondOutboundCall"].(map[string]any); ok {
					_, err = p.CallAgentFunction(str(sc["surfaceId"]), callOf(sc), str(sc["functionCallId"]), timeoutOf(sc))
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					_, err = call.Wait(ctx)
				}
				var re *a2ui.RPCError
				if !errors.As(err, &re) || re.Code != expErr["code"] {
					t.Errorf("error %v, want code %v", err, expErr["code"])
				}

			default:
				t.Fatal("a case this harness does not know")
			}
		})
	}
}

// rpcCatalog is the one catalog a handle_rpc case runs against: the
// case's functions with canned bodies, under the id and version the
// reference runners pick.
func rpcCatalog(c, args, exp map[string]any) *a2ui.Catalog {
	id := str(args["catalogId"])
	if id == "" {
		if m, ok := args["message"].(map[string]any); ok {
			if crf, ok := m["callRendererFunction"].(map[string]any); ok {
				cf, _ := crf["callFunction"].(map[string]any)
				resp, _ := exp["response"].(map[string]any)
				rfr, _ := resp["rendererFunctionResponse"].(map[string]any)
				e, _ := rfr["error"].(map[string]any)
				if !strings.Contains(str(e["message"]), "Catalog not found") {
					id = str(cf["catalogId"])
				}
			}
		}
	}
	if id == "" {
		if oc, ok := args["outboundCall"].(map[string]any); ok {
			cf, _ := oc["callFunction"].(map[string]any)
			id = str(cf["catalogId"])
		}
	}
	if id == "" {
		id = "basic"
	}
	version := str(args["catalogVersion"])
	if version == "" {
		version = caseVersion(c)
	}
	if version == "" {
		version = a2ui.Version
	}
	cat := &a2ui.Catalog{ID: id, ProtocolVersion: version, Components: map[string]*a2ui.ComponentType{}, Functions: map[string]*a2ui.Function{}}
	meta, _ := args["functionMetadata"].(map[string]any)
	for name, m := range meta {
		m, _ := m.(map[string]any)
		f := &a2ui.Function{
			Name:                   name,
			AllowedCallers:         str(m["allowedCallers"]),
			RequiresUserActivation: m["requiresUserActivation"] == true,
			Impl:                   cannedBody(name),
		}
		if f.AllowedCallers == "" {
			f.AllowedCallers = a2ui.RendererOrAgent
		}
		params := m["schema"]
		if params == nil {
			params = m["parameters"]
		}
		if params != nil {
			f.Schema = map[string]any{"properties": map[string]any{"args": params}}
		}
		cat.Functions[name] = f
	}
	return cat
}

// cannedBody is what a function of the suite does, by its name.
func cannedBody(name string) a2ui.FuncImpl {
	return func(_ *a2ui.Context, args map[string]any) (any, error) {
		switch name {
		case "playMedia":
			return map[string]any{"playing": true, "timestamp": 0.0}, nil
		case "openExternalUrl":
			return map[string]any{"opened": true}, nil
		case "failingFunction":
			return nil, errors.New("An error occurred during function execution.")
		case "calculateTax":
			amount, _ := args["amount"].(float64)
			return amount * 0.1, nil
		}
		return nil, nil
	}
}

func callOf(oc map[string]any) a2ui.CallFunction {
	cf, _ := oc["callFunction"].(map[string]any)
	args, _ := cf["args"].(map[string]any)
	return a2ui.CallFunction{Call: str(cf["@call"]), Catalog: str(cf["catalogId"]), Args: args}
}

func timeoutOf(oc map[string]any) time.Duration {
	if ms, ok := oc["timeoutMs"].(float64); ok {
		return time.Duration(ms * float64(time.Millisecond))
	}
	return 0
}
