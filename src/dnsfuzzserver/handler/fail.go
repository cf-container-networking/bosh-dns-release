package handler

import (
	"fmt"

	"github.com/miekg/dns"
)

var Fail = failHandler{}

type failHandler struct{}

func (failHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	var qname string

	if len(r.Question) > 0 {
		qname = r.Question[0].Name
	}

	fmt.Println(fmt.Errorf("dns-fuzz: ignored query: %s", qname))
}
