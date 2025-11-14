package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestServiceManagerTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(ServiceManagerTestSuite))
}

type ServiceManagerTestSuite struct {
	suite.Suite
	tempDir     string
	servicesDir string
	pluginsDir  string
	manager     *ServiceManager
}

func (s *ServiceManagerTestSuite) SetupSuite() {
	// Suite-level setup if needed
}

func (s *ServiceManagerTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "gokku-test-*")
	s.Require().NoError(err)

	s.servicesDir = filepath.Join(s.tempDir, "services")
	s.pluginsDir = filepath.Join(s.tempDir, "plugins")

	err = os.MkdirAll(s.servicesDir, 0755)
	s.Require().NoError(err)
	err = os.MkdirAll(s.pluginsDir, 0755)
	s.Require().NoError(err)

	s.manager = &ServiceManager{
		servicesDir: s.servicesDir,
		pluginsDir:  s.pluginsDir,
	}
}

func (s *ServiceManagerTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *ServiceManagerTestSuite) TestNewServiceManager() {
	manager := NewServiceManager()

	Expect(manager).ToNot(BeNil())
	Expect(manager.servicesDir).To(Equal("/opt/gokku/services"))
	Expect(manager.pluginsDir).To(Equal("/opt/gokku/plugins"))
}

func (s *ServiceManagerTestSuite) TestPluginExists_WhenPluginExists() {
	pluginDir := filepath.Join(s.pluginsDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	s.Require().NoError(err)

	exists := s.manager.pluginExists("test-plugin")
	Expect(exists).To(BeTrue())
}

func (s *ServiceManagerTestSuite) TestPluginExists_WhenPluginDoesNotExist() {
	exists := s.manager.pluginExists("non-existent-plugin")
	Expect(exists).To(BeFalse())
}

func (s *ServiceManagerTestSuite) TestServiceExists_WhenServiceExists() {
	serviceDir := filepath.Join(s.servicesDir, "test-service")
	err := os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	exists := s.manager.serviceExists("test-service")
	Expect(exists).To(BeTrue())
}

func (s *ServiceManagerTestSuite) TestServiceExists_WhenServiceDoesNotExist() {
	exists := s.manager.serviceExists("non-existent-service")
	Expect(exists).To(BeFalse())
}

func (s *ServiceManagerTestSuite) TestAppExists_WhenAppDoesNotExist() {
	exists := s.manager.appExists("non-existent-app", "")
	Expect(exists).To(BeFalse())
}

func (s *ServiceManagerTestSuite) TestIsAppLinked_WhenAppIsLinked() {
	linkedApps := []string{"app1", "app2", "app3"}
	result := s.manager.isAppLinked(linkedApps, "app2", "")
	Expect(result).To(BeTrue())
}

func (s *ServiceManagerTestSuite) TestIsAppLinked_WhenAppIsNotLinked() {
	linkedApps := []string{"app1", "app2", "app3"}
	result := s.manager.isAppLinked(linkedApps, "app4", "")
	Expect(result).To(BeFalse())
}

func (s *ServiceManagerTestSuite) TestIsAppLinked_WhenListIsEmpty() {
	linkedApps := []string{}
	result := s.manager.isAppLinked(linkedApps, "app1", "")
	Expect(result).To(BeFalse())
}

func (s *ServiceManagerTestSuite) TestSaveServiceConfig() {
	// Create service directory first
	serviceDir := filepath.Join(s.servicesDir, "test-service")

	err := os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	service := Service{
		Name:       "test-service",
		Plugin:     "postgres",
		Running:    false,
		LinkedApps: []string{},
		Config:     make(map[string]string),
	}

	err = s.manager.saveServiceConfig("test-service", service)
	Expect(err).To(BeNil())

	configPath := filepath.Join(s.servicesDir, "test-service", "config.json")
	Expect(configPath).To(BeAnExistingFile())
}

func (s *ServiceManagerTestSuite) TestGetServiceConfig_WhenServiceExists() {
	// Create service directory first
	serviceDir := filepath.Join(s.servicesDir, "test-service")
	err := os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	service := Service{
		Name:       "test-service",
		Plugin:     "postgres",
		Running:    false,
		LinkedApps: []string{"app1"},
		Config:     map[string]string{"version": "14"},
	}

	err = s.manager.saveServiceConfig("test-service", service)
	s.Require().NoError(err)

	retrieved, err := s.manager.getServiceConfig("test-service")

	Expect(err).To(BeNil())
	Expect(retrieved.Name).To(Equal("test-service"))
	Expect(retrieved.Plugin).To(Equal("postgres"))
	Expect(retrieved.LinkedApps).To(ContainElement("app1"))
	Expect(retrieved.Config["version"]).To(Equal("14"))
}

func (s *ServiceManagerTestSuite) TestGetServiceConfig_WhenServiceDoesNotExist() {
	_, err := s.manager.getServiceConfig("non-existent-service")
	Expect(err).ToNot(BeNil())
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvVars_Postgres() {
	config := map[string]string{
		"port":     "5432",
		"user":     "postgres",
		"password": "secret",
		"database": "mydb",
	}

	envVars := s.manager.getServiceEnvVars("test-service", config, "postgres")

	Expect(envVars).ToNot(BeEmpty())
	Expect(envVars["DATABASE_URL"]).To(ContainSubstring("postgres://"))
	Expect(envVars["POSTGRES_HOST"]).To(Equal("localhost"))
	Expect(envVars["POSTGRES_PORT"]).To(Equal("5432"))
	Expect(envVars["POSTGRES_USER"]).To(Equal("postgres"))
	Expect(envVars["POSTGRES_PASSWORD"]).To(Equal("secret"))
	Expect(envVars["POSTGRES_DB"]).To(Equal("mydb"))
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvVars_Postgres_IncompleteConfig() {
	config := map[string]string{
		"port": "5432",
		// Missing user, password, database
	}

	envVars := s.manager.getServiceEnvVars("test-service", config, "postgres")
	Expect(envVars).To(BeEmpty())
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvVars_Redis() {
	config := map[string]string{
		"port":     "6379",
		"password": "secret",
	}

	envVars := s.manager.getServiceEnvVars("test-service", config, "redis")

	Expect(envVars).ToNot(BeEmpty())
	Expect(envVars["REDIS_URL"]).To(ContainSubstring("redis://"))
	Expect(envVars["REDIS_HOST"]).To(Equal("localhost"))
	Expect(envVars["REDIS_PORT"]).To(Equal("6379"))
	Expect(envVars["REDIS_PASSWORD"]).To(Equal("secret"))
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvVars_Redis_IncompleteConfig() {
	config := map[string]string{
		"port": "6379",
		// Missing password
	}

	envVars := s.manager.getServiceEnvVars("test-service", config, "redis")
	Expect(envVars).To(BeEmpty())
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvVars_UnknownPlugin() {
	config := map[string]string{"key": "value"}
	envVars := s.manager.getServiceEnvVars("test-service", config, "unknown-plugin")
	Expect(envVars).To(BeEmpty())
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvKeys_Postgres() {
	keys := s.manager.getServiceEnvKeys("postgres")

	Expect(keys).To(ContainElement("DATABASE_URL"))
	Expect(keys).To(ContainElement("POSTGRES_HOST"))
	Expect(keys).To(ContainElement("POSTGRES_PORT"))
	Expect(keys).To(ContainElement("POSTGRES_USER"))
	Expect(keys).To(ContainElement("POSTGRES_PASSWORD"))
	Expect(keys).To(ContainElement("POSTGRES_DB"))
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvKeys_Redis() {
	keys := s.manager.getServiceEnvKeys("redis")

	Expect(keys).To(ContainElement("REDIS_URL"))
	Expect(keys).To(ContainElement("REDIS_HOST"))
	Expect(keys).To(ContainElement("REDIS_PORT"))
	Expect(keys).To(ContainElement("REDIS_PASSWORD"))
}

func (s *ServiceManagerTestSuite) TestGetServiceEnvKeys_UnknownPlugin() {
	keys := s.manager.getServiceEnvKeys("unknown-plugin")
	Expect(keys).To(BeEmpty())
}

func (s *ServiceManagerTestSuite) TestListServices_Empty() {
	services, err := s.manager.ListServices()
	Expect(err).To(BeNil())
	Expect(services).To(BeEmpty())
}

func (s *ServiceManagerTestSuite) TestListServices_WithServices() {
	// Create service directories first
	service1Dir := filepath.Join(s.servicesDir, "service1")
	service2Dir := filepath.Join(s.servicesDir, "service2")

	err := os.MkdirAll(service1Dir, 0755)
	s.Require().NoError(err)
	err = os.MkdirAll(service2Dir, 0755)
	s.Require().NoError(err)

	// Create service configs
	service1 := Service{
		Name:       "service1",
		Plugin:     "postgres",
		Running:    false,
		LinkedApps: []string{},
		Config:     make(map[string]string),
	}

	service2 := Service{
		Name:       "service2",
		Plugin:     "redis",
		Running:    true,
		LinkedApps: []string{"app1"},
		Config:     make(map[string]string),
	}

	err = s.manager.saveServiceConfig("service1", service1)
	s.Require().NoError(err)

	err = s.manager.saveServiceConfig("service2", service2)
	s.Require().NoError(err)

	services, err := s.manager.ListServices()
	Expect(err).To(BeNil())
	Expect(len(services)).To(Equal(2))
}

func (s *ServiceManagerTestSuite) TestGetService_WhenServiceExists() {
	// Create service directory first
	serviceDir := filepath.Join(s.servicesDir, "test-service")
	err := os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	service := Service{
		Name:       "test-service",
		Plugin:     "postgres",
		Running:    false,
		LinkedApps: []string{},
		Config:     make(map[string]string),
	}

	err = s.manager.saveServiceConfig("test-service", service)
	s.Require().NoError(err)

	retrieved, err := s.manager.GetService("test-service")
	Expect(err).To(BeNil())
	Expect(retrieved.Name).To(Equal("test-service"))
	Expect(retrieved.Plugin).To(Equal("postgres"))
}

func (s *ServiceManagerTestSuite) TestGetService_WhenServiceDoesNotExist() {
	_, err := s.manager.GetService("non-existent-service")
	Expect(err).ToNot(BeNil())
}

func (s *ServiceManagerTestSuite) TestUpdateServiceConfig() {
	// Create service directory first
	serviceDir := filepath.Join(s.servicesDir, "test-service")
	err := os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	service := Service{
		Name:       "test-service",
		Plugin:     "postgres",
		Running:    false,
		LinkedApps: []string{},
		Config:     make(map[string]string),
	}

	err = s.manager.saveServiceConfig("test-service", service)
	s.Require().NoError(err)

	newConfig := map[string]string{
		"version": "14",
		"port":    "5432",
	}

	err = s.manager.UpdateServiceConfig("test-service", newConfig)
	Expect(err).To(BeNil())

	updated, err := s.manager.getServiceConfig("test-service")
	Expect(err).To(BeNil())
	Expect(updated.Config["version"]).To(Equal("14"))
	Expect(updated.Config["port"]).To(Equal("5432"))
}

func (s *ServiceManagerTestSuite) TestCreateService_WhenPluginDoesNotExist() {
	err := s.manager.CreateService("non-existent-plugin", "test-service", "")

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("plugin"))
	Expect(err.Error()).To(ContainSubstring("not found"))
}

func (s *ServiceManagerTestSuite) TestCreateService_WhenServiceAlreadyExists() {
	// Create plugin directory
	pluginDir := filepath.Join(s.pluginsDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	s.Require().NoError(err)

	// Create service directory
	serviceDir := filepath.Join(s.servicesDir, "test-service")
	err = os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	err = s.manager.CreateService("test-plugin", "test-service", "")
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("already exists"))
}

func (s *ServiceManagerTestSuite) TestDestroyService_WhenServiceDoesNotExist() {
	err := s.manager.DestroyService("non-existent-service")

	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("not found"))
}

func (s *ServiceManagerTestSuite) TestDestroyService_WhenServiceExists() {
	// Create plugin directory
	pluginDir := filepath.Join(s.pluginsDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	s.Require().NoError(err)

	// Create service directory first
	serviceDir := filepath.Join(s.servicesDir, "test-service")
	err = os.MkdirAll(serviceDir, 0755)
	s.Require().NoError(err)

	// Create service
	service := Service{
		Name:       "test-service",
		Plugin:     "test-plugin",
		Running:    false,
		LinkedApps: []string{},
		Config:     make(map[string]string),
	}

	err = s.manager.saveServiceConfig("test-service", service)
	s.Require().NoError(err)

	// Verify service exists
	Expect(serviceDir).To(BeADirectory())

	// Destroy service
	err = s.manager.DestroyService("test-service")
	Expect(err).To(BeNil())

	// Verify service directory is removed
	_, err = os.Stat(serviceDir)
	Expect(os.IsNotExist(err)).To(BeTrue())
}

func (s *ServiceManagerTestSuite) TestServiceJSONSerialization() {
	service := Service{
		Name:        "test-service",
		Plugin:      "postgres",
		ContainerID: "container-123",
		Running:     true,
		LinkedApps:  []string{"app1", "app2"},
		CreatedAt:   "2024-01-01T00:00:00Z",
		Config:      map[string]string{"version": "14"},
	}

	data, err := json.Marshal(service)

	Expect(err).To(BeNil())
	Expect(data).ToNot(BeEmpty())

	var unmarshaled Service
	err = json.Unmarshal(data, &unmarshaled)

	Expect(err).To(BeNil())

	Expect(unmarshaled.Name).To(Equal(service.Name))
	Expect(unmarshaled.Plugin).To(Equal(service.Plugin))
	Expect(unmarshaled.Running).To(Equal(service.Running))
	Expect(unmarshaled.LinkedApps).To(Equal(service.LinkedApps))
	Expect(unmarshaled.Config).To(Equal(service.Config))
}
