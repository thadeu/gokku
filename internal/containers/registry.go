package containers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ContainerInfo represents information about a running container
type ContainerInfo struct {
	Name         string `json:"name"`
	AppName      string `json:"app_name"`
	ProcessType  string `json:"process_type"`
	Number       int    `json:"number"`
	HostPort     int    `json:"host_port"`
	InternalPort int    `json:"internal_port"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// ContainerRegistry manages container information
type ContainerRegistry struct {
	basePath string
}

// NewContainerRegistry creates a new container registry
func NewContainerRegistry() *ContainerRegistry {
	return &ContainerRegistry{
		basePath: "/opt/gokku/apps",
	}
}

// SaveContainerInfo saves container information to disk
func (cr *ContainerRegistry) SaveContainerInfo(info ContainerInfo) error {
	appPath := filepath.Join(cr.basePath, info.AppName)
	containersPath := filepath.Join(appPath, "containers", info.ProcessType)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(containersPath, 0755); err != nil {
		return fmt.Errorf("failed to create containers directory: %w", err)
	}

	// Create filename based on container number
	filename := fmt.Sprintf("%d.json", info.Number)
	filePath := filepath.Join(containersPath, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal container info: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write container info: %w", err)
	}

	return nil
}

// GetContainers returns all containers for a specific app and process type
func (cr *ContainerRegistry) GetContainers(appName, processType string) ([]ContainerInfo, error) {
	containersPath := filepath.Join(cr.basePath, appName, "containers", processType)

	// Check if directory exists
	if _, err := os.Stat(containersPath); os.IsNotExist(err) {
		return []ContainerInfo{}, nil
	}

	// Read directory
	files, err := os.ReadDir(containersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read containers directory: %w", err)
	}

	var containers []ContainerInfo

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(containersPath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip corrupted files
		}

		var info ContainerInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue // Skip corrupted files
		}

		containers = append(containers, info)
	}

	// Sort by container number
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Number < containers[j].Number
	})

	return containers, nil
}

// RemoveContainerInfo removes container information from disk
func (cr *ContainerRegistry) RemoveContainerInfo(appName, processType string, containerNumber int) error {
	filePath := filepath.Join(cr.basePath, appName, "containers", processType, fmt.Sprintf("%d.json", containerNumber))

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove container info: %w", err)
	}

	return nil
}

// GetNextContainerNumber returns the next available container number for a process type
func (cr *ContainerRegistry) GetNextContainerNumber(appName, processType string) int {
	containers, err := cr.GetContainers(appName, processType)
	if err != nil || len(containers) == 0 {
		return 1
	}

	// Find the highest number and add 1
	maxNumber := 0
	for _, container := range containers {
		if container.Number > maxNumber {
			maxNumber = container.Number
		}
	}

	return maxNumber + 1
}

// GetContainerByNumber returns a specific container by number
func (cr *ContainerRegistry) GetContainerByNumber(appName, processType string, number int) (*ContainerInfo, error) {
	filePath := filepath.Join(cr.basePath, appName, "containers", processType, fmt.Sprintf("%d.json", number))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("container not found: %w", err)
	}

	var info ContainerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal container info: %w", err)
	}

	return &info, nil
}

// UpdateContainerStatus updates the status of a container
func (cr *ContainerRegistry) UpdateContainerStatus(appName, processType string, number int, status string) error {
	info, err := cr.GetContainerByNumber(appName, processType, number)
	if err != nil {
		return err
	}

	info.Status = status
	return cr.SaveContainerInfo(*info)
}

// GetAllContainers returns all containers for an app across all process types
func (cr *ContainerRegistry) GetAllContainers(appName string) ([]ContainerInfo, error) {
	appPath := filepath.Join(cr.basePath, appName, "containers")

	// Check if directory exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return []ContainerInfo{}, nil
	}

	// Read process type directories
	processTypes, err := os.ReadDir(appPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read containers directory: %w", err)
	}

	var allContainers []ContainerInfo

	for _, processType := range processTypes {
		if !processType.IsDir() {
			continue
		}

		containers, err := cr.GetContainers(appName, processType.Name())
		if err != nil {
			continue // Skip process types with errors
		}

		allContainers = append(allContainers, containers...)
	}

	// Sort by process type, then by number
	sort.Slice(allContainers, func(i, j int) bool {
		if allContainers[i].ProcessType != allContainers[j].ProcessType {
			return allContainers[i].ProcessType < allContainers[j].ProcessType
		}
		return allContainers[i].Number < allContainers[j].Number
	})

	return allContainers, nil
}

// GetNextAvailablePort finds the next available port starting from 32768
func GetNextAvailablePort() (int, error) {
	// Start from port 32768 (Docker's default range)
	startPort := 32768
	endPort := 65535

	for port := startPort; port <= endPort; port++ {
		if isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports found")
}

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	portPattern := fmt.Sprintf(":%d ", port)

	// Try netstat with grep first
	netstatCmd := exec.Command("sh", "-c", fmt.Sprintf("netstat -ln 2>/dev/null | grep -q '%s'", portPattern))

	err := netstatCmd.Run()

	if err == nil {
		// grep found the port, so it's in use
		return false
	}

	// Try ss with grep as fallback
	ssCmd := exec.Command("sh", "-c", fmt.Sprintf("ss -ln 2>/dev/null | grep -q '%s'", portPattern))
	err = ssCmd.Run()

	if err == nil {
		// grep found the port, so it's in use
		return false
	}

	// If both commands fail or port is not found, consider it available
	return true
}

// CreateContainerInfo creates a new ContainerInfo struct
func CreateContainerInfo(appName, processType string, number int, hostPort, internalPort int) ContainerInfo {
	return ContainerInfo{
		Name:         fmt.Sprintf("%s-%s-%d", appName, processType, number),
		AppName:      appName,
		ProcessType:  processType,
		Number:       number,
		HostPort:     hostPort,
		InternalPort: internalPort,
		Status:       "running",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
}

// ParseScaleArgument parses scale arguments like "web=4" or "worker=2"
func ParseScaleArgument(arg string) (processType string, count int, err error) {
	parts := strings.Split(arg, "=")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid scale format: %s (expected: process=count)", arg)
	}

	processType = parts[0]
	count, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid count: %s", parts[1])
	}

	if count < 0 {
		return "", 0, fmt.Errorf("count must be non-negative: %d", count)
	}

	return processType, count, nil
}
