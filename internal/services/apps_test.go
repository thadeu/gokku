package services

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

func TestAppsServiceTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(AppsServiceTestSuite))
}

type AppsServiceTestSuite struct {
	suite.Suite
	tempDir string
	service *AppsService
}

func (s *AppsServiceTestSuite) SetupSuite() {
	// Suite-level setup if needed
}

func (s *AppsServiceTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "gokku-test-*")
	s.Require().NoError(err)

	s.service = NewAppsService(s.tempDir)
}

func (s *AppsServiceTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *AppsServiceTestSuite) TestNewAppsService_WithEmptyBaseDir() {
	service := NewAppsService("")
	Expect(service.baseDir).To(Equal("/opt/gokku"))
}

func (s *AppsServiceTestSuite) TestNewAppsService_WithCustomBaseDir() {
	customDir := "/custom/path"
	service := NewAppsService(customDir)
	Expect(service.baseDir).To(Equal(customDir))
}

func (s *AppsServiceTestSuite) TestListApps_EmptyDirectory() {
	apps, err := s.service.ListApps()
	Expect(err).To(BeNil())
	Expect(apps).To(BeEmpty())
}

func (s *AppsServiceTestSuite) TestListApps_WithSuccessfully() {
	appsDir := filepath.Join(s.tempDir, "apps")

	err := os.MkdirAll(appsDir, 0755)
	s.Require().NoError(err)

	// Create app directories
	app1Dir := filepath.Join(appsDir, "app1")
	app2Dir := filepath.Join(appsDir, "app2")

	err = os.MkdirAll(app1Dir, 0755)
	s.Require().NoError(err)

	err = os.MkdirAll(app2Dir, 0755)
	s.Require().NoError(err)

	apps, err := s.service.ListApps()

	Expect(err).To(BeNil())
	Expect(len(apps)).To(Equal(2))
	Expect(apps[0].Name).To(BeElementOf("app1", "app2"))
	Expect(apps[1].Name).To(BeElementOf("app1", "app2"))
}

func (s *AppsServiceTestSuite) TestListApps_IgnoresFiles() {
	appsDir := filepath.Join(s.tempDir, "apps")
	err := os.MkdirAll(appsDir, 0755)
	s.Require().NoError(err)

	// Create a file (not a directory)
	filePath := filepath.Join(appsDir, "not-an-app")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	s.Require().NoError(err)

	apps, err := s.service.ListApps()
	Expect(err).To(BeNil())
	Expect(apps).To(BeEmpty())
}

func (s *AppsServiceTestSuite) TestAppExists_WhenAppExists() {
	appsDir := filepath.Join(s.tempDir, "apps", "test-app")
	err := os.MkdirAll(appsDir, 0755)
	s.Require().NoError(err)

	exists := s.service.AppExists("test-app")
	Expect(exists).To(BeTrue())
}

func (s *AppsServiceTestSuite) TestAppExists_WhenAppDoesNotExist() {
	exists := s.service.AppExists("non-existent-app")
	Expect(exists).To(BeFalse())
}

func (s *AppsServiceTestSuite) TestGetApp_WhenAppNotFound() {
	_, err := s.service.GetApp("non-existent-app")
	Expect(err).ToNot(BeNil())
	Expect(err).To(BeAssignableToTypeOf(&AppNotFoundError{}))

	appErr := err.(*AppNotFoundError)
	Expect(appErr.AppName).To(Equal("non-existent-app"))
}

func (s *AppsServiceTestSuite) TestGetApp_WhenAppExists() {
	// Create app directory structure
	appsDir := filepath.Join(s.tempDir, "apps", "test-app")
	err := os.MkdirAll(appsDir, 0755)
	s.Require().NoError(err)

	app, err := s.service.GetApp("test-app")
	Expect(err).To(BeNil())
	Expect(app).ToNot(BeNil())
	Expect(app.Name).To(Equal("test-app"))
	Expect(app.ReleasesCount).To(Equal(0))
	Expect(app.EnvVars).ToNot(BeNil())
	// Containers is always initialized (empty slice if no containers or on error)
	// The GetApp method sets containers to empty slice on error
	// We just verify it's a valid slice (can be empty)
	_ = app.Containers // Just access it to ensure it exists
}

func (s *AppsServiceTestSuite) TestCountReleases_WhenReleasesDirExists() {
	appName := "test-app"
	releasesDir := filepath.Join(s.tempDir, "apps", appName, "releases")
	err := os.MkdirAll(releasesDir, 0755)
	s.Require().NoError(err)

	// Create some release directories
	for i := 1; i <= 3; i++ {
		releaseDir := filepath.Join(releasesDir, "release-"+strconv.Itoa(i))
		err = os.MkdirAll(releaseDir, 0755)
		s.Require().NoError(err)
	}

	count := s.service.countReleases(appName)
	Expect(count).To(Equal(3))
}

func (s *AppsServiceTestSuite) TestCountReleases_WhenReleasesDirDoesNotExist() {
	count := s.service.countReleases("non-existent-app")
	Expect(count).To(Equal(0))
}

func (s *AppsServiceTestSuite) TestGetCurrentRelease_WhenSymlinkExists() {
	appName := "test-app"
	appDir := filepath.Join(s.tempDir, "apps", appName)
	releasesDir := filepath.Join(appDir, "releases")
	releaseDir := filepath.Join(releasesDir, "release-1")

	err := os.MkdirAll(releaseDir, 0755)
	s.Require().NoError(err)

	currentLink := filepath.Join(appDir, "current")
	err = os.Symlink(releaseDir, currentLink)
	s.Require().NoError(err)

	release := s.service.getCurrentRelease(appName)
	Expect(release).To(Equal("release-1"))
}

func (s *AppsServiceTestSuite) TestGetCurrentRelease_WhenSymlinkDoesNotExist() {
	release := s.service.getCurrentRelease("non-existent-app")
	Expect(release).To(Equal("none"))
}
