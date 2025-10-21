package handlers

import (
	"fmt"
	"os"
)

// handleLogs shows application logs (legacy - should not be used)
func handleLogs(args []string) {
	fmt.Println("Error: This handler should not be called directly")
	os.Exit(1)
}
