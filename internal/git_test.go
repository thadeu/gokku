package internal

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type MockGitClient struct {
	mockedGetRemoteURL func(remoteName string) (string, error)
	mockedAddRemote    func(remoteName string, remoteURL string) (string, error)
	mockedRemoveRemote func(remoteName string) (string, error)
}

func (m *MockGitClient) AddRemote(remoteName string, remoteURL string) (string, error) {
	if m.mockedAddRemote != nil {
		return m.mockedAddRemote(remoteName, remoteURL)
	}

	return "", nil
}

func (m *MockGitClient) RemoveRemote(remoteName string) (string, error) {
	if m.mockedRemoveRemote != nil {
		return m.mockedRemoveRemote(remoteName)
	}

	return "", nil
}

func (m *MockGitClient) GetRemoteURL(remoteName string) (string, error) {
	if m.mockedGetRemoteURL != nil {
		return m.mockedGetRemoteURL(remoteName)
	}

	return "", nil
}

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(ConfigTestSuite))
}

func (c *ConfigTestSuite) TestGetRemoteInfo_WhenGitRemoteIsFound() {
	mockClient := &MockGitClient{
		mockedGetRemoteURL: func(remoteName string) (string, error) {
			return "user@server:api", nil
		},
	}

	remoteInfo, err := GetRemoteInfoWithClient(mockClient, "api")

	Expect(err).To(BeNil())

	Expect(remoteInfo).ToNot(BeNil())
	Expect(remoteInfo.Host).To(Equal("user@server"))
	Expect(remoteInfo.BaseDir).To(Equal("api"))
	Expect(remoteInfo.App).To(Equal("api"))
}

func (c *ConfigTestSuite) TestGetRemoteInfo_WhenGitRemoteNotFound() {
	mockClient := &MockGitClient{
		mockedGetRemoteURL: func(remoteName string) (string, error) {
			return "", errors.New("fatal: No such remote")
		},
	}

	remoteInfo, err := GetRemoteInfoWithClient(mockClient, "nonexistent")

	Expect(err).ToNot(BeNil())
	Expect(remoteInfo).To(BeNil())
	Expect(err.Error()).To(ContainSubstring("git remote 'nonexistent' not found"))
}

func (c *ConfigTestSuite) TestGetRemoteInfo_WithFullPath() {
	// Test with full path format: user@host:/opt/gokku/repos/api.git
	mockClient := &MockGitClient{
		mockedGetRemoteURL: func(remoteName string) (string, error) {
			return "user@server:/opt/gokku/repos/api.git", nil
		},
	}

	remoteInfo, err := GetRemoteInfoWithClient(mockClient, "api")

	Expect(err).To(BeNil())

	Expect(remoteInfo).ToNot(BeNil())
	Expect(remoteInfo.Host).To(Equal("user@server"))
	Expect(remoteInfo.BaseDir).To(Equal("/opt/gokku"))
	Expect(remoteInfo.App).To(Equal("api"))
}
