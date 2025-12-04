// Package proxy 提供一个高性能的 HTTP 代理服务器，用于将请求转发到 Unix 域套接字。
// 常用于将 Unix 套接字服务（如 Docker API）通过 HTTP 暴露出去。
//
// # 功能特性
//
// 本包包含以下功能：
//   - 连接池，实现客户端复用以提高性能
//   - 可配置的超时和连接数限制
//   - 访问日志中间件
//   - 健康检查和服务信息端点
//
// # 使用示例
//
// 创建并启动一个代理服务器：
//
//	cfg := &config.Config{
//	    Host:         "127.0.0.1",
//	    Port:         8080,
//	    Timeout:      30000,
//	    MaxConns:     100,
//	    MaxIdleConns: 10,
//	}
//	server, err := proxy.NewServer(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := server.Run(); err != nil {
//	    log.Fatal(err)
//	}
//
// # API 端点
//
// 服务器提供以下端点：
//   - GET /         - 返回服务信息和使用说明
//   - GET /health   - 健康检查端点
//   - GET /proxy    - 代理请求到 Unix 套接字
//
// 代理端点参数：
//   - path   (必需) Unix 套接字文件路径
//   - url    (可选) 目标 URL 路径，默认为 "/"
//   - method (可选) HTTP 方法，默认使用请求本身的方法
//
// 示例请求：
//
//	GET /proxy?path=/var/run/docker.sock&url=/containers/json
package proxy
