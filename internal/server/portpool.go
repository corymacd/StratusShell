package server

import (
	"errors"
	"sync"
)

type PortPool struct {
	minPort int
	maxPort int
	used    map[int]bool
	mu      sync.Mutex
}

func NewPortPool(minPort, maxPort int) *PortPool {
	return &PortPool{
		minPort: minPort,
		maxPort: maxPort,
		used:    make(map[int]bool),
	}
}

func (p *PortPool) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for port := p.minPort; port <= p.maxPort; port++ {
		if !p.used[port] {
			p.used[port] = true
			return port, nil
		}
	}
	return 0, errors.New("no available ports")
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
