package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleDeploy deploys applications via git push
func handleDeploy(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, remoteName string

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		remoteName = remote
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]
		remoteName = fmt.Sprintf("%s-%s", app, env)
	} else {
		fmt.Println("Usage: gokku deploy <app> <env>")
		fmt.Println("   or: gokku deploy --remote <git-remote>")
		os.Exit(1)
	}

	// Determine branch based on environment
	branch := "main"
	if env == "staging" {
		branch = "staging"
	} else if env == "develop" {
		branch = "develop"
	}

	// Check if remote exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("Error: git remote '%s' not found\n", remoteName)
		fmt.Println("\nAdd it with:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remoteName, app)
		os.Exit(1)
	}

	fmt.Printf("Deploying %s to %s environment...\n", app, env)
	fmt.Printf("Remote: %s\n", remoteName)
	fmt.Printf("Branch: %s\n\n", branch)

	// Push
	cmd := exec.Command("git", "push", remoteName, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nDeploy failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Deploy complete!")
}
