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

// Config represents the CLI configuration
type Config struct {
	Apps []App `yaml:"apps"`
}

// App represents an application configuration
type App struct {
	Name         string         `yaml:"name"`
	Lang         string         `yaml:"lang,omitempty"`
	Build        *Build         `yaml:"build,omitempty"`
	Deployment   *Deployment    `yaml:"deployment,omitempty"`
	Network      *NetworkConfig `yaml:"network"`
	Ports        []string       `yaml:"ports"`
	Environments []Environment  `yaml:"environments,omitempty"`
}

// RemoteInfo contains information about remote connection
type RemoteInfo struct {
	Host    string
	BaseDir string
	App     string
}

// Build represents build configuration
type Build struct {
	Type       string            `yaml:"type"` // "docker"
	Path       string            `yaml:"path"`
	BinaryName string            `yaml:"binary_name,omitempty"`
	GoVersion  string            `yaml:"go_version,omitempty"`
	Goos       string            `yaml:"goos,omitempty"`
	Goarch     string            `yaml:"goarch,omitempty"`
	CgoEnabled *bool             `yaml:"cgo_enabled,omitempty"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	BaseImage  string            `yaml:"base_image,omitempty"`
	Workdir    string            `yaml:"workdir,omitempty"`
	Entrypoint string            `yaml:"entrypoint,omitempty"`
	Command    string            `yaml:"command,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	Ports      []string          `yaml:"ports,omitempty"`
	Volumes    []string          `yaml:"volumes,omitempty"`
	Security   string            `yaml:"security,omitempty"`
}

type NetworkConfig struct {
	Mode string `yaml:"mode,omitempty"`
}

// Deployment represents deployment configuration
type Deployment struct {
	KeepReleases  int      `yaml:"keep_releases,omitempty"`
	KeepImages    int      `yaml:"keep_images,omitempty"`
	RestartPolicy string   `yaml:"restart_policy,omitempty"`
	RestartDelay  int      `yaml:"restart_delay,omitempty"`
	PostDeploy    []string `yaml:"post_deploy,omitempty"`
}

// Environment represents environment-specific app config
type Environment struct {
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

// LoadServerConfig loads the server configuration from gokku.yml
func LoadServerConfigByApp(appName string) (*ServerConfig, error) {
	configPath := fmt.Sprintf("/opt/gokku/apps/%s/gokku.yml", appName)

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

func LoadAppConfig(appName string) (*App, error) {
	filePath := fmt.Sprintf("/opt/gokku/apps/%s/gokku.yml", appName)

	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var serverConfig ServerConfig

	if err := yaml.Unmarshal(data, &serverConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Find the app by name
	for _, app := range serverConfig.Apps {
		if app.Name == appName {
			// Ensure the app name is set
			app.Name = appName
			return &app, nil
		}
	}

	return nil, fmt.Errorf("app '%s' not found in configuration", appName)
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

		if app.Build.Type != "docker" {
			return fmt.Errorf("app '%s' has invalid build type: %s (must be 'docker')", app.Name, app.Build.Type)
		}
	}

	return nil
}
