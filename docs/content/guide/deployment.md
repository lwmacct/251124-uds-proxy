# 部署指南

<!--TOC-->

- [二进制部署](#二进制部署) `:31+50`
  - [构建生产版本](#构建生产版本) `:33+13`
  - [Systemd 服务](#systemd-服务) `:46+35`
- [Docker 部署](#docker-部署) `:81+49`
  - [Dockerfile](#dockerfile) `:83+19`
  - [Docker Compose](#docker-compose) `:102+14`
  - [运行容器](#运行容器) `:116+14`
- [配置建议](#配置建议) `:130+22`
  - [生产环境配置](#生产环境配置) `:132+11`
  - [参数调优](#参数调优) `:143+9`
- [反向代理配置](#反向代理配置) `:152+34`
  - [Nginx](#nginx) `:154+24`
  - [Caddy](#caddy) `:178+8`
- [监控和日志](#监控和日志) `:186+30`
  - [健康检查](#健康检查) `:188+8`
  - [Prometheus 指标（规划中）](#prometheus-指标规划中) `:196+8`
  - [日志收集](#日志收集) `:204+12`
- [安全建议](#安全建议) `:216+21`
  - [访问控制](#访问控制) `:218+6`
  - [运行权限](#运行权限) `:224+10`
  - [TLS 加密](#tls-加密) `:234+3`

<!--TOC-->

本文档介绍 uds-proxy 的各种部署方式。

## 二进制部署

### 构建生产版本

```bash
# 标准构建
go build -o uds-proxy ./cmd/uds-proxy

# 优化构建（减小体积）
CGO_ENABLED=0 go build -ldflags="-s -w" -o uds-proxy ./cmd/uds-proxy

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o uds-proxy-linux-amd64 ./cmd/uds-proxy
```

### Systemd 服务

创建服务文件 `/etc/systemd/system/uds-proxy.service`：

```ini
[Unit]
Description=UDS Proxy - HTTP to Unix Domain Socket Proxy
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/uds-proxy --host 0.0.0.0 --port 8080
Restart=always
RestartSec=5

# 安全加固
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/run

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable uds-proxy
sudo systemctl start uds-proxy
sudo systemctl status uds-proxy
```

## Docker 部署

### Dockerfile

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o uds-proxy ./cmd/uds-proxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/uds-proxy .
EXPOSE 8080
ENTRYPOINT ["./uds-proxy"]
CMD ["--host", "0.0.0.0", "--port", "8080"]
```

### Docker Compose

```yaml
version: "3.8"
services:
  uds-proxy:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
```

### 运行容器

```bash
# 构建镜像
docker build -t uds-proxy .

# 运行（挂载 Docker socket）
docker run -d \
  --name uds-proxy \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  uds-proxy
```

## 配置建议

### 生产环境配置

```bash
./uds-proxy \
  --host 0.0.0.0 \
  --port 8080 \
  --timeout 60s \
  --max-conns 200 \
  --max-idle-conns 20
```

### 参数调优

| 场景         | 推荐配置                              |
| ------------ | ------------------------------------- |
| 低并发       | `--max-conns 50 --max-idle-conns 5`   |
| 高并发       | `--max-conns 500 --max-idle-conns 50` |
| 长连接服务   | `--timeout 120s`                      |
| 快速响应服务 | `--timeout 10s`                       |

## 反向代理配置

### Nginx

```nginx
upstream uds_proxy {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name proxy.example.com;

    location / {
        proxy_pass http://uds_proxy;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

### Caddy

```caddyfile
proxy.example.com {
    reverse_proxy localhost:8080
}
```

## 监控和日志

### 健康检查

配置负载均衡器或监控系统定期检查：

```bash
curl -f http://localhost:8080/health || exit 1
```

### Prometheus 指标（规划中）

未来版本将支持 `/metrics` 端点，提供：

- 请求计数
- 响应时间直方图
- 连接池状态

### 日志收集

访问日志输出到 stdout，可使用日志收集工具处理：

```bash
# 使用 journalctl 查看（systemd）
journalctl -u uds-proxy -f

# Docker 日志
docker logs -f uds-proxy
```

## 安全建议

### 访问控制

1. **网络隔离**：仅在内部网络暴露服务
2. **防火墙规则**：限制访问来源
3. **Unix Socket 权限**：确保代理进程有权限访问目标 socket

### 运行权限

```bash
# 创建专用用户
sudo useradd -r -s /bin/false uds-proxy

# 添加到 docker 组（如需访问 Docker socket）
sudo usermod -aG docker uds-proxy
```

### TLS 加密

建议在反向代理层启用 TLS，而非直接在 uds-proxy 配置。
