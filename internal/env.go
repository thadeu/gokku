package internal

import (
	"fmt"
	"os"
	"strings"
)

// EnvSet sets environment variables in the env file
func EnvSet(envFile string, pairs []string) {
	envVars := LoadEnvFile(envFile)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: invalid format '%s', expected KEY=VALUE\n", pair)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
		fmt.Printf("%s=%s\n", key, value)
	}

	if err := SaveEnvFile(envFile, envVars); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}
}

// EnvGet gets a specific environment variable from the env file
func EnvGet(envFile string, key string) {
	envVars := LoadEnvFile(envFile)

	if value, ok := envVars[key]; ok {
		fmt.Println(value)
	} else {
		fmt.Printf("Error: variable '%s' not found\n", key)
		os.Exit(1)
	}
}

// EnvList lists all environment variables in the env file
func EnvList(envFile string) {
	envVars := LoadEnvFile(envFile)

	if len(envVars) == 0 {
		fmt.Println("No environment variables set")
		return
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}

	// Sort alphabetically
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, envVars[key])
	}
}

// EnvUnset removes environment variables from the env file
func EnvUnset(envFile string, keys []string) {
	envVars := LoadEnvFile(envFile)

	for _, key := range keys {
		if _, ok := envVars[key]; ok {
			delete(envVars, key)
			fmt.Printf("Unset %s\n", key)
		} else {
			fmt.Printf("Warning: variable '%s' not found\n", key)
		}
	}

	if err := SaveEnvFile(envFile, envVars); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}
}

// LoadEnvFile loads environment variables from a file
func LoadEnvFile(envFile string) map[string]string {
	envVars := make(map[string]string)

	content, err := os.ReadFile(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			return envVars // Return empty map if file doesn't exist
		}
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	return envVars
}

// SaveEnvFile saves environment variables to a file
func SaveEnvFile(envFile string, envVars map[string]string) error {
	// Sort keys
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var content strings.Builder
	for _, key := range keys {
		content.WriteString(fmt.Sprintf("%s=%s\n", key, envVars[key]))
	}

	return os.WriteFile(envFile, []byte(content.String()), 0600)
}
