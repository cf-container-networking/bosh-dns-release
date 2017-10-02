// +build windows

package windows_test

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("windows tests", func() {
	It("should bind to tcp and udp", func() {
		cmd := exec.Command("powershell.exe", "-Command", "netstat -na | findstr :53")

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("TCP    169.254.0.2:53"))
		Expect(session.Out.Contents()).To(ContainSubstring("UDP    169.254.0.2:53"))
	})

	It("should respond to dns queries", func() {
		cmd := exec.Command("powershell.exe", "-Command", "Resolve-DnsName -DnsOnly -Name upcheck.bosh-dns. -Server 169.254.0.2 | Format-list")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("IPAddress  : 127.0.0.1"))
		Expect(session.Out.Contents()).To(ContainSubstring("Name       : upcheck.bosh-dns"))
	})

	Context("as the system-configured nameserver", func() {
		It("should respond to dns queries", func() {
			cmd := exec.Command("powershell.exe", "-Command", "Resolve-DnsName -DnsOnly -Name upcheck.bosh-dns. | Format-list")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("IPAddress  : 127.0.0.1"))
			Expect(session.Out.Contents()).To(ContainSubstring("Name       : upcheck.bosh-dns"))
		})
	})

	Context("when enabling system level caching TAG:os-cache-disabled", func() {
		It("caches dns recursed dns entries for the duration of the TTL", func() {
			cmd := exec.Command("powershell.exe", "-Command", "Get-Service -Dnscache | FormatList")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			output := string(session.Out.Contents())
			Expect(output).To(MatchRegexp(`Status\w*:\w*Stopped`))
			Expect(output).To(MatchRegexp(`CanStop\w*:\w*False`))
		})
	})

	Context("when enabling system level caching TAG:os-cache-enabled", func() {
		It("caches dns recursed dns entries for the duration of the TTL", func() {
			cmd := exec.Command("powershell.exe", "-Command", "Get-Service -Dnscache | FormatList")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			output := string(session.Out.Contents())
			Expect(output).To(MatchRegexp(`Status\w*:\w*Running`))
			Expect(output).To(MatchRegexp(`CanStop\w*:\w*True`))
		})
	})
})
