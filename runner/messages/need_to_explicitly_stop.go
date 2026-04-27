package messages

import (
	"github.com/JoshPattman/cg"

	"github.com/JoshPattman/jpf"
)

func NeedToExplicitlyStopMessage() Message {
	return needToEndMessage{}
}

type needToEndMessage struct{}

func (needToEndMessage) Role() jpf.Role {
	return jpf.UserRole
}

func (needToEndMessage) Content() cg.JsonObject {
	return cg.JsonObject{
		"explanation": "You called no tools, however you will continue iterating (calling no tools is not a useful thing to do). To stop iterating, please call the end_iteration tool by itself with no args.",
	}
}
