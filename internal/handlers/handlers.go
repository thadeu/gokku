package handlers

import "gokku/internal"

// HandleApps lists applications on the server
func HandleApps(args []string) {
	internal.TryCatch(func() {
		handleApps(args)
	})
}

// HandleConfigWithContext manages environment variable configuration using context
func HandleConfigWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() {
		handleConfigWithContext(ctx, args)
	})
}

// HandleRunWithContext executes arbitrary commands using context
func HandleRunWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() {
		handleRunWithContext(ctx, args)
	})
}

// HandleLogsWithContext shows application logs using context
func HandleLogsWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() {
		handleLogsWithContext(ctx, args)
	})
}

// HandleStatusWithContext shows service/container status using context
func HandleStatusWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() {
		handleStatusWithContext(ctx, args)
	})
}

// HandleRestartWithContext restarts services/containers using context
func HandleRestartWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() {
		handleRestartWithContext(ctx, args)
	})
}

// HandleRollbackWithContext rolls back to a previous release using context
func HandleRollbackWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() {
		handleRollbackWithContext(ctx, args)
	})
}

// HandleDeploy deploys applications via git push
func HandleDeploy(args []string) {
	internal.TryCatch(func() {
		handleDeploy(args)
	})
}

// HandleSSH establishes SSH connections to servers
func HandleSSH(args []string) {
	internal.TryCatch(func() {
		handleSSH(args)
	})
}

// HandleTool provides utility commands for scripts
func HandleTool(args []string) {
	internal.TryCatch(func() {
		handleTool(args)
	})
}

// HandleServer manages server connections and remotes
func HandleServer(args []string) {
	internal.TryCatch(func() {
		handleServer(args)
	})
}

// HandlePlugins manages plugin-related commands
func HandlePlugins(args []string) {
	handlePlugins(args)
}

// HandleServices manages service-related commands
func HandleServices(args []string) {
	handleServices(args)
}

// HandleAutoUpdate updates gokku based on current mode
func HandleAutoUpdate(args []string) {
	internal.TryCatch(func() {
		handleAutoUpdate(args)
	})
}
