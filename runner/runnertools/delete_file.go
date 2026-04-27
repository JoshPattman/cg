package runnertools

import (
	"fmt"

	"github.com/JoshPattman/cg/files"

	"github.com/JoshPattman/cg"
)

type deleteFileTool struct {
	fs files.FileSystem
}

func NewDeleteFileTool(fs files.FileSystem) cg.Tool {
	return &deleteFileTool{fs: fs}
}

func (t *deleteFileTool) Def() cg.ToolDef {
	return cg.ToolDef{
		Name: "delete_file",
		Desc: "Delete a file or directory at the given path. Args: 'path' (string).",
	}
}

func (t *deleteFileTool) Call(args map[string]any) (string, error) {
	pathAny, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("must specify 'path'")
	}
	path, ok := pathAny.(string)
	if !ok {
		return "", fmt.Errorf("'path' must be a string")
	}

	err := t.fs.Delete(path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted: %s", path), nil
}
