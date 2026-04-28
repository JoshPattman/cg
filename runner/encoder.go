package runner

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/runner/messages"

	"github.com/JoshPattman/jpf"
)

type encoderInput struct {
	Messages              []messages.Message
	ToolDefs              []cg.ToolDef
	Time                  time.Time
	WorkingMemoryLocation string
	WorkingMemory         string
}

type failedPlugin struct {
	Name  string
	Error error
}

// Build the encoder for use with the runner.
func buildEncoder(systemPrompt cg.JsonObject) jpf.Encoder[encoderInput] {
	return &encoder{
		systemPrompt,
	}
}

type encoder struct {
	systemPrompt cg.JsonObject
}

func (e *encoder) BuildInputMessages(input encoderInput) ([]jpf.Message, error) {
	messages := make([]jpf.Message, 0)

	messages = append(messages, jpf.Message{
		Role:    jpf.SystemRole,
		Content: objectContent(e.systemPrompt),
	})

	for _, msg := range input.Messages {
		messages = append(messages, jpf.Message{
			Role:    msg.Role(),
			Content: objectContent(msg.Content()),
		})
	}
	messages = append(messages, jpf.Message{
		Role:    jpf.UserRole,
		Content: objectContent(e.activeState(input)),
	})

	return messages, nil
}

func (e *encoder) activeState(input encoderInput) cg.JsonObject {
	return cg.JsonObject{
		"description":                  "This is a current state message. You will have been provided with them at previous points in the conversation too, however they have been removed for brevity. This state message is currently up-to-date and active.",
		"active_tools":                 input.ToolDefs,
		"current_datetime":             input.Time.Format(time.RFC1123),
		"working_memory_file_location": input.WorkingMemoryLocation,
		"working_memory":               input.WorkingMemory,
	}
}

func objectContent(obj cg.JsonObject) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	err := enc.Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
