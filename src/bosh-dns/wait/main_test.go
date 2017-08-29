package main_test

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("wait", func() {
	var (
		listenPort int
		cmd        string
	)

	BeforeEach(func() {
		var err error
		listenPort, err = getFreePort()
		Expect(err).NotTo(HaveOccurred())
		forwardedPort, err = getFreePort()
		Expect(err).NotTo(HaveOccurred())

		configContents, err := json.Marshal(map[string]interface{}{
			"address": "127.0.0.1",
			"port":    listenPort,
			// "records_file":     recordsFilePath,
			// "alias_files_glob": path.Join(aliasesDir, "*"),
			"upcheck_domains": []string{"health.check.bosh.", "health.check.ca."},
			"health": map[string]interface{}{
				"enabled":          true,
				"port":             2345 + config.GinkgoConfig.ParallelNode,
				"ca_file":          "../healthcheck/assets/test_certs/test_ca.pem",
				"certificate_file": "../healthcheck/assets/test_certs/test_client.pem",
				"private_key_file": "../healthcheck/assets/test_certs/test_client.key",
				"check_interval":   "1s",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		cmd = newCommandWithConfig(string(configContents))
	})

	It("passes when the check passes", func() {
		command := exec.Command(pathToBinary, `--command=true`, `--timeout=5ms`, `--checkDomain=google.com.`, `--nameServer=8.8.8.8:53`)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})

	It("fails when the check fails", func() {
		command := exec.Command(pathToBinary, `--command=true`, `--timeout=5ms`, `--checkDomain=something.does-not-exist.`, `--nameServer=127.0.0.1:1234`)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 100*time.Millisecond).Should(gexec.Exit(1))
	})

	It("starts the command and checks", func() {
		command := exec.Command(pathToBinary, `--command=`+cmd, `--timeout=500ms`, `--checkDomain=health.check.bosh.`, `--nameServer=127.0.0.1:`+strconv.Itoa(listenPort))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})

	It("allows for multiple commands", func() {
		command := exec.Command(pathToBinary,
			`--command=`+portForwarder,
			`--command=`+cmd,
			`--timeout=500ms`,
			`--checkDomain=health.check.bosh.`,
			`--nameServer=127.0.0.1:`+strconv.Itoa(forwardedPort),
		)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})
})

func newCommandWithConfig(config string) string {
	configFile, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())

	_, err = configFile.Write([]byte(config))

	Expect(err).NotTo(HaveOccurred())

	args := []string{
		"--config",
		configFile.Name(),
	}

	return strings.Join(append([]string{pathToServer}, args...), " ")
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}

	intPort, err := strconv.Atoi(port)
	if err != nil {
		return 0, err
	}

	return intPort, nil
}
