package messages

import (
	"bytes"
	"encoding/json"

	"github.com/JoshPattman/cg"

	"github.com/JoshPattman/jpf"
)

func EmptyToolCallsMessage() Message {
	return ToolCallsMessage{}
}

type ToolCallsMessage struct {
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolArg struct {
	ArgName string `json:"arg_name"`
	Value   any    `json:"value"`
}

type ToolCall struct {
	ToolName string    `json:"tool_name"`
	Args     []ToolArg `json:"args"`
}

func (m ToolCallsMessage) Role() jpf.Role {
	return jpf.AssistantRole
}

func (m ToolCallsMessage) Content() cg.JsonObject {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(m)
	if err != nil {
		panic(err)
	}
	var result cg.JsonObject
	err = json.NewDecoder(&buf).Decode(&result)
	if err != nil {
		panic(err)
	}
	return result
}
