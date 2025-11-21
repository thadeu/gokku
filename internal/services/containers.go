package services

import (
	"fmt"
	"os/exec"
	"strings"

	"gokku/internal"
)

// ContainerService provides operations for managing containers
type ContainerService struct {
	baseDir string
}

// NewContainerService creates a new ContainerService
func NewContainerService(baseDir string) *ContainerService {
	if baseDir == "" {
		baseDir = "/opt/gokku"
	}
	return &ContainerService{baseDir: baseDir}
}

// ListContainers returns containers based on filter
func (s *ContainerService) ListContainers(filter ContainerFilter) ([]internal.ContainerInfo, error) {
	allContainers, err := internal.ListContainers(filter.All)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	if filter.AppName == "" && filter.ProcessType == "" {
		return allContainers, nil
	}

	var filtered []internal.ContainerInfo
	for _, c := range allContainers {
		match := true

		if filter.AppName != "" {
			if !containsContainerName(c.Names, filter.AppName) {
				match = false
			}
		}

		if filter.ProcessType != "" && match {
			if !containsProcessType(c.Names, filter.ProcessType) {
				match = false
			}
		}

		if match {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

// RestartContainer restarts a container
func (s *ContainerService) RestartContainer(name string) error {
	cmd := exec.Command("docker", "restart", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart container %s: %v, output: %s", name, err, string(output))
	}
	return nil
}

// StopContainer stops a container
func (s *ContainerService) StopContainer(name string) error {
	return internal.StopContainer(name)
}

// StartContainer starts a container
func (s *ContainerService) StartContainer(name string) error {
	if !internal.ContainerExists(name) {
		return &ContainerNotFoundError{ContainerName: name}
	}
	cmd := exec.Command("docker", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container %s: %v, output: %s", name, err, string(output))
	}
	return nil
}

func (s *ContainerService) GetContainerInfo(name string) (*internal.ContainerInfo, error) {
	containers, err := internal.ListContainers(true)
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		if strings.Contains(c.Names, name) {
			return &c, nil
		}
	}

	return nil, &ContainerNotFoundError{ContainerName: name}
}

// containsContainerName checks if container name contains the app name
func containsContainerName(names, appName string) bool {
	return strings.Contains(names, appName)
}

// containsProcessType checks if container name contains the process type
func containsProcessType(names, processType string) bool {
	return strings.Contains(names, processType)
}
