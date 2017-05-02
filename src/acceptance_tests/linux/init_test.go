package linux_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/dns-release/src/acceptance_tests/gbosh"
	"github.com/onsi/gomega/gexec"
	"path/filepath"
	"testing"
)

func TestLinux(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "acceptance/linux")
}

var (
	boshDeployment gbosh.Deployment
)

var _ = BeforeSuite(func() {
	director := gbosh.NewDirectorFromEnv()
	boshDeployment = director.NewDeployment()

	dnsReleasePath, _ := filepath.Abs("../../../")
	aliasProvidingPath, _ := filepath.Abs("../dns-acceptance-release")

	boshDeployment.ExecuteDeploy(
		"../../../ci/assets/manifest.yml",
		[]string{
			"../../../ci/assets/two-instances-no-static-ips.yml",
			"../../../ci/assets/use-dns-release-default-bind-and-alias-addresses.yml",
			"../../../ci/assets/configure-recursor.yml",
		},
		map[string]string{
			"dns_release_path":        dnsReleasePath,
			"acceptance_release_path": aliasProvidingPath,
			"recursor_ip":             "172.17.0.1:9955",
		},
	)
})

var _ = AfterSuite(func() {
	boshDeployment.ExecuteDelete()
	gexec.CleanupBuildArtifacts()
})
