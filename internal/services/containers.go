package services

import (
	"fmt"
	"os/exec"
	"strings"

	"gokku/internal"
	"gokku/internal/containers"
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

// GetContainerInfo gets information about a specific container
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

// ScaleProcesses scales app processes
func (s *ContainerService) ScaleProcesses(appName string, scales map[string]int) error {
	registry := containers.NewContainerRegistry()

	for processType, count := range scales {
		// Get current containers
		currentContainers, err := registry.GetContainers(appName, processType)
		if err != nil {
			return fmt.Errorf("failed to get current containers for %s: %w", processType, err)
		}

		currentCount := len(currentContainers)

		if count > currentCount {
			// Scale up
			if err := s.scaleUp(appName, processType, count-currentCount, registry); err != nil {
				return fmt.Errorf("failed to scale up %s: %w", processType, err)
			}
		} else if count < currentCount {
			// Scale down
			if err := s.scaleDown(appName, processType, currentCount-count, registry); err != nil {
				return fmt.Errorf("failed to scale down %s: %w", processType, err)
			}
		}
	}

	return nil
}

// scaleUp creates new containers
func (s *ContainerService) scaleUp(appName, processType string, count int, registry *containers.ContainerRegistry) error {
	for i := 0; i < count; i++ {
		containerNum := registry.GetNextContainerNumber(appName, processType)
		containerName := fmt.Sprintf("%s-%s-%d", appName, processType, containerNum)

		hostPort, err := containers.GetNextAvailablePort()
		if err != nil {
			return fmt.Errorf("failed to get port for %s: %w", containerName, err)
		}

		// Check if app image exists
		appImage := fmt.Sprintf("%s:latest", appName)
		if !s.imageExists(appImage) {
			return fmt.Errorf("image %s not found, deploy the app first", appImage)
		}

		// Create container
		envFile := fmt.Sprintf("%s/apps/%s/shared/.env", s.baseDir, appName)
		if err := s.createContainer(containerName, appImage, envFile, hostPort); err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerName, err)
		}

		// Save container info
		info := containers.CreateContainerInfo(appName, processType, containerNum, hostPort, hostPort)
		if err := registry.SaveContainerInfo(info); err != nil {
			return fmt.Errorf("failed to save container info: %w", err)
		}
	}

	return nil
}

// scaleDown removes containers
func (s *ContainerService) scaleDown(appName, processType string, count int, registry *containers.ContainerRegistry) error {
	containerList, err := registry.GetContainers(appName, processType)
	if err != nil {
		return fmt.Errorf("failed to get containers: %w", err)
	}

	// Remove last N containers (highest numbers first)
	toRemove := count
	if toRemove > len(containerList) {
		toRemove = len(containerList)
	}

	for i := len(containerList) - toRemove; i < len(containerList); i++ {
		container := containerList[i]
		containerName := container.Name

		// Stop and remove container
		if err := internal.StopContainer(containerName); err != nil {
			return fmt.Errorf("failed to stop container %s: %w", containerName, err)
		}

		if err := internal.RemoveContainer(containerName, false); err != nil {
			return fmt.Errorf("failed to remove container %s: %w", containerName, err)
		}

		// Remove container info
		if err := registry.RemoveContainerInfo(appName, processType, container.Number); err != nil {
			return fmt.Errorf("failed to remove container info: %w", err)
		}
	}

	return nil
}

// createContainer creates a Docker container
func (s *ContainerService) createContainer(name, image, envFile string, port int) error {
	return internal.CreateContainer(internal.ContainerConfig{
		Name:          name,
		Image:         image,
		Ports:         []string{fmt.Sprintf("%d", port)},
		EnvFile:       envFile,
		RestartPolicy: "always",
	})
}

// containsContainerName checks if container name contains the app name
func containsContainerName(names, appName string) bool {
	return strings.Contains(names, appName)
}

// containsProcessType checks if container name contains the process type
func containsProcessType(names, processType string) bool {
	return strings.Contains(names, processType)
}

// imageExists checks if a Docker image exists
func (s *ContainerService) imageExists(imageName string) bool {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}
