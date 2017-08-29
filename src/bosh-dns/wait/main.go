package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/miekg/dns"
)

func main() {
	//start dns server
	//check dns server running
	//exit 0/1
	//5 minute timeout

	/*
		./wait -command "/var/vcap/jobs/bosh-dns/bin/bosh_dns_ctl start" \
					 -command "/var/vcap/jobs/bosh-dns/bin/bosh_dns_ctl start" \
			     -timeout 5m \
					 -checkDomain "upcheck.com"
	*/
	domain := flag.String("checkDomain", "", "dns address to confirm command success")
	command := flag.String("command", "", "command to execute")
	timeout := flag.Duration("timeout", time.Minute, "amount of time to wait for check to pass")
	nameServer := flag.String("nameServer", "", "dns server to talk to")

	flag.Parse()

	cmd := exec.Command("bash", "-c", *command)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	dnsClient := &dns.Client{}
	m := &dns.Msg{Question: []dns.Question{
		{Name: *domain},
	}}

	bomb := time.NewTimer(*timeout)
	ticker := time.NewTicker(*timeout / 10)
	fmt.Println(*timeout)
	for {
		select {
		case <-bomb.C:
			os.Exit(1)
		case <-ticker.C:
			_, _, err = dnsClient.Exchange(m, *nameServer)
			if err != nil {
				fmt.Println(err)
			} else {
				os.Exit(0)
			}
		}
	}
}
