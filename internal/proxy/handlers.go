package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// version 定义当前服务版本号。
const version = "0.1.0"

// handleRoot 处理根路径请求，返回服务信息。
// 响应包含服务名称、版本、描述和使用示例。
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]any{
		"service":     "uds-proxy",
		"version":     version,
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
	json.NewEncoder(w).Encode(info)
}

// handleHealth 处理健康检查请求。
// 返回 JSON 格式的健康状态信息，用于负载均衡器或监控系统。
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "uds-proxy",
	})
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
		s.errorResponse(w, http.StatusBadRequest, "缺少 path 参数")
		return
	}

	// Verify socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		s.errorResponse(w, http.StatusNotFound, fmt.Sprintf("Socket 文件不存在: %s", socketPath))
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

	log.Printf("Proxy request: %s %s -> %s", method, targetURL, socketPath)

	// Create backend request
	backendReq, err := http.NewRequestWithContext(r.Context(), method, targetURL, r.Body)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("创建请求失败: %v", err))
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
			s.errorResponse(w, http.StatusGatewayTimeout, fmt.Sprintf("请求超时: %v", err))
		} else {
			s.errorResponse(w, http.StatusServiceUnavailable, fmt.Sprintf("连接失败: %v", err))
		}
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	// Write status code and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// errorResponse 向客户端发送标准化的 JSON 错误响应。
func (s *Server) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"error":  true,
		"detail": message,
	})
}
