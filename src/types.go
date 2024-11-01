package main

type Action struct {
	Namespace	    string   `json:"name"`
	Shell           string   `json:"shell"`
	Commands        []string `json:"commands"`
	CancelOnFailure bool     `json:"cancel-on-failure"` // nullable (default: false); flag of critical point
}

type CommandResult struct {
	Namespace string
	Command string
	Success bool
	ExitCode int // -3 for error starting command, -2 for error creating pipe, -99 for context cancellation
}