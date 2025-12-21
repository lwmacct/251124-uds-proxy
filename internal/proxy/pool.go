package proxy

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// ClientPool 管理针对不同 Unix 域套接字的 HTTP 客户端池。
// 它提供线程安全的客户端管理，支持自动创建客户端和连接复用以提高性能。
//
// 每个 Unix 套接字路径都有独立的 HTTP 客户端，配置了专用的传输层用于
// Unix 套接字通信。客户端在首次访问时延迟创建，并缓存供后续请求使用。
//
// 此类型支持多个 goroutine 并发安全使用。
type ClientPool struct {
	clients      map[string]*http.Client
	mu           sync.RWMutex
	maxConns     int
	maxIdleConns int
	timeout      time.Duration
}

// NewClientPool 创建一个新的客户端池，使用指定的连接限制和超时设置。
//
// 参数：
//   - maxConns: 每个 Unix 套接字的最大总连接数
//   - maxIdleConns: 每个 Unix 套接字的最大空闲（保活）连接数
//   - timeout: 池中所有客户端的请求超时时间
//
// 返回的池可以立即使用。
func NewClientPool(maxConns, maxIdleConns int, timeout time.Duration) *ClientPool {
	return &ClientPool{
		clients:      make(map[string]*http.Client),
		maxConns:     maxConns,
		maxIdleConns: maxIdleConns,
		timeout:      timeout,
	}
}

// GetClient 返回针对指定 Unix 套接字路径配置的 HTTP 客户端。
// 如果该路径的客户端已存在，则从缓存中返回。
// 否则，创建一个新客户端，配置专用传输层用于 Unix 套接字通信。
//
// 此方法使用双重检查锁定来最小化锁竞争，同时确保线程安全。
// 返回的客户端可以并发使用。
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
			dialer := net.Dialer{Timeout: p.timeout}

			return dialer.DialContext(ctx, "unix", socketPath)
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

// RemoveClient 移除并关闭指定套接字路径的 HTTP 客户端。
// 当发生连接错误时应调用此方法，以便在下次请求时强制创建新客户端。
// 所有空闲连接都会被关闭。
func (p *ClientPool) RemoveClient(socketPath string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[socketPath]; exists {
		client.CloseIdleConnections()
		delete(p.clients, socketPath)
	}
}

// CloseAll 关闭池中所有 HTTP 客户端并清空客户端缓存。
// 应在服务器关闭时调用此方法以释放所有资源。
// 调用 CloseAll 后，池仍可继续使用，会根据需要创建新客户端。
func (p *ClientPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.CloseIdleConnections()
	}

	p.clients = make(map[string]*http.Client)
}
