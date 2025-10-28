package server

import (
	"net"
	"sync"
)

type PortPool struct {
	used map[int]bool
	mu   sync.Mutex
}

func NewPortPool(minPort, maxPort int) *PortPool {
	// minPort and maxPort are ignored now, keeping signature for compatibility
	return &PortPool{
		used: make(map[int]bool),
	}
}

// AllocateEphemeral allocates an OS-assigned ephemeral port by listening on port 0
func (p *PortPool) AllocateEphemeral() (int, error) {
	// Listen on port 0 to let OS assign an ephemeral port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	
	// Get the assigned port
	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port
	
	// Close the listener - there's a small race window here where another process
	// could claim the port before GoTTY binds to it. However, this is acceptable because:
	// 1. The window is extremely small (microseconds)
	// 2. GoTTY will return an error if binding fails, which we handle in SpawnTerminal
	// 3. The OS port assignment makes collisions unlikely in practice
	listener.Close()
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Mark as used
	p.used[port] = true
	return port, nil
}

// Allocate is kept for backward compatibility but now uses ephemeral ports
func (p *PortPool) Allocate() (int, error) {
	return p.AllocateEphemeral()
}

func (p *PortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.used, port)
}

func (p *PortPool) IsAllocated(port int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.used[port]
}
