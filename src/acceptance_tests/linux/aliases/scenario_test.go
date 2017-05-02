// +build linux darwin

package aliases

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"time"
)

var _ = Describe("aliases", func() {
	Context("custom alias endpoint", func() {
		It("aliases the request", func() {
			session := boshDeployment.StartSSH("dns/0", "-c", "dig +time=3 +tries=1 -t A healthiness.example.com.")

			Eventually(session, 10*time.Second).Should(gexec.Exit())

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring(";; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0"))
			Expect(output).To(MatchRegexp("healthcheck\\.bosh-dns\\.\\s+0\\s+IN\\s+A\\s+127\\.0\\.0\\.1"))
		})
	})
})
