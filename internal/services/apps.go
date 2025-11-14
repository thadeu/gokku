package services

import (
	"fmt"
	"os"
	"path/filepath"

	"gokku/internal"
)

// AppsService provides operations for managing apps
type AppsService struct {
	baseDir string
}

// NewAppsService creates a new AppsService
func NewAppsService(baseDir string) *AppsService {
	if baseDir == "" {
		baseDir = "/opt/gokku"
	}
	return &AppsService{baseDir: baseDir}
}

// ListApps returns all apps with their status
func (s *AppsService) ListApps() ([]AppInfo, error) {
	appsDir := filepath.Join(s.baseDir, "apps")

	// Check if apps directory exists
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		return []AppInfo{}, nil
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read apps directory: %w", err)
	}

	var apps []AppInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appName := entry.Name()
		status := s.getAppStatus(appName)
		releasesCount := s.countReleases(appName)
		currentRelease := s.getCurrentRelease(appName)

		apps = append(apps, AppInfo{
			Name:           appName,
			Status:         status,
			ReleasesCount:  releasesCount,
			CurrentRelease: currentRelease,
		})
	}

	return apps, nil
}

// GetApp returns detailed info for a specific app
func (s *AppsService) GetApp(name string) (*AppDetail, error) {
	// Check if app exists
	apps, err := s.ListApps()
	if err != nil {
		return nil, err
	}

	var appInfo *AppInfo

	for i := range apps {
		if apps[i].Name == name {
			appInfo = &apps[i]
			break
		}
	}

	if appInfo == nil {
		return nil, &AppNotFoundError{AppName: name}
	}

	// Load config
	config, err := internal.LoadAppConfig(name)
	if err != nil {
		// Config might not exist, continue without it
		config = nil
	}

	// Get containers
	containers, err := s.getAppContainers(name)
	if err != nil {
		containers = []internal.ContainerInfo{}
	}

	// Get env vars
	envVars := s.getAppEnvVars(name)
	if envVars == nil {
		envVars = make(map[string]string)
	}

	return &AppDetail{
		AppInfo:    *appInfo,
		Config:     config,
		Containers: containers,
		EnvVars:    envVars,
	}, nil
}

// AppExists checks if an app exists
func (s *AppsService) AppExists(name string) bool {
	appDir := filepath.Join(s.baseDir, "apps", name)
	_, err := os.Stat(appDir)
	return err == nil
}

// getAppStatus determines the status of an app based on Docker containers
func (s *AppsService) getAppStatus(appName string) string {
	if internal.ContainerIsRunning(appName) {
		return "running"
	} else if internal.ContainerExists(appName) {
		return "stopped"
	}
	return "not deployed"
}

// countReleases counts the number of releases for an app
func (s *AppsService) countReleases(appName string) int {
	releasesDir := filepath.Join(s.baseDir, "apps", appName, "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return 0
	}
	return len(entries)
}

// getCurrentRelease gets the current release symlink target
func (s *AppsService) getCurrentRelease(appName string) string {
	currentLink := filepath.Join(s.baseDir, "apps", appName, "current")
	linkTarget, err := os.Readlink(currentLink)
	if err != nil {
		return "none"
	}
	return filepath.Base(linkTarget)
}

// getAppContainers gets containers for an app
func (s *AppsService) getAppContainers(appName string) ([]internal.ContainerInfo, error) {
	containers, err := internal.ListContainers(false)
	if err != nil {
		return nil, err
	}

	var appContainers []internal.ContainerInfo
	for _, c := range containers {
		if containsContainerName(c.Names, appName) {
			appContainers = append(appContainers, c)
		}
	}

	return appContainers, nil
}

// getAppEnvVars loads environment variables from .env file
func (s *AppsService) getAppEnvVars(appName string) map[string]string {
	envFile := filepath.Join(s.baseDir, "apps", appName, "shared", ".env")
	return internal.LoadEnvFile(envFile)
}
