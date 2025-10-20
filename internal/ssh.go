package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

func Bash(command string) string {
	cmd := exec.Command("/bin/bash", "-c", command)

	output, err := cmd.Output()

	if err != nil {
		fmt.Println("Error running command: ", err)
		return ""
	}

	return strings.TrimSpace(string(output))
}
