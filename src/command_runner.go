package main

import (
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

func (cr *CommandRunner) RunCommand(ctx context.Context, act Action) bool {
	success := true

	// set valid shell to run the commands
	act.Shell = cr.getActionShell(act)

	for _, c := range act.Commands {
		select {
		case <-ctx.Done():
			fmt.Printf("Command '%s' canceled due to context cancellation.\n", c)
			return false
		default:
			command := exec.CommandContext(ctx, act.Shell, "-c", c)
			stdout, err := command.StdoutPipe()
			if err != nil {
				fmt.Printf("Error creating StdoutPipe for command '%s': %v\n", c, err)
				success = false
				continue
			}

			stderr, err := command.StderrPipe()
			if err != nil {
				fmt.Printf("Error creating StderrPipe for command '%s': %v\n", c, err)
				success = false
				continue
			}

			if err := command.Start(); err != nil {
				fmt.Printf("Error starting command '%s': %v\n", c, err)
				success = false
				continue
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
				if act.CancelOnFailure {
					fmt.Printf("FAIL: Command '%s' failed and cancel-on-failure is set. Returning failure.\n", c)
					return false
				}
			}
		}
	}

	return success
}

func (cr *CommandRunner) getActionShell(act Action) string {
	// checking if the shell is valid
	for _, shell := range cr.supportedShells() {
		if act.Shell == shell {
			return act.Shell
		}
	}

	fmt.Printf("Invalid shell type '%s'. Defaulting to '%s'\n", act.Shell, cr.defaultShell())
	return cr.defaultShell()
}

func (cr *CommandRunner) supportedShells() []string {
	validShells := []string{"bash", "sh"} // Github Actions supports bash and sh
	return validShells
}

func (cr *CommandRunner) defaultShell() string {
	return "bash"
}

