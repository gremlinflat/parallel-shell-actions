package main

type Action struct {
	Shell           string   `json:"shell"`
	Commands        []string `json:"commands"`
	CancelOnFailure bool     `json:"cancel-on-failure"`
}

