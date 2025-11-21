package services

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestContainerServiceTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(ContainerServiceTestSuite))
}

type ContainerServiceTestSuite struct {
	suite.Suite
	service *ContainerService
}

func (s *ContainerServiceTestSuite) SetupSuite() {
	// Suite-level setup if needed
}

func (s *ContainerServiceTestSuite) SetupTest() {
	s.service = NewContainerService("")
}

func (s *ContainerServiceTestSuite) TestNewContainerService_WithEmptyBaseDir() {
	service := NewContainerService("")
	Expect(service.baseDir).To(Equal("/opt/gokku"))
}

func (s *ContainerServiceTestSuite) TestNewContainerService_WithCustomBaseDir() {
	customDir := "/custom/path"
	service := NewContainerService(customDir)
	Expect(service.baseDir).To(Equal(customDir))
}

func (s *ContainerServiceTestSuite) TestContainsContainerName_WhenNameMatches() {
	names := "my-app-web-1"
	appName := "my-app"
	result := containsContainerName(names, appName)
	Expect(result).To(BeTrue())
}

func (s *ContainerServiceTestSuite) TestContainsContainerName_WhenNameDoesNotMatch() {
	names := "other-app-web-1"
	appName := "my-app"
	result := containsContainerName(names, appName)
	Expect(result).To(BeFalse())
}

func (s *ContainerServiceTestSuite) TestContainsContainerName_WhenNameIsSubstring() {
	names := "my-app-production-web-1"
	appName := "my-app"
	result := containsContainerName(names, appName)
	Expect(result).To(BeTrue())
}

func (s *ContainerServiceTestSuite) TestContainsProcessType_WhenTypeMatches() {
	names := "my-app-web-1"
	processType := "web"
	result := containsProcessType(names, processType)
	Expect(result).To(BeTrue())
}

func (s *ContainerServiceTestSuite) TestContainsProcessType_WhenTypeDoesNotMatch() {
	names := "my-app-worker-1"
	processType := "web"
	result := containsProcessType(names, processType)
	Expect(result).To(BeFalse())
}

func (s *ContainerServiceTestSuite) TestContainsProcessType_WhenTypeIsSubstring() {
	names := "my-app-web-production-1"
	processType := "web"
	result := containsProcessType(names, processType)
	Expect(result).To(BeTrue())
}

func (s *ContainerServiceTestSuite) TestListContainers_WithEmptyFilter() {
	filter := ContainerFilter{}
	_, err := s.service.ListContainers(filter)

	// This will depend on actual Docker state, so we just check no error
	// In a real scenario, you might want to mock this
	Expect(err).To(BeNil())
}

func (s *ContainerServiceTestSuite) TestListContainers_WithAppNameFilter() {
	filter := ContainerFilter{
		AppName: "test-app",
	}
	_, err := s.service.ListContainers(filter)

	// This will depend on actual Docker state
	Expect(err).To(BeNil())
}

func (s *ContainerServiceTestSuite) TestListContainers_WithProcessTypeFilter() {
	filter := ContainerFilter{
		ProcessType: "web",
	}
	_, err := s.service.ListContainers(filter)

	// This will depend on actual Docker state
	Expect(err).To(BeNil())
}

func (s *ContainerServiceTestSuite) TestListContainers_WithCombinedFilter() {
	filter := ContainerFilter{
		AppName:     "test-app",
		ProcessType: "web",
	}
	_, err := s.service.ListContainers(filter)

	// This will depend on actual Docker state
	Expect(err).To(BeNil())
}

func (s *ContainerServiceTestSuite) TestListContainers_WithAllFlag() {
	filter := ContainerFilter{
		All: true,
	}
	_, err := s.service.ListContainers(filter)

	// This will depend on actual Docker state
	Expect(err).To(BeNil())
}

func (s *ContainerServiceTestSuite) TestGetContainerInfo_WhenContainerNotFound() {
	_, err := s.service.GetContainerInfo("non-existent-container-12345")

	Expect(err).ToNot(BeNil())
	Expect(err).To(BeAssignableToTypeOf(&ContainerNotFoundError{}))

	containerErr := err.(*ContainerNotFoundError)
	Expect(containerErr.ContainerName).To(Equal("non-existent-container-12345"))
}

func (s *ContainerServiceTestSuite) TestStartContainer_WhenContainerDoesNotExist() {
	err := s.service.StartContainer("non-existent-container-12345")

	Expect(err).ToNot(BeNil())
	Expect(err).To(BeAssignableToTypeOf(&ContainerNotFoundError{}))
}
