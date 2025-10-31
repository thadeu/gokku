package services

import "gokku/internal"

// AppInfo represents information about an app
type AppInfo struct {
	Name           string
	Status         string
	ReleasesCount  int
	CurrentRelease string
}

// AppDetail represents detailed information about an app
type AppDetail struct {
	AppInfo
	Config     *internal.App
	Containers []internal.ContainerInfo
	EnvVars    map[string]string
}

// ContainerFilter represents filter options for listing containers
type ContainerFilter struct {
	AppName     string
	ProcessType string
	All         bool
}

// ConfigOperation represents a configuration operation
type ConfigOperation struct {
	Type  string // "set", "get", "list", "unset"
	Key   string
	Value string
	Keys  []string // for unset multiple
}

// Error types
type AppNotFoundError struct {
	AppName string
}

func (e *AppNotFoundError) Error() string {
	return "app not found: " + e.AppName
}

type ContainerNotFoundError struct {
	ContainerName string
}

func (e *ContainerNotFoundError) Error() string {
	return "container not found: " + e.ContainerName
}
