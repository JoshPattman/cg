package cg

import "encoding/json"

type JsonObject map[string]any

func ParseToolArgs[T any](args map[string]any) (T, error) {
	bs, err := json.Marshal(args)
	if err != nil {
		return *new(T), err
	}
	var t T
	err = json.Unmarshal(bs, &t)
	if err != nil {
		return *new(T), err
	}
	return t, nil
}
