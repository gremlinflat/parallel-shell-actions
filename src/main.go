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
	configPath := flag.String("i", "action.json", "Path to the JSON command configuration file")

	flag.Parse()

	// Read JSON file
	file, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		os.Exit(1)
	}

	// Parse JSON
	var actions []Action
	if err := json.Unmarshal(file, &actions); err != nil {
		fmt.Println("Error parsing JSON:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var commandRunner CommandRunner = *NewCommandRunner()

	resultChan := make(chan bool, len(actions))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run actions in parallel
	for _, act := range actions {
		wg.Add(1)
		go func(act Action) {
			success := commandRunner.RunCommand(ctx, act)
			if !success && act.CancelOnFailure {
				fmt.Printf("Critical action failed: %s\n", act.Commands[0])
				cancel() // Cancel all other goroutines
			}
			resultChan <- success
			wg.Done()
		}(act)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Check results
	overallSuccess := true
	for result := range resultChan {
		if !result {
			overallSuccess = false
		}
	}

	if overallSuccess {
		fmt.Println("All actions executed successfully.")
		os.Exit(0)
	} else {
		fmt.Println("Some actions failed.")
		os.Exit(1)
	}
}
