// Code generated by counterfeiter. DO NOT EDIT.
package handlersfakes

import (
	"bosh-dns/dns/server/handlers"
	"net/http"
	"sync"
)

type FakeHTTPClient struct {
	GetStub        func(endpoint string) (*http.Response, error)
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		endpoint string
	}
	getReturns struct {
		result1 *http.Response
		result2 error
	}
	getReturnsOnCall map[int]struct {
		result1 *http.Response
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeHTTPClient) Get(endpoint string) (*http.Response, error) {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		endpoint string
	}{endpoint})
	fake.recordInvocation("Get", []interface{}{endpoint})
	fake.getMutex.Unlock()
	if fake.GetStub != nil {
		return fake.GetStub(endpoint)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.getReturns.result1, fake.getReturns.result2
}

func (fake *FakeHTTPClient) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeHTTPClient) GetArgsForCall(i int) string {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return fake.getArgsForCall[i].endpoint
}

func (fake *FakeHTTPClient) GetReturns(result1 *http.Response, result2 error) {
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 *http.Response
		result2 error
	}{result1, result2}
}

func (fake *FakeHTTPClient) GetReturnsOnCall(i int, result1 *http.Response, result2 error) {
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 *http.Response
			result2 error
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 *http.Response
		result2 error
	}{result1, result2}
}

func (fake *FakeHTTPClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeHTTPClient) recordInvocation(key string, args []interface{}) {
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

var _ handlers.HTTPClient = new(FakeHTTPClient)
