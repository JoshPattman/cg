package runnertools

import (
	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/files"
)

func Plugin(fs files.FileSystem) cg.Plugin {
	events := make(chan cg.Event)
	return &runnerPlugin{
		[]cg.Tool{
			&doneTool{},
			&reminderTool{events},
			&listDirectoryTool{fs},
			&readFileTool{fs, 10000},
			&modifyFileTool{fs},
			&deleteFileTool{fs},
		},
		events,
	}
}

type runnerPlugin struct {
	tools  []cg.Tool
	events <-chan cg.Event
}

func (r *runnerPlugin) Name() string {
	return "internal"
}

func (r *runnerPlugin) Removable() bool {
	return false
}

func (r *runnerPlugin) Tools() []cg.Tool {
	return r.tools
}

func (r *runnerPlugin) Events() <-chan cg.Event {
	return r.events
}
