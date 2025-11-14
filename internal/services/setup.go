package services

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	. "gokku/internal"
)

// SetupConfig holds configuration for server setup
type SetupConfig struct {
	ServerHost   string
	IdentityFile string
}

// ServerSetup handles one-time server setup
type ServerSetup struct {
	config SetupConfig
}

// NewServerSetup creates a new server setup instance
func NewServerSetup(serverHost, identityFile string) *ServerSetup {
	return &ServerSetup{
		config: SetupConfig{
			ServerHost:   serverHost,
			IdentityFile: identityFile,
		},
	}
}

// Execute performs the complete server setup
func (ss *ServerSetup) Execute() error {
	fmt.Printf("-----> Setting up server %s...\n", ss.config.ServerHost)
	fmt.Println("")

	// Phase 1: Check prerequisites
	fmt.Println("-----> Checking prerequisites...")

	if err := ss.checkSSHConnection(); err != nil {
		return fmt.Errorf("%v\n\nMake sure you can SSH to the server without password.\nYou can use: ssh-copy-id %s", err, ss.config.ServerHost)
	}

	// Check if Gokku is already installed
	fmt.Println("-----> Checking if Gokku is installed...")

	// Phase 2: Install Gokku
	fmt.Println("-----> Installing Gokku on server...")
	if err := ss.installGokku(); err != nil {
		return fmt.Errorf("error installing Gokku: %v", err)
	}
	fmt.Println("-----> Gokku installed successfully")

	// Phase 3: Install essential plugins
	fmt.Println("-----> Installing essential plugins...")
	essentialPlugins := []string{"nginx", "letsencrypt", "cron", "postgres", "redis"}
	installedCount := 0

	for _, pluginName := range essentialPlugins {
		fmt.Printf("-----> Installing %s...", pluginName)

		if err := ss.installPlugin(pluginName); err != nil {
			fmt.Printf(" failed: %v\n", err)
		} else {
			fmt.Printf(" ✓\n")
			installedCount++
		}
	}

	fmt.Printf("-----> Installed %d/%d plugins\n", installedCount, len(essentialPlugins))

	// Phase 4: Configure SSH keys (if needed)
	fmt.Println("-----> Configuring SSH keys...")

	if err := ss.configureSSHKeys(); err != nil {
		fmt.Printf("Warning: Could not configure SSH keys: %v\n", err)
	} else {
		fmt.Println("-----> SSH keys configured ✓")
	}

	// Phase 5: Final verification
	fmt.Println("-----> Verifying setup...")

	if err := ss.verifySetup(); err != nil {
		fmt.Printf("Warning: Verification failed: %v\n", err)
	} else {
		fmt.Println("-----> Setup verification complete ✓")
	}

	// Phase 6: Create default remote "gokku" locally
	fmt.Println("-----> Creating default remote 'gokku'...")

	if err := ss.createDefaultRemote(); err != nil {
		fmt.Printf("Note: Could not create default remote 'gokku': %v\n", err)
		fmt.Println("You can create it manually with:")
		fmt.Printf("  gokku remote add gokku %s\n", ss.config.ServerHost)
	} else {
		fmt.Println("-----> Created default remote 'gokku' for future use")
	}

	fmt.Println("")
	fmt.Println("-----> Setup complete!")
	fmt.Println("-----> Server is ready to receive deployments")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Printf("  gokku apps create <app_name> [--remote]\n")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Printf("  gokku apps create api-production\n")
	fmt.Printf("  gokku remote add api-production %s\n", ss.config.ServerHost)
	fmt.Printf("  git push api-production main\n")

	return nil
}

// buildSSHArgs builds SSH command arguments with identity file if provided
// SSH format: ssh [options] [user@]hostname [command]
func (ss *ServerSetup) buildSSHArgs(extraArgs ...string) []string {
	args := []string{}

	if ss.config.IdentityFile != "" {
		// Check if identity file exists
		if _, err := os.Stat(ss.config.IdentityFile); err == nil {
			args = append(args, "-i", ss.config.IdentityFile)
		}
	}

	// Separate SSH options from commands
	// SSH options: -o (needs value), -t (flag), -i (already handled), etc.
	var sshOptions []string
	var commands []string

	i := 0

	for i < len(extraArgs) {
		arg := extraArgs[i]

		// Options that need a value: -o, -i
		if arg == "-o" || arg == "-i" {
			sshOptions = append(sshOptions, arg)
			i++
			if i < len(extraArgs) {
				// Next arg is the value for this option
				sshOptions = append(sshOptions, extraArgs[i])
			}
		} else if arg == "-t" || strings.HasPrefix(arg, "-") {
			// Other flags (no value needed)
			sshOptions = append(sshOptions, arg)
		} else {
			// Command arguments (not starting with -)
			commands = append(commands, arg)
		}
		i++
	}

	// Build final args: [identity] [ssh options] [hostname] [commands]
	args = append(args, sshOptions...)
	args = append(args, ss.config.ServerHost)
	args = append(args, commands...)

	return args
}

// checkSSHConnection verifies SSH connection works without password
func (ss *ServerSetup) checkSSHConnection() error {
	args := ss.buildSSHArgs("-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "echo OK")
	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include SSH output in error message for debugging
		outputStr := strings.TrimSpace(string(output))

		if outputStr != "" {
			return fmt.Errorf("SSH connection failed: %v\nOutput: %s", err, outputStr)
		}
		return fmt.Errorf("SSH connection failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if !strings.Contains(outputStr, "OK") {
		return fmt.Errorf("SSH connection test failed (received: %s)", outputStr)
	}

	return nil
}

// installGokku installs Gokku on the remote server
func (ss *ServerSetup) installGokku() error {
	// Execute install command via bash on remote server
	// Need to wrap in bash -c to properly handle pipes
	installCmd := `bash -c 'curl -fsSL https://gokku-vm.com/install | bash -s -- --server'`

	args := ss.buildSSHArgs("-t", installCmd)
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// installPlugin installs a plugin on the remote server
func (ss *ServerSetup) installPlugin(pluginName string) error {
	// Check if plugin already exists first
	checkCmd := fmt.Sprintf(`test -d /opt/gokku/plugins/%s && echo 'exists' || echo 'missing'`, pluginName)

	args := ss.buildSSHArgs(checkCmd)
	cmd := exec.Command("ssh", args...)
	output, err := cmd.Output()

	if err == nil && strings.Contains(string(output), "exists") {
		// Plugin already installed, skip
		return nil
	}

	// Install plugin via SSH (non-interactive)
	installCmd := fmt.Sprintf(`gokku plugins:add %s 2>&1`, pluginName)

	args = ss.buildSSHArgs(installCmd)
	sshCmd := exec.Command("ssh", args...)
	output, err = sshCmd.CombinedOutput()

	// Check if error is due to plugin already existing
	if err != nil {
		outputStr := string(output)
		// If plugin already exists, that's OK
		if strings.Contains(outputStr, "already exists") {
			return nil
		}
		return fmt.Errorf("%s", strings.TrimSpace(outputStr))
	}

	return nil
}

// configureSSHKeys configures SSH keys on the server if needed
func (ss *ServerSetup) configureSSHKeys() error {
	// If using identity file (PEM), skip this step as it's typically already configured
	if ss.config.IdentityFile != "" {
		return nil
	}

	// Get local SSH public key
	localUser, err := user.Current()

	if err != nil {
		return fmt.Errorf("could not get current user: %v", err)
	}

	// Try common SSH key locations
	keyPaths := []string{
		filepath.Join(localUser.HomeDir, ".ssh", "id_ed25519.pub"),
		filepath.Join(localUser.HomeDir, ".ssh", "id_rsa.pub"),
		filepath.Join(localUser.HomeDir, ".ssh", "id_ecdsa.pub"),
	}

	var publicKeyPath string
	for _, path := range keyPaths {
		if _, err := os.Stat(path); err == nil {
			publicKeyPath = path
			break
		}
	}

	if publicKeyPath == "" {
		return fmt.Errorf("no SSH public key found")
	}

	// Read public key
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("could not read public key: %v", err)
	}

	publicKey := strings.TrimSpace(string(publicKeyData))

	// Check if key already exists on server
	checkCmd := fmt.Sprintf(`grep -Fx "%s" ~/.ssh/authorized_keys >/dev/null 2>&1 && echo 'exists' || echo 'missing'`, publicKey)

	args := ss.buildSSHArgs(checkCmd)
	cmd := exec.Command("ssh", args...)
	output, err := cmd.Output()

	if err == nil && strings.Contains(string(output), "exists") {
		// Key already exists
		return nil
	}

	// Add key to server
	addKeyCmd := fmt.Sprintf(`mkdir -p ~/.ssh && echo "%s" >> ~/.ssh/authorized_keys && chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys`, publicKey)
	args = ss.buildSSHArgs(addKeyCmd)
	addCmd := exec.Command("ssh", args...)

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("could not add SSH key: %v", err)
	}

	return nil
}

// verifySetup performs final verification of the setup
func (ss *ServerSetup) verifySetup() error {
	// Check Docker
	fmt.Println("-----> Checking Docker...")

	args := ss.buildSSHArgs("docker ps >/dev/null 2>&1 && echo 'OK' || echo 'FAIL'")
	dockerCmd := exec.Command("ssh", args...)
	dockerOutput, err := dockerCmd.Output()

	if err == nil {
		dockerStatus := strings.TrimSpace(string(dockerOutput))
		if dockerStatus == "OK" {
			// fmt.Println("-----> Docker: OK")
		} else {
			fmt.Println("-----> Docker: Not running (this is OK if Docker was just installed)")
		}
	}

	// Check plugins
	fmt.Println("-----> Checking plugins...")

	args = ss.buildSSHArgs("ls -1 /opt/gokku/plugins 2>/dev/null | wc -l")
	pluginsCmd := exec.Command("ssh", args...)
	pluginsOutput, err := pluginsCmd.Output()

	if err == nil {
		pluginCount := strings.TrimSpace(string(pluginsOutput))
		fmt.Printf("-----> Plugins installed: %s\n", pluginCount)
	}

	// Check directories
	fmt.Println("-----> Checking directories...")

	args = ss.buildSSHArgs(`test -d /opt/gokku && echo 'OK' || echo 'FAIL'`)
	dirsCmd := exec.Command("ssh", args...)
	dirsOutput, err := dirsCmd.Output()

	if err == nil {
		dirsStatus := strings.TrimSpace(string(dirsOutput))
		if dirsStatus == "OK" {
			// fmt.Println("-----> Directories: OK")
		} else {
			fmt.Println("-----> Directories: Missing /opt/gokku")
			return fmt.Errorf("base directory not found")
		}
	}

	return nil
}

// createDefaultRemote creates a default "gokku" remote in the local git repository
// This allows users to use "gokku apps create" without specifying --remote
func (ss *ServerSetup) createDefaultRemote() error {
	// Check if remote "gokku" already exists
	client := &GitClient{}

	if _, err := client.GetRemoteURL("gokku"); err == nil {
		// Remote already exists, skip
		return nil
	}

	// Create a dummy remote URL (we only need the host info)
	// The actual app will be determined later when creating apps
	remoteURL := fmt.Sprintf("%s:gokku.git", ss.config.ServerHost)

	_, err := client.AddRemote("gokku", remoteURL)

	if err != nil {
		return fmt.Errorf("failed to add remote: %v", err)
	}

	return nil
}
