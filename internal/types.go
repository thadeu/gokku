package internal

// Config represents the CLI configuration
type Config struct {
	Servers []Server `yaml:"servers"`
}

// Server represents a deployment server
type Server struct {
	Name    string `yaml:"name"`
	Host    string `yaml:"host"`
	BaseDir string `yaml:"base_dir"`
	Default bool   `yaml:"default,omitempty"`
}

// RemoteInfo contains information about remote connection
type RemoteInfo struct {
	Host    string
	BaseDir string
	App     string
	Env     string
}
