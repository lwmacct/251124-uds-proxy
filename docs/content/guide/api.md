# API 文档

uds-proxy 提供简洁的 RESTful API 接口。

## 端点概览

| 端点 | 方法 | 说明 |
|------|------|------|
| `/` | GET | 服务信息 |
| `/health` | GET | 健康检查 |
| `/proxy` | ALL | 代理请求 |

## 服务信息

### `GET /`

返回服务基本信息和使用说明。

**响应示例：**

```json
{
  "service": "uds-proxy",
  "version": "0.1.0",
  "usage": {
    "endpoint": "/proxy",
    "params": {
      "path": "Unix socket 文件路径（必需）",
      "url": "目标 URL 路径（默认: /）",
      "method": "覆盖 HTTP 方法（可选）"
    },
    "example": "/proxy?path=/var/run/docker.sock&url=/version"
  }
}
```

## 健康检查

### `GET /health`

用于服务健康检查，适用于负载均衡器或监控系统。

**响应：**

```json
{
  "status": "ok"
}
```

**状态码：** `200 OK`

## 代理请求

### `[ALL] /proxy`

核心代理端点，支持所有 HTTP 方法。

### 请求参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `path` | string | **是** | Unix socket 文件路径 |
| `url` | string | 否 | 目标 URL 路径，默认 `/` |
| `method` | string | 否 | 覆盖 HTTP 方法 |
| `*` | any | 否 | 其他参数将转发到目标服务 |

### 请求头转发

所有请求头将自动转发到目标服务，以下头部除外：
- `Host`（将被替换为 `localhost`）

### 请求体转发

对于 POST、PUT、PATCH 等方法，请求体将完整转发。

### 响应

代理成功时，返回目标服务的原始响应，包括：
- 状态码
- 响应头
- 响应体

### 错误响应

| 状态码 | 说明 | 响应格式 |
|--------|------|----------|
| 400 | 缺少 `path` 参数 | `{"error": "missing required parameter: path"}` |
| 404 | Socket 文件不存在 | `{"error": "socket file not found: <path>"}` |
| 500 | 内部错误 | `{"error": "<错误信息>"}` |
| 503 | 连接失败 | `{"error": "failed to connect to socket: <path>"}` |
| 504 | 请求超时 | `{"error": "request timeout"}` |

## 使用示例

### 基本 GET 请求

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/version"
```

### 带查询参数的请求

```bash
# 额外参数自动转发
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/json&all=true&limit=10"
```

### POST 请求

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"Image": "nginx"}' \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/create"
```

### 覆盖 HTTP 方法

```bash
# 使用 method 参数覆盖实际请求方法
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/abc123&method=DELETE"
```

### 自定义请求头

```bash
curl -H "X-Custom-Header: value" \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/info"
```

## 调试技巧

### 查看详细请求信息

```bash
curl -v "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/version"
```

### 启用访问日志

默认情况下访问日志已启用，可在终端查看请求详情：

```
2024/01/15 10:30:45 GET /proxy?path=/var/run/docker.sock&url=/version 200 150ms
```

### 禁用访问日志

```bash
./uds-proxy --no-access-log
```
