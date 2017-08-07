package performance_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	metrics "github.com/rcrowley/go-metrics"
)

func TestPerformance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Tests")
}

var (
	healthSession *gexec.Session
	dnsSession    *gexec.Session
)

var _ = BeforeSuite(func() {
	// healthServerPath, err := gexec.Build("bosh-dns/healthcheck")
	// Expect(err).NotTo(HaveOccurred())

	dnsServerPath, err := gexec.Build("bosh-dns/dns")
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(2 * time.Second)

	// healthConfigFile, err := ioutil.TempFile("", "config.json")
	// Expect(err).ToNot(HaveOccurred())

	// healthFile, err := ioutil.TempFile("", "health.json")
	// Expect(err).ToNot(HaveOccurred())

	// healthPort := 8853

	// healthConfigContents, err := json.Marshal(healthserver.HealthCheckConfig{
	// 	Port:            healthPort,
	// 	CertificateFile: "../healthcheck/assets/test_certs/test_server.pem",
	// 	PrivateKeyFile:  "../healthcheck/assets/test_certs/test_server.key",
	// 	CAFile:          "../healthcheck/assets/test_certs/test_ca.pem",
	// 	HealthFileName:  healthFile.Name(),
	// })
	// Expect(err).NotTo(HaveOccurred())

	// err = ioutil.WriteFile(healthConfigFile.Name(), []byte(healthConfigContents), 0666)
	// Expect(err).ToNot(HaveOccurred())

	dnsConfigFile, err := ioutil.TempFile("", "config.json")
	Expect(err).ToNot(HaveOccurred())

	dnsPort := 9953

	dnsConfigContents, err := json.Marshal(map[string]interface{}{
		"address":          "127.0.0.1",
		"port":             dnsPort,
		"records_file":     "assets/records.json",
		"alias_files_glob": "assets/aliases.json",
		"upcheck_domains":  []string{"upcheck.bosh-dns."},
		"recursors":        []string{"8.8.8.8"},
		"recursor_timeout": "2s",
		"health": map[string]interface{}{
			"enabled": false,
			// "port":             healthPort,
			// "ca_file":          "../healthcheck/assets/test_certs/test_ca.pem",
			// "certificate_file": "../healthcheck/assets/test_certs/test_client.pem",
			// "private_key_file": "../healthcheck/assets/test_certs/test_client.key",
			// "check_interval":   "20s",
		},
	})
	Expect(err).NotTo(HaveOccurred())

	err = ioutil.WriteFile(dnsConfigFile.Name(), []byte(dnsConfigContents), 0666)
	Expect(err).ToNot(HaveOccurred())

	var cmd *exec.Cmd
	// cmd = exec.Command(healthServerPath, healthConfigFile.Name())
	// healthSession, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	// Expect(err).ToNot(HaveOccurred())

	// Expect(waitForServer(healthPort)).To(Succeed())

	cmd = exec.Command(dnsServerPath, "--config="+dnsConfigFile.Name())
	dnsSession, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	Expect(waitForServer(dnsPort)).To(Succeed())
})

func waitForServer(port int) error {
	var err error
	for i := 0; i < 20; i++ {
		var c net.Conn
		c, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", strconv.Itoa(port)))
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		_ = c.Close()
		return nil
	}

	return err //errors.New("dns server failed to start")
}

var _ = AfterSuite(func() {
	// if healthSession != nil && healthSession.Command.Process != nil {
	// 	Eventually(healthSession.Kill()).Should(gexec.Exit())
	// }

	if dnsSession != nil && dnsSession.Command.Process != nil {
		Eventually(dnsSession.Kill()).Should(gexec.Exit())
	}

	gexec.CleanupBuildArtifacts()
})

func assertEnvExists(envName string) string {
	val, found := os.LookupEnv(envName)
	if !found {
		Fail(fmt.Sprintf("Expected %s", envName))
	}
	return val
}

func GetPidFor(processName string) (int, bool) {
	pids := sigar.ProcList{}
	pids.Get()

	for _, pid := range pids.List {
		state := sigar.ProcState{}

		if err := state.Get(pid); err != nil {
			continue
		}

		if strings.Contains(state.Name, processName) {
			return pid, true
		}
	}

	return -1, false
}

func getProcessCPU(pid int) float64 {
	cmd := exec.Command("ps", []string{"-p", strconv.Itoa(pid), "-o", "%cpu"}...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(string(output) + err.Error())
	}

	percentString := strings.TrimSpace(strings.Split(string(output), "\n")[1])
	percent, err := strconv.ParseFloat(percentString, 64)
	Expect(err).ToNot(HaveOccurred())

	return percent
}

func setupWaitGroupWithSignaler(maxDnsRequests int) (*sync.WaitGroup, chan struct{}) {
	wg := &sync.WaitGroup{}
	wg.Add(maxDnsRequests)
	finishedDnsRequests := make(chan struct{})

	go func() {
		wg.Wait()
		close(finishedDnsRequests)
	}()

	return wg, finishedDnsRequests
}

func printStatsForHistogram(hist metrics.Histogram, label string, unit string, scalingDivisor float64) {
	fmt.Printf("\n~~~~~~~~~~~~~~~%s~~~~~~~~~~~~~~~\n", label)
	printStatNamed("Std Deviation", hist.StdDev()/scalingDivisor, unit)
	printStatNamed("Median", hist.Percentile(0.5)/scalingDivisor, unit)
	printStatNamed("Mean", hist.Mean()/scalingDivisor, unit)
	printStatNamed("Max", float64(hist.Max())/scalingDivisor, unit)
	printStatNamed("Min", float64(hist.Min())/scalingDivisor, unit)
	printStatNamed("90th Percentile", hist.Percentile(0.9)/scalingDivisor, unit)
	printStatNamed("95th Percentile", hist.Percentile(0.95)/scalingDivisor, unit)
	printStatNamed("99th Percentile", hist.Percentile(0.99)/scalingDivisor, unit)
}

func printStatNamed(label string, value float64, unit string) {
	fmt.Printf("%s: %3.3f%s\n", label, value, unit)
}
