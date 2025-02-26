package backend

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	JiraBaseUrl      string
	AdditionalIssues string
}

func LoadConfig() (*Config, error) {
	var config Config

	configPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, errors.New("Couldn't get default config path")
	}
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, errors.New("Couldn't decode config file")
	}

	return &config, nil
}

func getDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Couldn't determine user's home dir!")
	}
	return path.Join(homeDir, ".config/jsl.toml"), nil
}

func AddIssueToConfig(issues string, config *Config) {
	configPath, err := getDefaultConfigPath()

	file, err := os.Create(configPath)
	if err != nil {
		panic(fmt.Sprintf("Couldn't load config file: %s", configPath))
	}

	for _, issue := range strings.Split(issues, ",") {
		issue = strings.TrimSpace(issue)
		if strings.Contains(config.AdditionalIssues, issue) {
			continue
		}
		if "" == config.AdditionalIssues {
			config.AdditionalIssues += fmt.Sprintf("%s", issue)
		} else {
			config.AdditionalIssues += fmt.Sprintf(",%s", issue)
		}
	}
	if err := toml.NewEncoder(file).Encode(config); err != nil {
		panic(fmt.Sprintf("Couldn't update config file: %s", configPath))
	}
	if err = file.Close(); err != nil {
		panic("Couldn't properly close config file")
	}
}

func RemoveIssueFromConfig(issue string, config *Config) {
	configPath, err := getDefaultConfigPath()

	file, err := os.Create(configPath)
	if err != nil {
		panic(fmt.Sprintf("Couldn't load config file: %s", configPath))
	}

	if strings.Contains(config.AdditionalIssues, ","+issue) {
		config.AdditionalIssues = strings.Replace(config.AdditionalIssues, ","+issue, "", -1)
	} else {
		config.AdditionalIssues = strings.Replace(config.AdditionalIssues, issue, "", -1)
	}

	if err := toml.NewEncoder(file).Encode(config); err != nil {
		panic(fmt.Sprintf("Couldn't update config file: %s", configPath))
	}
	if err = file.Close(); err != nil {
		panic("Couldn't properly close config file")
	}
}
