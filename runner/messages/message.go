package messages

import (
	"github.com/JoshPattman/cg"

	"github.com/JoshPattman/jpf"
)

type Message interface {
	Role() jpf.Role
	Content() cg.JsonObject
}

type ShrinkableMessage interface {
	Message
	Shrunk() Message
	IsShrunk() bool
}
