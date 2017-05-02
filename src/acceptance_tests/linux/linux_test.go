package linux_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"time"
)

var _ = Describe("Alias address binding", func() {
	It("should start a dns server on port 53", func() {
		session := boshDeployment.StartSSH("dns/0", "-c", "sudo lsof -n -i :53")

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Expect(output).To(MatchRegexp("dns.*TCP .*:domain"))
		Expect(output).To(MatchRegexp("dns.*UDP .*:domain"))
	})

	It("should respond to tcp dns queries", func() {
		session := boshDeployment.StartSSH("dns/0", "-c", "dig +tcp healthcheck.bosh-dns. @169.254.0.2")

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("Got answer:"))
		Eventually(session.Out).Should(gbytes.Say("flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0"))
		Eventually(session.Out).Should(gbytes.Say("healthcheck\\.bosh-dns\\.\\s+0\\s+IN\\s+A\\s+127\\.0\\.0\\.1"))
		Eventually(session.Out).Should(gbytes.Say("SERVER: 169.254.0.2#53"))
	})

	It("should respond to udp dns queries", func() {
		session := boshDeployment.StartSSH("dns/0", "-c", "dig +notcp healthcheck.bosh-dns. @169.254.0.2")

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("Got answer:"))
		Eventually(session.Out).Should(gbytes.Say("flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0"))
		Eventually(session.Out).Should(gbytes.Say(";healthcheck\\.bosh-dns\\.\\s+IN\\s+A"))
		Eventually(session.Out).Should(gbytes.Say("SERVER: 169.254.0.2#53"))
	})

	Context("as the system-configured nameserver", func() {
		It("resolves the bosh-dns healthcheck", func() {
			session := boshDeployment.StartSSH("dns/0", "-c", "dig -t A healthcheck.bosh-dns.")

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring(";; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0"))
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
					session = boshDeployment.StartSSH("dns/0", "-c", "dig +time=3 +tries=1 -t A healthcheck.bosh-dns.")
					Eventually(session, 10*time.Second).Should(gexec.Exit())

					return session
				}, 20*time.Second, time.Second*2).Should(gexec.Exit(0))

				output := string(session.Out.Contents())
				Expect(output).To(ContainSubstring(";; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0"))
				Expect(output).To(MatchRegexp("healthcheck\\.bosh-dns\\.\\s+0\\s+IN\\s+A\\s+127\\.0\\.0\\.1"))
			})
		})
	})
})
