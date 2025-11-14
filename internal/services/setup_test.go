package services

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestServerSetupTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(ServerSetupTestSuite))
}

type ServerSetupTestSuite struct {
	suite.Suite
}

func (s *ServerSetupTestSuite) TestNewServerSetup_WithHostOnly() {
	setup := NewServerSetup("example.com", "")
	Expect(setup).ToNot(BeNil())
	Expect(setup.config.ServerHost).To(Equal("example.com"))
	Expect(setup.config.IdentityFile).To(BeEmpty())
}

func (s *ServerSetupTestSuite) TestNewServerSetup_WithHostAndIdentityFile() {
	identityFile := "/path/to/key.pem"
	setup := NewServerSetup("example.com", identityFile)
	Expect(setup).ToNot(BeNil())
	Expect(setup.config.ServerHost).To(Equal("example.com"))
	Expect(setup.config.IdentityFile).To(Equal(identityFile))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_WithoutIdentityFile() {
	setup := NewServerSetup("example.com", "")
	args := setup.buildSSHArgs("echo", "test")

	Expect(args).To(ContainElement("example.com"))
	Expect(args).To(ContainElement("echo"))
	Expect(args).To(ContainElement("test"))
	Expect(args).ToNot(ContainElement("-i"))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_WithIdentityFile() {
	// Create a temporary identity file
	tempFile, err := os.CreateTemp("", "test-key-*.pem")
	s.Require().NoError(err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	setup := NewServerSetup("example.com", tempFile.Name())
	args := setup.buildSSHArgs("echo", "test")

	Expect(args).To(ContainElement("example.com"))
	Expect(args).To(ContainElement("-i"))
	Expect(args).To(ContainElement(tempFile.Name()))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_WithIdentityFile_WhenFileDoesNotExist() {
	setup := NewServerSetup("example.com", "/non/existent/key.pem")
	args := setup.buildSSHArgs("echo", "test")

	// Should not include -i if file doesn't exist
	Expect(args).ToNot(ContainElement("-i"))
	Expect(args).ToNot(ContainElement("/non/existent/key.pem"))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_WithSSHOptions() {
	setup := NewServerSetup("example.com", "")
	args := setup.buildSSHArgs("-o", "BatchMode=yes", "-o", "ConnectTimeout=5", "echo", "OK")

	Expect(args).To(ContainElement("-o"))
	Expect(args).To(ContainElement("BatchMode=yes"))
	Expect(args).To(ContainElement("ConnectTimeout=5"))
	Expect(args).To(ContainElement("echo"))
	Expect(args).To(ContainElement("OK"))
	Expect(args).To(ContainElement("example.com"))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_WithTFlag() {
	setup := NewServerSetup("example.com", "")
	args := setup.buildSSHArgs("-t", "bash", "-c", "command")

	Expect(args).To(ContainElement("-t"))
	Expect(args).To(ContainElement("bash"))
	Expect(args).To(ContainElement("-c"))
	Expect(args).To(ContainElement("command"))
	Expect(args).To(ContainElement("example.com"))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_ComplexCommand() {
	setup := NewServerSetup("example.com", "")
	args := setup.buildSSHArgs("-o", "StrictHostKeyChecking=no", "docker", "ps")

	Expect(args).To(ContainElement("-o"))
	Expect(args).To(ContainElement("StrictHostKeyChecking=no"))
	Expect(args).To(ContainElement("docker"))
	Expect(args).To(ContainElement("ps"))
	Expect(args).To(ContainElement("example.com"))
}

func (s *ServerSetupTestSuite) TestBuildSSHArgs_Order() {
	setup := NewServerSetup("example.com", "")
	args := setup.buildSSHArgs("command")

	// Host should be before command
	hostIndex := -1
	commandIndex := -1
	for i, arg := range args {
		if arg == "example.com" {
			hostIndex = i
		}
		if arg == "command" {
			commandIndex = i
		}
	}

	Expect(hostIndex).To(BeNumerically(">=", 0))
	Expect(commandIndex).To(BeNumerically(">=", 0))
	Expect(hostIndex).To(BeNumerically("<", commandIndex))
}

func (s *ServerSetupTestSuite) TestSetupConfig_Structure() {
	config := SetupConfig{
		ServerHost:   "example.com",
		IdentityFile: "/path/to/key.pem",
	}

	Expect(config.ServerHost).To(Equal("example.com"))
	Expect(config.IdentityFile).To(Equal("/path/to/key.pem"))
}
