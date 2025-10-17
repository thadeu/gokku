package internal

// Config represents the CLI configuration
type Config struct {
	Apps []AppConfig `yaml:"apps"`
}

// App represents a deployment app
type AppConfig struct {
	Name       string      `yaml:"name"`
	Lang       string      `yaml:"lang"`
	Build      *Build      `yaml:"build"`
	Deployment *Deployment `yaml:"deployment"`
}

// RemoteInfo contains information about remote connection
type RemoteInfo struct {
	Host    string
	BaseDir string
	App     string
}
