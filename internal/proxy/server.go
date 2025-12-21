package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lwmacct/251124-uds-proxy/internal/config"
)

// Server 表示 HTTP 代理服务器实例。
// 它管理 HTTP 服务器、客户端连接池和服务器生命周期。
type Server struct {
	config     *config.Config
	httpServer *http.Server
	pool       *ClientPool
	actualPort int
}

// NewServer 创建一个新的代理服务器实例。
// 它使用提供的配置初始化服务器和客户端连接池。
func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		config: cfg,
		pool:   NewClientPool(cfg.MaxConns, cfg.MaxIdleConns, time.Duration(cfg.Timeout)*time.Millisecond),
	}

	return s, nil
}

// Run 启动 HTTP 服务器。
// 它会设置路由、启动监听，并阻塞直到服务器关闭。
// 如果配置中 Port 为 0，会自动分配可用端口。
func (s *Server) Run() error {
	// Get available port
	port, err := s.getAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to get available port: %w", err)
	}

	s.actualPort = port

	// Write port to file
	if err := s.writePortInfo(); err != nil {
		slog.Warn("写入端口文件失败", "error", err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/proxy", s.handleProxy)

	var handler http.Handler = mux
	if !s.config.NoAccessLog {
		handler = s.accessLogMiddleware(mux)
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.actualPort)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Print startup info
	slog.Info("PORT", "port", s.actualPort)
	slog.Info("服务器启动", "addr", addr)

	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅地关闭服务器。
// 它会等待正在处理的请求完成（最多 5 秒），关闭所有客户端连接，
// 并清理端口文件（如果配置了的话）。
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			slog.Warn("服务器关闭时出错", "error", err)
		}
	}

	s.pool.CloseAll()

	// Clean up port file
	if s.config.PortFile != "" {
		_ = os.Remove(s.config.PortFile)
	}

	slog.Info("服务器关闭完成")
}

// getAvailablePort 返回可用的端口号。
// 如果配置中指定了端口，则返回该端口；否则自动查找可用端口。
func (s *Server) getAvailablePort() (int, error) {
	if s.config.Port != 0 {
		return s.config.Port, nil
	}

	// Find available port
	lc := net.ListenConfig{}

	listener, err := lc.Listen(context.Background(), "tcp", s.config.Host+":0")
	if err != nil {
		return 0, err
	}

	defer func() { _ = listener.Close() }()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected address type: %T", listener.Addr())
	}

	return tcpAddr.Port, nil
}

// writePortInfo 将实际端口号写入配置的端口文件。
func (s *Server) writePortInfo() error {
	if s.config.PortFile == "" {
		return nil
	}

	return os.WriteFile(s.config.PortFile, fmt.Appendf(nil, "%d\n", s.actualPort), 0600)
}

// accessLogMiddleware 返回一个访问日志中间件。
// 它记录每个请求的方法、路径、状态码和处理时间。
func (s *Server) accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		slog.Info("access",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start),
		)
	})
}

// responseWriter 是一个包装 http.ResponseWriter 的结构体，
// 用于捕获响应状态码以便记录日志。
type responseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
