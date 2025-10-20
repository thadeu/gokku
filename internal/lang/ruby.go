package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "infra/internal"
)

type Ruby struct {
	app *App
}

func (l *Ruby) Build(app *App, releaseDir string) error {
	fmt.Println("-----> Building Ruby application...")

	// Ensure Dockerfile exists
	if err := l.EnsureDockerfile(releaseDir, app); err != nil {
		return fmt.Errorf("failed to ensure Dockerfile: %v", err)
	}

	// Build Docker image
	imageTag := fmt.Sprintf("%s:latest", app.Name)

	// Check if custom Dockerfile path is specified
	var cmd *exec.Cmd
	if app.Build != nil && app.Build.Dockerfile != "" {
		// Use custom Dockerfile path
		dockerfilePath := filepath.Join(releaseDir, app.Build.Dockerfile)
		// Check if Dockerfile exists in workdir
		if app.Build.Workdir != "" {
			workdirDockerfilePath := filepath.Join(releaseDir, app.Build.Workdir, app.Build.Dockerfile)
			if _, err := os.Stat(workdirDockerfilePath); err == nil {
				dockerfilePath = workdirDockerfilePath
			}
		}
		fmt.Printf("-----> Using custom Dockerfile: %s\n", dockerfilePath)
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

	fmt.Println("-----> Ruby build complete!")
	return nil
}

func (l *Ruby) Deploy(app *App, releaseDir string) error {
	fmt.Println("-----> Deploying Ruby application...")

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

	// Create deployment config
	return dc.DeployContainer(DeploymentConfig{
		AppName:     app.Name,
		ImageTag:    "latest",
		EnvFile:     envFile,
		ReleaseDir:  releaseDir,
		NetworkMode: networkMode,
	})
}

func (l *Ruby) Restart(app *App) error {
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

func (l *Ruby) Cleanup(app *App) error {
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

func (l *Ruby) DetectLanguage(releaseDir string) (string, error) {
	if _, err := os.Stat(filepath.Join(releaseDir, "Gemfile")); err == nil {
		return "ruby", nil
	}
	return "", fmt.Errorf("not a Ruby project")
}

func (l *Ruby) EnsureDockerfile(releaseDir string, app *App) error {
	// Check if custom Dockerfile is specified
	if app.Build != nil && app.Build.Dockerfile != "" {
		customDockerfilePath := filepath.Join(releaseDir, app.Build.Dockerfile)
		if _, err := os.Stat(customDockerfilePath); err == nil {
			fmt.Printf("-----> Using custom Dockerfile: %s\n", app.Build.Dockerfile)
			return nil
		}
		// If custom Dockerfile not found, try relative to workdir
		if app.Build.Workdir != "" {
			workdirDockerfilePath := filepath.Join(releaseDir, app.Build.Workdir, app.Build.Dockerfile)
			if _, err := os.Stat(workdirDockerfilePath); err == nil {
				fmt.Printf("-----> Using custom Dockerfile in workdir: %s/%s\n", app.Build.Workdir, app.Build.Dockerfile)
				return nil
			}
		}
		return fmt.Errorf("custom Dockerfile not found: %s or %s", customDockerfilePath, filepath.Join(releaseDir, app.Build.Workdir, app.Build.Dockerfile))
	}

	// Check if default Dockerfile exists
	dockerfilePath := filepath.Join(releaseDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		fmt.Println("-----> Using existing Dockerfile")
		return nil
	}

	fmt.Println("-----> Generating Dockerfile for Ruby...")

	// Get build configuration
	build := l.GetDefaultConfig()
	if app.Build != nil {
		// Merge with app-specific config
		if app.Build.BaseImage != "" {
			build.BaseImage = app.Build.BaseImage
		}
		if app.Build.Entrypoint != "" {
			build.Entrypoint = app.Build.Entrypoint
		}
	}

	// Generate Dockerfile content
	dockerfileContent := l.generateDockerfile(build, app)

	// Write Dockerfile
	return os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)
}

func (l *Ruby) GetDefaultConfig() *Build {
	return &Build{
		Type:       "docker",
		BaseImage:  "ruby:3.2-alpine",
		Entrypoint: "app.rb",
		Workdir:    ".",
	}
}

func (l *Ruby) generateDockerfile(build *Build, app *App) string {
	// Determine entrypoint
	entrypoint := build.Entrypoint
	if entrypoint == "" {
		entrypoint = "app.rb"
	}

	// Determine base image
	baseImage := build.BaseImage
	if baseImage == "" {
		baseImage = "ruby:3.2-alpine"
	}

	return fmt.Sprintf(`# Generated Dockerfile for Ruby application
# App: %s
# Entrypoint: %s

FROM %s

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache build-base

# Copy Gemfile
COPY Gemfile* ./
RUN bundle install --without development test

# Copy application code
COPY . .

# Expose port (will be set via env var)
EXPOSE ${PORT:-8080}

# Run the application
CMD ["ruby", "%s"]
`, app.Name, entrypoint, baseImage, entrypoint)
}
