package a2ui

// Validator checks what an agent sends against A2UI's JSON schemas: each
// message's envelope, each component's definition against its catalog's,
// and each call of a catalog's function. Each returns a *ValidationError
// that says what failed and where.
//
// A Processor without one, as NewProcessor makes it, still checks what it
// needs to apply a message (one action, its surface, its components' types
// and references, reserved keys, nesting) and trusts the rest to its
// agent: for a program that is its own agent, which then links neither
// the schemas nor jsonschema. A renderer whose agent is another program
// sets one; package a2ui/schema has A2UI's.
type Validator interface {
	// Message checks a v1.0 message whose action is action; what names it
	// in the error ("Invalid v1.0 message: message 2 (createSurface)").
	Message(action string, m map[string]any, what string) error
	// Component checks a component's definition against its catalog's.
	Component(c *Catalog, def map[string]any) error
	// Call checks a call, {"@call", "args", …}, of the catalog's function
	// name.
	Call(c *Catalog, name string, call map[string]any) error
}
