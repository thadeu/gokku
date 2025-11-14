package services

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestConfigServiceTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(ConfigServiceTestSuite))
}

type ConfigServiceTestSuite struct {
	suite.Suite
	tempDir string
	service *ConfigService
	appName string
}

func (s *ConfigServiceTestSuite) SetupSuite() {
	// Suite-level setup if needed
}

func (s *ConfigServiceTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "gokku-test-*")
	s.Require().NoError(err)

	s.appName = "test-app"
	s.service = NewConfigService(s.tempDir)

	// Create app directory structure
	appDir := filepath.Join(s.tempDir, "apps", s.appName, "shared")
	err = os.MkdirAll(appDir, 0755)

	s.Require().NoError(err)
}

func (s *ConfigServiceTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *ConfigServiceTestSuite) TestNewConfigService_WithEmptyBaseDir() {
	service := NewConfigService("")

	Expect(service.baseDir).To(Equal("/opt/gokku"))
}

func (s *ConfigServiceTestSuite) TestNewConfigService_WithCustomBaseDir() {
	customDir := "/custom/path"
	service := NewConfigService(customDir)

	Expect(service.baseDir).To(Equal(customDir))
}

func (s *ConfigServiceTestSuite) TestSetEnvVar_SingleVariable() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars["KEY1"]).To(Equal("value1"))
}

func (s *ConfigServiceTestSuite) TestSetEnvVar_MultipleVariables() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1", "KEY2=value2", "KEY3=value3"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars["KEY1"]).To(Equal("value1"))
	Expect(envVars["KEY2"]).To(Equal("value2"))
	Expect(envVars["KEY3"]).To(Equal("value3"))
}

func (s *ConfigServiceTestSuite) TestSetEnvVar_OverwritesExisting() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1"})
	Expect(err).To(BeNil())

	err = s.service.SetEnvVar(s.appName, []string{"KEY1=value2"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars["KEY1"]).To(Equal("value2"))
}

func (s *ConfigServiceTestSuite) TestSetEnvVar_WithSpaces() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value with spaces"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars["KEY1"]).To(Equal("value with spaces"))
}

func (s *ConfigServiceTestSuite) TestSetEnvVar_TrimsWhitespace() {
	err := s.service.SetEnvVar(s.appName, []string{"  KEY1  =  value1  "})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars["KEY1"]).To(Equal("value1"))
}

func (s *ConfigServiceTestSuite) TestSetEnvVar_InvalidFormat() {
	err := s.service.SetEnvVar(s.appName, []string{"INVALID_FORMAT"})
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("invalid format"))
}

func (s *ConfigServiceTestSuite) TestGetEnvVar_WhenVariableExists() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1"})
	s.Require().NoError(err)

	value, err := s.service.GetEnvVar(s.appName, "KEY1")
	Expect(err).To(BeNil())
	Expect(value).To(Equal("value1"))
}

func (s *ConfigServiceTestSuite) TestGetEnvVar_WhenVariableDoesNotExist() {
	_, err := s.service.GetEnvVar(s.appName, "NON_EXISTENT")
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("not found"))
}

func (s *ConfigServiceTestSuite) TestListEnvVars_Empty() {
	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars).ToNot(BeNil())
	Expect(envVars).To(BeEmpty())
}

func (s *ConfigServiceTestSuite) TestListEnvVars_WithVariables() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1", "KEY2=value2"})
	s.Require().NoError(err)

	envVars := s.service.ListEnvVars(s.appName)
	Expect(len(envVars)).To(Equal(2))
	Expect(envVars["KEY1"]).To(Equal("value1"))
	Expect(envVars["KEY2"]).To(Equal("value2"))
}

func (s *ConfigServiceTestSuite) TestUnsetEnvVar_SingleVariable() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1", "KEY2=value2"})
	s.Require().NoError(err)

	err = s.service.UnsetEnvVar(s.appName, []string{"KEY1"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars).ToNot(HaveKey("KEY1"))
	Expect(envVars["KEY2"]).To(Equal("value2"))
}

func (s *ConfigServiceTestSuite) TestUnsetEnvVar_MultipleVariables() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1", "KEY2=value2", "KEY3=value3"})
	s.Require().NoError(err)

	err = s.service.UnsetEnvVar(s.appName, []string{"KEY1", "KEY3"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars).ToNot(HaveKey("KEY1"))
	Expect(envVars).ToNot(HaveKey("KEY3"))
	Expect(envVars["KEY2"]).To(Equal("value2"))
}

func (s *ConfigServiceTestSuite) TestUnsetEnvVar_NonExistentVariable() {
	err := s.service.SetEnvVar(s.appName, []string{"KEY1=value1"})
	s.Require().NoError(err)

	// Unsetting a non-existent variable should not error
	err = s.service.UnsetEnvVar(s.appName, []string{"NON_EXISTENT"})
	Expect(err).To(BeNil())

	envVars := s.service.ListEnvVars(s.appName)
	Expect(envVars["KEY1"]).To(Equal("value1"))
}

func (s *ConfigServiceTestSuite) TestGetEnvFilePath() {
	expectedPath := filepath.Join(s.tempDir, "apps", s.appName, "shared", ".env")
	actualPath := s.service.getEnvFilePath(s.appName)

	Expect(actualPath).To(Equal(expectedPath))
}
