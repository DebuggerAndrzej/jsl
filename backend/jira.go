package backend

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

type Issue struct {
	Key              string
	Status           string
	ShortDescription string
	OriginalEstimate string
	LoggedTime       string
}

func GetJiraClient(config *Config) (*jira.Client, error) {
	jiraPat := os.Getenv("JIRA_PAT")
	if "" == jiraPat {
		return nil, errors.New("Please set JIRA_PAT env variable")
	}

	tp := jira.BearerAuthTransport{Token: jiraPat}
	client, err := jira.NewClient(tp.Client(), config.JiraBaseUrl)
	if err != nil {
		return nil, errors.New("Couldn't connect to jira server. Please check config and internet connection.")
	}
	return client, nil
}

func GetAllJiraIssuesForAssignee(config *Config, client *jira.Client) ([]Issue, error) {
	var jql string
	if config.AdditionalIssues != "" {
		jql = fmt.Sprintf("assignee = currentuser() OR key in (%s)", config.AdditionalIssues)
	} else {
		jql = "assignee = currentuser()"
	}

	issues, _, err := client.Issue.Search(jql, &jira.SearchOptions{MaxResults: 1000})
	if err != nil {
		return nil, errors.New("Couldn't get issues from Jira API. Check internet connection and vpn if applicable.")
	}

	var mappedIssues []Issue
	for _, issue := range issues {
		if isActiveIssue(issue.Fields.Status.Name) {
			mappedIssues = append(
				mappedIssues,
				Issue{
					Key:              issue.Key,
					Status:           issue.Fields.Status.Name,
					ShortDescription: issue.Fields.Summary,
					OriginalEstimate: getJiraDurationAsString(issue.Fields.TimeOriginalEstimate),
					LoggedTime:       getJiraDurationAsString(issue.Fields.TimeSpent),
				},
			)
		}
	}

	return mappedIssues, nil
}

func LogHoursForIssue(client *jira.Client, id, time string) error {
	_, _, err := client.Issue.AddWorklogRecord(id, &jira.WorklogRecord{TimeSpent: time})
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't add worklog to %s issue", id))
	}
	return nil
}

func LogHoursForIssuesScrumMeetings(client *jira.Client, issueId, timeToLog string) error {
	issueCustomFields, _, err := client.Issue.GetCustomFields(issueId)
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't get %s issue's custom fields", issueId))
	}
	issuesEpic, _, err := client.Issue.Get(issueCustomFields["customfield_12790"], nil)
	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't find epic for %s issue", issueId))
	}
	var scrumIssue string
	for _, issueLink := range issuesEpic.Fields.IssueLinks {
		outwardIssue := issueLink.OutwardIssue
		if outwardIssue != nil && strings.Contains(issueLink.OutwardIssue.Fields.Summary, "Scrum meetings") {
			scrumIssue = issueLink.OutwardIssue.Key
		}
	}
	if scrumIssue != "" {
		return LogHoursForIssue(client, scrumIssue, timeToLog)
	} else {
		return errors.New(fmt.Sprintf("Couldn't find scrum issue for %s issue under %v epic", issueId, issuesEpic))
	}
}

func TransitionToStatus(client *jira.Client, id, status string) error {
	var transitionID string
	possibleTransitions, _, _ := client.Issue.GetTransitions(id)
	for _, v := range possibleTransitions {
		if v.Name == status {
			transitionID = v.ID
			break
		}
	}

	if transitionID != "" {
		_, err := client.Issue.DoTransition(id, transitionID)
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't change status for %s issue", id))
		}
		return nil
	}
	return errors.New(fmt.Sprintf("Couldn't find transitionID required to change issues status for %s issue", id))
}

func isActiveIssue(status string) bool {
	return slices.Contains([]string{"Open", "In Progress", "In Review"}, status)
}

func getJiraDurationAsString(estimate int) string {
	est, _ := time.ParseDuration(fmt.Sprintf("%ds", estimate))
	strEstimate := fmt.Sprintf(
		"%sh",
		strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", est.Hours()), "0"), "."),
	)
	return strEstimate
}
