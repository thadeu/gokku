package plugins

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type PluginManagerTestSuite struct {
	suite.Suite
	manager *PluginManager
}

func TestPluginManagerTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(PluginManagerTestSuite))
}

func (s *PluginManagerTestSuite) SetupTest() {
	s.manager = NewPluginManager()

	tempDir, err := os.MkdirTemp("", "gokku-test-*")
	s.Require().NoError(err)

	pluginsDir := filepath.Join(tempDir, "plugins")
	err = os.MkdirAll(pluginsDir, 0755)
	s.Require().NoError(err)

	s.manager.pluginsDir = pluginsDir
}

func (s *PluginManagerTestSuite) TestListPlugins_WhenNoPluginsAreNotInstalled() {
	plugins, err := s.manager.ListPlugins()

	Expect(err).To(BeNil())
	Expect(plugins).To(BeEmpty())
}

func (s *PluginManagerTestSuite) TestListPlugins_WhenPluginsAreInstalled() {
	plugin1 := filepath.Join(s.manager.pluginsDir, "nginx")
	err := os.MkdirAll(plugin1, 0755)
	s.Require().NoError(err)

	plugin2 := filepath.Join(s.manager.pluginsDir, "letsencrypt")
	err = os.MkdirAll(plugin2, 0755)
	s.Require().NoError(err)

	plugins, err := s.manager.ListPlugins()

	Expect(err).To(BeNil())
	Expect(plugins).ToNot(BeEmpty())
	Expect(plugins).To(ContainElement("nginx"))
	Expect(plugins).To(ContainElement("letsencrypt"))
}
