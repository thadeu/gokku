package handlers

import "infra/internal"

// HandleApps lists applications on the server
func HandleApps(args []string) {
	handleApps(args)
}

// HandleConfig manages environment variable configuration (legacy)
func HandleConfig(args []string) {
	handleConfig(args)
}

// HandleConfigWithContext manages environment variable configuration using context
func HandleConfigWithContext(ctx *internal.ExecutionContext, args []string) {
	handleConfigWithContext(ctx, args)
}

// HandleRunWithContext executes arbitrary commands using context
func HandleRunWithContext(ctx *internal.ExecutionContext, args []string) {
	handleRunWithContext(ctx, args)
}

// HandleLogsWithContext shows application logs using context
func HandleLogsWithContext(ctx *internal.ExecutionContext, args []string) {
	handleLogsWithContext(ctx, args)
}

// HandleStatusWithContext shows service/container status using context
func HandleStatusWithContext(ctx *internal.ExecutionContext, args []string) {
	handleStatusWithContext(ctx, args)
}

// HandleRestartWithContext restarts services/containers using context
func HandleRestartWithContext(ctx *internal.ExecutionContext, args []string) {
	handleRestartWithContext(ctx, args)
}

// HandleRollbackWithContext rolls back to a previous release using context
func HandleRollbackWithContext(ctx *internal.ExecutionContext, args []string) {
	handleRollbackWithContext(ctx, args)
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
