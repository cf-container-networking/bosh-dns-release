package main_test

import (
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWait(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "wait")
}

var (
	pathToBinary string
	pathToServer string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	waitPath, err := gexec.Build("bosh-dns/wait")
	Expect(err).NotTo(HaveOccurred())
	dnsPath, err := gexec.Build("bosh-dns/dns")
	Expect(err).NotTo(HaveOccurred())
	SetDefaultEventuallyTimeout(2 * time.Second)

	return []byte(strings.Join([]string{waitPath, dnsPath}, ","))
}, func(data []byte) {
	paths := strings.Split(string(data), ",")
	pathToBinary = paths[0]
	pathToServer = paths[1]
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})
