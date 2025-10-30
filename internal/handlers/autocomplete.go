package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// handleAutocomplete installs shell completion for gokku
func handleAutocomplete(args []string) {
	if len(args) < 1 {
		showAutocompleteHelp()
		os.Exit(1)
	}

	shell := args[0]

	switch shell {
	case "bash":
		installBashCompletion()
	case "zsh":
		installZshCompletion()
	case "fish":
		installFishCompletion()
	default:
		fmt.Printf("Unknown shell: %s\n", shell)
		fmt.Println("Supported shells: bash, zsh, fish")
		os.Exit(1)
	}
}

func showAutocompleteHelp() {
	fmt.Println("Install shell completion for gokku")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  gokku autocomplete <shell>")
	fmt.Println("")
	fmt.Println("Supported shells:")
	fmt.Println("  bash    Install bash completion")
	fmt.Println("  zsh     Install zsh completion")
	fmt.Println("  fish    Install fish completion")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gokku autocomplete bash")
	fmt.Println("  gokku autocomplete zsh")
	fmt.Println("  gokku autocomplete fish")
}

func installBashCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	completionDir := filepath.Join(homeDir, ".bash_completion.d")
	completionDest := filepath.Join(completionDir, "completion.sh")

	var shellRc string
	if runtime.GOOS == "darwin" {
		shellRc = filepath.Join(homeDir, ".bash_profile")
	} else {
		shellRc = filepath.Join(homeDir, ".bashrc")
	}

	// Create directory
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		fmt.Printf("Error creating completion directory: %v\n", err)
		os.Exit(1)
	}

	// Download completion script
	if err := downloadCompletionScript("scripts/completion.sh", completionDest); err != nil {
		fmt.Printf("Error downloading completion script: %v\n", err)
		os.Exit(1)
	}

	// Make executable
	if err := os.Chmod(completionDest, 0755); err != nil {
		fmt.Printf("Error making completion script executable: %v\n", err)
		os.Exit(1)
	}

	// Add to shell rc
	if err := addToBashRc(shellRc, completionDest); err != nil {
		fmt.Printf("Warning: Failed to add completion to %s: %v\n", shellRc, err)
		fmt.Printf("Please manually add this line to %s:\n", shellRc)
		fmt.Printf("  [ -f %s ] && source %s 2>/dev/null || true\n", completionDest, completionDest)
	} else {
		fmt.Printf("Bash completion installed successfully!\n")
		fmt.Printf("Run 'source %s' or restart your terminal.\n", shellRc)
	}
}

func installZshCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	completionDir := filepath.Join(homeDir, ".zsh-completions")
	completionDest := filepath.Join(completionDir, "_gokku")
	shellRc := filepath.Join(homeDir, ".zshrc")

	// Create directory
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		fmt.Printf("Error creating completion directory: %v\n", err)
		os.Exit(1)
	}

	// Download completion script
	if err := downloadCompletionScript("scripts/completion.sh", completionDest); err != nil {
		fmt.Printf("Error downloading completion script: %v\n", err)
		os.Exit(1)
	}

	// Make executable
	if err := os.Chmod(completionDest, 0755); err != nil {
		fmt.Printf("Error making completion script executable: %v\n", err)
		os.Exit(1)
	}

	// Add to shell rc
	if err := addToZshRc(shellRc, completionDir); err != nil {
		fmt.Printf("Warning: Failed to add completion to %s: %v\n", shellRc, err)
		fmt.Printf("Please manually add this line to %s:\n", shellRc)
		fmt.Printf("  fpath=($HOME/.zsh-completions $fpath)\n")
	} else {
		fmt.Printf("Zsh completion installed successfully!\n")
		fmt.Printf("Run 'source %s' or restart your terminal.\n", shellRc)
	}
}

func installFishCompletion() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	completionDir := filepath.Join(homeDir, ".config", "fish", "completions")
	completionDest := filepath.Join(completionDir, "gokku.fish")

	// Create directory
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		fmt.Printf("Error creating completion directory: %v\n", err)
		os.Exit(1)
	}

	// Download completion script
	if err := downloadCompletionScript("scripts/completion.fish", completionDest); err != nil {
		fmt.Printf("Error downloading completion script: %v\n", err)
		os.Exit(1)
	}

	// Fish completions don't need to be executable
	if err := os.Chmod(completionDest, 0644); err != nil {
		fmt.Printf("Error setting completion file permissions: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fish completion installed successfully!\n")
	fmt.Printf("Restart your terminal or run 'fish -c \"complete -C gokku\"' to reload.\n")
}

func downloadCompletionScript(scriptPath, destPath string) error {
	repoURL := os.Getenv("GOKKU_REPO_URL")
	if repoURL == "" {
		repoURL = "https://raw.githubusercontent.com/thadeu/gokku/refs/heads/main"
	}

	url := fmt.Sprintf("%s/%s", repoURL, scriptPath)

	// Use curl to download
	cmd := exec.Command("curl", "-fsSL", url, "-o", destPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download completion script: %w", err)
	}

	return nil
}

func addToBashRc(shellRc, completionDest string) error {
	file, err := os.OpenFile(shellRc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if already added
	content, err := os.ReadFile(shellRc)
	if err == nil && contains(string(content), "completion.sh") {
		return nil // Already added
	}

	_, err = file.WriteString(fmt.Sprintf("\n[ -f %s ] && source %s 2>/dev/null || true\n", completionDest, completionDest))
	return err
}

func addToZshRc(shellRc, completionDir string) error {
	file, err := os.OpenFile(shellRc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if already added
	content, err := os.ReadFile(shellRc)
	if err == nil && contains(string(content), ".zsh-completions") {
		return nil // Already added
	}

	_, err = file.WriteString(fmt.Sprintf("\nfpath=($HOME/.zsh-completions $fpath)\n"))
	return err
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
