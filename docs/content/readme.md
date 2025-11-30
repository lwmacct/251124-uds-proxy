# uds-proxy

一个高性能的 HTTP 到 Unix Domain Socket 代理服务器，使用 Go 语言构建。

## 功能特性

- 🚀 高性能异步代理，基于 Go 标准库 `net/http`
- 🔌 HTTP 请求代理到 Unix Socket（如 Docker API）
- 🔄 支持所有 HTTP 方法（GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS）
- 📊 内置健康检查和服务信息端点
- 🔗 完整的查询参数和请求头转发
- 🌐 连接池管理和自动重连
- 📝 可配置的访问日志

## 快速开始

### 安装

```bash
# 从源码构建
go build -o bin/uds-proxy ./cmd/uds-proxy

# 或直接运行
go run ./cmd/uds-proxy
```

### 运行

```bash
# 自动分配端口（默认写入 /tmp/uds-proxy.port）
bin/uds-proxy

# 指定端口和主机
bin/uds-proxy --host 127.0.0.1 --port 8080

# 禁用访问日志
bin/uds-proxy --port 8080 --no-access-log
```

### 基本用法

```bash
# 获取服务端口（如果使用了自动分配）
PORT=$(cat /tmp/uds-proxy.port)

# 代理请求到 Docker socket
curl "http://127.0.0.1:$PORT/proxy?path=/var/run/docker.sock&url=/version"

# 健康检查
curl "http://127.0.0.1:$PORT/health"

# 服务信息
curl "http://127.0.0.1:$PORT/"
```

## API 端点

### GET /

返回服务信息和使用示例。

**响应示例：**

```json
{
  "service": "uds-proxy",
  "version": "0.1.0",
  "description": "HTTP server that proxies requests to Unix domain sockets",
  "usage": "GET /proxy?path=/var/run/docker.sock&url=/containers/json"
}
```

### GET /health

健康检查端点，返回服务状态。

**响应示例：**

```json
{
  "status": "healthy",
  "service": "uds-proxy"
}
```

### [ALL METHODS] /proxy

核心代理端点，将 HTTP 请求转发到 Unix socket。

**必需参数：**

| 参数   | 说明                 |
| ------ | -------------------- |
| `path` | Unix socket 文件路径 |

**可选参数：**

| 参数     | 说明           | 默认值   |
| -------- | -------------- | -------- |
| `url`    | 目标 URL 路径  | `/`      |
| `method` | 指定 HTTP 方法 | 请求方法 |

其他查询参数会自动转发到目标服务。

## 使用示例

### Docker API 代理

```bash
# 获取 Docker 版本信息
curl "http://127.0.0.1:8080/proxy?path=/var/run/docker.sock&url=/version"

# 列出所有容器
curl "http://127.0.0.1:8080/proxy?path=/var/run/docker.sock&url=/containers/json"

# 列出运行中的容器
curl "http://127.0.0.1:8080/proxy?path=/var/run/docker.sock&url=/containers/json&all=false"

# 获取系统信息
curl "http://127.0.0.1:8080/proxy?path=/var/run/docker.sock&url=/info"

# 列出镜像
curl "http://127.0.0.1:8080/proxy?path=/var/run/docker.sock&url=/images/json"

# POST 请求 - 创建容器
curl -X POST "http://127.0.0.1:8080/proxy?path=/var/run/docker.sock&url=/containers/create" \
  -H "Content-Type: application/json" \
  -d '{"Image":"alpine","Cmd":["echo","hello"]}'
```

### 其他 Unix Socket 服务

```bash
# 代理到自定义服务
curl "http://127.0.0.1:8080/proxy?path=/tmp/myservice.sock&url=/api/status"

# 带查询参数的请求
curl "http://127.0.0.1:8080/proxy?path=/tmp/service.sock&url=/api/search&q=test&limit=10"
```

## 命令行参数

| 参数               | 短名 | 默认值                | 说明                       |
| ------------------ | ---- | --------------------- | -------------------------- |
| `--host`           | `-H` | `127.0.0.1`           | 监听主机地址               |
| `--port`           | `-p` | `0`                   | 监听端口（0 为自动分配）   |
| `--port-file`      |      | `/tmp/uds-proxy.port` | 端口号写入文件             |
| `--timeout`        |      | `10000`               | 请求超时（毫秒）           |
| `--max-conns`      |      | `10`                  | 每个 socket 最大连接数     |
| `--max-idle-conns` |      | `5`                   | 每个 socket 最大空闲连接数 |
| `--no-access-log`  |      | `false`               | 禁用访问日志               |
| `--version`        | `-v` |                       | 打印版本号                 |
| `--help`           | `-h` |                       | 显示帮助信息               |

## 错误处理

| 状态码 | 说明              |
| ------ | ----------------- |
| 200    | 请求成功          |
| 400    | 缺少 path 参数    |
| 404    | Socket 文件不存在 |
| 500    | 内部服务器错误    |
| 503    | 连接失败          |
| 504    | 请求超时          |

## 项目结构

```
251124-uds-proxy/
├── cmd/
│   └── uds-proxy/
│       └── main.go          # CLI 入口
├── internal/
│   └── proxy/
│       ├── config.go        # 配置结构体
│       ├── server.go        # HTTP 服务器
│       ├── handlers.go      # 路由处理器
│       └── pool.go          # 连接池管理
├── go.mod
└── go.sum
```

## 开发

### 初始化开发环境

```bash
pre-commit install
```

### 常用命令

```bash
# 查看所有可用任务
task -a

# 构建项目
go build -o bin/uds-proxy ./cmd/uds-proxy

# 运行测试
go test ./...
```

## 相关链接

- 使用 [Taskfile](https://taskfile.dev) 管理项目 CLI
- 使用 [Pre-commit](https://pre-commit.com/) 管理 Git hooks
- 使用 [urfave/cli](https://github.com/urfave/cli) 构建 CLI

## 许可证

MIT License
