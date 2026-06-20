package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/illusion-engine/chronos/engine/internal/engine"
	"github.com/illusion-engine/chronos/engine/internal/models"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	sendEvent(models.LogEvent{
		Type:    "log",
		Level:   "info",
		Message: "Chronos Engine ready",
	})

	for scanner.Scan() {
		var msg struct {
			Action string          `json:"action"`
			Params json.RawMessage `json:"params"`
		}
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			continue
		}

		switch msg.Action {
		case "execute":
			var cfg models.Config
			json.Unmarshal(msg.Params, &cfg)

			eng := engine.New(cfg)
			if err := eng.Run(); err != nil {
				sendEvent(models.LogEvent{
					Type:    "log",
					Level:   "error",
					Message: fmt.Sprintf("Execution failed: %v", err),
				})
			}
		case "ping":
			sendEvent(models.LogEvent{
				Type:    "log",
				Level:   "info",
				Message: "pong",
			})
		}
	}
}

func sendEvent(event interface{}) {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}
