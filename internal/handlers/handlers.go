package handlers

// HandleApps lists applications on the server
func HandleApps(args []string) {
	handleApps(args)
}

// HandleConfig manages environment variable configuration
func HandleConfig(args []string) {
	handleConfig(args)
}

// HandleRun executes arbitrary commands on remote servers
func HandleRun(args []string) {
	handleRun(args)
}

// HandleLogs shows application logs
func HandleLogs(args []string) {
	handleLogs(args)
}

// HandleStatus shows service/container status
func HandleStatus(args []string) {
	handleStatus(args)
}

// HandleRestart restarts services/containers
func HandleRestart(args []string) {
	handleRestart(args)
}

// HandleDeploy deploys applications via git push
func HandleDeploy(args []string) {
	handleDeploy(args)
}

// HandleRollback rolls back to a previous release
func HandleRollback(args []string) {
	handleRollback(args)
}

// HandleSSH establishes SSH connections to servers
func HandleSSH(args []string) {
	handleSSH(args)
}

// HandleTool provides utility commands for scripts
func HandleTool(args []string) {
	handleTool(args)
}

// HandleServer manages server connections and remotes
func HandleServer(args []string) {
	handleServer(args)
}

// HandlePlugins manages plugin-related commands
func HandlePlugins(args []string) {
	handlePlugins(args)
}

// HandleServices manages service-related commands
func HandleServices(args []string) {
	handleServices(args)
}
