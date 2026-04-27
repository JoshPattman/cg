package runner

import (
	"encoding/json"

	"github.com/JoshPattman/cg/runner/messages"

	"github.com/invopop/jsonschema"
)

func GetResponseSchema() map[string]any {
	s, err := getSchema(messages.ToolCallsMessage{})
	if err != nil {
		panic(err)
	}
	return s
}

func getSchema(obj any) (map[string]any, error) {
	r := &jsonschema.Reflector{
		BaseSchemaID:   "Anonymous",
		Anonymous:      true,
		DoNotReference: true,
	}
	s := r.Reflect(obj)
	schemaBs, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}
	schema := make(map[string]any)
	err = json.Unmarshal(schemaBs, &schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
