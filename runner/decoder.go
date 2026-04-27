package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/JoshPattman/jpf"
)

var _ jpf.Parser[int] = &firstJsonObjectParser[int]{}

type firstJsonObjectParser[T any] struct{}

func (f *firstJsonObjectParser[T]) ParseResponseText(response string) (T, error) {
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(response)
	if match == "" {
		var zero T
		return zero, errors.Join(jpf.ErrInvalidResponse, errors.New("response did not contain a json object"))
	}
	var result T
	readBuf := bytes.NewBufferString(match)
	dec := json.NewDecoder(readBuf)
	err := dec.Decode(&result)
	if err != nil {
		var zero T
		return zero, errors.Join(errors.Join(err, jpf.ErrInvalidResponse), errors.New("llm returned an invalid json object"))
	}
	return result, nil
}
