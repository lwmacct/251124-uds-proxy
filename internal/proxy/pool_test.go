package proxy

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClientPool 测试创建新的客户端池
func TestNewClientPool(t *testing.T) {
	tests := []struct {
		name         string
		maxConns     int
		maxIdleConns int
		timeout      time.Duration
	}{
		{
			name:         "默认配置",
			maxConns:     100,
			maxIdleConns: 10,
			timeout:      30 * time.Second,
		},
		{
			name:         "零值配置",
			maxConns:     0,
			maxIdleConns: 0,
			timeout:      0,
		},
		{
			name:         "高并发配置",
			maxConns:     1000,
			maxIdleConns: 100,
			timeout:      5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewClientPool(tt.maxConns, tt.maxIdleConns, tt.timeout)

			require.NotNil(t, pool, "池不应为 nil")
			assert.NotNil(t, pool.clients, "clients map 不应为 nil")
			assert.Equal(t, tt.maxConns, pool.maxConns, "maxConns 应匹配")
			assert.Equal(t, tt.maxIdleConns, pool.maxIdleConns, "maxIdleConns 应匹配")
			assert.Equal(t, tt.timeout, pool.timeout, "timeout 应匹配")
		})
	}
}

// TestClientPool_GetClient 测试获取客户端
func TestClientPool_GetClient(t *testing.T) {
	pool := NewClientPool(100, 10, 30*time.Second)

	t.Run("获取新客户端", func(t *testing.T) {
		socketPath := "/var/run/test.sock"
		client := pool.GetClient(socketPath)

		require.NotNil(t, client, "客户端不应为 nil")
		assert.NotNil(t, client.Transport, "Transport 不应为 nil")
	})

	t.Run("获取缓存的客户端", func(t *testing.T) {
		socketPath := "/var/run/cached.sock"

		// 第一次获取
		client1 := pool.GetClient(socketPath)
		require.NotNil(t, client1)

		// 第二次获取应返回相同的客户端
		client2 := pool.GetClient(socketPath)
		require.NotNil(t, client2)

		assert.Same(t, client1, client2, "应返回相同的缓存客户端")
	})

	t.Run("不同路径返回不同客户端", func(t *testing.T) {
		client1 := pool.GetClient("/var/run/sock1.sock")
		client2 := pool.GetClient("/var/run/sock2.sock")

		require.NotNil(t, client1)
		require.NotNil(t, client2)
		assert.NotSame(t, client1, client2, "不同路径应返回不同客户端")
	})
}

// TestClientPool_GetClient_Concurrent 测试并发获取客户端
func TestClientPool_GetClient_Concurrent(t *testing.T) {
	pool := NewClientPool(100, 10, 30*time.Second)
	socketPath := "/var/run/concurrent.sock"

	var wg sync.WaitGroup

	clients := make(chan *struct{}, 100)

	// 启动 100 个 goroutine 并发获取同一个客户端
	for range 100 {
		wg.Go(func() {
			client := pool.GetClient(socketPath)
			if client != nil {
				clients <- &struct{}{}
			}
		})
	}

	wg.Wait()
	close(clients)

	// 验证所有 goroutine 都成功获取了客户端
	count := 0
	for range clients {
		count++
	}

	assert.Equal(t, 100, count, "所有 goroutine 都应成功获取客户端")

	// 验证只创建了一个客户端
	pool.mu.RLock()
	clientCount := len(pool.clients)
	pool.mu.RUnlock()
	assert.Equal(t, 1, clientCount, "应该只创建一个客户端")
}

// TestClientPool_RemoveClient 测试移除客户端
func TestClientPool_RemoveClient(t *testing.T) {
	pool := NewClientPool(100, 10, 30*time.Second)
	socketPath := "/var/run/remove.sock"

	t.Run("移除存在的客户端", func(t *testing.T) {
		// 先创建客户端
		_ = pool.GetClient(socketPath)

		pool.mu.RLock()
		_, exists := pool.clients[socketPath]
		pool.mu.RUnlock()
		require.True(t, exists, "客户端应存在")

		// 移除客户端
		pool.RemoveClient(socketPath)

		pool.mu.RLock()
		_, exists = pool.clients[socketPath]
		pool.mu.RUnlock()
		assert.False(t, exists, "客户端应已被移除")
	})

	t.Run("移除不存在的客户端不报错", func(t *testing.T) {
		// 移除不存在的客户端不应 panic
		assert.NotPanics(t, func() {
			pool.RemoveClient("/var/run/nonexistent.sock")
		})
	})
}

// TestClientPool_CloseAll 测试关闭所有客户端
func TestClientPool_CloseAll(t *testing.T) {
	pool := NewClientPool(100, 10, 30*time.Second)

	// 创建多个客户端
	paths := []string{
		"/var/run/sock1.sock",
		"/var/run/sock2.sock",
		"/var/run/sock3.sock",
	}
	for _, path := range paths {
		_ = pool.GetClient(path)
	}

	pool.mu.RLock()
	initialCount := len(pool.clients)
	pool.mu.RUnlock()
	require.Equal(t, 3, initialCount, "应有 3 个客户端")

	// 关闭所有客户端
	pool.CloseAll()

	pool.mu.RLock()
	finalCount := len(pool.clients)
	pool.mu.RUnlock()
	assert.Equal(t, 0, finalCount, "关闭后应没有客户端")
}

// TestClientPool_CloseAll_CanReusePool 测试关闭后池可继续使用
func TestClientPool_CloseAll_CanReusePool(t *testing.T) {
	pool := NewClientPool(100, 10, 30*time.Second)
	socketPath := "/var/run/reuse.sock"

	// 创建客户端
	client1 := pool.GetClient(socketPath)
	require.NotNil(t, client1)

	// 关闭所有
	pool.CloseAll()

	// 再次获取客户端
	client2 := pool.GetClient(socketPath)
	require.NotNil(t, client2, "关闭后应能创建新客户端")
	assert.NotSame(t, client1, client2, "应是新创建的客户端")
}
