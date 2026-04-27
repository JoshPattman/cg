package files

import (
	"os"
	"path/filepath"
)

func OSFileSystem() FileSystem {
	return &osFileSystem{}
}

type osFileSystem struct{}

func (o *osFileSystem) ListDir(loc string) ([]string, error) {
	entries, err := os.ReadDir(loc)
	if err != nil {
		return nil, err
	}
	results := make([]string, len(entries))
	for i, entry := range entries {
		results[i] = filepath.Join(loc, entry.Name())
		if entry.IsDir() {
			results[i] += "/"
		}
	}
	return results, nil
}

func (o *osFileSystem) Overwrite(loc string, content []byte) error {
	return os.WriteFile(loc, content, 0644)
}

func (o *osFileSystem) Read(loc string) ([]byte, error) {
	return os.ReadFile(loc)
}

func (o *osFileSystem) Delete(loc string) error {
	return os.RemoveAll(loc)
}
