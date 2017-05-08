package handlers

import (
	"errors"
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/dns-release/src/dns/server/aliases"

	"github.com/cloudfoundry/dns-release/src/dns/clock"
	"github.com/cloudfoundry/dns-release/src/dns/server/records/dnsresolver"
	"github.com/miekg/dns"
	"net"
)

type AliasResolvingHandler struct {
	nonAliasHandler dns.Handler
	config          aliases.Config
	domainResolver  DomainResolver
	logger          logger.Logger
	logTag          string
	clock           clock.Clock
}

//go:generate counterfeiter . DomainResolver
type DomainResolver interface {
	Resolve(aliasDomains []string, protocol dnsresolver.Protocol, requestMsg *dns.Msg) *dns.Msg
}

func NewAliasResolvingHandler(nonAliasedHandler dns.Handler, config aliases.Config, domainResolver DomainResolver, clock clock.Clock, logger logger.Logger) (AliasResolvingHandler, error) {
	if !config.IsReduced() {
		return AliasResolvingHandler{}, errors.New("must configure with non-recursing alias config")
	}

	return AliasResolvingHandler{
		nonAliasHandler: nonAliasedHandler,
		config:          config,
		domainResolver:  domainResolver,
		clock:           clock,
		logger:          logger,
		logTag:          "AliasResolvingHandler",
	}, nil
}

func (h AliasResolvingHandler) ServeDNS(resp dns.ResponseWriter, requestMsg *dns.Msg) {
	if len(requestMsg.Question) > 0 {
		if aliasDomains := h.config.Resolutions(requestMsg.Question[0].Name); len(aliasDomains) > 0 {
			aliasedDomainsHandler := NewAliasedDomainsHandler(h.domainResolver, aliasDomains, h.logger)
			loggingHandler := NewRequestLoggerHandler(aliasedDomainsHandler, h.clock, h.logger)
			loggingHandler.ServeDNS(resp, requestMsg)
			return
		}
	}

	h.nonAliasHandler.ServeDNS(resp, requestMsg)
}

type aliasedDomainsHandler struct {
	domainResolver DomainResolver
	aliasDomains   []string
	logger         logger.Logger
	logTag         string
}

func NewAliasedDomainsHandler(domainResolver DomainResolver, aliasDomains []string, logger logger.Logger) aliasedDomainsHandler {
	return aliasedDomainsHandler{
		domainResolver: domainResolver,
		aliasDomains:   aliasDomains,
		logger:         logger,
		logTag:         "AliasedHandler",
	}
}

func (a aliasedDomainsHandler) ServeDNS(resp dns.ResponseWriter, requestMsg *dns.Msg) {
	protocol := dnsresolver.UDP
	if _, ok := resp.RemoteAddr().(*net.TCPAddr); ok {
		protocol = dnsresolver.TCP
	}

	// TODO: add tests for protocol, or refactor?

	responseMsg := a.domainResolver.Resolve(a.aliasDomains, protocol, requestMsg)

	if err := resp.WriteMsg(responseMsg); err != nil {
		a.logger.Error(a.logTag, "error writing response %s", err.Error())
	}
}
