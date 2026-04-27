package runnertools

import (
	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/files"
)

func Plugin(events chan<- cg.Event, fs files.FileSystem) cg.Plugin {
	return runnerPlugin{events: events, fs: fs}
}

type runnerPlugin struct {
	events chan<- cg.Event
	fs     files.FileSystem
}

func (r runnerPlugin) Load() ([]cg.Tool, <-chan cg.Event, func(), error) {
	return []cg.Tool{
		&doneTool{},
		&reminderTool{r.events},
		&listDirectoryTool{r.fs},
		&readFileTool{r.fs, 10000},
		&modifyFileTool{r.fs},
		&deleteFileTool{r.fs},
	}, nil, nil, nil
}

func (r runnerPlugin) Name() string {
	return "internal"
}
