package internal

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

// TryCatch executes a function with panic recovery (elegant version)
func TryCatch(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error: %v\n", r)
			fmt.Println("")
			fmt.Println("This is likely a bug in gokku. Please report this issue with the following details:")
			fmt.Println("")

			// Get stack trace
			stack := debug.Stack()
			fmt.Printf("Stack trace:\n%s\n", stack)

			// Get runtime info
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

			os.Exit(1)
		}
	}()

	fn()
}

// TryCatchE executes a function with panic recovery and returns error (elegant version)
func TryCatchE(fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error: %v\n", r)
			fmt.Println("")
			fmt.Println("This is likely a bug in gokku. Please report this issue with the following details:")
			fmt.Println("")

			// Get stack trace
			stack := debug.Stack()
			fmt.Printf("Stack trace:\n%s\n", stack)

			// Get runtime info
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

			os.Exit(1)
		}
	}()

	return fn()
}
