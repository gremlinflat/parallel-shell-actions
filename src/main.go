package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
)

func main() {
	// Define command-line flag for JSON file path
	inputPath := flag.String("i", "action.json", "Path to the JSON file containing the actions to execute")

	flag.Parse()

	// Read JSON file
	file, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Println("Error reading Input file:", err)
		os.Exit(1)
	}

	// Parse JSON
	var actions []Action
	if err := json.Unmarshal(file, &actions); err != nil {
		fmt.Println("Error parsing Input file:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var commandRunner CommandRunner = *NewCommandRunner()

	resultChan := make(chan []CommandResult, len(actions))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run actions in parallel
	for idx, act := range actions {
		if act.Namespace == "" {
			act.Namespace = fmt.Sprintf("Action-%d", idx+1)
		}

		wg.Add(1)
		
		// Run each action in a separate goroutine
		go func(act Action) {
			defer wg.Done()
			results := commandRunner.RunCommand(ctx, act)
			resultChan <- results

			// Check if any command failed and should cancel other goroutines
			for _, result := range results {
				if !result.Success && act.CancelOnFailure {
					fmt.Printf("Cancelling all goroutines due to failure in command: %s\n", result.Command)
					cancel()
					return
				}
			}
		}(act)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Check results
	var failedCommands []CommandResult
	for results := range resultChan {
		for _, result := range results {
			if !result.Success {
				failedCommands = append(failedCommands, result)
			}
		}
	}

	// Log failed commands
	if len(failedCommands) > 0 {
		fmt.Println("\nFailed commands:")
		for _, failedCmd := range failedCommands {
			fmt.Printf("Command: %s\nExit Code: %d\n\n", failedCmd.Command, failedCmd.ExitCode)
		}
		fmt.Printf("Total failed commands: %d\n", len(failedCommands))
		os.Exit(1)
	} else {
		fmt.Println("All actions executed successfully.")
		os.Exit(0)
	}
}
