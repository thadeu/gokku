package internal

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func (c *Config) GetAppConfig(appName string) *App {
	if app, exists := c.Apps[appName]; exists {
		return &app
	}
	return nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)

	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Apps: make(map[string]App)}, nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	configPath := GetConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
