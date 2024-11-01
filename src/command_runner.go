package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type CommandRunner struct{}

func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

func (cr *CommandRunner) RunCommand(ctx context.Context, act Action) []CommandResult {
	results := []CommandResult{}

	// set valid shell to run the commands
	shell := cr.getActionShell(act)

	// Create a heredoc script
	script := fmt.Sprintf(`#!/bin/%s
	set -e
	%s
	`, shell, strings.Join(act.Commands, "\n"))

	// Create a temporary file for the script
	tmpfile, err := os.CreateTemp("", "script-*.sh")
	if err != nil {
		fmt.Printf("[%s] Error creating temporary file: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script creation", Success: false, ExitCode: -1})
		return results
	}
	defer os.Remove(tmpfile.Name())

	// Write the script to the temporary file
	if _, err := tmpfile.Write([]byte(script)); err != nil {
		fmt.Printf("[%s] Error writing to temporary file: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script writing", Success: false, ExitCode: -1})
		return results
	}
	if err := tmpfile.Close(); err != nil {
		fmt.Printf("[%s] Error closing temporary file: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script closing", Success: false, ExitCode: -1})
		return results
	}

	// Make the script executable
	if err := os.Chmod(tmpfile.Name(), 0700); err != nil {
		fmt.Printf("[%s] Error making script executable: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script permissions", Success: false, ExitCode: -1})
		return results
	}

	// Execute the script
	command := exec.CommandContext(ctx, tmpfile.Name())

	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Printf("[%s] Error creating StdoutPipe: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script execution", Success: false, ExitCode: -2})
		return results
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		fmt.Printf("[%s] Error creating StderrPipe: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script execution", Success: false, ExitCode: -2})
		return results
	}

	if err := command.Start(); err != nil {
		fmt.Printf("[%s] Error starting script: %v\n", act.Namespace, err)
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script execution", Success: false, ExitCode: -3})
		return results
	}

	go cr.pipeOutput(stdout, os.Stdout, act.Namespace)
	go cr.pipeOutput(stderr, os.Stderr, act.Namespace)

	if err := command.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			fmt.Printf("[%s] Script failed with exit code %d\n", act.Namespace, exitCode)
			results = append(results, CommandResult{Namespace: act.Namespace, Command: "script execution", Success: false, ExitCode: exitCode})
		} else {
			fmt.Printf("[%s] Script failed: %v\n", act.Namespace, err)
			results = append(results, CommandResult{Namespace: act.Namespace, Command: "script execution", Success: false, ExitCode: -1})
		}
	} else {
		results = append(results, CommandResult{Namespace: act.Namespace, Command: "script execution", Success: true, ExitCode: 0})
	}

	return results
}

func (cr *CommandRunner) pipeOutput(input io.Reader, output *os.File, namespace string) {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		fmt.Fprintf(output, "[%s] %s\n", namespace, scanner.Text())
	}
}

func (cr *CommandRunner) getActionShell(act Action) string {
	// checking if the shell is valid
	for _, shell := range cr.supportedShells() {
		if act.Shell == shell {
			return act.Shell
		}
	}

	fmt.Printf("[%s] Invalid shell type '%s'. Defaulting to '%s'\n", act.Namespace, act.Shell, cr.defaultShell())
	return cr.defaultShell()
}

func (cr *CommandRunner) supportedShells() []string {
	validShells := []string{"bash", "sh"} // Github Actions supports bash and sh
	return validShells
}

func (cr *CommandRunner) defaultShell() string {
	return "bash"
}
