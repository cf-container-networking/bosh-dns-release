// Code generated by counterfeiter. DO NOT EDIT.
package dnsresolverfakes

import (
	"sync"

	"github.com/cloudfoundry/dns-release/src/dns/server/records/dnsresolver"
	"github.com/miekg/dns"
)

type FakeAnswerShuffler struct {
	ShuffleStub        func(src []dns.RR) []dns.RR
	shuffleMutex       sync.RWMutex
	shuffleArgsForCall []struct {
		src []dns.RR
	}
	shuffleReturns struct {
		result1 []dns.RR
	}
	shuffleReturnsOnCall map[int]struct {
		result1 []dns.RR
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeAnswerShuffler) Shuffle(src []dns.RR) []dns.RR {
	var srcCopy []dns.RR
	if src != nil {
		srcCopy = make([]dns.RR, len(src))
		copy(srcCopy, src)
	}
	fake.shuffleMutex.Lock()
	ret, specificReturn := fake.shuffleReturnsOnCall[len(fake.shuffleArgsForCall)]
	fake.shuffleArgsForCall = append(fake.shuffleArgsForCall, struct {
		src []dns.RR
	}{srcCopy})
	fake.recordInvocation("Shuffle", []interface{}{srcCopy})
	fake.shuffleMutex.Unlock()
	if fake.ShuffleStub != nil {
		return fake.ShuffleStub(src)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.shuffleReturns.result1
}

func (fake *FakeAnswerShuffler) ShuffleCallCount() int {
	fake.shuffleMutex.RLock()
	defer fake.shuffleMutex.RUnlock()
	return len(fake.shuffleArgsForCall)
}

func (fake *FakeAnswerShuffler) ShuffleArgsForCall(i int) []dns.RR {
	fake.shuffleMutex.RLock()
	defer fake.shuffleMutex.RUnlock()
	return fake.shuffleArgsForCall[i].src
}

func (fake *FakeAnswerShuffler) ShuffleReturns(result1 []dns.RR) {
	fake.ShuffleStub = nil
	fake.shuffleReturns = struct {
		result1 []dns.RR
	}{result1}
}

func (fake *FakeAnswerShuffler) ShuffleReturnsOnCall(i int, result1 []dns.RR) {
	fake.ShuffleStub = nil
	if fake.shuffleReturnsOnCall == nil {
		fake.shuffleReturnsOnCall = make(map[int]struct {
			result1 []dns.RR
		})
	}
	fake.shuffleReturnsOnCall[i] = struct {
		result1 []dns.RR
	}{result1}
}

func (fake *FakeAnswerShuffler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.shuffleMutex.RLock()
	defer fake.shuffleMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeAnswerShuffler) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ dnsresolver.AnswerShuffler = new(FakeAnswerShuffler)
