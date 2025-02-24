package backend

import (
	"errors"
	"os"
	"path"

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
