package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"gokku/internal"
)

// ConfigService provides operations for managing app configuration
type ConfigService struct {
	baseDir string
}

// NewConfigService creates a new ConfigService
func NewConfigService(baseDir string) *ConfigService {
	if baseDir == "" {
		baseDir = "/opt/gokku"
	}
	return &ConfigService{baseDir: baseDir}
}

// SetEnvVar sets one or more environment variables
func (s *ConfigService) SetEnvVar(appName string, keyValues []string) error {
	envFile := s.getEnvFilePath(appName)
	envVars := internal.LoadEnvFile(envFile)

	// Parse and update env vars
	for _, pair := range keyValues {
		parts := strings.SplitN(pair, "=", 2)

		if len(parts) != 2 {
			return fmt.Errorf("invalid format '%s', expected KEY=VALUE", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}

	return internal.SaveEnvFile(envFile, envVars)
}

// GetEnvVar gets an environment variable value
func (s *ConfigService) GetEnvVar(appName, key string) (string, error) {
	envVars := s.ListEnvVars(appName)

	value, ok := envVars[key]

	if !ok {
		return "", fmt.Errorf("variable '%s' not found", key)
	}

	return value, nil
}

// ListEnvVars lists all environment variables
func (s *ConfigService) ListEnvVars(appName string) map[string]string {
	envFile := s.getEnvFilePath(appName)

	return internal.LoadEnvFile(envFile)
}

// UnsetEnvVar removes one or more environment variables
func (s *ConfigService) UnsetEnvVar(appName string, keys []string) error {
	envFile := s.getEnvFilePath(appName)
	envVars := internal.LoadEnvFile(envFile)

	for _, key := range keys {
		delete(envVars, key)
	}

	return internal.SaveEnvFile(envFile, envVars)
}

// ReloadApp restarts/recreates the app container to apply config changes
func (s *ConfigService) ReloadApp(appName string) error {
	envFile := s.getEnvFilePath(appName)
	appDir := fmt.Sprintf("%s/apps/%s/current", s.baseDir, appName)
	return internal.RecreateActiveContainer(appName, envFile, appDir)
}

// getEnvFilePath returns the path to the .env file for an app
func (s *ConfigService) getEnvFilePath(appName string) string {
	return filepath.Join(s.baseDir, "apps", appName, "shared", ".env")
}
