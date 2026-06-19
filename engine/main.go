package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type LogMessage struct {
	Type    string `json:"type"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type Command struct {
	Action string          `json:"action"`
	Params json.RawMessage `json:"params"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	sendEvent(LogMessage{
		Type:    "log",
		Level:   "info",
		Message: "Chronos Engine started",
	})

	for scanner.Scan() {
		var cmd Command
		err := json.Unmarshal(scanner.Bytes(), &cmd)
		if err != nil {
			sendEvent(LogMessage{
				Type:    "log",
				Level:   "error",
				Message: fmt.Sprintf("Failed to parse command: %v", err),
			})
			continue
		}

		handleCommand(cmd)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading standard input: %v\n", err)
	}
}

func handleCommand(cmd Command) {
	switch cmd.Action {
	case "ping":
		sendEvent(LogMessage{
			Type:    "log",
			Level:   "info",
			Message: "pong",
		})
	default:
		sendEvent(LogMessage{
			Type:    "log",
			Level:   "warn",
			Message: fmt.Sprintf("Unknown command: %s", cmd.Action),
		})
	}
}

func sendEvent(event interface{}) {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}
