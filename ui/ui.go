package ui

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/DebuggerAndrzej/jsl/backend"
	jira "github.com/andygrunwald/go-jira"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

type LogData struct {
	Key              string
	ShortDescription string
	StandardTime     string
	ScrumTime        string
	Status           string
	OriginalStatus   string
	IsSelected       bool
}

var ISSUE_STATUSES = [5]string{"Open", "In Progress", "In Review", "Done", "Obsolete"}

func StartUi(client *jira.Client, config *backend.Config) {
	var issues []backend.Issue
	var chosenIndexes []int
	var err error
	var confirm bool
	fetchingContext, fetchingCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer fetchingCancel()
	err = spinner.New().
		Title("Fetching jira issues...").
		Context(fetchingContext).
		Action(func() {
			issues = getJiraIssues(client, config)
		}).
		Context(fetchingContext).
		Run()

	if err != nil {
		fmt.Println("Fetching jira issues failed, probably due to 10s timeout...")
	}

	logData := prepareLogData(issues)
	var loggingForm, confirmationForm, pickingForm *huh.Form
	for !confirm {
		confirm = true
		pickingForm = huh.NewForm(huh.NewGroup(huh.NewMultiSelect[int]().
			Title("Pick issues to edit:").
			Options(getOptions(issues)...).
			Value(&chosenIndexes)),
		)

		err = pickingForm.Run()
		if err != nil {
			fmt.Println("Issue picking form failed.")
			os.Exit(1)
		}
		for index := 0; index < len(logData); index++ {
			logEntry := &logData[index]
			if slices.Contains(chosenIndexes, index) {
				logEntry.IsSelected = true
			} else {
				logEntry.IsSelected = false
			}
		}

		loggingForm = huh.NewForm(getGroupsForIssues(chosenIndexes, &logData)...)
		err = loggingForm.Run()
		if err != nil {
			fmt.Println("Logging form failed.")
			os.Exit(1)
		}

		confirmationForm = huh.NewForm(
			huh.NewGroup(huh.NewConfirm().Title("Logging summary:").Description(getSummary(logData)).Value(&confirm)),
		)
		err = confirmationForm.Run()
		if err != nil {
			fmt.Println("Confirmation form failed.")
			os.Exit(1)
		}
	}

	loggingContext, loggingCancel := context.WithTimeout(context.Background(), time.Second*20)
	defer loggingCancel()
	err = spinner.New().
		Title("Logging...").
		Context(loggingContext).
		Action(func() {
			logOnJira(client, logData)
		}).
		Context(loggingContext).
		Run()

	if err != nil {
		fmt.Println("Logging jira issues failed, probably due to 20s timeout...")
	}

}

func prepareLogData(issues []backend.Issue) []LogData {
	logData := make([]LogData, len(issues))
	for index, issue := range issues {
		logData[index] = LogData{
			Key:              issue.Key,
			Status:           issue.Status + " (current)",
			OriginalStatus:   issue.Status,
			ShortDescription: issue.ShortDescription,
		}
	}
	return logData
}

func getSummary(logData []LogData) string {
	var summary string
	var loggedInSession float64
	for _, logEntry := range logData {
		if !logEntry.IsSelected {
			continue
		}
		var issueData string
		if "" != logEntry.StandardTime {
			logged := getParsedTime(logEntry.StandardTime, logEntry.Key, "Standard")
			issueData += fmt.Sprintf("StandardTime: %.2fh ", logged)
			loggedInSession += logged
		}
		if "" != logEntry.ScrumTime {
			logged := getParsedTime(logEntry.ScrumTime, logEntry.Key, "Scrum")
			issueData += fmt.Sprintf("ScrumTime: %.2fh ", logged)
			loggedInSession += logged
		}
		if !strings.HasSuffix(logEntry.Status, " (current)") {
			issueData += fmt.Sprintf("Status: %s -> %s ", logEntry.OriginalStatus, logEntry.Status)
		}
		if "" != issueData {
			summary += fmt.Sprintf("%s\n%s\n", logEntry.Key, issueData)
		}
	}

	if "" != summary {
		summary = fmt.Sprintf("Total logged time: %.2fh\n", loggedInSession) + summary
	}
	return summary
}

func getJiraIssues(client *jira.Client, config *backend.Config) []backend.Issue {
	issues, err := backend.GetAllJiraIssuesForAssignee(config, client)
	if err != nil {
		fmt.Println("error")
		os.Exit(1)
	}
	return issues
}

func getOptions(issues []backend.Issue) []huh.Option[int] {
	var issueChoices []huh.Option[int]
	longestKey, longestEstimate, longestStatus := getTextPaddings(issues)
	for index, issue := range issues {
		estimate := fmt.Sprintf("(%s of %s)", issue.LoggedTime, issue.OriginalEstimate)
		issueChoices = append(
			issueChoices,
			huh.NewOption(
				fmt.Sprintf(
					"%-*s - %-*s - %-*s - %s",
					longestKey,
					issue.Key,
					longestStatus,
					issue.Status,
					longestEstimate,
					estimate,
					issue.ShortDescription,
				),
				index,
			),
		)
	}
	return issueChoices
}

func getParsedTime(timeStr, key, logType string) float64 {
	logged, err := time.ParseDuration(timeStr)
	if err != nil {
		logged, err = time.ParseDuration(timeStr + "h")
		if err != nil {
			fmt.Printf("Couldn't parse time duration for %s %s: %s", key, logType, timeStr)
		}
	}
	return logged.Hours()
}

func getGroupsForIssues(issuesIndexes []int, logData *[]LogData) []*huh.Group {
	var issueGroups []*huh.Group
	for index, issue := range *logData {
		if slices.Contains(issuesIndexes, index) {
			issueGroups = append(issueGroups, huh.NewGroup(
				huh.NewNote().Title(fmt.Sprintf("---- %s ----", issue.Key)).Description(issue.ShortDescription),
				huh.NewInput().Title("Standard hours: ").Value(&(*logData)[index].StandardTime),
				huh.NewInput().Title("Scrum hours: ").Value(&(*logData)[index].ScrumTime),
				huh.NewSelect[string]().Title("Status: ").
					Options(getOptionsForStatus(issue.OriginalStatus)...).
					Value(&(*logData)[index].Status),
			))
		}
	}
	return issueGroups
}

func getOptionsForStatus(currentStatus string) []huh.Option[string] {
	var statusChoices []huh.Option[string]
	for _, status := range ISSUE_STATUSES {
		if status == currentStatus {
			status += " (current)"
		}
		statusChoices = append(statusChoices, huh.NewOption(status, status))
	}
	return statusChoices
}

func getTextPaddings(issues []backend.Issue) (int, int, int) {
	var longestKey, longestEstimate, longestStatus int

	for _, issue := range issues {
		estimate := fmt.Sprintf("(%s of %s)", issue.LoggedTime, issue.OriginalEstimate)
		if len(issue.Key) > longestKey {
			longestKey = len(issue.Key)
		}
		if len(estimate) > longestEstimate {
			longestEstimate = len(estimate)
		}
		if len(issue.Status) > longestStatus {
			longestStatus = len(issue.Status)
		}
	}
	return longestKey, longestEstimate, longestStatus
}

func printSuccessLog(message string) {
	hours, minutes, seconds := time.Now().Clock()
	fmt.Println(fmt.Sprintf("%d:%d:%d\033[32m - %s\033[0m", hours, minutes, seconds, message))
}

func printErrorLog(message string) {
	hours, minutes, seconds := time.Now().Clock()
	fmt.Println(fmt.Sprintf("%d:%d:%d\033[31m - %s\033[0m", hours, minutes, seconds, message))
}

func logOnJira(client *jira.Client, logData []LogData) {
	var wg sync.WaitGroup
	var err error
	for _, logEntry := range logData {
		if !logEntry.IsSelected {
			continue
		}
		if "" != logEntry.StandardTime {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = backend.LogHoursForIssue(client, logEntry.Key, logEntry.StandardTime)
				if err != nil {
					printErrorLog(fmt.Sprintf("Couldn't log %s under %s: %v", logEntry.StandardTime, logEntry.Key, err))
				} else {
					printSuccessLog(fmt.Sprintf("Successfully logged %s standard time under %s", logEntry.StandardTime, logEntry.Key))
				}
			}()
		}
		if "" != logEntry.ScrumTime {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = backend.LogHoursForIssuesScrumMeetings(client, logEntry.Key, logEntry.ScrumTime)
				if err != nil {
					printErrorLog(
						fmt.Sprintf(
							"Couldn't log %s under %s scrum meetings: %v",
							logEntry.ScrumTime,
							logEntry.Key,
							err,
						),
					)
				} else {
					printSuccessLog(fmt.Sprintf("Successfully logged %s scrum time under %s scrum meetings", logEntry.ScrumTime, logEntry.Key))
				}
			}()
		}
		if !strings.HasSuffix(logEntry.Status, " (current)") {
			wg.Add(1)
			go func() {
				defer wg.Done()
				backend.TransitionToStatus(client, logEntry.Key, logEntry.Status)
				if err != nil {
					printErrorLog(
						fmt.Sprintf(
							"Couldn't transition issue %s to %s status: %v",
							logEntry.Key,
							logEntry.Status,
							err,
						),
					)
				} else {
					printSuccessLog(fmt.Sprintf("Successfully transitioned issue %s to %s status", logEntry.Key, logEntry.Status))
				}
			}()
		}
	}
	wg.Wait()
}
