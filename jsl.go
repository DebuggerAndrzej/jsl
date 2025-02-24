package main

import (
	"fmt"
	"os"

	"github.com/DebuggerAndrzej/jsl/backend"
	"github.com/DebuggerAndrzej/jsl/ui"
)

func main() {
	config, err := backend.LoadConfig()
	if err != nil {
		exitWithErrorMessage(err)
	}

	client, err := backend.GetJiraClient(config)
	if err != nil {
		exitWithErrorMessage(err)
	}

	ui.StartUi(client, config)
}

func exitWithErrorMessage(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
