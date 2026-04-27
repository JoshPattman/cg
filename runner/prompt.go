package runner

import (
	"bytes"
	"encoding/json"

	"github.com/JoshPattman/cg"

	_ "embed"
)

//go:embed system.json
var defaultPrompt []byte

func getDefaultPrompt() cg.JsonObject {
	var res cg.JsonObject
	err := json.NewDecoder(bytes.NewReader(defaultPrompt)).Decode(&res)
	if err != nil {
		panic(err)
	}
	return res
}
