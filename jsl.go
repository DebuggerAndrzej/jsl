package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DebuggerAndrzej/jsl/backend"
	"github.com/DebuggerAndrzej/jsl/ui"
)

func main() {
	var dayDelta = flag.Int("d", 0, "Day delta from today (might be negative for past days)")
	var issuesToAdd = flag.String("a", "", "Issue(s) to add. Multiple issues must be comma separated: IS_SUE-22,IS_SUE-11")
	var issueToDelete = flag.String("r", "", "Issue to delete. Only one issue can be passed. Done issues are automatically deleted")
	flag.Parse()

	config, err := backend.LoadConfig()
	if err != nil {
		exitWithErrorMessage(err)
	}

	if "" != *issuesToAdd {
		backend.AddIssueToConfig(*issuesToAdd, config)
	}
	if "" != *issueToDelete {
		backend.RemoveIssueFromConfig(*issueToDelete, config)
	}
	client, err := backend.GetJiraClient(config)
	if err != nil {
		exitWithErrorMessage(err)
	}

	ui.StartUi(client, config, *dayDelta)
}

func exitWithErrorMessage(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
