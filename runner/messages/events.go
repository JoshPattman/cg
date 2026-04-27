package messages

import (
	"github.com/JoshPattman/cg"

	"github.com/JoshPattman/jpf"
)

func EventsMessage(events ...cg.Event) Message {
	return eventsMessage{events}
}

type eventsMessage struct {
	events []cg.Event
}

func (eventsMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (m eventsMessage) Content() cg.JsonObject {
	eventObjects := make([]cg.JsonObject, len(m.events))
	for i, e := range m.events {
		eventObjects[i] = cg.JsonObject{
			"kind":    e.Kind(),
			"content": e.Content(),
		}
	}
	return cg.JsonObject{
		"explanation": "Some events have occured since your last message. They might be relevant or you might be able to ignore them.",
		"events":      eventObjects,
	}
}
