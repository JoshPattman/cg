package cg

import (
	_ "embed"
)

// An event is somthing that happens that can trigger the agent to respond.
type Event interface {
	Kind() string
	Content() JsonObject
}

// A plugin is somthing that provides tools to the agent.
// A plugin can be loaded all-or-nothing, so if one tool fails none will be added.
// A plugin may also have an event channel (can be nil) that will be forwarded to the agent.
type Plugin interface {
	Name() string
	// Load the plugin, returning tools, a channel of events, a function to call to cleanup the plugin, and an error if the plugin failed to load.
	Load() ([]Tool, <-chan Event, func(), error)
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
	AddPlugin(Plugin)
	RemovePlugin(string) bool
	RemoveAllPlugins()
	Events() chan<- Event
	Run() error
	CleanStop()
}
