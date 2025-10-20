package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "infra/internal"
)

type Generic struct {
	app *App
}

func (l *Generic) Build(app *App, releaseDir string) error {
	fmt.Println("-----> Building generic application...")

	// Check if custom Dockerfile is specified
	var dockerfilePath string
	if app.Build != nil && app.Build.Dockerfile != "" {
		dockerfilePath = filepath.Join(releaseDir, app.Build.Dockerfile)
		// Check if Dockerfile exists in workdir
		if app.Build.Workdir != "" {
			workdirDockerfilePath := filepath.Join(releaseDir, app.Build.Workdir, app.Build.Dockerfile)
			if _, err := os.Stat(workdirDockerfilePath); err == nil {
				dockerfilePath = workdirDockerfilePath
				fmt.Printf("-----> Using custom Dockerfile in workdir: %s/%s\n", app.Build.Workdir, app.Build.Dockerfile)
			} else {
				fmt.Printf("-----> Using custom Dockerfile: %s\n", app.Build.Dockerfile)
			}
		} else {
			fmt.Printf("-----> Using custom Dockerfile: %s\n", app.Build.Dockerfile)
		}
	} else {
		dockerfilePath = filepath.Join(releaseDir, "Dockerfile")
		fmt.Println("-----> Using default Dockerfile")
	}

	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("no Dockerfile found and no language-specific strategy available")
	}

	// Build Docker image
	imageTag := fmt.Sprintf("%s:latest", app.Name)

	// Build Docker image with the determined Dockerfile path
	var cmd *exec.Cmd
	if app.Build != nil && app.Build.Dockerfile != "" {
		// Use custom Dockerfile path
		cmd = exec.Command("docker", "build", "-f", dockerfilePath, "-t", imageTag, releaseDir)
	} else {
		// Use default Dockerfile in release directory
		cmd = exec.Command("docker", "build", "-t", imageTag, releaseDir)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	fmt.Println("-----> Generic build complete!")
	return nil
}

func (l *Generic) Deploy(app *App, releaseDir string) error {
	fmt.Println("-----> Deploying generic application...")

	// Get Docker client
	dc, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	// Get environment file
	envFile := filepath.Join("/opt/gokku/apps", app.Name, "shared", ".env")

	networkMode := "bridge"

	if app.Network != nil && app.Network.Mode != "" {
		networkMode = app.Network.Mode
	}

	// Deploy using Docker client
	return dc.DeployContainer(DeploymentConfig{
		AppName:     app.Name,
		ImageTag:    "latest",
		EnvFile:     envFile,
		ReleaseDir:  releaseDir,
		NetworkMode: networkMode,
		DockerPorts: app.Ports,
	})
}

func (l *Generic) Restart(app *App) error {
	fmt.Printf("-----> Restarting %s...\n", app.Name)

	dc, err := NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	// Find active container
	containerName := app.Name
	if !dc.ContainerExists(containerName) {
		containerName = app.Name + "-green"
	}

	if !dc.ContainerExists(containerName) {
		return fmt.Errorf("no active container found for %s", app.Name)
	}

	// Restart container
	cmd := exec.Command("docker", "restart", containerName)
	return cmd.Run()
}

func (l *Generic) Cleanup(app *App) error {
	fmt.Printf("-----> Cleaning up old releases for %s...\n", app.Name)

	appDir := filepath.Join("/opt/gokku/apps", app.Name)
	releasesDir := filepath.Join(appDir, "releases")

	// Read all release directories
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return err
	}

	// Keep only the last 5 releases
	keepReleases := 5
	if len(entries) <= keepReleases {
		return nil
	}

	// Remove old releases
	toRemove := len(entries) - keepReleases
	for i := 0; i < toRemove; i++ {
		entry := entries[i]
		releasePath := filepath.Join(releasesDir, entry.Name())
		if err := os.RemoveAll(releasePath); err != nil {
			fmt.Printf("Warning: Failed to remove old release %s: %v\n", entry.Name(), err)
		} else {
			fmt.Printf("-----> Removed old release: %s\n", entry.Name())
		}
	}

	return nil
}

func (l *Generic) DetectLanguage(releaseDir string) (string, error) {
	// Check for existing Dockerfile
	if _, err := os.Stat(filepath.Join(releaseDir, "Dockerfile")); err == nil {
		return "docker", nil
	}
	return "generic", nil
}

func (l *Generic) EnsureDockerfile(releaseDir string, app *App) error {
	dockerfilePath := filepath.Join(releaseDir, "Dockerfile")

	// Check if Dockerfile already exists
	if _, err := os.Stat(dockerfilePath); err == nil {
		fmt.Println("-----> Using existing Dockerfile")
		return nil
	}

	return fmt.Errorf("no Dockerfile found and no language-specific strategy available")
}

func (l *Generic) GetDefaultConfig() *Build {
	return &Build{
		Type: "docker",
	}
}
