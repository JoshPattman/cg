package messages

import (
	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/jpf"
)

func ToolResponseMessage(responses []string) Message {
	return toolResponseMessage{responses, false}
}

type toolResponseMessage struct {
	responses []string
	shrink    bool
}

func (m toolResponseMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m toolResponseMessage) Content() cg.JsonObject {
	if m.shrink {
		return cg.JsonObject{
			"explanation": "This tool response has been hidden to save tokens. You should have written any useful info into your scratchpad.",
		}
	} else {
		return cg.JsonObject{
			"explanation":    "Here are the responses from your tool calls",
			"tool_responses": m.responses,
		}
	}
}

func (m toolResponseMessage) Shrunk() Message {
	m.shrink = true
	return m
}

func (m toolResponseMessage) IsShrunk() bool {
	return m.shrink
}
