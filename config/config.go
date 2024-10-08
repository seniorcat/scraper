package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config описывает структуру конфигурационного файла
type Config struct {
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`

	Worker struct {
		Type          int `yaml:"type"`
		MaxRecipes    int `yaml:"maxRecipes"`
		Timeout       int `yaml:"timeout"`
		MaxRetries    int `yaml:"maxRetries"`
		RetryInterval int `yaml:"retryInterval"`
		Concurrency   int `yaml:"concurrency"`
		RPS           int `yaml:"rps"`
	} `yaml:"worker"`
}

// LoadConfig загружает конфигурацию из файла YAML
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
