package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Command struct {
	Shell           string   `json:"shell"`
	Command         []string `json:"command"`
	CancelOnFailure bool     `json:"cancel-on-failure"`
}

type Config struct {
	Commands []Command `json:"command"`
}

func runCommand(ctx context.Context, cmd Command, wg *sync.WaitGroup, resultChan chan<- bool) {
	defer wg.Done()
	success := true

	for _, c := range cmd.Command {
		select {
		case <-ctx.Done():
			fmt.Printf("Command '%s' canceled due to context cancellation.\n", c)
			resultChan <- false
			return
		default:
			command := exec.CommandContext(ctx, cmd.Shell, "-c", c)
			stdout, err := command.StdoutPipe()
			if err != nil {
				fmt.Printf("Error creating StdoutPipe for command '%s': %v\n", c, err)
				success = false
				break
			}

			stderr, err := command.StderrPipe()
			if err != nil {
				fmt.Printf("Error creating StderrPipe for command '%s': %v\n", c, err)
				success = false
				break
			}

			if err := command.Start(); err != nil {
				fmt.Printf("Error starting command '%s': %v\n", c, err)
				success = false
				break
			}

			go io.Copy(os.Stdout, stdout)
			go io.Copy(os.Stderr, stderr)

			if err := command.Wait(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					fmt.Printf("Command '%s' failed with exit code %d\n", c, exitErr.ExitCode())
				} else {
					fmt.Printf("Command '%s' failed: %v\n", c, err)
				}
				success = false
				if cmd.CancelOnFailure {
					break
				}
			}
		}

		if !success && cmd.CancelOnFailure {
			break
		}
	}

	resultChan <- success
}

func main() {
	// Read JSON file
	file, err := os.ReadFile("sample.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		os.Exit(1)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		fmt.Println("Error parsing JSON:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	resultChan := make(chan bool, len(config.Commands))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run commands in parallel
	for _, cmd := range config.Commands {
		wg.Add(1)
		go func(cmd Command) {
			runCommand(ctx, cmd, &wg, resultChan)
		}(cmd)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Check results
	success := true
	for result := range resultChan {
		if !result {
			success = false
			cancel() // Cancel all ongoing operations
		}
	}

	if success {
		fmt.Println("All commands executed successfully.")
		os.Exit(0)
	} else {
		fmt.Println("Some commands failed.")
		os.Exit(1)
	}
}
