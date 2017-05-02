package override_nameserver

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"time"
)

var _ = Describe("dns job: override_nameserver", func() {
	Describe("disabled", func() {
		Context("as the system-configured nameserver", func() {
			It("resolves the bosh-dns healthcheck", func() {
				session := boshDeployment.StartSSH("dns/0", "-c", "dig +time=3 +tries=1 -t A healthcheck.bosh-dns.")

				Eventually(session, 10*time.Second).Should(gexec.Exit())
				output := string(session.Out.Contents())
				Expect(output).To(ContainSubstring("status: NXDOMAIN"))
			})

			Context("external processes changing /etc/resolv.conf", func() {
				BeforeEach(func() {
					session := boshDeployment.StartSSH("dns/0", "-c", "sudo cp /etc/resolv.conf /tmp/resolv.conf.backup")
					Eventually(session, 10*time.Second).Should(gexec.Exit(0))
				})

				AfterEach(func() {
					session := boshDeployment.StartSSH("dns/0", "-c", "sudo mv /tmp/resolv.conf.backup /etc/resolv.conf")
					Eventually(session, 10*time.Second).Should(gexec.Exit(0))
				})

				It("rewrites the nameserver configuration back to our dns server", func() {
					session := boshDeployment.StartSSH("dns/0", "-c", "echo 'nameserver 192.0.2.100' | sudo tee /etc/resolv.conf > /dev/null")
					Eventually(session, 10*time.Second).Should(gexec.Exit(0))

					Eventually(func() *gexec.Session {
						session := boshDeployment.StartSSH("dns/0", "-c", "dig +time=3 +tries=1 -t A healthcheck.bosh-dns.")
						Eventually(session, 10*time.Second).Should(gexec.Exit())

						return session
					}, 15*time.Second, time.Second*2).ShouldNot(gexec.Exit(0))
				})
			})
		})
	})
})
