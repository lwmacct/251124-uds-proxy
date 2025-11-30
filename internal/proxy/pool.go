package proxy

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// ClientPool manages HTTP clients for different Unix sockets
type ClientPool struct {
	clients      map[string]*http.Client
	mu           sync.RWMutex
	maxConns     int
	maxIdleConns int
	timeout      time.Duration
}

// NewClientPool creates a new client pool
func NewClientPool(maxConns, maxIdleConns int, timeout time.Duration) *ClientPool {
	return &ClientPool{
		clients:      make(map[string]*http.Client),
		maxConns:     maxConns,
		maxIdleConns: maxIdleConns,
		timeout:      timeout,
	}
}

// GetClient returns an HTTP client for the given socket path
func (p *ClientPool) GetClient(socketPath string) *http.Client {
	p.mu.RLock()
	client, exists := p.clients[socketPath]
	p.mu.RUnlock()

	if exists {
		return client
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists = p.clients[socketPath]; exists {
		return client
	}

	// Create new client with Unix socket transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, p.timeout)
		},
		MaxConnsPerHost:     p.maxConns,
		MaxIdleConnsPerHost: p.maxIdleConns,
		IdleConnTimeout:     90 * time.Second,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   p.timeout,
	}

	p.clients[socketPath] = client
	return client
}

// RemoveClient removes and closes a client for the given socket path
func (p *ClientPool) RemoveClient(socketPath string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[socketPath]; exists {
		client.CloseIdleConnections()
		delete(p.clients, socketPath)
	}
}

// CloseAll closes all clients in the pool
func (p *ClientPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.CloseIdleConnections()
	}
	p.clients = make(map[string]*http.Client)
}
