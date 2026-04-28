package runner

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/JoshPattman/cg/runner/messages"
	"github.com/JoshPattman/cg/runner/runnertools"

	"github.com/JoshPattman/cg"
	"github.com/JoshPattman/cg/files"

	"github.com/JoshPattman/jpf"

	_ "embed"
)

var (
	ErrPluginNotExist     = errors.New("that plugin did not exist")
	ErrPluginAlreadyExist = errors.New("that plugin already existed")
	ErrPluginNotRemovable = errors.New("that plugin cannot be removed")
)

type agentRunner struct {
	logger             *slog.Logger
	events             chan cg.Event
	history            []messages.Message
	encoder            jpf.Encoder[encoderInput]
	pipeline           jpf.Pipeline[encoderInput, messages.ToolCallsMessage]
	collectionDuration time.Duration
	plugins            []cg.Plugin
	pluginLock         *sync.Mutex
	memoryLoc          string
	fs                 files.FileSystem
	running            *sync.Mutex
	cleanStop          chan struct{}
	shrinkBoundary     int
	truncateBoundary   int
	tokenLimit         int
}

func (a *agentRunner) AddPlugin(p cg.Plugin) error {
	a.pluginLock.Lock()
	defer a.pluginLock.Unlock()

	_, p2 := a.getPluginWithName(p.Name())
	if p2 != nil {
		return errors.Join(fmt.Errorf("plugin '%s'", p.Name()), ErrPluginAlreadyExist)
	}

	a.plugins = append(a.plugins, p)
	a.logger.Info("loaded plugin", "plugin", p.Name(), "num_tools", len(p.Tools()))
	return nil
}

func (a *agentRunner) RemovePlugin(name string) error {
	a.pluginLock.Lock()
	defer a.pluginLock.Unlock()
	return a.removePluginWithoutLock(name)
}

func (a *agentRunner) removePluginWithoutLock(name string) error {
	i, p := a.getPluginWithName(name)
	if i == -1 {
		return errors.Join(fmt.Errorf("plugin '%s'", name), ErrPluginNotExist)
	}
	if !p.Removable() {
		return errors.Join(fmt.Errorf("plugin '%s'", name), ErrPluginNotRemovable)
	}
	a.plugins = slices.Delete(a.plugins, i, i+1)
	a.logger.Info("removed plugin", "plugin", name)
	return nil
}

func (a *agentRunner) AllPlugins() iter.Seq[cg.Plugin] {
	return func(yield func(cg.Plugin) bool) {
		a.pluginLock.Lock()
		defer a.pluginLock.Unlock()

		for _, p := range a.plugins {
			if !yield(p) {
				return
			}
		}
	}
}

func (a *agentRunner) getPluginWithName(name string) (int, cg.Plugin) {
	for i, p := range a.plugins {
		if p.Name() == name {
			return i, p
		}
	}
	return -1, nil
}

func (a *agentRunner) Events() chan<- cg.Event { return a.events }

func (a *agentRunner) Run() error {
	if !a.running.TryLock() {
		return fmt.Errorf("agent is already running")
	}
	defer a.running.Unlock()

	a.logger.Info("starting event forwarder")
	done := make(chan struct{}, 1)
	defer func() {
		done <- struct{}{}
	}()
	go a.eventForwarder(done)

	a.logger.Info("running event loop")

	var eventBuffer []cg.Event
	var dispatch <-chan time.Time

	for {
		select {
		case event := <-a.events:
			eventBuffer = append(eventBuffer, event)
			if len(eventBuffer) == 1 {
				a.logger.Info("initial event occured, waiting to collect more")
				dispatch = time.After(a.collectionDuration)
			} else {
				a.logger.Info("additional event occured")
			}

		case <-dispatch:
			a.addEventsMessage(eventBuffer...)
			err := a.processUntilDone()
			if err != nil {
				return err
			}
			eventBuffer = nil
			dispatch = nil
		case <-a.cleanStop:
			a.logger.Info("clean stop signal received, stopping event loop")
			return nil
		}
	}
}

func (a *agentRunner) CleanStop() {
	a.logger.Info("clean stop initiated, stopping event forwarder and waiting for current processing to finish")
	a.cleanStop <- struct{}{}
	a.running.Lock()
	a.logger.Info("clean stop complete, agent has stopped")
	a.running.Unlock()
}

func (a *agentRunner) eventForwarder(stop <-chan struct{}) {
	for {
		a.pluginLock.Lock()
		for _, p := range a.plugins {
			if p.Events() == nil {
				continue
			}
			finishedPluginEvents := false
			for !finishedPluginEvents {
				// If stop is used, immediately stop without processing further events.
				select {
				case <-stop:
					return
				default:
				}
				// Recv an event.
				select {
				case event, ok := <-p.Events():
					// Try to send but if stop is used, stop immediately (before blocking on events chan).
					if !ok {
						continue
					}
					select {
					case <-stop:
						return
					case a.events <- event:
					}
				// If no events, continue
				default:
					finishedPluginEvents = true
				}
			}
		}
		a.pluginLock.Unlock()
		// If stop after we are done with events, stop now.
		select {
		case <-stop:
			return
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (a *agentRunner) processUntilDone() error {
	a.logger.Info("processing events")
	for {
		a.ensureNotTooManyTokens()
		inputData := a.encoderInput()
		result, _, err := a.pipeline.Call(context.Background(), inputData)
		if err != nil {
			a.logger.Error("failed to process events", "err", err)
			return err
		}
		a.addHistory(result)

		if len(result.ToolCalls) == 0 {
			a.logger.Info("agent called no tools so we need to remind it this is not how it stops processing")
			a.addHistory(messages.NeedToExplicitlyStopMessage())
			continue
		}
		if len(result.ToolCalls) == 1 && result.ToolCalls[0].ToolName == runnertools.DoneToolName() {
			approxTokens, err := a.countApproxTokens()
			if err != nil {
				a.logger.Error("failed to count tokens", "err", err)
				approxTokens = -1
			}
			a.logger.Info("agent called end iteration tool so we can stop", "conversation_tokens_approx", approxTokens)
			return nil
		}
		a.logToolCalls(result.ToolCalls)

		responses := make([]string, 0, len(result.ToolCalls))

		for _, call := range result.ToolCalls {
			tool := a.lookupTool(call.ToolName)
			if tool == nil {
				responses = append(responses,
					fmt.Sprintf("Tool '%s' not found", call.ToolName),
				)
				a.logger.Warn("a tool was not found", "tool", call.ToolName)
				continue
			}

			args := make(map[string]any)
			for _, arg := range call.Args {
				args[arg.ArgName] = arg.Value
			}

			out, err := tool.Call(args)
			if err != nil {
				responses = append(responses,
					fmt.Sprintf("Tool '%s' error: %v", call.ToolName, err),
				)
				a.logger.Warn("a tool errored", "tool", call.ToolName, "err", err)
				continue
			}

			responses = append(responses, out)
			a.logger.Info("Tool added tokens to context", "tool", call.ToolName, "tokens", countApproxTokens(out))
		}
		a.addHistory(messages.ToolResponseMessage(responses))
	}
}

func (a *agentRunner) ensureNotTooManyTokens() {
	for {
		tokens, err := a.countApproxTokens()
		if err != nil {
			a.logger.Error("failed to count tokens", "err", err)
			return
		}
		if tokens <= a.tokenLimit {
			return
		}
		couldShrink := a.shrinkAMessage()
		if couldShrink {
			continue
		}
		couldRemove := a.removeAMessage()
		if couldRemove {
			continue
		}
		a.logger.Warn("too many tokens and cannot shrink or remove any more messages", "token_count_approx", tokens)
		return
	}
}

func (a *agentRunner) removeAMessage() bool {
	if len(a.history) <= a.truncateBoundary {
		return false
	}
	a.logger.Info("removing a message to reduce token count", "message_type", fmt.Sprintf("%T", a.history[0]))
	a.history = a.history[1:]
	return true
}

func (a *agentRunner) shrinkAMessage() bool {
	for i, msg := range a.history {
		if len(a.history)-i <= a.shrinkBoundary {
			break
		}
		shrinkable, ok := msg.(messages.ShrinkableMessage)
		if !ok {
			continue

		}
		if shrinkable.IsShrunk() {
			continue
		}
		a.logger.Info("shrinking a message to reduce token count", "index", i, "message_type", fmt.Sprintf("%T", msg))
		a.history[i] = shrinkable.Shrunk()
		return true
	}
	return false
}

func (a *agentRunner) countApproxTokens() (int, error) {
	messages, err := a.encoder.BuildInputMessages(a.encoderInput())
	if err != nil {
		return 0, err
	}
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += 10 // roughly for stuff like role etc
		totalTokens += countApproxTokens(msg.Content)
	}
	return totalTokens, nil
}

func (a *agentRunner) encoderInput() encoderInput {
	return encoderInput{
		a.history,
		a.toolDefs(),
		time.Now(),
		a.memoryLoc,
		a.workingMemory(),
	}
}

func (a *agentRunner) logToolCalls(toolCalls []messages.ToolCall) {
	toolNames := make([]string, len(toolCalls))
	for i, t := range toolCalls {
		toolNames[i] = t.ToolName
	}
	a.logger.Info("agent called tools", "tool_names", strings.Join(toolNames, ";"))
}

func (a *agentRunner) workingMemory() string {
	bs, err := a.fs.Read(a.memoryLoc)
	if err != nil {
		return fmt.Sprintf("There was an error loading your working memory: %s", err.Error())
	}
	return string(bs)
}

func (a *agentRunner) addEventsMessage(events ...cg.Event) {
	a.addHistory(messages.EventsMessage(events...))
	eventNames := make([]string, len(events))
	for i, e := range events {
		eventNames[i] = e.Kind()
	}
	a.logger.Info("events occured", "n", len(events), "names", strings.Join(eventNames, ";"))
}

func (a *agentRunner) addHistory(messages ...messages.Message) {
	a.history = append(a.history, messages...)
}

func (a *agentRunner) lookupTool(name string) cg.Tool {
	for _, p := range a.plugins {
		for _, t := range p.Tools() {
			if t.Def().Name == name {
				return t
			}
		}
	}
	return nil
}

func (a *agentRunner) toolExists(name string) bool {
	return a.lookupTool(name) != nil
}

func (a *agentRunner) toolNames() []string {
	names := make([]string, 0)
	for _, p := range a.plugins {
		for _, t := range p.Tools() {
			names = append(names, t.Def().Name)
		}
	}
	return names
}

func (a *agentRunner) toolDefs() []cg.ToolDef {
	defs := make([]cg.ToolDef, 0)
	for _, p := range a.plugins {
		for _, t := range p.Tools() {
			defs = append(defs, t.Def())
		}
	}
	return defs
}
