package handlers

import (
	"fmt"
	"os"
)

// handleRollback rolls back to a previous release (legacy - should not be used)
func handleRollback(args []string) {
	fmt.Println("Error: This handler should not be called directly")
	os.Exit(1)
}
