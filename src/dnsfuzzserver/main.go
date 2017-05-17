package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/dns-release/src/dnsfuzzserver/handler"
	"github.com/miekg/dns"
)

// $ go run ./main.go &
// $ dig -p 35053 +tcp @127.0.0.1 rcode-servfail.delay-8s.size-4.ttl-7.hdr-tc.hdr-aa.test
func main() {
	net := "tcp"
	addr := "127.0.0.1:35053"

	if len(os.Args) > 1 {
		net = os.Args[1]

		if len(os.Args) > 2 {
			addr = os.Args[2]
		}
	}

	server := &dns.Server{
		Addr:    addr,
		Net:     net,
		UDPSize: dns.MaxMsgSize,
	}

	dns.Handle("test.", handler.NewDynamicFuzz("test."))
	dns.Handle("custom.test.", handler.NewStaticFuzz("size-8.ttl-16.answer.ttl-4.answer"))
	dns.Handle(".", handler.Fail)

	fmt.Println(fmt.Sprintf("listening on %s (%s)", addr, net))

	if err := server.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}
