# 快速开始

本指南帮助你快速安装和使用 uds-proxy。

## 安装

### 从源码构建

```bash
# 克隆项目
git clone <repository-url>
cd uds-proxy

# 构建
go build -o uds-proxy ./cmd/uds-proxy

# 或使用 task
task build
```

### 环境要求

- Go 1.23 或更高版本
- Linux/macOS（Unix Socket 支持）

## 基本使用

### 启动代理服务

```bash
# 使用默认配置启动（监听 0.0.0.0:8080）
./uds-proxy

# 指定端口
./uds-proxy --port 9000

# 自动分配端口（端口号写入文件）
./uds-proxy --port 0 --port-file /tmp/proxy.port
```

### 发送代理请求

代理请求通过 `/proxy` 端点发送：

```bash
# 基本格式
curl "http://localhost:8080/proxy?path=<socket路径>&url=<目标URL>"

# 示例：访问 Docker API
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/version"
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--host` | 监听地址 | `0.0.0.0` |
| `--port` | 监听端口（0 表示自动分配） | `8080` |
| `--port-file` | 端口号输出文件 | - |
| `--timeout` | 请求超时时间 | `30s` |
| `--max-conns` | 每个 socket 最大连接数 | `100` |
| `--max-idle-conns` | 每个 socket 最大空闲连接数 | `10` |
| `--no-access-log` | 禁用访问日志 | `false` |

## 验证安装

启动服务后，访问以下端点验证：

```bash
# 服务信息
curl http://localhost:8080/

# 健康检查
curl http://localhost:8080/health
```

## 下一步

- 阅读 [项目架构](/guide/architecture) 了解内部实现
- 查看 [API 文档](/guide/api) 了解完整接口
- 参考 [Docker API 示例](/examples/docker-api) 了解实际用法
