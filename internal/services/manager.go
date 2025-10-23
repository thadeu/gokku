package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"infra/internal"
)

// ServiceManager manages services and their lifecycle
type ServiceManager struct {
	servicesDir string
	pluginsDir  string
}

// Service represents a service instance
type Service struct {
	Name        string            `json:"name"`
	Plugin      string            `json:"plugin"`
	ContainerID string            `json:"container_id"`
	Running     bool              `json:"running"`
	LinkedApps  []string          `json:"linked_apps"`
	CreatedAt   string            `json:"created_at"`
	Config      map[string]string `json:"config"`
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		servicesDir: "/opt/gokku/services",
		pluginsDir:  "/opt/gokku/plugins",
	}
}

// CreateService creates a new service from a plugin
func (sm *ServiceManager) CreateService(pluginName, serviceName, version string) error {
	// Check if plugin exists
	if !sm.pluginExists(pluginName) {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// Check if service already exists
	if sm.serviceExists(serviceName) {
		return fmt.Errorf("service '%s' already exists", serviceName)
	}

	// Create service directory
	serviceDir := filepath.Join(sm.servicesDir, serviceName)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %v", err)
	}

	// Save service configuration
	service := Service{
		Name:       serviceName,
		Plugin:     pluginName,
		Running:    false,
		LinkedApps: []string{},
		CreatedAt:  time.Now().Format(time.RFC3339),
		Config:     make(map[string]string),
	}

	// Store version if provided
	if version != "" {
		service.Config["version"] = version
	}

	if err := sm.saveServiceConfig(serviceName, service); err != nil {
		// Cleanup on error
		os.RemoveAll(serviceDir)
		return fmt.Errorf("failed to save service config: %v", err)
	}

	// Execute plugin install script
	installScript := filepath.Join(sm.pluginsDir, pluginName, "bin", "install")

	if _, err := os.Stat(installScript); err == nil {
		cmd := exec.Command("bash", installScript, serviceName, version)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Cleanup on error
			os.RemoveAll(serviceDir)
			return fmt.Errorf("failed to execute install script: %v", err)
		}
	}

	return nil
}

// LinkService links a service to an app
func (sm *ServiceManager) LinkService(serviceName, appName, env string) error {
	// Check if service exists
	service, err := sm.getServiceConfig(serviceName)
	if err != nil {
		return fmt.Errorf("service '%s' not found: %v", serviceName, err)
	}

	// Check if app exists
	if !sm.appExists(appName, env) {
		return fmt.Errorf("app '%s' not found", appName)
	}

	// Update service config
	if !sm.isAppLinked(service.LinkedApps, appName, env) {
		service.LinkedApps = append(service.LinkedApps, appName)
		if err := sm.saveServiceConfig(serviceName, service); err != nil {
			return fmt.Errorf("failed to update service config: %v", err)
		}
	}

	// Add environment variables to app
	if err := sm.addServiceEnvVars(serviceName, appName, service.Plugin); err != nil {
		return fmt.Errorf("failed to add environment variables: %v", err)
	}

	return nil
}

// UnlinkService unlinks a service from an app
func (sm *ServiceManager) UnlinkService(serviceName, appName, env string) error {
	// Check if service exists
	service, err := sm.getServiceConfig(serviceName)
	if err != nil {
		return fmt.Errorf("service '%s' not found: %v", serviceName, err)
	}

	// Remove environment variables from app
	if err := sm.removeServiceEnvVars(serviceName, appName, service.Plugin); err != nil {
		return fmt.Errorf("failed to remove environment variables: %v", err)
	}

	// Remove app from linked apps
	var newLinkedApps []string
	for _, linked := range service.LinkedApps {
		if linked != appName {
			newLinkedApps = append(newLinkedApps, linked)
		}
	}

	service.LinkedApps = newLinkedApps
	if err := sm.saveServiceConfig(serviceName, service); err != nil {
		return fmt.Errorf("failed to update service config: %v", err)
	}

	return nil
}

// ListServices returns a list of all services
func (sm *ServiceManager) ListServices() ([]Service, error) {
	var services []Service

	entries, err := os.ReadDir(sm.servicesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			service, err := sm.getServiceConfig(entry.Name())
			if err != nil {
				continue // Skip invalid services
			}
			services = append(services, service)
		}
	}

	return services, nil
}

// GetService returns a specific service
func (sm *ServiceManager) GetService(serviceName string) (Service, error) {
	return sm.getServiceConfig(serviceName)
}

// DestroyService destroys a service
func (sm *ServiceManager) DestroyService(serviceName string) error {
	// Check if service exists
	if !sm.serviceExists(serviceName) {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Get service config to know which plugin it belongs to
	service, err := sm.getServiceConfig(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service config: %v", err)
	}

	// Execute plugin uninstall script
	uninstallScript := filepath.Join(sm.pluginsDir, service.Plugin, "uninstall")
	if _, err := os.Stat(uninstallScript); err == nil {
		cmd := exec.Command("bash", uninstallScript, serviceName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute uninstall script: %v", err)
		}
	}

	// Remove service directory
	serviceDir := filepath.Join(sm.servicesDir, serviceName)
	if err := os.RemoveAll(serviceDir); err != nil {
		return fmt.Errorf("failed to remove service directory: %v", err)
	}

	return nil
}

// UpdateServiceConfig updates service configuration
func (sm *ServiceManager) UpdateServiceConfig(serviceName string, config map[string]string) error {
	service, err := sm.getServiceConfig(serviceName)
	if err != nil {
		return err
	}

	service.Config = config
	return sm.saveServiceConfig(serviceName, service)
}

// Helper methods
func (sm *ServiceManager) pluginExists(pluginName string) bool {
	pluginPath := filepath.Join(sm.pluginsDir, pluginName)
	_, err := os.Stat(pluginPath)
	return !os.IsNotExist(err)
}

func (sm *ServiceManager) serviceExists(serviceName string) bool {
	servicePath := filepath.Join(sm.servicesDir, serviceName)
	_, err := os.Stat(servicePath)
	return !os.IsNotExist(err)
}

func (sm *ServiceManager) appExists(appName, env string) bool {
	appPath := filepath.Join("/opt/gokku/apps", appName, env)
	_, err := os.Stat(appPath)
	return !os.IsNotExist(err)
}

func (sm *ServiceManager) saveServiceConfig(serviceName string, service Service) error {
	configPath := filepath.Join(sm.servicesDir, serviceName, "config.json")
	data, err := json.Marshal(service)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func (sm *ServiceManager) getServiceConfig(serviceName string) (Service, error) {
	configPath := filepath.Join(sm.servicesDir, serviceName, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Service{}, err
	}

	var service Service
	if err := json.Unmarshal(data, &service); err != nil {
		return Service{}, err
	}

	return service, nil
}

func (sm *ServiceManager) isAppLinked(linkedApps []string, appName, env string) bool {
	for _, linked := range linkedApps {
		if linked == appName {
			return true
		}
	}
	return false
}

// addServiceEnvVars adds service environment variables to an app
func (sm *ServiceManager) addServiceEnvVars(serviceName, appName, pluginName string) error {
	// Get service configuration
	service, err := sm.getServiceConfig(serviceName)
	if err != nil {
		return err
	}

	// Get environment variables based on plugin type
	envVars := sm.getServiceEnvVars(serviceName, service.Config, pluginName)
	if len(envVars) == 0 {
		return nil
	}

	// Get app env file path
	envFile := filepath.Join("/opt/gokku/apps", appName, "shared", ".env")

	// Load existing env vars
	existingVars := internal.LoadEnvFile(envFile)

	// Add service env vars
	for key, value := range envVars {
		existingVars[key] = value
	}

	// Save updated env vars
	return internal.SaveEnvFile(envFile, existingVars)
}

// removeServiceEnvVars removes service environment variables from an app
func (sm *ServiceManager) removeServiceEnvVars(serviceName, appName, pluginName string) error {
	// Get environment variable keys based on plugin type
	envKeys := sm.getServiceEnvKeys(pluginName)
	if len(envKeys) == 0 {
		return nil
	}

	// Get app env file path
	envFile := filepath.Join("/opt/gokku/apps", appName, "shared", ".env")

	// Load existing env vars
	existingVars := internal.LoadEnvFile(envFile)

	// Remove service env vars
	for _, key := range envKeys {
		delete(existingVars, key)
	}

	// Save updated env vars
	return internal.SaveEnvFile(envFile, existingVars)
}

// getServiceEnvVars returns environment variables for a service based on plugin type
func (sm *ServiceManager) getServiceEnvVars(serviceName string, config map[string]string, pluginName string) map[string]string {
	envVars := make(map[string]string)

	switch pluginName {
	case "postgres":
		host := "localhost"
		port := config["port"]
		user := config["user"]
		password := config["password"]
		database := config["database"]

		if port == "" || user == "" || password == "" || database == "" {
			return envVars
		}

		envVars["DATABASE_URL"] = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, database)
		envVars["POSTGRES_HOST"] = host
		envVars["POSTGRES_PORT"] = port
		envVars["POSTGRES_USER"] = user
		envVars["POSTGRES_PASSWORD"] = password
		envVars["POSTGRES_DB"] = database

	case "redis":
		host := "localhost"
		port := config["port"]
		password := config["password"]

		if port == "" || password == "" {
			return envVars
		}

		envVars["REDIS_URL"] = fmt.Sprintf("redis://:%s@%s:%s", password, host, port)
		envVars["REDIS_HOST"] = host
		envVars["REDIS_PORT"] = port
		envVars["REDIS_PASSWORD"] = password
	}

	return envVars
}

// getServiceEnvKeys returns environment variable keys for a plugin type
func (sm *ServiceManager) getServiceEnvKeys(pluginName string) []string {
	switch pluginName {
	case "postgres":
		return []string{
			"DATABASE_URL",
			"POSTGRES_HOST",
			"POSTGRES_PORT",
			"POSTGRES_USER",
			"POSTGRES_PASSWORD",
			"POSTGRES_DB",
		}
	case "redis":
		return []string{
			"REDIS_URL",
			"REDIS_HOST",
			"REDIS_PORT",
			"REDIS_PASSWORD",
		}
	default:
		return []string{}
	}
}

// getServiceContainerInfo gets container information from Docker
func (sm *ServiceManager) getServiceContainerInfo(serviceName string) (map[string]string, error) {
	info := make(map[string]string)

	// Check if container exists
	checkCmd := exec.Command("docker", "ps", "-aq", "-f", fmt.Sprintf("name=^%s$", serviceName))
	output, err := checkCmd.Output()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return info, fmt.Errorf("container not found")
	}

	// Check if container is running
	runningCmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=^%s$", serviceName))
	runningOutput, err := runningCmd.Output()
	if err != nil || len(strings.TrimSpace(string(runningOutput))) == 0 {
		info["running"] = "false"
		return info, nil
	}

	info["running"] = "true"
	return info, nil
}
