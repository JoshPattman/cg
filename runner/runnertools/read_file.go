package runnertools

import (
	"fmt"

	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/files"
)

type readFileTool struct {
	fs        files.FileSystem
	maxLength int
}

func (t *readFileTool) Def() cg.ToolDef {
	return cg.ToolDef{
		Name: "read_file",
		Desc: fmt.Sprintf(
			"Read a number of characters from a file. Args: 'path' (string), 'offset' (int), 'max_characters' (int). Max allowed characters: %d.",
			t.maxLength,
		),
	}
}

func (t *readFileTool) Call(args map[string]any) (string, error) {
	pathAny, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("must specify a 'path'.")
	}
	path, ok := pathAny.(string)
	if !ok {
		return "", fmt.Errorf("'path' must be a string")
	}
	offsetAny, ok := args["offset"]
	if !ok {
		return "", fmt.Errorf("must specify an 'offset' (use 0 if you want no offset).")
	}
	offsetFloat, ok := offsetAny.(float64)
	if !ok {
		return "", fmt.Errorf("'offset' must be an integer")
	}
	maxCharactersAny, ok := args["max_characters"]
	if !ok {
		return "", fmt.Errorf("must specify a 'max_characters' (use the maximum if you want as much as possible).")
	}
	maxCharactersFloat, ok := maxCharactersAny.(float64)
	if !ok {
		return "", fmt.Errorf("'max_characters' must be a number")
	}

	offset := int(offsetFloat)
	maxCharacters := int(maxCharactersFloat)

	if maxCharacters > t.maxLength {
		maxCharacters = t.maxLength
	}

	data, err := t.fs.Read(path)
	if err != nil {
		return "", err
	}

	realOffset := min(max(0, offset), len(data))
	realEnd := min(max(0, offset+maxCharacters), len(data))

	resultText := string(data[realOffset:realEnd])

	return fmt.Sprintf(
		"File contents of %s from %d to %d (%d chars):\n%s",
		path,
		realOffset,
		realEnd,
		realEnd-realOffset,
		resultText,
	), nil
}
