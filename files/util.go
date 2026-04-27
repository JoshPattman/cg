package files

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func ReplaceText(fs FileSystem, loc string, oldText string, newText string) error {
	originalData, err := fs.Read(loc)
	if errors.Is(err, os.ErrNotExist) {
		if oldText != "" {
			return fmt.Errorf("must specify empty old text to write a new file")
		}
		return fs.Overwrite(loc, []byte(newText))
	} else if err != nil {
		return err
	}
	originalText := string(originalData)
	num := strings.Count(originalText, oldText)
	if num != 1 {
		return fmt.Errorf("must have exactly one occurence of the string in the file, but got %d", num)
	}
	updatedText := strings.ReplaceAll(originalText, oldText, newText)
	return fs.Overwrite(loc, []byte(updatedText))
}
