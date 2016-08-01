// This file was generated by counterfeiter
package schedulerfakes

import (
	"sync"

	"github.com/concourse/atc"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/db/algorithm"
	"github.com/concourse/atc/scheduler"
	"github.com/pivotal-golang/lager"
)

type FakeBuildScheduler struct {
	ScheduleStub        func(logger lager.Logger, versions *algorithm.VersionsDB, jobConfig atc.JobConfig, resourceConfigs atc.ResourceConfigs, resourceTypes atc.ResourceTypes) error
	scheduleMutex       sync.RWMutex
	scheduleArgsForCall []struct {
		logger          lager.Logger
		versions        *algorithm.VersionsDB
		jobConfig       atc.JobConfig
		resourceConfigs atc.ResourceConfigs
		resourceTypes   atc.ResourceTypes
	}
	scheduleReturns struct {
		result1 error
	}
	TriggerImmediatelyStub        func(logger lager.Logger, jobConfig atc.JobConfig, resourceConfigs atc.ResourceConfigs, resourceTypes atc.ResourceTypes) (db.Build, scheduler.Waiter, error)
	triggerImmediatelyMutex       sync.RWMutex
	triggerImmediatelyArgsForCall []struct {
		logger          lager.Logger
		jobConfig       atc.JobConfig
		resourceConfigs atc.ResourceConfigs
		resourceTypes   atc.ResourceTypes
	}
	triggerImmediatelyReturns struct {
		result1 db.Build
		result2 scheduler.Waiter
		result3 error
	}
	SaveNextInputMappingStub        func(logger lager.Logger, job atc.JobConfig) error
	saveNextInputMappingMutex       sync.RWMutex
	saveNextInputMappingArgsForCall []struct {
		logger lager.Logger
		job    atc.JobConfig
	}
	saveNextInputMappingReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBuildScheduler) Schedule(logger lager.Logger, versions *algorithm.VersionsDB, jobConfig atc.JobConfig, resourceConfigs atc.ResourceConfigs, resourceTypes atc.ResourceTypes) error {
	fake.scheduleMutex.Lock()
	fake.scheduleArgsForCall = append(fake.scheduleArgsForCall, struct {
		logger          lager.Logger
		versions        *algorithm.VersionsDB
		jobConfig       atc.JobConfig
		resourceConfigs atc.ResourceConfigs
		resourceTypes   atc.ResourceTypes
	}{logger, versions, jobConfig, resourceConfigs, resourceTypes})
	fake.recordInvocation("Schedule", []interface{}{logger, versions, jobConfig, resourceConfigs, resourceTypes})
	fake.scheduleMutex.Unlock()
	if fake.ScheduleStub != nil {
		return fake.ScheduleStub(logger, versions, jobConfig, resourceConfigs, resourceTypes)
	} else {
		return fake.scheduleReturns.result1
	}
}

func (fake *FakeBuildScheduler) ScheduleCallCount() int {
	fake.scheduleMutex.RLock()
	defer fake.scheduleMutex.RUnlock()
	return len(fake.scheduleArgsForCall)
}

func (fake *FakeBuildScheduler) ScheduleArgsForCall(i int) (lager.Logger, *algorithm.VersionsDB, atc.JobConfig, atc.ResourceConfigs, atc.ResourceTypes) {
	fake.scheduleMutex.RLock()
	defer fake.scheduleMutex.RUnlock()
	return fake.scheduleArgsForCall[i].logger, fake.scheduleArgsForCall[i].versions, fake.scheduleArgsForCall[i].jobConfig, fake.scheduleArgsForCall[i].resourceConfigs, fake.scheduleArgsForCall[i].resourceTypes
}

func (fake *FakeBuildScheduler) ScheduleReturns(result1 error) {
	fake.ScheduleStub = nil
	fake.scheduleReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBuildScheduler) TriggerImmediately(logger lager.Logger, jobConfig atc.JobConfig, resourceConfigs atc.ResourceConfigs, resourceTypes atc.ResourceTypes) (db.Build, scheduler.Waiter, error) {
	fake.triggerImmediatelyMutex.Lock()
	fake.triggerImmediatelyArgsForCall = append(fake.triggerImmediatelyArgsForCall, struct {
		logger          lager.Logger
		jobConfig       atc.JobConfig
		resourceConfigs atc.ResourceConfigs
		resourceTypes   atc.ResourceTypes
	}{logger, jobConfig, resourceConfigs, resourceTypes})
	fake.recordInvocation("TriggerImmediately", []interface{}{logger, jobConfig, resourceConfigs, resourceTypes})
	fake.triggerImmediatelyMutex.Unlock()
	if fake.TriggerImmediatelyStub != nil {
		return fake.TriggerImmediatelyStub(logger, jobConfig, resourceConfigs, resourceTypes)
	} else {
		return fake.triggerImmediatelyReturns.result1, fake.triggerImmediatelyReturns.result2, fake.triggerImmediatelyReturns.result3
	}
}

func (fake *FakeBuildScheduler) TriggerImmediatelyCallCount() int {
	fake.triggerImmediatelyMutex.RLock()
	defer fake.triggerImmediatelyMutex.RUnlock()
	return len(fake.triggerImmediatelyArgsForCall)
}

func (fake *FakeBuildScheduler) TriggerImmediatelyArgsForCall(i int) (lager.Logger, atc.JobConfig, atc.ResourceConfigs, atc.ResourceTypes) {
	fake.triggerImmediatelyMutex.RLock()
	defer fake.triggerImmediatelyMutex.RUnlock()
	return fake.triggerImmediatelyArgsForCall[i].logger, fake.triggerImmediatelyArgsForCall[i].jobConfig, fake.triggerImmediatelyArgsForCall[i].resourceConfigs, fake.triggerImmediatelyArgsForCall[i].resourceTypes
}

func (fake *FakeBuildScheduler) TriggerImmediatelyReturns(result1 db.Build, result2 scheduler.Waiter, result3 error) {
	fake.TriggerImmediatelyStub = nil
	fake.triggerImmediatelyReturns = struct {
		result1 db.Build
		result2 scheduler.Waiter
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeBuildScheduler) SaveNextInputMapping(logger lager.Logger, job atc.JobConfig) error {
	fake.saveNextInputMappingMutex.Lock()
	fake.saveNextInputMappingArgsForCall = append(fake.saveNextInputMappingArgsForCall, struct {
		logger lager.Logger
		job    atc.JobConfig
	}{logger, job})
	fake.recordInvocation("SaveNextInputMapping", []interface{}{logger, job})
	fake.saveNextInputMappingMutex.Unlock()
	if fake.SaveNextInputMappingStub != nil {
		return fake.SaveNextInputMappingStub(logger, job)
	} else {
		return fake.saveNextInputMappingReturns.result1
	}
}

func (fake *FakeBuildScheduler) SaveNextInputMappingCallCount() int {
	fake.saveNextInputMappingMutex.RLock()
	defer fake.saveNextInputMappingMutex.RUnlock()
	return len(fake.saveNextInputMappingArgsForCall)
}

func (fake *FakeBuildScheduler) SaveNextInputMappingArgsForCall(i int) (lager.Logger, atc.JobConfig) {
	fake.saveNextInputMappingMutex.RLock()
	defer fake.saveNextInputMappingMutex.RUnlock()
	return fake.saveNextInputMappingArgsForCall[i].logger, fake.saveNextInputMappingArgsForCall[i].job
}

func (fake *FakeBuildScheduler) SaveNextInputMappingReturns(result1 error) {
	fake.SaveNextInputMappingStub = nil
	fake.saveNextInputMappingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBuildScheduler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.scheduleMutex.RLock()
	defer fake.scheduleMutex.RUnlock()
	fake.triggerImmediatelyMutex.RLock()
	defer fake.triggerImmediatelyMutex.RUnlock()
	fake.saveNextInputMappingMutex.RLock()
	defer fake.saveNextInputMappingMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeBuildScheduler) recordInvocation(key string, args []interface{}) {
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

var _ scheduler.BuildScheduler = new(FakeBuildScheduler)
