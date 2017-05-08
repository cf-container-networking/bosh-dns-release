package handlers

import (
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/dns-release/src/dns/server/records/dnsresolver"
	"github.com/miekg/dns"
	"net"
)

type DiscoveryHandler struct {
	logger      logger.Logger
	logTag      string
	localDomain dnsresolver.LocalDomain
}

func NewDiscoveryHandler(logger logger.Logger, localDomain dnsresolver.LocalDomain) DiscoveryHandler {
	return DiscoveryHandler{
		logger:      logger,
		logTag:      "DiscoveryHandler",
		localDomain: localDomain,
	}
}

func (d DiscoveryHandler) ServeDNS(responseWriter dns.ResponseWriter, req *dns.Msg) {
	responseMsg := d.buildResponseMsg(responseWriter, req)
	if err := responseWriter.WriteMsg(responseMsg); err != nil {
		d.logger.Error(d.logTag, err.Error())
	}
}

func (d DiscoveryHandler) buildResponseMsg(responseWriter dns.ResponseWriter, req *dns.Msg) *dns.Msg {
	defaultMessage := &dns.Msg{}
	defaultMessage.Authoritative = true
	defaultMessage.RecursionAvailable = false
	defaultMessage.SetRcode(req, dns.RcodeSuccess)

	if len(req.Question) > 0 {
		switch req.Question[0].Qtype {
		case dns.TypeA, dns.TypeANY:
			return d.buildARecords(responseWriter, req)
		case dns.TypeMX, dns.TypeAAAA:
			return defaultMessage
		default:
			defaultMessage.SetRcode(req, dns.RcodeServerFailure)
		}
	}
	return defaultMessage
}

func (d DiscoveryHandler) buildARecords(responseWriter dns.ResponseWriter, requestMsg *dns.Msg) *dns.Msg {
	protocol := dnsresolver.UDP
	if _, ok := responseWriter.RemoteAddr().(*net.TCPAddr); ok {
		protocol = dnsresolver.TCP
	}
	return d.localDomain.Resolve([]string{requestMsg.Question[0].Name}, protocol, requestMsg)
}
