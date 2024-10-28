package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func main() {
	// Read the JSON file
	jsonFile, err := ioutil.ReadFile("/Users/fahri.novaldi/Desktop/work/clone/parallel-shell-actions/src/sample.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Parse the JSON
	var config Config
	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a channel to signal cancellation
	cancelChan := make(chan struct{})

	// Run commands in parallel
	for i, cmd := range config.Commands {
		wg.Add(1)
		go func(index int, command Command) {
			defer wg.Done()
			runCommand(index, command, cancelChan)
		}(i, cmd)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func runCommand(index int, command Command, cancelChan chan struct{}) {
	fmt.Printf("Starting command set %d\n", index+1)

	for _, cmdStr := range command.Command {
		select {
		case <-cancelChan:
			fmt.Printf("Command set %d cancelled\n", index+1)
			return
		default:
			cmd := exec.Command(command.Shell, "-c", cmdStr)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error in command set %d: %v\n", index+1, err)
				if command.CancelOnFailure {
					close(cancelChan)
					return
				}
			}
		}
	}

	fmt.Printf("Finished command set %d\n", index+1)
}
