package pkg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const (
	GokkuLabelKey   = "createdby"
	GokkuLabelValue = "gokku"
)

func GetGokkuLabels() []string {
	return []string{fmt.Sprintf("%s=%s", GokkuLabelKey, GokkuLabelValue)}
}

// ListContainers returns list of containers in JSON format
// By default, only lists containers with Gokku labels to avoid conflicts
func ListContainers(all bool) ([]ContainerInfo, error) {
	args := []string{"ps", "--format", "json", "--filter", fmt.Sprintf("label=%s=%s", GokkuLabelKey, GokkuLabelValue)}

	if all {
		args = []string{"ps", "-a", "--format", "json", "--filter", fmt.Sprintf("label=%s=%s", GokkuLabelKey, GokkuLabelValue)}
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	var containers []ContainerInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var container ContainerInfo

		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue // Skip invalid JSON lines
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// ContainerExists checks if a container exists
// First checks containers with Gokku labels, then falls back to direct name check for backwards compatibility
func ContainerExists(name string) bool {
	// First try with label filter
	containers, err := ListContainers(true)
	if err == nil {
		for _, container := range containers {
			if strings.Contains(container.Name, name) {
				return true
			}
		}
	}

	// Fallback: check if container exists by trying to get its info
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == name {
			return true
		}
	}

	return false
}

// ContainerIsRunning checks if a container is currently running
func ContainerIsRunning(name string) bool {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", name))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == name {
			return true
		}
	}

	return false
}

// StopContainer stops a running container
func StopContainer(name string) error {
	cmd := exec.Command("docker", "stop", name)
	return cmd.Run()
}

// RemoveContainer removes a container
func RemoveContainer(name string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	cmd := exec.Command("docker", args...)
	return cmd.Run()
}

// DeploymentConfig represents deployment configuration
type DeploymentConfig struct {
	AppName     string
	ImageTag    string
	EnvFile     string
	ReleaseDir  string
	NetworkMode string
	DockerPorts []string
	Volumes     []string
}

// DeployContainer deploys a container using zero-downtime or standard deployment
// TODO: Implement full logic from internal/docker.go
func DeployContainer(config DeploymentConfig) error {
	return fmt.Errorf("DeployContainer not implemented yet - needs migration from internal/docker.go")
}
