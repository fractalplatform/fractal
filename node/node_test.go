// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/p2p"
	"github.com/fractalplatform/fractal/rpc"
)

var (
	testNodeKey, _ = crypto.GenerateKey()
)

func TestNodeLifeCycle(t *testing.T) {
	// Create a temporary folder to use as the data directory
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary data directory: %v", err)
	}
	defer os.RemoveAll(dir)

	var config = &Config{
		DataDir:   dir,
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}

	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	// Ensure that a stopped node can be stopped again
	for i := 0; i < 3; i++ {
		if err := stack.Stop(); err != ErrNodeStopped {
			t.Fatalf("iter %d: stop failure mismatch: have %v, want %v", i, err, ErrNodeStopped)
		}
	}
	// Ensure that a node can be successfully started, but only once
	if err := stack.Start(); err != nil {
		t.Fatalf("failed to start node: %v", err)
	}
	if err := stack.Start(); err != ErrNodeRunning {
		t.Fatalf("start failure mismatch: have %v, want %v ", err, ErrNodeRunning)
	}

	// Ensure that a node can be restarted arbitrarily many times
	for i := 0; i < 3; i++ {
		if err := stack.Restart(); err != nil {
			t.Fatalf("iter %d: failed to restart node: %v", i, err)
		}
	}
	// Ensure that a node can be stopped, but only once
	if err := stack.Stop(); err != nil {
		t.Fatalf("failed to stop node: %v", err)
	}
	if err := stack.Stop(); err != ErrNodeStopped {
		t.Fatalf("stop failure mismatch: have %v, want %v ", err, ErrNodeStopped)
	}
}

// Tests that if the data dir is already in use, an appropriate error is returned.
func TestNodeUsedDataDir(t *testing.T) {

	// Create a temporary folder to use as the data directory
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary data directory: %v", err)
	}
	defer os.RemoveAll(dir)

	var config = &Config{
		DataDir:   dir,
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}

	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.Start(); err != nil {
		t.Fatalf("failed to start original protocol stack: %v", err)
	}
	defer stack.Stop()

	// Create a second node based on the same data directory and ensure failure
	duplicate, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := duplicate.Start(); err != ErrDatadirUsed {
		t.Fatalf("duplicate datadir failure mismatch: have %v, want %v", err, ErrDatadirUsed)
	}
}

// NoopService is a trivial implementation of the Service interface.
type NoopService struct{}

func (s *NoopService) APIs() []rpc.API           { return nil }
func (s *NoopService) Start() error              { return nil }
func (s *NoopService) Stop() error               { return nil }
func (s *NoopService) Protocols() []p2p.Protocol { return nil }

func NewNoopService(*ServiceContext) (Service, error) { return new(NoopService), nil }

// Set of services all wrapping the base NoopService resulting in the same method
// signatures but different outer types.
type NoopServiceA struct{ NoopService }
type NoopServiceB struct{ NoopService }
type NoopServiceC struct{ NoopService }

func NewNoopServiceA(*ServiceContext) (Service, error) { return new(NoopServiceA), nil }
func NewNoopServiceB(*ServiceContext) (Service, error) { return new(NoopServiceB), nil }
func NewNoopServiceC(*ServiceContext) (Service, error) { return new(NoopServiceC), nil }

// Tests whether services can be registered and duplicates caught.
func TestServiceRegistry(t *testing.T) {

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary data directory: %v", err)
	}
	defer os.RemoveAll(dir)

	var config = &Config{
		DataDir:   dir,
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}

	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	// Register a batch of unique services and ensure they start successfully
	services := []ServiceConstructor{NewNoopServiceA, NewNoopServiceB, NewNoopServiceC}
	for i, constructor := range services {
		if err := stack.Register(constructor); err != nil {
			t.Fatalf("service #%d: registration failed: %v", i, err)
		}
	}
	if err := stack.Start(); err != nil {
		t.Fatalf("failed to start original service stack: %v", err)
	}
	if err := stack.Stop(); err != nil {
		t.Fatalf("failed to stop original service stack: %v", err)
	}
	// Duplicate one of the services and retry starting the node
	if err := stack.Register(NewNoopServiceB); err != nil {
		t.Fatalf("duplicate registration failed: %v", err)
	}
	if err := stack.Start(); err == nil {
		t.Fatalf("duplicate service started")
	} else {
		if !strings.Contains(err.Error(), "duplicate service") {
			t.Fatalf("duplicate error mismatch: %v", err)
		}
	}
}

// InstrumentedService is an implementation of Service for which all interface
// methods can be instrumented both return value as well as event hook wise.
type InstrumentedService struct {
	protocols []p2p.Protocol
	apis      []rpc.API
	start     error
	stop      error

	protocolsHook func()
	startHook     func()
	stopHook      func()
}

func NewInstrumentedService(*ServiceContext) (Service, error) { return new(InstrumentedService), nil }

func (s *InstrumentedService) Protocols() []p2p.Protocol {
	if s.protocolsHook != nil {
		s.protocolsHook()
	}
	return s.protocols
}

func (s *InstrumentedService) APIs() []rpc.API {
	return s.apis
}

func (s *InstrumentedService) Start() error {
	if s.startHook != nil {
		s.startHook()
	}
	return s.start
}

func (s *InstrumentedService) Stop() error {
	if s.stopHook != nil {
		s.stopHook()
	}
	return s.stop
}

// InstrumentingWrapper is a method to specialize a service constructor returning
// a generic InstrumentedService into one returning a wrapping specific one.
type InstrumentingWrapper func(base ServiceConstructor) ServiceConstructor

func InstrumentingWrapperMaker(base ServiceConstructor, kind reflect.Type) ServiceConstructor {
	return func(ctx *ServiceContext) (Service, error) {
		obj, err := base(ctx)
		if err != nil {
			return nil, err
		}
		wrapper := reflect.New(kind)
		wrapper.Elem().Field(0).Set(reflect.ValueOf(obj).Elem())

		return wrapper.Interface().(Service), nil
	}
}

// Set of services all wrapping the base InstrumentedService resulting in the
// same method signatures but different outer types.
type InstrumentedServiceA struct{ InstrumentedService }
type InstrumentedServiceB struct{ InstrumentedService }
type InstrumentedServiceC struct{ InstrumentedService }

func InstrumentedServiceMakerA(base ServiceConstructor) ServiceConstructor {
	return InstrumentingWrapperMaker(base, reflect.TypeOf(InstrumentedServiceA{}))
}

func InstrumentedServiceMakerB(base ServiceConstructor) ServiceConstructor {
	return InstrumentingWrapperMaker(base, reflect.TypeOf(InstrumentedServiceB{}))
}

func InstrumentedServiceMakerC(base ServiceConstructor) ServiceConstructor {
	return InstrumentingWrapperMaker(base, reflect.TypeOf(InstrumentedServiceC{}))
}

// Tests that registered services get started and stopped correctly.
func TestServiceLifeCycle(t *testing.T) {

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary data directory: %v", err)
	}
	defer os.RemoveAll(dir)

	var config = &Config{
		DataDir:   dir,
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}

	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	// Register a batch of life-cycle instrumented services
	services := map[string]InstrumentingWrapper{
		"A": InstrumentedServiceMakerA,
		"B": InstrumentedServiceMakerB,
		"C": InstrumentedServiceMakerC,
	}
	started := make(map[string]bool)
	stopped := make(map[string]bool)

	for id, maker := range services {
		id := id // Closure for the constructor
		constructor := func(*ServiceContext) (Service, error) {
			return &InstrumentedService{
				startHook: func() { started[id] = true },
				stopHook:  func() { stopped[id] = true },
			}, nil
		}
		if err := stack.Register(maker(constructor)); err != nil {
			t.Fatalf("service %s: registration failed: %v", id, err)
		}
	}
	// Start the node and check that all services are running
	if err := stack.Start(); err != nil {
		t.Fatalf("failed to start protocol stack: %v", err)
	}
	for id := range services {
		if !started[id] {
			t.Fatalf("service %s: freshly started service not running", id)
		}
		if stopped[id] {
			t.Fatalf("service %s: freshly started service already stopped", id)
		}
	}
	// Stop the node and check that all services have been stopped
	if err := stack.Stop(); err != nil {
		t.Fatalf("failed to stop protocol stack: %v", err)
	}
	for id := range services {
		if !stopped[id] {
			t.Fatalf("service %s: freshly terminated service still running", id)
		}
	}
}

// Tests that services are restarted cleanly as new instances.
func TestServiceRestarts(t *testing.T) {

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary data directory: %v", err)
	}
	defer os.RemoveAll(dir)

	var config = &Config{
		DataDir:   dir,
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}

	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	// Define a service that does not support restarts
	var (
		running bool
		started int
	)
	constructor := func(*ServiceContext) (Service, error) {
		running = false
		return &InstrumentedService{
			startHook: func() {
				if running {
					panic("already running")
				}
				running = true
				started++
			},
		}, nil
	}
	// Register the service and start the protocol stack
	if err := stack.Register(constructor); err != nil {
		t.Fatalf("failed to register the service: %v", err)
	}
	if err := stack.Start(); err != nil {
		t.Fatalf("failed to start protocol stack: %v", err)
	}
	defer stack.Stop()

	if !running || started != 1 {
		t.Fatalf("running/started mismatch: have %v/%d, want true/1", running, started)
	}
	// Restart the stack a few times and check successful service restarts
	for i := 0; i < 3; i++ {
		if err := stack.Restart(); err != nil {
			t.Fatalf("iter %d: failed to restart stack: %v", i, err)
		}
	}
	if !running || started != 4 {
		t.Fatalf("running/started mismatch: have %v/%d, want true/4", running, started)
	}
}

// Tests that if a service fails to initialize itself, none of the other services
// will be allowed to even start.
func TestServiceConstructionAbortion(t *testing.T) {

	var config = &Config{
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}
	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	// Define a batch of good services
	services := map[string]InstrumentingWrapper{
		"A": InstrumentedServiceMakerA,
		"B": InstrumentedServiceMakerB,
		"C": InstrumentedServiceMakerC,
	}
	started := make(map[string]bool)
	for id, maker := range services {
		id := id // Closure for the constructor
		constructor := func(*ServiceContext) (Service, error) {
			return &InstrumentedService{
				startHook: func() { started[id] = true },
			}, nil
		}
		if err := stack.Register(maker(constructor)); err != nil {
			t.Fatalf("service %s: registration failed: %v", id, err)
		}
	}
	// Register a service that fails to construct itself
	failure := errors.New("fail")
	failer := func(*ServiceContext) (Service, error) {
		return nil, failure
	}
	if err := stack.Register(failer); err != nil {
		t.Fatalf("failer registration failed: %v", err)
	}
	// Start the protocol stack and ensure none of the services get started
	for i := 0; i < 100; i++ {
		if err := stack.Start(); err != failure {
			t.Fatalf("iter %d: stack startup failure mismatch: have %v, want %v", i, err, failure)
		}
		for id := range services {
			if started[id] {
				t.Fatalf("service %s: started should not have", id)
			}
			delete(started, id)
		}
	}
}

// Tests that if a service fails to start, all others started before it will be
// shut down.
func TestServiceStartupAbortion(t *testing.T) {

	var config = &Config{
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}

	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	// Register a batch of good services
	services := map[string]InstrumentingWrapper{
		"A": InstrumentedServiceMakerA,
		"B": InstrumentedServiceMakerB,
		"C": InstrumentedServiceMakerC,
	}
	started := make(map[string]bool)
	stopped := make(map[string]bool)

	for id, maker := range services {
		id := id // Closure for the constructor
		constructor := func(*ServiceContext) (Service, error) {
			return &InstrumentedService{
				startHook: func() { started[id] = true },
				stopHook:  func() { stopped[id] = true },
			}, nil
		}
		if err := stack.Register(maker(constructor)); err != nil {
			t.Fatalf("service %s: registration failed: %v", id, err)
		}
	}
	// Register a service that fails to start
	failure := errors.New("fail")
	failer := func(*ServiceContext) (Service, error) {
		return &InstrumentedService{
			start: failure,
		}, nil
	}
	if err := stack.Register(failer); err != nil {
		t.Fatalf("failer registration failed: %v", err)
	}
	// Start the protocol stack and ensure all started services stop
	for i := 0; i < 100; i++ {
		if err := stack.Start(); err != failure {
			t.Fatalf("iter %d: stack startup failure mismatch: have %v, want %v", i, err, failure)
		}
		for id := range services {
			if started[id] && !stopped[id] {
				t.Fatalf("service %s: started but not stopped", id)
			}
			delete(started, id)
			delete(stopped, id)
		}
	}
}

// Tests that even if a registered service fails to shut down cleanly, it does
// not influece the rest of the shutdown invocations.
func TestServiceTerminationGuarantee(t *testing.T) {

	var config = &Config{
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}
	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	// Register a batch of good services
	services := map[string]InstrumentingWrapper{
		"A": InstrumentedServiceMakerA,
		"B": InstrumentedServiceMakerB,
		"C": InstrumentedServiceMakerC,
	}
	started := make(map[string]bool)
	stopped := make(map[string]bool)

	for id, maker := range services {
		id := id // Closure for the constructor
		constructor := func(*ServiceContext) (Service, error) {
			return &InstrumentedService{
				startHook: func() { started[id] = true },
				stopHook:  func() { stopped[id] = true },
			}, nil
		}
		if err := stack.Register(maker(constructor)); err != nil {
			t.Fatalf("service %s: registration failed: %v", id, err)
		}
	}
	// Register a service that fails to shot down cleanly
	failure := errors.New("fail")
	failer := func(*ServiceContext) (Service, error) {
		return &InstrumentedService{
			stop: failure,
		}, nil
	}
	if err := stack.Register(failer); err != nil {
		t.Fatalf("failer registration failed: %v", err)
	}
	// Start the protocol stack, and ensure that a failing shut down terminates all
	for i := 0; i < 100; i++ {
		// Start the stack and make sure all is online
		if err := stack.Start(); err != nil {
			t.Fatalf("iter %d: failed to start protocol stack: %v", i, err)
		}
		for id := range services {
			if !started[id] {
				t.Fatalf("iter %d, service %s: service not running", i, id)
			}
			if stopped[id] {
				t.Fatalf("iter %d, service %s: service already stopped", i, id)
			}
		}

		// Stop the stack, verify failure and check all terminations
		err := stack.Stop()
		if err, ok := err.(*StopError); !ok {
			t.Fatalf("iter %d: termination failure mismatch: have %v, want StopError", i, err)
		} else {
			failer := reflect.TypeOf(&InstrumentedService{})
			if err.Services[failer] != failure {
				t.Fatalf("iter %d: failer termination failure mismatch: have %v, want %v", i, err.Services[failer], failure)
			}
			if len(err.Services) != 1 {
				t.Fatalf("iter %d: failure count mismatch: have %d, want %d", i, len(err.Services), 1)
			}
		}
		for id := range services {
			if !stopped[id] {
				t.Fatalf("iter %d, service %s: service not terminated", i, id)
			}
			delete(started, id)
			delete(stopped, id)
		}
	}
}

// TestServiceRetrieval tests that individual services can be retrieved.
func TestServiceRetrieval(t *testing.T) {

	// Create a simple stack and register two service types
	var config = &Config{
		Logger:    log.New(),
		Name:      "ft",
		P2PConfig: &p2p.Config{PrivateKey: testNodeKey},
	}
	stack, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	if err := stack.Register(NewNoopService); err != nil {
		t.Fatalf("noop service registration failed: %v", err)
	}
	if err := stack.Register(NewInstrumentedService); err != nil {
		t.Fatalf("instrumented service registration failed: %v", err)
	}
	// Make sure none of the services can be retrieved until started
	var noopServ *NoopService
	if err := stack.Service(&noopServ); err != ErrNodeStopped {
		t.Fatalf("noop service retrieval mismatch: have %v, want %v", err, ErrNodeStopped)
	}
	var instServ *InstrumentedService
	if err := stack.Service(&instServ); err != ErrNodeStopped {
		t.Fatalf("instrumented service retrieval mismatch: have %v, want %v", err, ErrNodeStopped)
	}
	// Start the stack and ensure everything is retrievable now
	if err := stack.Start(); err != nil {
		t.Fatalf("failed to start stack: %v", err)
	}
	defer stack.Stop()

	if err := stack.Service(&noopServ); err != nil {
		t.Fatalf("noop service retrieval mismatch: have %v, want %v", err, nil)
	}
	if err := stack.Service(&instServ); err != nil {
		t.Fatalf("instrumented service retrieval mismatch: have %v, want %v", err, nil)
	}
}
