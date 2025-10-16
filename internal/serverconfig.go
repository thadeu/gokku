package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig represents the gokku.yml configuration on the server
type ServerConfig struct {
	Project      *Project      `yaml:"project,omitempty"`
	Apps         []App         `yaml:"apps"`
	Defaults     *Defaults     `yaml:"defaults,omitempty"`
	Docker       *Docker       `yaml:"docker,omitempty"`
	Environments []Environment `yaml:"environments,omitempty"`
}

// Project represents project-level configuration
type Project struct {
	Name string `yaml:"name"`
}

// App represents an application configuration
type App struct {
	Name         string           `yaml:"name"`
	Lang         string           `yaml:"lang,omitempty"`
	Build        *Build           `yaml:"build,omitempty"`
	Deployment   *Deployment      `yaml:"deployment,omitempty"`
	Environments []AppEnvironment `yaml:"environments,omitempty"`
}

// Build represents build configuration
type Build struct {
	Type       string `yaml:"type"` // "systemd" or "docker"
	Path       string `yaml:"path"`
	BinaryName string `yaml:"binary_name,omitempty"`
	GoVersion  string `yaml:"go_version,omitempty"`
	Dockerfile string `yaml:"dockerfile,omitempty"`
	BaseImage  string `yaml:"base_image,omitempty"`
}

// Deployment represents deployment configuration
type Deployment struct {
	KeepReleases  int      `yaml:"keep_releases,omitempty"`
	KeepImages    int      `yaml:"keep_images,omitempty"`
	RestartPolicy string   `yaml:"restart_policy,omitempty"`
	RestartDelay  int      `yaml:"restart_delay,omitempty"`
	PostDeploy    []string `yaml:"post_deploy,omitempty"`
}

// AppEnvironment represents environment-specific app config
type AppEnvironment struct {
	Name           string            `yaml:"name"`
	Branch         string            `yaml:"branch,omitempty"`
	DefaultEnvVars map[string]string `yaml:"default_env_vars,omitempty"`
}

// Defaults represents default configurations
type Defaults struct {
	BuildType string `yaml:"build_type,omitempty"`
}

// Docker represents Docker-related configurations
type Docker struct {
	Registry   string            `yaml:"registry,omitempty"`
	BaseImages map[string]string `yaml:"base_images,omitempty"`
}

// Environment represents global environment configuration
type Environment struct {
	Name string `yaml:"name"`
}

// LoadServerConfig loads the server configuration from gokku.yml
func LoadServerConfig() (*ServerConfig, error) {
	configPath := "/opt/gokku/gokku.yml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ServerConfig{
			Apps: []App{},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// GetApp finds an app by name
func (c *ServerConfig) GetApp(name string) (*App, error) {
	for _, app := range c.Apps {
		if app.Name == name {
			return &app, nil
		}
	}
	return nil, fmt.Errorf("app '%s' not found", name)
}

// Validate validates the server configuration
func (c *ServerConfig) Validate() error {
	if len(c.Apps) == 0 {
		return fmt.Errorf("no apps defined")
	}

	appNames := make(map[string]bool)
	for _, app := range c.Apps {
		if app.Name == "" {
			return fmt.Errorf("app name cannot be empty")
		}
		if appNames[app.Name] {
			return fmt.Errorf("duplicate app name: %s", app.Name)
		}
		appNames[app.Name] = true

		if app.Build == nil {
			return fmt.Errorf("app '%s' missing build configuration", app.Name)
		}

		if app.Build.Type != "systemd" && app.Build.Type != "docker" {
			return fmt.Errorf("app '%s' has invalid build type: %s (must be 'systemd' or 'docker')", app.Name, app.Build.Type)
		}
	}

	return nil
}
