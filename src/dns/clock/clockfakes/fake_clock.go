// Code generated by counterfeiter. DO NOT EDIT.
package clockfakes

import (
	"sync"
	"time"

	"github.com/cloudfoundry/dns-release/src/dns/clock"
)

type FakeClock struct {
	NowStub        func() time.Time
	nowMutex       sync.RWMutex
	nowArgsForCall []struct{}
	nowReturns     struct {
		result1 time.Time
	}
	nowReturnsOnCall map[int]struct {
		result1 time.Time
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClock) Now() time.Time {
	fake.nowMutex.Lock()
	ret, specificReturn := fake.nowReturnsOnCall[len(fake.nowArgsForCall)]
	fake.nowArgsForCall = append(fake.nowArgsForCall, struct{}{})
	fake.recordInvocation("Now", []interface{}{})
	fake.nowMutex.Unlock()
	if fake.NowStub != nil {
		return fake.NowStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.nowReturns.result1
}

func (fake *FakeClock) NowCallCount() int {
	fake.nowMutex.RLock()
	defer fake.nowMutex.RUnlock()
	return len(fake.nowArgsForCall)
}

func (fake *FakeClock) NowReturns(result1 time.Time) {
	fake.NowStub = nil
	fake.nowReturns = struct {
		result1 time.Time
	}{result1}
}

func (fake *FakeClock) NowReturnsOnCall(i int, result1 time.Time) {
	fake.NowStub = nil
	if fake.nowReturnsOnCall == nil {
		fake.nowReturnsOnCall = make(map[int]struct {
			result1 time.Time
		})
	}
	fake.nowReturnsOnCall[i] = struct {
		result1 time.Time
	}{result1}
}

func (fake *FakeClock) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.nowMutex.RLock()
	defer fake.nowMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeClock) recordInvocation(key string, args []interface{}) {
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

var _ clock.Clock = new(FakeClock)
