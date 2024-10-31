package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type CommandRunner struct{}

func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

func (cr *CommandRunner) RunCommand(ctx context.Context, act Action) []CommandResult {
	results := []CommandResult{}

	// set valid shell to run the commands
	shell := cr.getActionShell(act)

	for _, c := range act.Commands {
		select {
		case <-ctx.Done():
			fmt.Printf("Command '%s' canceled due to context cancellation.\n", c)
			results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: false, ExitCode: -99})
			return results
		default:
			command := exec.CommandContext(ctx, shell, "-c", c)
			stdout, err := command.StdoutPipe()
			if err != nil {
				fmt.Printf("[%s] Error creating StdoutPipe for command '%s': %v\n", act.Namespace, c, err)
				results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: false, ExitCode: -2})
				continue
			}

			stderr, err := command.StderrPipe()
			if err != nil {
				fmt.Printf("[%s] Error creating StderrPipe for command '%s': %v\n", act.Namespace, c, err)
				results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: false, ExitCode: -2})
				continue
			}

			if err := command.Start(); err != nil {
				fmt.Printf("[%s] Error starting command '%s': %v\n", act.Namespace, c, err)
				results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: false, ExitCode: -3})
				continue
			}

			go cr.pipeOutput(stdout, os.Stdout, act.Namespace)
			go cr.pipeOutput(stderr, os.Stderr, act.Namespace)

			if err := command.Wait(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode := exitErr.ExitCode()
					fmt.Printf("[%s] Command '%s' failed with exit code %d\n", act.Namespace, c, exitCode)
					results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: false, ExitCode: exitCode})
				} else { 
					fmt.Printf("[%s] Command '%s' failed: %v\n", act.Namespace, c, err)
					results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: false, ExitCode: -1})
				}
				if act.CancelOnFailure {
					fmt.Printf("[%s] FAIL: Command '%s' failed on critical action. Cancelling further actions.\n", act.Namespace, c)
					return results
				}
			} else {
				results = append(results, CommandResult{Namespace:act.Namespace, Command: c, Success: true, ExitCode: 0})
			}
		}
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
