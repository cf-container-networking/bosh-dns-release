// +build linux darwin

package override_nameserver

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
	"github.com/cloudfoundry/dns-release/src/acceptance_tests/gbosh"
	"os"
	"path/filepath"
	"testing"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "override_nameserver/disabled")
}

var (
	boshDeployment gbosh.Deployment
)

var _ = BeforeSuite(func() {
	director := gbosh.NewDirectorFromEnv()
	boshDeployment = director.NewDeployment()

	dnsReleasePath, _ := filepath.Abs("../../../../../")
	aliasProvidingPath, _ := filepath.Abs("../../../dns-acceptance-release")

	boshDeployment.ExecuteDeploy(
		"../../../../../ci/assets/manifest.yml",
		[]string{
			"../../../../../ci/assets/use-dns-release-default-bind-and-alias-addresses.yml",
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
