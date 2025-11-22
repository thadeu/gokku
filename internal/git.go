package internal

import (
	"context"
	"fmt"

	"os"
	"os/exec"
	"strings"
)

type AbstractGitClient interface {
	AddRemote(remoteName string, remoteURL string) (string, error)
	RemoveRemote(remoteName string) (string, error)
	GetRemoteURL(remoteName string) (string, error)
}

type GitClient struct{}

func (c *GitClient) ExecuteCommand(command ...string) ([]byte, error) {
	cmd := exec.Command("git", command...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *GitClient) ExecuteCommandWithContext(ctx context.Context, command ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", command...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *GitClient) GetRemoteURL(remoteName string) (string, error) {
	output, err := c.ExecuteCommand("remote", "get-url", remoteName)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *GitClient) AddRemote(remoteName string, remoteURL string) (string, error) {
	output, err := c.ExecuteCommand("remote", "add", remoteName, remoteURL)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *GitClient) AddRemoteWithClient(client AbstractGitClient, remoteName string, remoteURL string) (string, error) {
	return client.AddRemote(remoteName, remoteURL)
}

func (c *GitClient) RemoveRemote(remoteName string) (string, error) {
	output, err := c.ExecuteCommand("remote", "remove", remoteName)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *GitClient) RemoveRemoteWithClient(client AbstractGitClient, remoteName string) (string, error) {
	return client.RemoveRemote(remoteName)
}

func GetRemoteInfoWithClient(client AbstractGitClient, remoteName string) (*RemoteInfo, error) {
	// Get remote URL
	remoteURL, err := client.GetRemoteURL(remoteName)

	if err != nil {
		return nil, fmt.Errorf("git remote '%s' not found. Add it with: git remote add %s user@host:/opt/gokku/repos/<app>.git", remoteName, remoteName)
	}

	remoteURL = strings.TrimSpace(remoteURL)

	// Parse:
	// user@host:/opt/gokku/repos/app-name.git
	// or
	// user@host:app-name
	parts := strings.Split(remoteURL, ":")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid remote URL format: %s (expected user@host:/path)", remoteURL)
	}

	host := parts[0]
	path := parts[1]

	// Extract app name from path: api -> api
	pathParts := strings.Split(path, "/")

	if len(pathParts) == 1 {
		pathParts = append(pathParts, "/opt/gokku/repos", pathParts[0])
	} else if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid remote path: %s", path)
	}

	appFile := pathParts[len(pathParts)-1]         // api.git
	appName := strings.TrimSuffix(appFile, ".git") // api

	// Extract base dir: api -> /opt/gokku
	baseDir := strings.TrimSuffix(path, "/repos/"+appFile)

	return &RemoteInfo{
		Host:    host,
		BaseDir: baseDir,
		App:     appName,
	}, nil
}

// GetRemoteInfo extracts info from git remote
// Example: ubuntu@server:api
// Returns: RemoteInfo{Host: "ubuntu@server", BaseDir: "/opt/gokku", App: "api"}
func GetRemoteInfo(remoteName string) (*RemoteInfo, error) {
	return GetRemoteInfoWithClient(&GitClient{}, remoteName)
}
