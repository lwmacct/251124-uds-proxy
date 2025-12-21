package proxy

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/lwmacct/251207-go-pkg-version/pkg/version"
)

// handleRoot 处理根路径请求，返回服务信息。
// 响应包含服务名称、版本、描述和使用示例。
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	info := map[string]any{
		"service":     "uds-proxy",
		"version":     version.GetVersion(),
		"description": "HTTP server that proxies requests to Unix domain sockets",
		"usage":       "GET /proxy?path=/var/run/docker.sock&url=/containers/json",
		"examples": map[string]string{
			"获取容器列表": "GET /proxy?path=/var/run/docker.sock&url=/containers/json",
			"获取镜像列表": "GET /proxy?path=/var/run/docker.sock&url=/images/json",
			"获取系统信息": "GET /proxy?path=/var/run/docker.sock&url=/info",
			"获取版本信息": "GET /proxy?path=/var/run/docker.sock&url=/version",
		},
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(info); err != nil {
		slog.Error("JSON编码失败", "error", err)
	}
}

// handleHealth 处理健康检查请求。
// 返回 JSON 格式的健康状态信息，用于负载均衡器或监控系统。
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "uds-proxy",
	}); err != nil {
		slog.Error("JSON编码失败", "error", err)
	}
}

// handleProxy 是核心代理处理函数，将 HTTP 请求转发到 Unix 域套接字。
//
// 请求参数：
//   - path: (必需) Unix 套接字文件路径，如 /var/run/docker.sock
//   - url: (可选) 目标 URL 路径，默认为 "/"
//   - method: (可选) HTTP 方法，默认使用请求本身的方法
//
// 其他查询参数会被透传到后端请求。请求头（除 hop-by-hop 头）也会被复制。
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Get socket path from query parameter
	socketPath := r.URL.Query().Get("path")
	if socketPath == "" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Verify socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		slog.Warn("socket文件不存在", "path", socketPath)
		w.WriteHeader(http.StatusBadGateway)

		return
	}

	// Get target URL path
	targetPath := r.URL.Query().Get("url")
	if targetPath == "" {
		targetPath = "/"
	}

	// Get HTTP method (allow override via query parameter)
	method := r.URL.Query().Get("method")
	if method == "" {
		method = r.Method
	}

	method = strings.ToUpper(method)

	// Build query parameters (excluding proxy-specific ones)
	queryParams := url.Values{}

	for key, values := range r.URL.Query() {
		if key != "path" && key != "url" && key != "method" {
			for _, v := range values {
				queryParams.Add(key, v)
			}
		}
	}

	// Build full target URL
	targetURL := "http://localhost" + targetPath
	if len(queryParams) > 0 {
		targetURL += "?" + queryParams.Encode()
	}

	slog.Debug("代理请求", "method", method, "url", targetURL, "socket", socketPath)

	// Create backend request
	backendReq, err := http.NewRequestWithContext(r.Context(), method, targetURL, r.Body)
	if err != nil {
		slog.Error("创建请求失败", "error", err)
		w.WriteHeader(http.StatusBadGateway)

		return
	}

	// Copy headers (excluding hop-by-hop headers)
	for key, values := range r.Header {
		lowerKey := strings.ToLower(key)
		if lowerKey == "host" || lowerKey == "content-length" || lowerKey == "transfer-encoding" {
			continue
		}

		for _, v := range values {
			backendReq.Header.Add(key, v)
		}
	}

	// Get client from pool and make request
	client := s.pool.GetClient(socketPath)

	resp, err := client.Do(backendReq)
	if err != nil {
		// Remove client from pool on connection error
		s.pool.RemoveClient(socketPath)

		if os.IsTimeout(err) {
			slog.Warn("请求超时", "socket", socketPath, "error", err)
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			slog.Warn("连接失败", "socket", socketPath, "error", err)
			w.WriteHeader(http.StatusBadGateway)
		}

		return
	}

	defer func() { _ = resp.Body.Close() }()

	// Copy response headers
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	// Write status code and body
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
