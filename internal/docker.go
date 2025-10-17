package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/client"
)

// ContainerInfo represents container information from docker ps --format json
type ContainerInfo struct {
	ID      string `json:"ID"`
	Names   string `json:"Names"`
	Image   string `json:"Image"`
	Status  string `json:"Status"`
	Ports   string `json:"Ports"`
	Command string `json:"Command"`
	Created string `json:"CreatedAt"`
}

// ContainerConfig represents configuration for creating a container
type ContainerConfig struct {
	Name          string
	Image         string
	Ports         []string
	EnvFile       string
	NetworkMode   string
	RestartPolicy string
	Volumes       []string
	WorkingDir    string
	Command       []string
}

// DeploymentConfig represents configuration for deployment
type DeploymentConfig struct {
	AppName       string
	ImageTag      string
	EnvFile       string
	ReleaseDir    string
	ZeroDowntime  bool
	HealthTimeout int
	NetworkMode   string
	DockerPorts   []string
}

// DockerClient wraps the Docker client with additional functionality
type DockerClient struct {
	client *client.Client
	ctx    context.Context
}

// NewDockerClient creates a new Docker client
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &DockerClient{
		client: cli,
		ctx:    context.Background(),
	}, nil
}

// ListContainers returns list of containers in JSON format
func (dc *DockerClient) ListContainers(all bool) ([]ContainerInfo, error) {
	cmd := exec.Command("docker", "ps", "--format", "json")
	if all {
		cmd = exec.Command("docker", "ps", "-a", "--format", "json")
	}

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
func (dc *DockerClient) ContainerExists(name string) bool {
	containers, err := dc.ListContainers(true)
	if err != nil {
		return false
	}

	for _, container := range containers {
		if strings.Contains(container.Names, name) {
			return true
		}
	}
	return false
}

// ContainerIsRunning checks if a container is running
func (dc *DockerClient) ContainerIsRunning(name string) bool {
	containers, err := dc.ListContainers(false)
	if err != nil {
		return false
	}

	for _, container := range containers {
		if strings.Contains(container.Names, name) {
			return true
		}
	}
	return false
}

// StopContainer stops a container
func (dc *DockerClient) StopContainer(name string) error {
	cmd := exec.Command("docker", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %v, output: %s", name, err, string(output))
	}
	return nil
}

// RemoveContainer removes a container
func (dc *DockerClient) RemoveContainer(name string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %v, output: %s", name, err, string(output))
	}
	return nil
}

// CreateContainer creates a new container with the given configuration
func (dc *DockerClient) CreateContainer(config ContainerConfig) error {
	args := []string{"run", "-d", "--name", config.Name}

	// Add restart policy
	if config.RestartPolicy != "" {
		args = append(args, "--restart", config.RestartPolicy)
	}

	// Add network mode
	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}

	// Add port mappings
	for _, port := range config.Ports {
		args = append(args, "-p", port)
	}

	// Add environment file
	if config.EnvFile != "" && fileExists(config.EnvFile) {
		args = append(args, "--env-file", config.EnvFile)
	}

	// Add volumes
	for _, volume := range config.Volumes {
		args = append(args, "-v", volume)
	}

	// Add working directory
	if config.WorkingDir != "" {
		args = append(args, "-w", config.WorkingDir)
	}

	// Add image
	args = append(args, config.Image)

	// Add command if specified
	if len(config.Command) > 0 {
		args = append(args, config.Command...)
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create container %s: %v, output: %s", config.Name, err, string(output))
	}

	return nil
}

// GetContainerPort extracts port from environment file
func GetContainerPort(envFile string, defaultPort int) int {
	if !fileExists(envFile) {
		return defaultPort
	}

	content, err := os.ReadFile(envFile)
	if err != nil {
		return defaultPort
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PORT=") {
			portStr := strings.TrimSpace(strings.TrimPrefix(line, "PORT="))
			if port, err := strconv.Atoi(portStr); err == nil {
				return port
			}
		}
	}

	return defaultPort
}

// IsZeroDowntimeEnabled checks if zero downtime deployment is enabled
func IsZeroDowntimeEnabled(envFile string) bool {
	if !fileExists(envFile) {
		return true // Default: enabled
	}

	content, err := os.ReadFile(envFile)
	if err != nil {
		return true
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ZERO_DOWNTIME=") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "ZERO_DOWNTIME="))
			switch strings.ToLower(value) {
			case "0", "false", "no", "off", "n":
				return false
			case "1", "true", "yes", "on", "y":
				return true
			default:
				return true // Default: enabled
			}
		}
	}

	return true // Default: enabled
}

// WaitForContainerHealth waits for container to be healthy
func (dc *DockerClient) WaitForContainerHealth(name string, timeout int) error {
	startTime := time.Now()
	maxWait := time.Duration(timeout) * time.Second

	fmt.Printf("-----> Waiting for container to be healthy (max %ds)...\n", timeout)

	for {
		if time.Since(startTime) > maxWait {
			return fmt.Errorf("container failed to become healthy within %ds", timeout)
		}

		// Check container status using docker inspect
		cmd := exec.Command("docker", "inspect", name, "--format", "{{.State.Health.Status}}")
		output, err := cmd.Output()
		if err != nil {
			// No health check configured, assume ready after a short wait
			time.Sleep(3 * time.Second)
			fmt.Println("-----> Container ready (no health check configured)")
			return nil
		}

		status := strings.TrimSpace(string(output))
		elapsed := int(time.Since(startTime).Seconds())

		switch status {
		case "healthy":
			fmt.Println("-----> Container is healthy!")
			return nil
		case "starting":
			fmt.Printf("       Starting... (%d/%ds)\n", elapsed, timeout)
			time.Sleep(2 * time.Second)
		case "unhealthy":
			// Get container logs for debugging
			logCmd := exec.Command("docker", "logs", name)
			logOutput, _ := logCmd.Output()
			return fmt.Errorf("container is unhealthy, logs: %s", string(logOutput))
		default:
			// Unknown status, wait a bit more
			time.Sleep(2 * time.Second)
		}
	}
}

// StandardDeploy performs standard deployment (kill and restart)
func (dc *DockerClient) StandardDeploy(config DeploymentConfig) error {
	fmt.Println("=====> Starting Standard Deployment")

	blueName := config.AppName + "-blue"

	// Stop and remove old container
	if dc.ContainerExists(blueName) {
		fmt.Printf("-----> Stopping old container: %s\n", blueName)
		if err := dc.StopContainer(blueName); err != nil {
			fmt.Printf("Warning: Failed to stop container: %v\n", err)
		}
		if err := dc.RemoveContainer(blueName, true); err != nil {
			fmt.Printf("Warning: Failed to remove container: %v\n", err)
		}
		time.Sleep(2 * time.Second)
	}

	// Get container port
	containerPort := GetContainerPort(config.EnvFile, 8080)

	// Build container configuration
	containerConfig := ContainerConfig{
		Name:          blueName,
		Image:         fmt.Sprintf("%s:%s", config.AppName, config.ImageTag),
		NetworkMode:   config.NetworkMode,
		RestartPolicy: "no",
		WorkingDir:    "/app",
		Volumes:       []string{fmt.Sprintf("%s:/app", config.ReleaseDir)},
	}

	// Add port mappings
	if config.NetworkMode != "host" {
		if len(config.DockerPorts) > 0 {
			containerConfig.Ports = config.DockerPorts
			fmt.Println("-----> Using ports from gokku.yml")
		} else {
			containerConfig.Ports = []string{fmt.Sprintf("%d:%d", containerPort, containerPort)}
		}
	} else {
		fmt.Println("-----> Using host network (all ports exposed)")
	}

	// Add environment file
	if fileExists(config.EnvFile) {
		containerConfig.EnvFile = config.EnvFile
	}

	// Create new container
	fmt.Printf("-----> Starting new container: %s\n", blueName)
	if err := dc.CreateContainer(containerConfig); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for container to be ready
	fmt.Println("-----> Waiting for container to be ready...")
	time.Sleep(5 * time.Second)

	// Check if container is running
	if !dc.ContainerIsRunning(blueName) {
		// Get container logs for debugging
		logCmd := exec.Command("docker", "logs", blueName)
		logOutput, _ := logCmd.Output()
		return fmt.Errorf("container failed to start, logs: %s", string(logOutput))
	}

	fmt.Println("=====> Standard Deployment Complete!")
	fmt.Printf("-----> Active container: %s\n", blueName)
	fmt.Printf("-----> Running image: %s\n", config.ImageTag)
	fmt.Printf("-----> Port: %d\n", containerPort)

	return nil
}

// BlueGreenDeploy performs blue/green deployment
func (dc *DockerClient) BlueGreenDeploy(config DeploymentConfig) error {
	fmt.Println("=====> Starting Blue/Green Deployment")

	// Get container port
	containerPort := GetContainerPort(config.EnvFile, 8080)

	// Start green container
	if err := dc.startGreenContainer(config, containerPort); err != nil {
		return fmt.Errorf("failed to start green container: %v", err)
	}

	// Wait for green to be healthy
	if err := dc.WaitForContainerHealth(config.AppName+"-green", config.HealthTimeout); err != nil {
		// Cleanup green container on failure
		dc.StopContainer(config.AppName + "-green")
		dc.RemoveContainer(config.AppName+"-green", true)
		return fmt.Errorf("green container failed health check: %v", err)
	}

	// Check if we have an existing blue container
	blueName := config.AppName + "-blue"
	if dc.ContainerExists(blueName) {
		// Switch traffic: blue → green
		if err := dc.switchTrafficBlueToGreen(config.AppName, containerPort); err != nil {
			// Cleanup green container on failure
			dc.StopContainer(config.AppName + "-green")
			dc.RemoveContainer(config.AppName+"-green", true)
			return fmt.Errorf("failed to switch traffic: %v", err)
		}

		// Cleanup old blue
		dc.cleanupOldBlueContainer(config.AppName)
	} else {
		// First deployment, just rename green to blue
		fmt.Println("-----> First deployment, activating green as blue")
		if err := dc.renameContainer(config.AppName+"-green", blueName); err != nil {
			return fmt.Errorf("failed to rename green to blue: %v", err)
		}
		dc.updateContainerRestartPolicy(blueName, "always")
	}

	fmt.Println("=====> Blue/Green Deployment Complete!")
	fmt.Printf("-----> Active container: %s\n", blueName)
	fmt.Printf("-----> Running image: %s\n", config.ImageTag)
	fmt.Printf("-----> Port: %d\n", containerPort)

	return nil
}

// DeployContainer determines and executes deployment strategy
func (dc *DockerClient) DeployContainer(config DeploymentConfig) error {
	if IsZeroDowntimeEnabled(config.EnvFile) {
		fmt.Println("=====> ZERO_DOWNTIME deployment enabled")
		return dc.BlueGreenDeploy(config)
	} else {
		fmt.Println("=====> ZERO_DOWNTIME deployment disabled")
		return dc.StandardDeploy(config)
	}
}

// Helper functions

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (dc *DockerClient) startGreenContainer(config DeploymentConfig, containerPort int) error {
	greenName := config.AppName + "-green"
	fmt.Printf("-----> Starting green container: %s\n", greenName)

	// Stop and remove old green container if exists
	if dc.ContainerExists(greenName) {
		fmt.Println("       Removing old green container...")
		dc.StopContainer(greenName)
		dc.RemoveContainer(greenName, true)
	}

	// Build container configuration
	containerConfig := ContainerConfig{
		Name:          greenName,
		Image:         fmt.Sprintf("%s:%s", config.AppName, config.ImageTag),
		NetworkMode:   config.NetworkMode,
		RestartPolicy: "no",
		WorkingDir:    "/app",
		Volumes:       []string{fmt.Sprintf("%s:/app", config.ReleaseDir)},
	}

	// Add port mappings
	if config.NetworkMode != "host" {
		if len(config.DockerPorts) > 0 {
			containerConfig.Ports = config.DockerPorts
		} else {
			containerConfig.Ports = []string{fmt.Sprintf("%d:%d", containerPort, containerPort)}
		}
	}

	// Add environment file
	if fileExists(config.EnvFile) {
		containerConfig.EnvFile = config.EnvFile
	}

	// Create green container
	if err := dc.CreateContainer(containerConfig); err != nil {
		return fmt.Errorf("failed to start green container: %v", err)
	}

	fmt.Printf("-----> Green container started (%s)\n", greenName)
	return nil
}

func (dc *DockerClient) switchTrafficBlueToGreen(appName string, containerPort int) error {
	blueName := appName + "-blue"
	greenName := appName + "-green"

	fmt.Println("-----> Switching traffic: blue → green")

	// Stop accepting connections on blue
	if dc.ContainerIsRunning(blueName) {
		fmt.Println("       Pausing blue container...")
		cmd := exec.Command("docker", "pause", blueName)
		cmd.Run() // Ignore errors
		time.Sleep(2 * time.Second)
	}

	// Rename containers (atomic swap)
	fmt.Println("       Swapping container names...")

	// Temporary rename old blue to blue-old
	if dc.ContainerExists(blueName) {
		dc.renameContainer(blueName, blueName+"-old")
	}

	// Rename green to blue
	if err := dc.renameContainer(greenName, blueName); err != nil {
		return fmt.Errorf("failed to rename green container to blue: %v", err)
	}

	// Set proper restart policy for new blue container
	dc.updateContainerRestartPolicy(blueName, "always")

	fmt.Println("-----> Traffic switch complete (green → blue)")
	return nil
}

func (dc *DockerClient) cleanupOldBlueContainer(appName string) {
	oldBlueName := appName + "-blue-old"

	fmt.Println("-----> Cleaning up old blue container...")

	if dc.ContainerExists(oldBlueName) {
		// Give it time to drain connections
		fmt.Println("       Waiting 5s before removing old container...")
		time.Sleep(5 * time.Second)

		fmt.Println("       Removing old blue container...")
		dc.StopContainer(oldBlueName)
		dc.RemoveContainer(oldBlueName, true)

		fmt.Println("-----> Old container cleaned up")
	}
}

func (dc *DockerClient) renameContainer(oldName, newName string) error {
	cmd := exec.Command("docker", "rename", oldName, newName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rename container %s to %s: %v, output: %s", oldName, newName, err, string(output))
	}
	return nil
}

func (dc *DockerClient) updateContainerRestartPolicy(containerName, policy string) {
	cmd := exec.Command("docker", "update", "--restart", policy, containerName)
	cmd.Run() // Ignore errors
}

// RecreateActiveContainer recreates the active container with new environment variables
func (dc *DockerClient) RecreateActiveContainer(appName, envFile, appDir string) error {
	// Determine which container is active
	var activeContainer string
	if dc.ContainerExists(appName + "-blue") {
		activeContainer = appName + "-blue"
	} else if dc.ContainerExists(appName + "-green") {
		activeContainer = appName + "-green"
	} else {
		return fmt.Errorf("no active container found for %s", appName)
	}

	fmt.Printf("-----> Recreating container: %s\n", activeContainer)

	config, _ := LoadConfig()
	appConfig := config.GetAppConfig(appName)

	// Get current image
	cmd := exec.Command("docker", "inspect", activeContainer, "--format", "{{.Config.Image}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not determine container image: %v", err)
	}

	image := strings.TrimSpace(string(output))
	fmt.Printf("       Using image: %s\n", image)

	// Get Docker configuration from gokku.yml (simplified for now)
	networkMode := "bridge" // TODO: Get from gokku.yml

	if appConfig.Network != nil && appConfig.Network.Mode != "" {
		networkMode = appConfig.Network.Mode
	}

	containerPort := GetContainerPort(envFile, 8080)

	fmt.Printf("       Network mode: %s\n", networkMode)

	// Stop and remove old container
	fmt.Println("       Stopping old container...")
	dc.StopContainer(activeContainer)
	dc.RemoveContainer(activeContainer, true)

	// Build container configuration
	containerConfig := ContainerConfig{
		Name:          activeContainer,
		Image:         image,
		NetworkMode:   networkMode,
		RestartPolicy: "always",
		WorkingDir:    "/app",
		Volumes:       []string{fmt.Sprintf("%s:/app", appDir)},
	}

	// Add port mappings
	if networkMode != "host" {
		containerConfig.Ports = []string{fmt.Sprintf("%d:%d", containerPort, containerPort)}
	} else {
		fmt.Println("       Using host network (all ports exposed)")
	}

	// Add environment file
	if fileExists(envFile) {
		containerConfig.EnvFile = envFile
	}

	// Start new container with same name and updated env
	fmt.Println("       Starting new container with updated configuration...")
	if err := dc.CreateContainer(containerConfig); err != nil {
		return fmt.Errorf("failed to recreate container: %v", err)
	}

	fmt.Println("✓ Container recreated successfully with new environment")
	return nil
}

// BlueGreenRollback performs rollback to previous blue container
func (dc *DockerClient) BlueGreenRollback(appName string) error {
	blueName := appName + "-blue"
	oldBlueName := appName + "-blue-old"

	fmt.Println("=====> Starting Blue/Green Rollback")

	// Check if old blue exists
	if !dc.ContainerExists(oldBlueName) {
		return fmt.Errorf("no previous blue container found for rollback")
	}

	fmt.Println("-----> Stopping current blue container...")
	dc.StopContainer(blueName)

	fmt.Println("-----> Restoring previous blue container...")
	if err := dc.renameContainer(oldBlueName, blueName); err != nil {
		return fmt.Errorf("failed to restore previous blue container: %v", err)
	}

	fmt.Println("-----> Starting previous blue container...")
	cmd := exec.Command("docker", "start", blueName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start previous blue container: %v, output: %s", err, string(output))
	}

	// Wait for container to be ready
	time.Sleep(5 * time.Second)

	fmt.Println("=====> Blue/Green Rollback Complete!")
	fmt.Printf("-----> Active container: %s\n", blueName)

	return nil
}

// Legacy functions for backward compatibility

func ListContainers(remoteInfo *RemoteInfo, format string) string {
	dc, err := NewDockerClient()
	if err != nil {
		return fmt.Sprintf("Error creating Docker client: %v", err)
	}

	containers, err := dc.ListContainers(true)
	if err != nil {
		return fmt.Sprintf("Error listing containers: %v", err)
	}

	var result strings.Builder
	for _, container := range containers {
		result.WriteString(fmt.Sprintf("%s %s %s %s\n", container.ID, container.Names, container.Status, container.Ports))
	}

	return result.String()
}

func ListImages(remoteInfo *RemoteInfo, format string) string {
	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf("docker images --format %s", format))
	output, err := cmd.Output()

	if err != nil {
		fmt.Printf("Error listing images: %v\n", err)
		return ""
	}

	return string(output)
}

func RemoveContainer(remoteInfo *RemoteInfo, containerName string) string {
	dc, err := NewDockerClient()
	if err != nil {
		return fmt.Sprintf("Error creating Docker client: %v", err)
	}

	if err := dc.RemoveContainer(containerName, true); err != nil {
		return fmt.Sprintf("Error removing container: %v", err)
	}

	return "Container removed successfully"
}

func CreateContainer(remoteInfo *RemoteInfo, options map[string]string) string {
	dc, err := NewDockerClient()
	if err != nil {
		return fmt.Sprintf("Error creating Docker client: %v", err)
	}

	config := ContainerConfig{
		Name:    options["name"],
		Image:   options["image"],
		Command: []string{options["command"]},
	}

	if err := dc.CreateContainer(config); err != nil {
		return fmt.Sprintf("Error creating container: %v", err)
	}

	return "Container created successfully"
}
