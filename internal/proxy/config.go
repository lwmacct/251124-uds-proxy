package proxy

// Config holds the server configuration
type Config struct {
	Host         string // Listen host address
	Port         int    // Listen port (0 for auto-assign)
	PortFile     string // File to write actual port
	Timeout      int    // Request timeout in milliseconds
	MaxConns     int    // Maximum connections per socket
	MaxIdleConns int    // Maximum idle connections per socket
	NoAccessLog  bool   // Disable access logging
}
