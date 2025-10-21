package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
func (sm *ServiceManager) CreateService(pluginName, serviceName string) error {
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

	if err := sm.saveServiceConfig(serviceName, service); err != nil {
		// Cleanup on error
		os.RemoveAll(serviceDir)
		return fmt.Errorf("failed to save service config: %v", err)
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

	return nil
}

// UnlinkService unlinks a service from an app
func (sm *ServiceManager) UnlinkService(serviceName, appName, env string) error {
	// Check if service exists
	service, err := sm.getServiceConfig(serviceName)
	if err != nil {
		return fmt.Errorf("service '%s' not found: %v", serviceName, err)
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
