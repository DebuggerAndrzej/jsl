<div align="center" width="100%">
    <img src="https://github.com/user-attachments/assets/9cecb781-c7ad-40e6-963c-5cfd95f3e335" width="300">
</div>
<h2 align="center">JSL - Jira Shell Logger</h2>

JSL is a simple shell form written to assist me in daily logging in jira, spiritual successor to [jtl](https://github.com/DebuggerAndrzej/jtl). Project is written in go, it leverages [huh](https://github.com/charmbracelet/huh) for forms generation, and [lipgloss](https://github.com/charmbracelet/lipgloss) for styling.

# Installation
```
 go install github.com/DebuggerAndrzej/jsl@latest
```
Requirements:
- go >= 1.23
- unix system

> default installation path is ~/go/bin so in order to have jsl command available this path has to be added to shell user paths

# Configuration
As of now config file path is hardcoded to `~/.config/jsl.toml`.

Config template:
```
jiraBaseUrl = ""
additionalIssues = "" # comma separated list of Issue Keys. This is an optional argument
```

Also `JIRA_PAT` (jira Personal Access Token) environment variable is required.
