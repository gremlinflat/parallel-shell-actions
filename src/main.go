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

func runCommand(ctx context.Context, cmd Command, wg *sync.WaitGroup, successChan chan bool) {
	defer wg.Done()
	for _, c := range cmd.Command {
		select {
		case <-ctx.Done():
			fmt.Printf("Command '%s' canceled due to failure in another command.\n", c)
			successChan <- false
			return
		default:
			command := exec.Command(cmd.Shell, "-c", c)
			stdout, err := command.StdoutPipe()
			if err != nil {
				fmt.Printf("Error creating StdoutPipe for command '%s': %v\n", c, err)
				if cmd.CancelOnFailure {
					successChan <- false
					return
				}
			}

			stderr, err := command.StderrPipe()
			if err != nil {
				fmt.Printf("Error creating StderrPipe for command '%s': %v\n", c, err)
				if cmd.CancelOnFailure {
					successChan <- false
					return
				}
			}

			if err := command.Start(); err != nil {
				fmt.Printf("Error starting command '%s': %v\n", c, err)
				if cmd.CancelOnFailure {
					successChan <- false
					return
				}
			}

			go io.Copy(os.Stdout, stdout)
			go io.Copy(os.Stderr, stderr)

			if err := command.Wait(); err != nil {
				fmt.Printf("Error waiting for command '%s' to finish: %v\n", c, err)
				if cmd.CancelOnFailure {
					successChan <- false
					return
				}
			}
		}
	}
	successChan <- true
}

func main() {
	// Read JSON file
	file, err := os.ReadFile("sample.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		fmt.Println("Error parsing JSON:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	successChan := make(chan bool, len(config.Commands))
	ctx, cancel := context.WithCancel(context.Background())

	// Run commands in parallel
	for _, cmd := range config.Commands {
		wg.Add(1)
		go func(cmd Command) {
			defer wg.Done()
			runCommand(ctx, cmd, &wg, successChan)
			if !<-successChan && cmd.CancelOnFailure {
				cancel()
			}
		}(cmd)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(successChan)

	// Check success flag
	success := true
	for result := range successChan {
		if !result {
			success = false
			break
		}
	}

	if success {
		fmt.Println("All commands executed successfully.")
	} else {
		fmt.Println("Some commands failed.")
	}
}
