package server_test

import (
	"fmt"

	"github.com/cloudfoundry/dns-release/src/dns/server"

	"errors"

	"net"
	"time"

	"sync"

	"github.com/cloudfoundry/dns-release/src/dns/server/internal/internalfakes"
	"github.com/cloudfoundry/dns-release/src/dns/server/serverfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getFreePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", err
	}

	return port, nil
}

func tcpServerStub(bindAddress string, stop chan struct{}) func() error {
	return func() error {
		_, err := net.Listen("tcp", bindAddress)
		if err != nil {
			return err
		}

		select {
		case <-stop:
		}

		return nil
	}
}

func udpServerStub(bindAddress string, timeout time.Duration, stop chan struct{}) func() error {
	return func() error {
		listener, err := net.ListenPacket("udp", bindAddress)

		if err != nil {
			return err
		}
		listener.SetDeadline(time.Now().Add(timeout + (10 * time.Second)))

		buf := make([]byte, 1)
		for {
			n, addr, err := listener.ReadFrom(buf)
			if err != nil {
				return err
			}

			if n == 0 {
				return errors.New("not enough bytes read")
			}

			if _, err := listener.WriteTo(buf, addr); err != nil {
				return err
			}

			select {
			case <-stop:
			default:
			}
		}
	}
}

func notListeningStub(stop chan struct{}) func() error {
	return func() error {
		select {
		case <-stop:
		}

		return nil
	}
}

func passthroughCheck(reactiveAnswerChan chan error) *serverfakes.FakeHealthCheck {
	return &serverfakes.FakeHealthCheck{
		IsHealthyStub: func() error {
			return <-reactiveAnswerChan
		},
	}
}

func healthyCheck() *serverfakes.FakeHealthCheck {
	return &serverfakes.FakeHealthCheck{
		IsHealthyStub: func() error {
			return nil
		},
	}
}

func unhealthyCheck() *serverfakes.FakeHealthCheck {
	return &serverfakes.FakeHealthCheck{
		IsHealthyStub: func() error {
			return errors.New("fake unhealthy")
		},
	}
}

func shutdownStub(err error) func() error {
	return func() error {
		return err
	}
}

var _ = Describe("Server", func() {
	var (
		dnsServer       server.Server
		fakeTCPServer   *serverfakes.FakeDNSServer
		fakeUDPServer   *serverfakes.FakeDNSServer
		tcpHealthCheck  *serverfakes.FakeHealthCheck
		udpHealthCheck  *serverfakes.FakeHealthCheck
		fakeDialer      server.Dialer
		timeout         time.Duration
		bindAddress     string
		lock            sync.Mutex
		shutdownChannel chan struct{}
		stopFakeServer  chan struct{}
	)

	BeforeEach(func() {
		stopFakeServer = make(chan struct{})
		shutdownChannel = make(chan struct{})
		timeout = 1 * time.Second

		port, err := getFreePort()
		Expect(err).NotTo(HaveOccurred())

		bindAddress = fmt.Sprintf("127.0.0.1:%s", port)

		fakeTCPServer = &serverfakes.FakeDNSServer{}
		fakeUDPServer = &serverfakes.FakeDNSServer{}
		fakeTCPServer.ListenAndServeStub = tcpServerStub(bindAddress, stopFakeServer)
		fakeUDPServer.ListenAndServeStub = udpServerStub(bindAddress, timeout, stopFakeServer)

		tcpHealthCheck = healthyCheck()
		udpHealthCheck = healthyCheck()

		SetDefaultEventuallyTimeout(timeout + 2*time.Second)
		SetDefaultConsistentlyDuration(timeout + 2*time.Second)

		fakeDialer = net.Dial
	})

	JustBeforeEach(func() {
		dnsServer = server.New(
			[]server.DNSServer{fakeTCPServer, fakeUDPServer},
			[]server.HealthCheck{tcpHealthCheck, udpHealthCheck},
			timeout,
			shutdownChannel,
		)
	})

	AfterEach(func() {
		close(stopFakeServer)
	})

	Context("Run", func() {
		Context("when the timeout has been reached", func() {
			Context("and the servers are not up", func() {
				BeforeEach(func() {
					tcpHealthCheck = unhealthyCheck()
					udpHealthCheck = unhealthyCheck()
				})
				It("returns an error", func() {

					fakeTCPServer.ListenAndServeStub = notListeningStub(stopFakeServer)
					fakeUDPServer.ListenAndServeStub = notListeningStub(stopFakeServer)

					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Eventually(dnsServerFinished).Should(Receive(Equal(errors.New("timed out waiting for server to bind"))))
				})
			})
		})

		Context("when a provided tcp server cannot listen and serve", func() {
			BeforeEach(func() {
				tcpHealthCheck = unhealthyCheck()
				udpHealthCheck = unhealthyCheck()
			})

			It("should return an error when the tcp server cannot listen and serve", func() {
				fakeTCPServer.ListenAndServeReturns(errors.New("some-fake-tcp-error"))

				err := dnsServer.Run()
				Expect(err).To(MatchError("some-fake-tcp-error"))
			})

			It("should return an error when the udp server cannot listen and serve", func() {
				fakeUDPServer.ListenAndServeReturns(errors.New("some-fake-udp-error"))

				err := dnsServer.Run()
				Expect(err).To(MatchError("some-fake-udp-error"))
			})
		})

		Context("health checking", func() {
			var fakeConn *internalfakes.FakeConn

			BeforeEach(func() {
				fakeConn = &internalfakes.FakeConn{}
			})

			Context("when both servers are up", func() {
				var fakeProtocolConn map[string]*internalfakes.FakeConn
				var fakeProtocolDialConn map[string]int

				BeforeEach(func() {
					fakeProtocolConn = map[string]*internalfakes.FakeConn{
						"tcp": {},
						"udp": {},
					}

					fakeProtocolDialConn = map[string]int{
						"tcp": 0,
						"udp": 0,
					}

					fakeDialer = func(protocol, address string) (net.Conn, error) {
						lock.Lock()
						fakeProtocolDialConn[protocol]++
						lock.Unlock()

						return fakeProtocolConn[protocol], nil
					}
				})

				It("blocks forever", func() {
					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Consistently(dnsServerFinished).ShouldNot(Receive())

					Expect(fakeProtocolConn["tcp"].CloseCallCount()).To(Equal(fakeProtocolDialConn["tcp"]))
					Expect(fakeProtocolConn["udp"].CloseCallCount()).To(Equal(fakeProtocolDialConn["udp"]))
				})
			})

			Context("when the TCP server suddenly stops working", func() {
				var tcpHealthChan chan error

				BeforeEach(func() {
					tcpHealthChan = make(chan error)
					udpHealthCheck = healthyCheck()
					tcpHealthCheck = passthroughCheck(tcpHealthChan)
				})

				It("kills itself after five failures in a row", func() {
					startFailing := make(chan struct{})
					go func() {
						defer GinkgoRecover()
					dance:
						for {
							select {
							case <-startFailing:
								break dance
							default:
								tcpHealthChan <- nil
							}
						}

						for i := 0; i < 5; i++ {
							tcpHealthChan <- errors.New("deadbeef")
						}
					}()

					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Consistently(dnsServerFinished).ShouldNot(Receive())
					close(startFailing)
					Eventually(dnsServerFinished, time.Second*30).Should(Receive())
				})
			})

			Context("when the UDP server suddenly stops working", func() {
				var udpHealthChan chan error

				BeforeEach(func() {
					udpHealthChan = make(chan error)
					udpHealthCheck = passthroughCheck(udpHealthChan)
					tcpHealthCheck = healthyCheck()
				})

				It("kills itself after five failures in a row", func() {
					startFailing := make(chan struct{})
					go func() {
					dance:
						for {
							select {
							case <-startFailing:
								break dance
							default:
								udpHealthChan <- nil
							}
						}

						for i := 0; i < 5; i++ {
							udpHealthChan <- errors.New("deadbeef")
						}
					}()

					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Consistently(dnsServerFinished).ShouldNot(Receive())
					close(startFailing)
					Eventually(dnsServerFinished, 30*time.Second).Should(Receive())
				})
			})

			Context("when the udp server never binds to a port", func() {
				BeforeEach(func() {
					udpHealthCheck = unhealthyCheck()
				})

				It("returns an error", func() {
					fakeUDPServer.ListenAndServeStub = notListeningStub(stopFakeServer)

					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Eventually(dnsServerFinished).Should(Receive(Equal(errors.New("timed out waiting for server to bind"))))
				})
			})

			Context("when the tcp server never binds to a port", func() {
				BeforeEach(func() {
					tcpHealthCheck = unhealthyCheck()
				})

				It("returns an error", func() {
					fakeTCPServer.ListenAndServeStub = notListeningStub(stopFakeServer)

					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Eventually(dnsServerFinished).Should(Receive(Equal(errors.New("timed out waiting for server to bind"))))
				})
			})

			Context("shutdown signal", func() {
				It("gracefully terminates the servers when the shutdown signal has been fired", func() {
					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Consistently(dnsServerFinished).ShouldNot(Receive())

					close(shutdownChannel)
					Eventually(dnsServerFinished).Should(Receive(nil))

					Expect(fakeTCPServer.ShutdownCallCount()).To(Equal(1))
					Expect(fakeUDPServer.ShutdownCallCount()).To(Equal(1))
				})

				It("returns an error if the servers were unable to shutdown cleanly", func() {
					err := errors.New("failed to shutdown tcp server")
					fakeTCPServer.ShutdownReturns(err)
					dnsServerFinished := make(chan error)
					go func() {
						dnsServerFinished <- dnsServer.Run()
					}()

					Consistently(dnsServerFinished).ShouldNot(Receive())

					close(shutdownChannel)
					Eventually(dnsServerFinished).Should(Receive(Equal(err)))

					Expect(fakeTCPServer.ShutdownCallCount()).To(Equal(1))
					Expect(fakeUDPServer.ShutdownCallCount()).To(Equal(1))
				})
			})
		})
	})
})
