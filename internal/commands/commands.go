package commands

import "gokku/internal"

func Apps(args []string) {
	internal.TryCatch(func() { useApps(args) })
}

func Processes(args []string) {
	internal.TryCatch(func() { useProcesses(args) })
}

func ConfigWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() { useConfigWithContext(ctx, args) })
}

func RunWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() { useRunWithContext(ctx, args) })
}

func LogsWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() { useLogsWithContext(ctx, args) })
}

func StatusWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() { useStatusWithContext(ctx, args) })
}

func RestartWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() { useRestartWithContext(ctx, args) })
}

func RollbackWithContext(ctx *internal.ExecutionContext, args []string) {
	internal.TryCatch(func() { useRollbackWithContext(ctx, args) })
}

func Deploy(args []string) {
	internal.TryCatch(func() { useDeploy(args) })
}

func Tool(args []string) {
	internal.TryCatch(func() { useTool(args) })
}

func Remote(args []string) {
	internal.TryCatch(func() { useRemote(args) })
}

func Uninstall(args []string) {
	internal.TryCatch(func() { useUninstall(args) })
}

func Plugins(args []string) {
	internal.TryCatch(func() { usePlugins(args) })
}

func Services(args []string) {
	internal.TryCatch(func() { useServices(args) })
}

func AutoUpdate(args []string) {
	internal.TryCatch(func() { useAutoUpdate(args) })
}
