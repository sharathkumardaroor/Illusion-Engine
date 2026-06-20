package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	exitCode := 0

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
				sendEvent(models.LogEvent{
					Type: "state",
					Payload: models.State{
						Status: "error",
					},
				})
				exitCode = 1
			}
		case "test-prep":
			tempDir, err := os.MkdirTemp("", "chronos-test-source-*")
			if err != nil {
				sendEvent(models.LogEvent{Type: "log", Level: "error", Message: "Failed to create test source"})
				continue
			}
			// Create dummy project files
			os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
			os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test Project\n"), 0644)
			os.Mkdir(filepath.Join(tempDir, "pkg"), 0755)
			os.WriteFile(filepath.Join(tempDir, "pkg", "utils.go"), []byte("package pkg\n"), 0644)

			sendEvent(models.LogEvent{
				Type: "log",
				Level: "info",
				Message: fmt.Sprintf("Test source prepared at %s", tempDir),
				Payload: map[string]string{"path": tempDir},
			})
		case "ping":
			sendEvent(models.LogEvent{
				Type:    "log",
				Level:   "info",
				Message: "pong",
			})
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading standard input: %v\n", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func sendEvent(event interface{}) {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}
