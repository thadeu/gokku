package lang

import (
	"fmt"
	"os"
	"path/filepath"

	. "infra/internal"
)

type Lang interface {
	Build(app *App, releaseDir string) error
	Deploy(app *App, releaseDir string) error
	Restart(app *App) error
	Cleanup(app *App) error
	DetectLanguage(releaseDir string) (string, error)
	EnsureDockerfile(releaseDir string, app *App) error
	GetDefaultConfig() *Build
}

// DetectLanguage automatically detects the programming language based on project files
func DetectLanguage(releaseDir string) (string, error) {
	// Check for Go
	if _, err := os.Stat(filepath.Join(releaseDir, "go.mod")); err == nil {
		return "go", nil
	}

	// Check for Node.js
	if _, err := os.Stat(filepath.Join(releaseDir, "package.json")); err == nil {
		return "nodejs", nil
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(releaseDir, "requirements.txt")); err == nil {
		return "python", nil
	}
	if _, err := os.Stat(filepath.Join(releaseDir, "pyproject.toml")); err == nil {
		return "python", nil
	}

	// Check for Ruby
	if _, err := os.Stat(filepath.Join(releaseDir, "Gemfile")); err == nil {
		return "ruby", nil
	}

	// Check for existing Dockerfile
	if _, err := os.Stat(filepath.Join(releaseDir, "Dockerfile")); err == nil {
		return "docker", nil
	}

	// Default to generic
	return "generic", nil
}

// NewLang creates a language handler based on detected or configured language
func NewLang(app *App, releaseDir string) (Lang, error) {
	var langType string
	var err error

	// Use configured language or auto-detect
	if app.Lang != "" {
		langType = app.Lang
	} else {
		langType, err = DetectLanguage(releaseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to detect language: %v", err)
		}
	}

	// Update app.Lang with detected language
	app.Lang = langType

	switch langType {
	case "go":
		return &Golang{app: app}, nil
	case "python":
		return &Python{app: app}, nil
	case "nodejs":
		return &Nodejs{app: app}, nil
	case "ruby":
		return &Ruby{app: app}, nil
	case "docker":
		return &Generic{app: app}, nil
	default:
		return &Generic{app: app}, nil
	}
}
