package services

import (
	"gokku/internal"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestCommonTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(CommonTestSuite))
}

type CommonTestSuite struct {
	suite.Suite
}

func (s *CommonTestSuite) SetupSuite() {
	// Suite-level setup if needed
}

func (s *CommonTestSuite) SetupTest() {
	// Test-level setup if needed
}

func (s *CommonTestSuite) TestAppNotFoundError_Error() {
	err := &AppNotFoundError{AppName: "test-app"}
	errorMsg := err.Error()
	Expect(errorMsg).To(ContainSubstring("app not found"))
	Expect(errorMsg).To(ContainSubstring("test-app"))
}

func (s *CommonTestSuite) TestAppNotFoundError_Structure() {
	err := &AppNotFoundError{AppName: "my-app"}
	Expect(err.AppName).To(Equal("my-app"))
}

func (s *CommonTestSuite) TestContainerNotFoundError_Error() {
	err := &ContainerNotFoundError{ContainerName: "test-container"}
	errorMsg := err.Error()
	Expect(errorMsg).To(ContainSubstring("container not found"))
	Expect(errorMsg).To(ContainSubstring("test-container"))
}

func (s *CommonTestSuite) TestContainerNotFoundError_Structure() {
	err := &ContainerNotFoundError{ContainerName: "my-container"}
	Expect(err.ContainerName).To(Equal("my-container"))
}

func (s *CommonTestSuite) TestAppInfo_Structure() {
	appInfo := AppInfo{
		Name:           "test-app",
		Status:         "running",
		ReleasesCount:  5,
		CurrentRelease: "release-3",
	}

	Expect(appInfo.Name).To(Equal("test-app"))
	Expect(appInfo.Status).To(Equal("running"))
	Expect(appInfo.ReleasesCount).To(Equal(5))
	Expect(appInfo.CurrentRelease).To(Equal("release-3"))
}

func (s *CommonTestSuite) TestAppDetail_Structure() {
	appInfo := AppInfo{
		Name:           "test-app",
		Status:         "running",
		ReleasesCount:  2,
		CurrentRelease: "release-1",
	}

	appDetail := AppDetail{
		AppInfo:    appInfo,
		Config:     nil,
		Containers: []internal.ContainerInfo{},
		EnvVars:    make(map[string]string),
	}

	Expect(appDetail.Name).To(Equal("test-app"))
	Expect(appDetail.Status).To(Equal("running"))
	Expect(appDetail.ReleasesCount).To(Equal(2))
	Expect(appDetail.CurrentRelease).To(Equal("release-1"))
	Expect(appDetail.Config).To(BeNil())
	Expect(appDetail.Containers).ToNot(BeNil())
	Expect(appDetail.EnvVars).ToNot(BeNil())
}

func (s *CommonTestSuite) TestContainerFilter_Structure() {
	filter := ContainerFilter{
		AppName:     "test-app",
		ProcessType: "web",
		All:         true,
	}

	Expect(filter.AppName).To(Equal("test-app"))
	Expect(filter.ProcessType).To(Equal("web"))
	Expect(filter.All).To(BeTrue())
}

func (s *CommonTestSuite) TestContainerFilter_Empty() {
	filter := ContainerFilter{}

	Expect(filter.AppName).To(BeEmpty())
	Expect(filter.ProcessType).To(BeEmpty())
	Expect(filter.All).To(BeFalse())
}

func (s *CommonTestSuite) TestConfigOperation_Structure() {
	op := ConfigOperation{
		Type:  "set",
		Key:   "DATABASE_URL",
		Value: "postgres://localhost/db",
		Keys:  []string{},
	}

	Expect(op.Type).To(Equal("set"))
	Expect(op.Key).To(Equal("DATABASE_URL"))
	Expect(op.Value).To(Equal("postgres://localhost/db"))
	Expect(op.Keys).To(BeEmpty())
}

func (s *CommonTestSuite) TestConfigOperation_Unset() {
	op := ConfigOperation{
		Type:  "unset",
		Key:   "",
		Value: "",
		Keys:  []string{"KEY1", "KEY2", "KEY3"},
	}

	Expect(op.Type).To(Equal("unset"))
	Expect(op.Keys).To(HaveLen(3))
	Expect(op.Keys).To(ContainElement("KEY1"))
	Expect(op.Keys).To(ContainElement("KEY2"))
	Expect(op.Keys).To(ContainElement("KEY3"))
}
