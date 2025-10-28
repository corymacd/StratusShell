package server

import (
	"testing"
)

func TestPortPoolAllocation(t *testing.T) {
	pool := NewPortPool(8081, 8085)

	// Allocate all ports
	ports := make([]int, 0)
	for i := 0; i < 5; i++ {
		port, err := pool.Allocate()
		if err != nil {
			t.Fatalf("failed to allocate port %d: %v", i, err)
		}
		ports = append(ports, port)
	}

	// Should fail when exhausted
	_, err := pool.Allocate()
	if err == nil {
		t.Fatal("expected error when pool exhausted, got nil")
	}

	// Release and reallocate
	pool.Release(ports[0])
	port, err := pool.Allocate()
	if err != nil {
		t.Fatalf("failed to reallocate: %v", err)
	}
	if port != ports[0] {
		t.Fatalf("expected port %d, got %d", ports[0], port)
	}
}

func TestPortPoolConcurrency(t *testing.T) {
	pool := NewPortPool(9000, 9010)

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			port, err := pool.Allocate()
			if err != nil {
				t.Errorf("concurrent allocation failed: %v", err)
			}
			pool.Release(port)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
