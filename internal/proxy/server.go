package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// Server represents the HTTP proxy server
type Server struct {
	config     Config
	httpServer *http.Server
	pool       *ClientPool
	actualPort int
}

// NewServer creates a new proxy server instance
func NewServer(cfg Config) (*Server, error) {
	s := &Server{
		config: cfg,
		pool:   NewClientPool(cfg.MaxConns, cfg.MaxIdleConns, time.Duration(cfg.Timeout)*time.Millisecond),
	}
	return s, nil
}

// Run starts the HTTP server
func (s *Server) Run() error {
	// Get available port
	port, err := s.getAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to get available port: %w", err)
	}
	s.actualPort = port

	// Write port to file
	if err := s.writePortInfo(); err != nil {
		log.Printf("Warning: failed to write port file: %v", err)
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
	fmt.Printf("PORT=%d\n", s.actualPort)
	log.Printf("uds-proxy server starting on %s", addr)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}

	s.pool.CloseAll()

	// Clean up port file
	if s.config.PortFile != "" {
		os.Remove(s.config.PortFile)
	}

	log.Println("Server shutdown complete")
}

func (s *Server) getAvailablePort() (int, error) {
	if s.config.Port != 0 {
		return s.config.Port, nil
	}

	// Find available port
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:0", s.config.Host))
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func (s *Server) writePortInfo() error {
	if s.config.PortFile == "" {
		return nil
	}

	return os.WriteFile(s.config.PortFile, []byte(fmt.Sprintf("%d\n", s.actualPort)), 0644)
}

func (s *Server) accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
