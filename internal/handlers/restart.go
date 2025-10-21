package handlers

import (
	"fmt"
	"os"
)

// handleRestart restarts services/containers (legacy - should not be used)
func handleRestart(args []string) {
	fmt.Println("Error: This handler should not be called directly")
	os.Exit(1)
}
