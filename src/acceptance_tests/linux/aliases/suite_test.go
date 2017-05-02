// +build linux darwin

package aliases

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/dns-release/src/acceptance_tests/gbosh"
	"path/filepath"
	"testing"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "aliases")
}

var (
	boshDeployment gbosh.Deployment
)

var _ = BeforeSuite(func() {
	director := gbosh.NewDirectorFromEnv()
	boshDeployment = director.NewDeployment()

	dnsReleasePath, _ := filepath.Abs("../../../../")
	aliasProvidingPath, _ := filepath.Abs("../../dns-acceptance-release")

	boshDeployment.ExecuteDeploy(
		"../../../../ci/assets/manifest.yml",
		[]string{
			"../../../../ci/assets/use-dns-release-default-bind-and-alias-addresses.yml",
			"scenario.yml",
		},
		map[string]string{
			"dns_release_path":        dnsReleasePath,
			"acceptance_release_path": aliasProvidingPath,
		},
	)
})

var _ = AfterSuite(func() {
	boshDeployment.ExecuteDelete()
})
