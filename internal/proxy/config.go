package proxy

// Config 保存服务器配置参数。
//
// 所有字段在零值时都有合理的默认行为：
//   - Host 为空时绑定到所有网络接口
//   - Port=0 时启用自动端口分配
//   - Timeout=0 时使用默认的 HTTP 客户端超时
type Config struct {
	// Host 指定监听的网络接口地址。
	// 使用 "127.0.0.1" 仅监听本地，使用 "0.0.0.0" 监听所有接口。
	Host string

	// Port 指定监听的 TCP 端口。
	// 设置为 0 时自动分配可用端口。
	Port int

	// PortFile 指定写入实际端口号的文件路径。
	// 当 Port 设为 0 进行自动分配时特别有用。
	// 服务器关闭时会自动清理此文件。
	PortFile string

	// Timeout 指定请求超时时间，单位为毫秒。
	// 此超时同时应用于 Unix 套接字连接和请求本身。
	Timeout int

	// MaxConns 指定每个 Unix 套接字的最大连接数。
	// 用于限制到每个后端套接字的总并发连接数。
	MaxConns int

	// MaxIdleConns 指定每个 Unix 套接字的最大空闲连接数。
	// 空闲连接保持活跃状态以便复用，提高性能。
	MaxIdleConns int

	// NoAccessLog 设为 true 时禁用访问日志中间件。
	// 默认情况下，所有请求都会记录方法、路径、状态码和耗时。
	NoAccessLog bool
}
