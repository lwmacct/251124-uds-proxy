// Package config 提供应用配置管理。
//
// 配置加载优先级 (从低到高)：
//  1. 默认值 - DefaultConfig() 函数中定义
//  2. 配置文件 - 按 configPaths 顺序搜索
//  3. CLI flags - 最高优先级
package config

// Config UDS 代理服务配置
type Config struct {
	Host         string `koanf:"host" comment:"监听地址，如 '127.0.0.1' 仅本地，'0.0.0.0' 所有接口"`
	Port         int    `koanf:"port" comment:"监听端口，0 表示自动分配"`
	PortFile     string `koanf:"port_file" comment:"写入实际端口号的文件路径"`
	Timeout      int    `koanf:"timeout" comment:"请求超时时间 (毫秒)"`
	MaxConns     int    `koanf:"max_conns" comment:"每个 Unix 套接字的最大连接数"`
	MaxIdleConns int    `koanf:"max_idle_conns" comment:"每个 Unix 套接字的最大空闲连接数"`
	NoAccessLog  bool   `koanf:"no_access_log" comment:"禁用访问日志"`
}

// DefaultConfig 返回默认配置
// 这是配置默认值的唯一来源 (Single Source of Truth)
// CLI flags 从此函数读取默认值，--help 显示与代码自动一致
func DefaultConfig() Config {
	return Config{
		Host:         "127.0.0.1",
		Port:         0,
		PortFile:     "/tmp/uds-proxy.port",
		Timeout:      10000,
		MaxConns:     10,
		MaxIdleConns: 5,
		NoAccessLog:  false,
	}
}
