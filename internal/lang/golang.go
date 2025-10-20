package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "infra/internal"
)

type Golang struct {
	app *App
}

func (l *Golang) Build(app *App, releaseDir string) error {
	fmt.Println("-----> Building Go application...")

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

	fmt.Println("-----> Go build complete!")
	return nil
}

func (l *Golang) Deploy(app *App, releaseDir string) error {
	fmt.Println("-----> Deploying Go application...")

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
		DockerPorts: app.Ports,
	})
}

func (l *Golang) Restart(app *App) error {
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

func (l *Golang) Cleanup(app *App) error {
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

func (l *Golang) DetectLanguage(releaseDir string) (string, error) {
	if _, err := os.Stat(filepath.Join(releaseDir, "go.mod")); err == nil {
		return "go", nil
	}
	return "", fmt.Errorf("not a Go project")
}

func (l *Golang) EnsureDockerfile(releaseDir string, app *App) error {
	fmt.Printf("-----> EnsureDockerfile called for app: %s\n", app.Name)

	// Check if custom Dockerfile is specified
	if app.Build != nil && app.Build.Dockerfile != "" {
		customDockerfilePath := filepath.Join(releaseDir, app.Build.Dockerfile)
		fmt.Printf("-----> Custom Dockerfile path: %s\n", customDockerfilePath)
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
	fmt.Printf("-----> Default Dockerfile path: %s\n", dockerfilePath)
	if _, err := os.Stat(dockerfilePath); err == nil {
		fmt.Println("-----> Using existing Dockerfile")
		return nil
	}

	fmt.Println("-----> Generating Dockerfile for Go...")

	// Get build configuration
	build := l.GetDefaultConfig()
	if app.Build != nil {
		// Merge with app-specific config
		if app.Build.BaseImage != "" {
			build.BaseImage = app.Build.BaseImage
		}
		// Determine working directory
		workDir := "."
		if app.Build.Workdir != "" {
			workDir = app.Build.Workdir
		}
		fmt.Printf("-----> Working directory from config: '%s'\n", app.Build.Workdir)
		fmt.Printf("-----> Using workDir: '%s'\n", workDir)

		// Build path is relative to work_dir
		if app.Build.Path != "" {
			// Since we COPY workdir ., the build path should be relative to workdir
			build.Path = "./" + strings.TrimPrefix(app.Build.Path, "./")
			fmt.Printf("-----> Configured path: '%s'\n", app.Build.Path)
		} else {
			// Default to working directory root
			build.Path = "."
		}
		fmt.Printf("-----> Final build path: '%s'\n", build.Path)
	}

	// Generate Dockerfile content
	dockerfileContent := l.generateDockerfile(build, app)

	// Write Dockerfile
	return os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)
}

func (l *Golang) GetDefaultConfig() *Build {
	return &Build{
		Type:      "docker",
		BaseImage: "golang:1.25-alpine",
		Path:      "",
		Workdir:   ".",
	}
}

func (l *Golang) generateDockerfile(build *Build, app *App) string {
	// Determine build path
	buildPath := build.Path

	if buildPath == "" {
		buildPath = "."
	}

	fmt.Printf("-----> Dockerfile build path: %s\n", buildPath)

	// Determine base image
	baseImage := build.BaseImage

	if baseImage == "" {
		baseImage = "golang:1.25-alpine"
	}

	// Get the workdir for COPY
	workDir := "."

	if app.Build != nil && app.Build.Workdir != "" {
		workDir = app.Build.Workdir
	}

	fmt.Printf("-----> Using workdir: %s\n", workDir)

	return fmt.Sprintf(`# Generated Dockerfile for Go application
# App: %s
# Build path: %s

FROM %s AS builder

WORKDIR /app

# Copy workdir
COPY %s .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -ldflags="-w -s" -o app %s

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/app .

# Expose port
EXPOSE ${PORT:-8080}

# Run the application
CMD ["/root/app"]
`, app.Name, buildPath, baseImage, workDir, buildPath)
}
