package main

import (
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/files"
	"github.com/JoshPattman/cg/runner"
	"github.com/JoshPattman/jpf/models"
)

func main() {
	model := models.NewAPIModel(models.OpenAI, "gpt-5.4-mini", os.Getenv("OPENAI_KEY"))
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	agent := runner.New(model, path.Join(wd, "craig.md"), files.OSFileSystem(), runner.WithLogger(slog.Default()))

	go agent.Run()

	agent.Events() <- chatEvent{"Hi there! Please create a file called hi now, and then a file called bye after 10 seconds. Also, once you are done, write to your memory file that you have completed your task."}

	time.Sleep(time.Second * 30)
	agent.CleanStop()
}

type chatEvent struct {
	message string
}

func (e chatEvent) Kind() string {
	return "user_message"
}

func (e chatEvent) Content() cg.JsonObject {
	return cg.JsonObject{
		"user_message": e.message,
		"explanation":  "The user has sent you a message",
	}
}
