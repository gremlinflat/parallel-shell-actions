package main

type Action struct {
	Shell           string   `json:"shell"`
	Commands        []string `json:"commands"`
	CancelOnFailure bool     `json:"cancel-on-failure"` // nullable (default: false); flag of critical point
}

type CommandResult struct {
	Command string
	Success bool
	ExitCode int
}