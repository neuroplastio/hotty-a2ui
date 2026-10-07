package a2ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// DefaultAgentTimeout is how long a call of an agent's function that
// evaluation makes waits for the agent's answer.
const DefaultAgentTimeout = 30 * time.Second

// RPCError is a call of an agent's function that failed: the agent's
// error, or the renderer's TIMEOUT or DUPLICATE.
type RPCError struct{ Code, Message string }

func (e *RPCError) Error() string { return "a2ui: " + e.Code + ": " + e.Message }

// AgentCall is a call of one of the agent's functions, a callAgentFunction,
// until the agent answers it with an agentFunctionResponse.
type AgentCall struct {
	ID        string
	SurfaceID string
	Function  CallFunction

	// cached: evaluation made it, and keeps its result for the same
	// arguments.
	cached bool
	done   chan struct{}
	value  any
	err    error
	timer  *time.Timer
}

// Done is closed when the call has its result.
func (c *AgentCall) Done() <-chan struct{} { return c.done }

// Result is the call's result, once Done is closed.
func (c *AgentCall) Result() (any, error) {
	select {
	case <-c.done:
		return c.value, c.err
	default:
		return nil, fmt.Errorf("a2ui: call %s is pending", c.ID)
	}
}

// Wait waits for the call's result.
func (c *AgentCall) Wait(ctx context.Context) (any, error) {
	select {
	case <-c.done:
		return c.value, c.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// agentCalls are the calls of agent functions that wait for an answer,
// and the results of those evaluation made, by their arguments. Timers
// finish calls on their own goroutines, so this is locked.
type agentCalls struct {
	mu      sync.Mutex
	pending map[string]*AgentCall
	results map[string]*AgentCall
	seq     int
}

func (a *agentCalls) start(surfaceID string, f CallFunction, id string, timeout time.Duration, cached bool) (*AgentCall, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pending == nil {
		a.pending = map[string]*AgentCall{}
	}
	if id == "" {
		a.seq++
		id = "fn-" + strconv.Itoa(a.seq)
		for a.pending[id] != nil {
			a.seq++
			id = "fn-" + strconv.Itoa(a.seq)
		}
	}
	if a.pending[id] != nil {
		return nil, &RPCError{Code: "DUPLICATE", Message: fmt.Sprintf("a call with functionCallId '%s' is already pending", id)}
	}
	c := &AgentCall{ID: id, SurfaceID: surfaceID, Function: f, cached: cached, done: make(chan struct{})}
	a.pending[id] = c
	if timeout > 0 {
		c.timer = time.AfterFunc(timeout, func() {
			if a.take(id, c) {
				c.finish(nil, &RPCError{Code: "TIMEOUT", Message: fmt.Sprintf("the agent did not answer call '%s' (%s) within %s", id, f.Call, timeout)})
			}
		})
	}
	return c, nil
}

// take removes a pending call: whoever takes it finishes it.
func (a *agentCalls) take(id string, c *AgentCall) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pending[id] != c || c == nil {
		return false
	}
	delete(a.pending, id)
	return true
}

func (c *AgentCall) finish(v any, err error) {
	if c.timer != nil {
		c.timer.Stop()
	}
	c.value, c.err = v, err
	close(c.done)
}

// answer finishes the call an agentFunctionResponse answers: nil when
// none waits for it.
func (a *agentCalls) answer(body map[string]any) *AgentCall {
	id := str(body["functionCallId"])
	a.mu.Lock()
	c := a.pending[id]
	a.mu.Unlock()
	if !a.take(id, c) {
		return nil
	}
	if e, ok := body["error"].(map[string]any); ok {
		c.finish(nil, &RPCError{Code: str(e["code"]), Message: str(e["message"])})
	} else {
		c.finish(body["value"], nil)
	}
	return c
}

// cachedCall finds the call evaluation made with these arguments.
func (a *agentCalls) cachedCall(key string) *AgentCall {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.results[key]
}

func (a *agentCalls) cache(key string, c *AgentCall) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.results == nil {
		a.results = map[string]*AgentCall{}
	}
	a.results[key] = c
}

// forget drops what evaluation keeps for a deleted surface.
func (a *agentCalls) forget(surfaceID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for k, c := range a.results {
		if c.SurfaceID == surfaceID {
			delete(a.results, k)
		}
	}
}

// CallAgentFunction calls one of the agent's functions: the
// callAgentFunction goes to the agent now, and the call finishes when the
// agent answers, or fails with TIMEOUT after timeout (none when it is not
// positive). id is the call's functionCallId, made up when empty; one
// already pending is a DUPLICATE.
func (p *Processor) CallAgentFunction(surfaceID string, f CallFunction, id string, timeout time.Duration) (*AgentCall, error) {
	c, err := p.calls.start(surfaceID, f, id, timeout, false)
	if err != nil {
		return nil, err
	}
	p.send(Outbound{CallAgentFunction: &CallAgentFunctionMessage{SurfaceID: surfaceID, FunctionCallID: c.ID, CallFunction: f}})
	return c, nil
}

// agentFunction is evaluation's call of an agent function: its result
// once the agent answers, nil until then. The answer resolves the
// surface again.
func (p *Processor) agentFunction(s *Surface, f *Function, catalog string, args map[string]any) (any, error) {
	b, _ := json.Marshal(args)
	key := s.ID + "\x00" + catalog + "\x00" + f.Name + "\x00" + string(b)
	if c := p.calls.cachedCall(key); c != nil {
		select {
		case <-c.done:
			return c.value, c.err
		default:
			return nil, nil
		}
	}
	call := CallFunction{Call: f.Name, Catalog: catalog, Args: args}
	c, err := p.calls.start(s.ID, call, "", DefaultAgentTimeout, true)
	if err != nil {
		return nil, err
	}
	p.calls.cache(key, c)
	p.send(Outbound{CallAgentFunction: &CallAgentFunctionMessage{SurfaceID: s.ID, FunctionCallID: c.ID, CallFunction: call}})
	return nil, nil
}
