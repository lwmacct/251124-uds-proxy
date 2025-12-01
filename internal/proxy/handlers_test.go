package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer 创建用于测试的服务器实例
func newTestServer() *Server {
	server, _ := NewServer(Config{
		Host:         "127.0.0.1",
		Port:         0,
		Timeout:      30000,
		MaxConns:     100,
		MaxIdleConns: 10,
	})
	return server
}

// TestServer_handleRoot 测试根路径处理
func TestServer_handleRoot(t *testing.T) {
	server := newTestServer()

	t.Run("根路径返回服务信息", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		server.handleRoot(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "uds-proxy", resp["service"])
		assert.Equal(t, version, resp["version"])
		assert.NotEmpty(t, resp["description"])
		assert.NotEmpty(t, resp["usage"])
		assert.NotNil(t, resp["examples"])
	})

	t.Run("非根路径返回 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
		rec := httptest.NewRecorder()

		server.handleRoot(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// TestServer_handleHealth 测试健康检查
func TestServer_handleHealth(t *testing.T) {
	server := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", resp["status"])
	assert.Equal(t, "uds-proxy", resp["service"])
}

// TestServer_handleProxy 测试代理处理
func TestServer_handleProxy(t *testing.T) {
	server := newTestServer()

	t.Run("缺少 path 参数返回 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
		rec := httptest.NewRecorder()

		server.handleProxy(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["error"].(bool))
		assert.Contains(t, resp["detail"], "path")
	})

	t.Run("socket 文件不存在返回 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/proxy?path=/nonexistent/socket.sock", nil)
		rec := httptest.NewRecorder()

		server.handleProxy(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var resp map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["error"].(bool))
		assert.Contains(t, resp["detail"], "不存在")
	})
}

// TestServer_errorResponse 测试错误响应
func TestServer_errorResponse(t *testing.T) {
	server := newTestServer()

	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			message:    "参数错误",
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			message:    "资源未找到",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			message:    "内部服务器错误",
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
			message:    "服务不可用",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			server.errorResponse(rec, tt.statusCode, tt.message)

			assert.Equal(t, tt.statusCode, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var resp map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.True(t, resp["error"].(bool))
			assert.Equal(t, tt.message, resp["detail"])
		})
	}
}

// TestServer_handleRoot_Methods 测试不同 HTTP 方法
func TestServer_handleRoot_Methods(t *testing.T) {
	server := newTestServer()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()

			server.handleRoot(rec, req)

			// 根路径应接受所有方法
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// TestServer_handleHealth_Methods 测试健康检查的不同 HTTP 方法
func TestServer_handleHealth_Methods(t *testing.T) {
	server := newTestServer()

	methods := []string{
		http.MethodGet,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			rec := httptest.NewRecorder()

			server.handleHealth(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
