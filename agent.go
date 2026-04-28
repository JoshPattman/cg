package cg

import (
	_ "embed"
	"iter"
)

// An event is somthing that happens that can trigger the agent to respond.
type Event interface {
	Kind() string
	Content() JsonObject
}

type Plugin interface {
	// Unique name of this loaded plugin
	Name() string
	// Once added to the agent, can this plugin ever be removed again?
	Removable() bool
	// What tools does this plugin provide?
	Tools() []Tool
	// What is the channel (if any, otherwise nil) that events from this plugin come in on?
	Events() <-chan Event
}

// A tool is somthing the agent can call to perform an action.
type Tool interface {
	Def() ToolDef
	Call(map[string]any) (string, error)
}

// A tool definition specifies how a tool should be used.
type ToolDef struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// An agent can run (blocking) and respond to events with tool calls.
type Agent interface {
	AddPlugin(Plugin) error
	RemovePlugin(string) error
	AllPlugins() iter.Seq[Plugin]
	Events() chan<- Event
	Run() error
	CleanStop()
}
