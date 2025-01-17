package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type FileMapping struct {
	GistFile  string `yaml:"gist_file"`
	LocalPath string `yaml:"local_path"`
	Exec     string `yaml:"exec,omitempty"`
}

type ClientConfig struct {
	GistID        string        `yaml:"gist_id"`
	CheckInterval string        `yaml:"check_interval"`
	Mappings      []FileMapping `yaml:"mappings"`
}

func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ClientConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *ClientConfig) GetCheckInterval() (time.Duration, error) {
	return time.ParseDuration(c.CheckInterval)
}
