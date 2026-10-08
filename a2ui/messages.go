package a2ui

import (
	"encoding/json"
	"errors"
)

// Version is the A2UI protocol version this core speaks, as messages
// write it.
const Version = "v1.0"

// ActionMessage is an action for the agent: the user did what a component's
// event action names.
type ActionMessage struct {
	Name              string         `json:"name"`
	SurfaceID         string         `json:"surfaceId"`
	SourceComponentID string         `json:"sourceComponentId"`
	Timestamp         string         `json:"timestamp"`
	Context           map[string]any `json:"context"`
	UserMessage       string         `json:"userMessage,omitempty"`
}

// ErrorMessage is an error for the agent. Validation errors carry
// SurfaceID and Path; others SurfaceID or FunctionCallID.
type ErrorMessage struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	SurfaceID      string `json:"surfaceId,omitempty"`
	FunctionCallID string `json:"functionCallId,omitempty"`
	Path           string `json:"path,omitempty"`
}

// CallFunction is a function call on the wire, in callAgentFunction and
// callRendererFunction.
type CallFunction struct {
	Call    string         `json:"@call"`
	Catalog string         `json:"catalogId,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
}

// CallAgentFunctionMessage asks the agent to run one of its functions.
type CallAgentFunctionMessage struct {
	SurfaceID      string       `json:"surfaceId"`
	FunctionCallID string       `json:"functionCallId"`
	CallFunction   CallFunction `json:"callFunction"`
}

// FunctionResponse answers a function call: Value, or Error.
type FunctionResponse struct {
	FunctionCallID string         `json:"functionCallId"`
	Value          any            `json:"value,omitempty"`
	Error          *ResponseError `json:"error,omitempty"`
}

// MarshalJSON writes the response with "value" always, null included,
// unless it is an error: the schema requires one of them.
func (r FunctionResponse) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		return json.Marshal(struct {
			FunctionCallID string         `json:"functionCallId"`
			Error          *ResponseError `json:"error"`
		}{r.FunctionCallID, r.Error})
	}
	return json.Marshal(struct {
		FunctionCallID string `json:"functionCallId"`
		Value          any    `json:"value"`
	}{r.FunctionCallID, r.Value})
}

// ResponseError is a function's failure.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Outbound is one message from the renderer to the agent: exactly one of
// its fields is set.
type Outbound struct {
	Action                   *ActionMessage            `json:"action,omitempty"`
	CallAgentFunction        *CallAgentFunctionMessage `json:"callAgentFunction,omitempty"`
	RendererFunctionResponse *FunctionResponse         `json:"rendererFunctionResponse,omitempty"`
	Error                    *ErrorMessage             `json:"error,omitempty"`
}

// MarshalJSON writes the message with its version.
func (o Outbound) MarshalJSON() ([]byte, error) {
	type plain Outbound
	return json.Marshal(struct {
		Version string `json:"version"`
		plain
	}{Version, plain(o)})
}

// The categories of error, as A2UI's conformance suites name them.

// ValidationError is a message that does not have the shape A2UI says.
// Code is the wire error code to report, when it is not A2UI's default
// "VALIDATION_FAILED": a composition constraint violation is
// "UNALLOWED_PARENT" or "UNALLOWED_CHILD" (a2ui_protocol.md, Composition
// validation rules).
type ValidationError struct {
	Msg  string
	Path string
	Code string
}

func (e *ValidationError) Error() string { return "a2ui: " + e.Msg }

// IntegrityError is a message that does not fit the surfaces' state: an
// unknown surface, a duplicate id, a cycle, a missing root.
type IntegrityError struct{ Msg string }

func (e *IntegrityError) Error() string { return "a2ui: " + e.Msg }

// CatalogError is a catalog that cannot be used: unknown, or of another
// protocol version.
type CatalogError struct{ Msg string }

func (e *CatalogError) Error() string { return "a2ui: " + e.Msg }

// RecursionError is a message nested deeper than A2UI allows.
type RecursionError struct{ Msg string }

func (e *RecursionError) Error() string { return "a2ui: " + e.Msg }

// Code is the wire error code for an error the renderer reports.
func Code(err error) string {
	var (
		v *ValidationError
		i *IntegrityError
		c *CatalogError
		r *RecursionError
		d *DataError
		e *ExpressionError
		k *ReservedKeyError
	)
	switch {
	case errors.As(err, &r):
		return "RECURSION_ERROR"
	case errors.As(err, &i):
		return "INTEGRITY_ERROR"
	case errors.As(err, &v):
		if v.Code != "" {
			return v.Code
		}
		return "VALIDATION_FAILED"
	case errors.As(err, &c):
		return "CATALOG_ERROR"
	case errors.As(err, &d):
		return "DATA_ERROR"
	case errors.As(err, &e), errors.As(err, &k):
		return "EXPRESSION_ERROR"
	}
	return "EXECUTION_ERROR"
}
