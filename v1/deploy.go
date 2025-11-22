package v1

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gokku/internal"

	"gopkg.in/yaml.v3"
)

// DeployCommand gerencia deploy de aplicações
type DeployCommand struct {
	output  Output
	baseDir string
}

// NewDeployCommand cria uma nova instância de DeployCommand
func NewDeployCommand(output Output) *DeployCommand {
	return &DeployCommand{
		output:  output,
		baseDir: "/opt/gokku",
	}
}

// Execute executa o deploy de uma aplicação
func (c *DeployCommand) Execute(appName string) error {
	c.output.Print(fmt.Sprintf("Deploying app '%s'...", appName))

	repoDir := filepath.Join(c.baseDir, "repos", appName+".git")

	// Verificar se o repositório tem commits
	if !c.hasCommits(repoDir) {
		c.output.Error("Repository is empty, cannot deploy")
		return fmt.Errorf("repository is empty")
	}

	// Criar release directory
	releaseID := time.Now().Format("20060102150405")
	releaseDir := filepath.Join(c.baseDir, "apps", appName, "releases", releaseID)

	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		c.output.Error(fmt.Sprintf("Failed to create release directory: %v", err))
		return err
	}

	c.output.Print(fmt.Sprintf("-----> Creating release %s", releaseID))

	// Extrair código do repositório
	if err := c.extractCodeFromRepo(appName, repoDir, releaseDir); err != nil {
		c.output.Error(fmt.Sprintf("Failed to extract code: %v", err))
		return err
	}

	// Verificar se existe gokku.yml
	gokkuYmlPath := filepath.Join(releaseDir, "gokku.yml")
	if _, err := os.Stat(gokkuYmlPath); err == nil {
		c.output.Print("-----> Found gokku.yml, processing configuration...")
		if err := c.initialSetup(appName, gokkuYmlPath, releaseDir); err != nil {
			c.output.Print(fmt.Sprintf("Warning: Failed to process gokku.yml: %v", err))
		}
	}

	// Detectar linguagem e fazer build
	if err := c.buildRelease(appName, releaseDir); err != nil {
		c.output.Error(fmt.Sprintf("Failed to build release: %v", err))
		return err
	}

	// Atualizar symlink current
	currentLink := filepath.Join(c.baseDir, "apps", appName, "current")
	if err := c.updateCurrentSymlink(currentLink, releaseDir); err != nil {
		c.output.Error(fmt.Sprintf("Failed to update current symlink: %v", err))
		return err
	}

	// Executar comandos post-deploy
	if err := c.executePostDeployCommands(appName, releaseDir); err != nil {
		c.output.Print(fmt.Sprintf("Warning: Post-deploy commands failed: %v", err))
	}

	// Iniciar/reiniciar containers
	if err := c.startContainers(appName, releaseDir); err != nil {
		c.output.Error(fmt.Sprintf("Failed to start containers: %v", err))
		return err
	}

	c.output.Success(fmt.Sprintf("Deploy completed successfully for '%s'", appName))
	return nil
}

// Métodos privados

func (c *DeployCommand) hasCommits(repoDir string) bool {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--verify", "HEAD")
	return cmd.Run() == nil
}

func (c *DeployCommand) extractCodeFromRepo(appName, repoDir, releaseDir string) error {
	// Get current HEAD branch
	headCmd := exec.Command("git", "-C", repoDir, "symbolic-ref", "HEAD")
	headOutput, err := headCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get HEAD branch: %w", err)
	}

	branch := strings.TrimSpace(string(headOutput))
	branch = strings.TrimPrefix(branch, "refs/heads/")

	c.output.Print(fmt.Sprintf("-----> Extracting code from branch: %s", branch))

	// Clone repository to release directory
	cmd := exec.Command("git", "clone", "--branch", branch, "--depth", "1", repoDir, releaseDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %v (output: %s)", err, string(output))
	}

	// Remove .git directory from release
	gitDir := filepath.Join(releaseDir, ".git")
	os.RemoveAll(gitDir)

	return nil
}

func (c *DeployCommand) initialSetup(appName, gokkuYmlPath, releaseDir string) error {
	// Carregar gokku.yml
	data, err := os.ReadFile(gokkuYmlPath)
	if err != nil {
		return err
	}

	var config internal.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Atualizar arquivo de ambiente
	envFile := filepath.Join(c.baseDir, "apps", appName, "shared", ".env")
	if err := c.updateEnvironmentFile(envFile, appName); err != nil {
		return err
	}

	return nil
}

func (c *DeployCommand) updateEnvironmentFile(envFile, appName string) error {
	// Carregar configuração
	config, err := internal.LoadConfig()
	if err != nil {
		return err
	}

	appConfig := config.GetAppConfig(appName)
	if appConfig == nil {
		return nil // No app config
	}

	// Carregar env vars existentes
	envVars := internal.LoadEnvFile(envFile)

	// Adicionar/atualizar env vars do config
	if appConfig.Env != nil {
		for key, value := range appConfig.Env {
			envVars[key] = value
		}
	}

	// Salvar env file
	return internal.SaveEnvFile(envFile, envVars)
}

func (c *DeployCommand) buildRelease(appName, releaseDir string) error {
	c.output.Print("-----> Building release...")

	// TODO: Implementar detecção de linguagem e build
	// Por enquanto, apenas verificar se existe Dockerfile
	dockerfilePath := filepath.Join(releaseDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		c.output.Print("-----> Found Dockerfile")
		// Build com Docker será feito no startContainers
	}

	c.output.Print("-----> Build completed")
	return nil
}

func (c *DeployCommand) updateCurrentSymlink(currentLink, releaseDir string) error {
	// Remover symlink antigo se existir
	os.Remove(currentLink)

	// Criar novo symlink
	if err := os.Symlink(releaseDir, currentLink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func (c *DeployCommand) executePostDeployCommands(appName, releaseDir string) error {
	config, err := internal.LoadAppConfig(appName)
	if err != nil {
		return nil // No config
	}

	if config.Deployment == nil || config.Deployment.PostDeploy == nil || len(config.Deployment.PostDeploy) == 0 {
		return nil // No post-deploy commands
	}

	c.output.Print("-----> Running post-deploy commands...")

	for _, command := range config.Deployment.PostDeploy {
		c.output.Print(fmt.Sprintf("       $ %s", command))

		cmd := exec.Command("bash", "-c", command)
		cmd.Dir = releaseDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("post-deploy command failed: %w", err)
		}
	}

	return nil
}

func (c *DeployCommand) startContainers(appName, releaseDir string) error {
	c.output.Print("-----> Starting containers...")

	envFile := filepath.Join(c.baseDir, "apps", appName, "shared", ".env")

	// Recriar container ativo
	if err := internal.RecreateActiveContainer(appName, envFile, releaseDir); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	c.output.Print("-----> Containers started")
	return nil
}

func (c *DeployCommand) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
