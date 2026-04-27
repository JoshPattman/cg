package runnertools

import "github.com/JoshPattman/cg"

func DoneToolName() string {
	return (&doneTool{}).Def().Name
}

type doneTool struct{}

func (*doneTool) Def() cg.ToolDef {
	return cg.ToolDef{
		Name: "end_iteration",
		Desc: "Special tool: Call this tool to end iteration until the next set of events come in. This is like awaiting the next events. You will not be able to do anything else until the next events are received. You must always call this when you think you do not ned to do anything alse to handle the current events.",
	}
}
func (*doneTool) Call(map[string]any) (string, error) {
	return "You can only call this tool in isolation (without any other tool calls at the same time). Please finish what you are currently doing, then you may call this.", nil
}
