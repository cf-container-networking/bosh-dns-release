package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"bosh-dns/healthcheck/healthserver"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHealthCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "healthcheck")
}

var (
	pathToServer         string
	sess                 *gexec.Session
	cmd                  *exec.Cmd
	healthFile           *os.File
	configFile           *os.File
	healthExecutableFile *os.File
	configPort           int
)

var _ = SynchronizedBeforeSuite(func() []byte {
	serverPath, err := gexec.Build("bosh-dns/healthcheck")
	Expect(err).NotTo(HaveOccurred())
	SetDefaultEventuallyTimeout(2 * time.Second)

	return []byte(serverPath)
}, func(data []byte) {
	pathToServer = string(data)

	var err error

	configFile, err = ioutil.TempFile("", "config.json")
	Expect(err).ToNot(HaveOccurred())

	tempDir := os.TempDir()
	healthFile, err = ioutil.TempFile("", "health.json")
	Expect(err).ToNot(HaveOccurred())

	healthExecutableFile, err = ioutil.TempFile(tempDir, "health_executable")
	Expect(err).ToNot(HaveOccurred())

	err = os.Chmod(healthExecutableFile.Name(), 0700)
	Expect(err).ToNot(HaveOccurred())

	configPort = 1234 + config.GinkgoConfig.ParallelNode

	configContents, err := json.Marshal(healthserver.HealthCheckConfig{
		Port:                  configPort,
		CertificateFile:       "assets/test_certs/test_server.pem",
		PrivateKeyFile:        "assets/test_certs/test_server.key",
		CAFile:                "assets/test_certs/test_ca.pem",
		HealthFileName:        healthFile.Name(),
		HealthExecutablesGlob: filepath.Join(tempDir, "*"),
	})
	Expect(err).NotTo(HaveOccurred())

	err = ioutil.WriteFile(configFile.Name(), []byte(configContents), 0666)
	Expect(err).ToNot(HaveOccurred())

	// run the server
	cmd = exec.Command(pathToServer, configFile.Name())
	sess, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	Expect(waitForServer(configPort)).To(Succeed())
})

var _ = BeforeEach(func() {
	err := ioutil.WriteFile(healthExecutableFile.Name(), []byte("#!/bin/bash"), 0700)
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
	if cmd.Process != nil {
		Eventually(sess.Kill()).Should(gexec.Exit())
	}
}, func() {
	gexec.CleanupBuildArtifacts()
})
