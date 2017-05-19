package manager_test

import (
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"

	boshsysfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/cloudfoundry/dns-release/src/dns/manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolvConfManager", func() {
	var (
		dnsManager    manager.DNSManager
		fs            *boshsysfakes.FakeFileSystem
		clock         *fakeclock.FakeClock
		fakeCmdRunner *boshsysfakes.FakeCmdRunner
	)

	BeforeEach(func() {
		clock = fakeclock.NewFakeClock(time.Now())
		fakeCmdRunner = boshsysfakes.NewFakeCmdRunner()
		fs = boshsysfakes.NewFakeFileSystem()
		dnsManager = manager.NewResolvConfManager(clock, fs, fakeCmdRunner)
	})

	Describe("Read", func() {
		Context("When resolv.conf is empty", func() {
			BeforeEach(func() {
				_ = fs.WriteFile("/etc/resolv.conf", []byte(""))
			})

			It("returns an empty array", func() {
				nameservers, err := dnsManager.Read()

				Expect(err).ToNot(HaveOccurred())
				Expect(nameservers).To(HaveLen(0))
			})
		})

		Context("When resolv.conf has multiple nameservers", func() {
			BeforeEach(func() {
				_ = fs.WriteFile("/etc/resolv.conf", []byte(fmt.Sprintf(`
# Generated by dhcpcd from eth0.dhcp
# /etc/resolv.conf.head can replace this line
domain sf.pivotallabs.com
search sf.pivotallabs.com pivotallabs.com
nameserver ns-1
nameserver ns-2
# /etc/resolv.conf.tail can replace this line
`)))
			})

			It("returns all entries", func() {
				nameservers, err := dnsManager.Read()

				Expect(err).ToNot(HaveOccurred())
				Expect(nameservers).To(HaveLen(2))
				Expect(nameservers).To(ConsistOf("ns-1", "ns-2"))
			})

			Context("When there are malformed entries", func() {
				Context("nameserver is missing a value", func() {
					BeforeEach(func() {
						_ = fs.WriteFile("/etc/resolv.conf", []byte(fmt.Sprintf(`
# Generated by dhcpcd from eth0.dhcp
# /etc/resolv.conf.head can replace this line
domain sf.pivotallabs.com
search sf.pivotallabs.com pivotallabs.com
nameserver ns-1
nameserver
# /etc/resolv.conf.tail can replace this line
`)))
					})

					It("returns all complete entries", func() {
						nameservers, err := dnsManager.Read()

						Expect(err).ToNot(HaveOccurred())
						Expect(nameservers).To(HaveLen(1))
						Expect(nameservers).To(ConsistOf("ns-1"))
					})
				})

				Context("nameserver entry has spaces or other text before 'nameserver'", func() {
					BeforeEach(func() {
						_ = fs.WriteFile("/etc/resolv.conf", []byte(fmt.Sprintf(`
# Generated by dhcpcd from eth0.dhcp
# /etc/resolv.conf.head can replace this line
domain sf.pivotallabs.com
search sf.pivotallabs.com pivotallabs.com
 nameserver   ns-1
nameserver ns-2
1 nameserver foo
# /etc/resolv.conf.tail can replace this line
`)))
					})

					It("returns all entries", func() {
						nameservers, err := dnsManager.Read()

						Expect(err).ToNot(HaveOccurred())
						Expect(nameservers).To(HaveLen(2))
						Expect(nameservers).To(ConsistOf("ns-1", "ns-2"))
					})
				})
			})
		})

		Context("When resolv.conf is not readable", func() {
			BeforeEach(func() {
				_ = fs.WriteFile("/etc/resolv.conf", []byte(""))
				fs.RegisterReadFileError("/etc/resolv.conf", errors.New("unable to read /etc/resolv.conf"))
			})

			It("returns an error", func() {
				_, err := dnsManager.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("attempting to read dns nameservers"))
			})
		})
	})

	Describe("SetPrimary", func() {
		Context("filesystem fails", func() {
			It("errors", func() {
				fakeCmdRunner.AddCmdResult("resolvconf -u", boshsysfakes.FakeCmdResult{})
				fs.WriteFileError = errors.New("fake-err1")

				go clock.WaitForWatcherAndIncrement(time.Second * 2)
				err := dnsManager.SetPrimary("192.0.2.100")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Writing "))
				Expect(err.Error()).To(ContainSubstring("fake-err1"))
			})
		})

		Context("resolvconf update fails", func() {
			It("errors", func() {
				fakeCmdRunner.AddCmdResult("resolvconf -u", boshsysfakes.FakeCmdResult{ExitStatus: 1, Error: errors.New("fake-err1")})

				go clock.WaitForWatcherAndIncrement(time.Second * 2)
				err := dnsManager.SetPrimary("192.0.2.100")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Executing "))
				Expect(err.Error()).To(ContainSubstring("fake-err1"))
			})
		})

		Context("resolvconf fails to rewrite /etc/resolv.conf", func() {
			It("errors if resolvconf update fails", func() {
				fakeCmdRunner.AddCmdResult("resolvconf -u", boshsysfakes.FakeCmdResult{})

				go func() {
					for i := 0; i < manager.MaxResolvConfRetries; i++ {
						clock.WaitForWatcherAndIncrement(time.Second * 2)
					}
				}()
				err := dnsManager.SetPrimary("192.0.2.100")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to confirm nameserver "))
			})
		})

		It("skips if resolvconf already has our server", func() {
			_ = fs.WriteFileString("/etc/resolv.conf", `nameserver 192.0.2.100`)

			err := dnsManager.SetPrimary("192.0.2.100")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmdRunner.RunCommands).To(HaveLen(0))
		})

		It("creates /etc/resolvconf/resolv.conf.d/head with our DNS server", func() {
			fakeCmdRunner.AddCmdResult("resolvconf -u", boshsysfakes.FakeCmdResult{})
			fakeCmdRunner.SetCmdCallback("resolvconf -u", func() {
				_ = fs.WriteFileString("/etc/resolv.conf", `nameserver 192.0.2.100`)
			})

			go clock.WaitForWatcherAndIncrement(time.Second * 2)
			err := dnsManager.SetPrimary("192.0.2.100")
			Expect(err).NotTo(HaveOccurred())

			contents, err := fs.ReadFileString("/etc/resolvconf/resolv.conf.d/head")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(`# This file was automatically updated by bosh-dns
nameserver 192.0.2.100
`))
		})

		It("avoids prepending itself more than once (in case resolvconf is slower than our check interval)", func() {
			fakeCmdRunner.AddCmdResult("resolvconf -u", boshsysfakes.FakeCmdResult{})

			fakeCmdRunner.SetCmdCallback("resolvconf -u", func() {
				_ = fs.WriteFileString("/etc/resolv.conf", `nameserver 192.0.2.100`)
			})

			err := fs.WriteFileString("/etc/resolvconf/resolv.conf.d/head", `
nameserver 192.0.2.100
nameserver 8.8.8.8
`)
			Expect(err).NotTo(HaveOccurred())

			go clock.WaitForWatcherAndIncrement(time.Second * 2)
			err = dnsManager.SetPrimary("192.0.2.100")
			Expect(err).NotTo(HaveOccurred())

			contents, err := fs.ReadFileString("/etc/resolvconf/resolv.conf.d/head")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(`# This file was automatically updated by bosh-dns
nameserver 192.0.2.100
`))
		})

		It("prepends /etc/resolvconf/resolv.conf.d/head with our DNS server", func() {
			_ = fs.WriteFileString("/etc/resolvconf/resolv.conf.d/head", `# some comment
nameserver 192.0.3.1
nameserver 192.0.3.2
`)

			fakeCmdRunner.SetCmdCallback("resolvconf -u", func() {
				_ = fs.WriteFileString("/etc/resolv.conf", `nameserver 192.0.2.100`)
			})

			go clock.WaitForWatcherAndIncrement(time.Second * 2)
			err := dnsManager.SetPrimary("192.0.2.100")
			Expect(err).NotTo(HaveOccurred())

			contents, err := fs.ReadFileString("/etc/resolvconf/resolv.conf.d/head")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(`# This file was automatically updated by bosh-dns
nameserver 192.0.2.100

# some comment
nameserver 192.0.3.1
nameserver 192.0.3.2
`))
		})
	})
})
