// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/concourse/atc/radar"
	"github.com/tedsuo/ifrit"
)

type FakeScanner struct {
	ScanStub        func(string) ifrit.Runner
	scanMutex       sync.RWMutex
	scanArgsForCall []struct {
		arg1 string
	}
	scanReturns struct {
		result1 ifrit.Runner
	}
}

func (fake *FakeScanner) Scan(arg1 string) ifrit.Runner {
	fake.scanMutex.Lock()
	fake.scanArgsForCall = append(fake.scanArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.scanMutex.Unlock()
	if fake.ScanStub != nil {
		return fake.ScanStub(arg1)
	} else {
		return fake.scanReturns.result1
	}
}

func (fake *FakeScanner) ScanCallCount() int {
	fake.scanMutex.RLock()
	defer fake.scanMutex.RUnlock()
	return len(fake.scanArgsForCall)
}

func (fake *FakeScanner) ScanArgsForCall(i int) string {
	fake.scanMutex.RLock()
	defer fake.scanMutex.RUnlock()
	return fake.scanArgsForCall[i].arg1
}

func (fake *FakeScanner) ScanReturns(result1 ifrit.Runner) {
	fake.ScanStub = nil
	fake.scanReturns = struct {
		result1 ifrit.Runner
	}{result1}
}

var _ radar.Scanner = new(FakeScanner)
