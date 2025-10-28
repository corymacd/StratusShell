package server

import (
	"testing"
)

func TestPortPoolAllocation(t *testing.T) {
	pool := NewPortPool(0, 0) // minPort/maxPort no longer used

	// Allocate ephemeral ports - OS assigns these dynamically
	ports := make([]int, 0)
	for i := 0; i < 5; i++ {
		port, err := pool.Allocate()
		if err != nil {
			t.Fatalf("failed to allocate port %d: %v", i, err)
		}
		if port <= 0 {
			t.Fatalf("expected valid port, got %d", port)
		}
		ports = append(ports, port)
	}

	// All ports should be unique
	seen := make(map[int]bool)
	for _, port := range ports {
		if seen[port] {
			t.Fatalf("duplicate port allocated: %d", port)
		}
		seen[port] = true
	}

	// Verify ports are tracked as allocated
	for _, port := range ports {
		if !pool.IsAllocated(port) {
			t.Fatalf("port %d should be allocated", port)
		}
	}

	// Release and verify
	pool.Release(ports[0])
	if pool.IsAllocated(ports[0]) {
		t.Fatalf("port %d should be released", ports[0])
	}
}

func TestPortPoolConcurrency(t *testing.T) {
	pool := NewPortPool(0, 0)

	done := make(chan bool)
	errors := make(chan error, 5)
	
	for i := 0; i < 5; i++ {
		go func() {
			port, err := pool.Allocate()
			if err != nil {
				errors <- err
				done <- false
				return
			}
			if port <= 0 {
				t.Errorf("got invalid port: %d", port)
			}
			pool.Release(port)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		success := <-done
		if !success {
			err := <-errors
			t.Errorf("concurrent allocation failed: %v", err)
		}
	}
}
