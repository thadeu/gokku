package handlers

import (
	"fmt"
	"os"
)

// handleStatus shows service/container status (legacy - should not be used)
func handleStatus(args []string) {
	fmt.Println("Error: This handler should not be called directly")
	os.Exit(1)
}
