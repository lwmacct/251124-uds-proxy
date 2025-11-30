---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "uds-proxy"
  text: "Unix Domain Socket Proxy"
  tagline: 高性能 HTTP 到 Unix Domain Socket 代理服务
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/quick-start
    - theme: alt
      text: API 文档
      link: /guide/api

features:
  - title: 高性能代理
    details: 基于 Go 标准库 net/http 构建，支持所有 HTTP 方法的透明代理
  - title: 连接池管理
    details: 智能连接池自动管理 Unix Socket 连接，支持并发安全和自动重连
  - title: 灵活配置
    details: 支持命令行参数、自动端口分配、可配置超时和连接数限制
---

## 特性概览

- **完整 HTTP 支持**：代理 GET、POST、PUT、DELETE、PATCH、HEAD、OPTIONS 等所有方法
- **查询参数转发**：自动转发所有查询参数到目标 Unix Socket
- **请求头保留**：完整保留并转发原始请求头
- **健康检查**：内置 `/health` 端点，便于服务监控
- **访问日志**：可配置的请求日志记录
- **错误处理**：完善的错误码和错误信息返回

## 典型使用场景

### Docker API 代理

通过 HTTP 接口访问 Docker Unix Socket API：

```bash
# 启动代理
uds-proxy

# 获取 Docker 版本
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/version"

# 列出所有容器
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/json?all=true"
```

### 自定义服务代理

代理任意 Unix Socket 服务：

```bash
# 代理到自定义 socket
curl "http://localhost:8080/proxy?path=/tmp/my-service.sock&url=/api/status"
```
