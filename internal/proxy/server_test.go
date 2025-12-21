package proxy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lwmacct/251124-uds-proxy/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewServer 测试创建服务器
func TestNewServer(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "默认配置",
			config: &config.Config{
				Host:         "127.0.0.1",
				Port:         8080,
				Timeout:      30000,
				MaxConns:     100,
				MaxIdleConns: 10,
			},
		},
		{
			name: "自动分配端口",
			config: &config.Config{
				Host:         "127.0.0.1",
				Port:         0,
				Timeout:      30000,
				MaxConns:     100,
				MaxIdleConns: 10,
			},
		},
		{
			name: "最小配置",
			config: &config.Config{
				Host: "127.0.0.1",
			},
		},
		{
			name:   "零值配置",
			config: &config.Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(tt.config)

			require.NoError(t, err)
			require.NotNil(t, server)
			assert.Equal(t, tt.config, server.config)
			assert.NotNil(t, server.pool)
		})
	}
}

// TestServer_getAvailablePort 测试获取可用端口
func TestServer_getAvailablePort(t *testing.T) {
	t.Run("返回配置的端口", func(t *testing.T) {
		server, _ := NewServer(&config.Config{
			Host: "127.0.0.1",
			Port: 8888,
		})

		port, err := server.getAvailablePort()

		require.NoError(t, err)
		assert.Equal(t, 8888, port)
	})

	t.Run("自动分配可用端口", func(t *testing.T) {
		server, _ := NewServer(&config.Config{
			Host: "127.0.0.1",
			Port: 0,
		})

		port, err := server.getAvailablePort()

		require.NoError(t, err)
		assert.Positive(t, port, "端口号应大于 0")
		assert.Less(t, port, 65536, "端口号应小于 65536")
	})
}

// TestServer_writePortInfo 测试写入端口信息
func TestServer_writePortInfo(t *testing.T) {
	t.Run("不配置端口文件时不写入", func(t *testing.T) {
		server, _ := NewServer(&config.Config{
			Host:     "127.0.0.1",
			Port:     8080,
			PortFile: "",
		})
		server.actualPort = 8080

		err := server.writePortInfo()

		assert.NoError(t, err)
	})

	t.Run("写入端口到文件", func(t *testing.T) {
		tmpDir := t.TempDir()
		portFile := filepath.Join(tmpDir, "port.txt")

		server, _ := NewServer(&config.Config{
			Host:     "127.0.0.1",
			Port:     8080,
			PortFile: portFile,
		})
		server.actualPort = 12345

		err := server.writePortInfo()

		require.NoError(t, err)

		content, err := os.ReadFile(portFile) //nolint:gosec // test file with controlled path
		require.NoError(t, err)
		assert.Equal(t, "12345\n", string(content))
	})
}

// TestServer_accessLogMiddleware 测试访问日志中间件
func TestServer_accessLogMiddleware(t *testing.T) {
	server, _ := NewServer(&config.Config{
		Host: "127.0.0.1",
		Port: 0,
	})

	// 创建一个简单的处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 包装中间件
	wrapped := server.accessLogMiddleware(handler)

	t.Run("正常请求", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("不同状态码", func(t *testing.T) {
		statusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		wrappedStatus := server.accessLogMiddleware(statusHandler)

		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		rec := httptest.NewRecorder()

		wrappedStatus.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// TestResponseWriter 测试响应写入器
func TestResponseWriter(t *testing.T) {
	t.Run("默认状态码为 200 且可正常写入", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			statusCode:     http.StatusOK,
		}

		// 验证初始状态码
		assert.Equal(t, http.StatusOK, rw.statusCode)

		// 验证可以正常写入响应体
		n, err := rw.Write([]byte("test"))
		require.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "test", rec.Body.String())
	})

	t.Run("WriteHeader 设置状态码", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, rw.statusCode)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// TestServer_Shutdown 测试服务器关闭
func TestServer_Shutdown(t *testing.T) {
	t.Run("关闭未启动的服务器", func(t *testing.T) {
		server, _ := NewServer(&config.Config{
			Host: "127.0.0.1",
			Port: 0,
		})

		// 关闭未启动的服务器不应 panic
		assert.NotPanics(t, func() {
			server.Shutdown()
		})
	})

	t.Run("关闭时清理端口文件", func(t *testing.T) {
		tmpDir := t.TempDir()
		portFile := filepath.Join(tmpDir, "port.txt")

		// 创建端口文件
		err := os.WriteFile(portFile, []byte("8080\n"), 0600)
		require.NoError(t, err)

		server, _ := NewServer(&config.Config{
			Host:     "127.0.0.1",
			Port:     0,
			PortFile: portFile,
		})

		server.Shutdown()

		// 验证端口文件已被删除
		_, err = os.Stat(portFile)
		assert.True(t, os.IsNotExist(err), "端口文件应已被删除")
	})
}

// TestServer_PoolIntegration 测试服务器与连接池集成
func TestServer_PoolIntegration(t *testing.T) {
	server, err := NewServer(&config.Config{
		Host:         "127.0.0.1",
		Port:         0,
		Timeout:      1000,
		MaxConns:     50,
		MaxIdleConns: 5,
	})

	require.NoError(t, err)
	require.NotNil(t, server.pool)

	// 验证连接池配置正确
	assert.Equal(t, 50, server.pool.maxConns)
	assert.Equal(t, 5, server.pool.maxIdleConns)
	assert.Equal(t, time.Duration(1000)*time.Millisecond, server.pool.timeout)
}
