# API 文档

<!--TOC-->

- [端点概览](#端点概览) `:32+8`
- [服务信息](#服务信息) `:40+24`
  - [GET /](#get) `:42+22`
- [健康检查](#健康检查) `:64+16`
  - [GET /health](#get-health) `:66+14`
- [代理请求](#代理请求) `:80+46`
  - [[ALL] /proxy](#all-proxy) `:82+4`
  - [请求参数](#请求参数) `:86+9`
  - [请求头转发](#请求头转发) `:95+6`
  - [请求体转发](#请求体转发) `:101+4`
  - [响应](#响应) `:105+8`
  - [错误响应](#错误响应) `:113+13`
- [使用示例](#使用示例) `:126+38`
  - [基本 GET 请求](#基本-get-请求) `:128+6`
  - [带查询参数的请求](#带查询参数的请求) `:134+7`
  - [POST 请求](#post-请求) `:141+9`
  - [覆盖 HTTP 方法](#覆盖-http-方法) `:150+7`
  - [自定义请求头](#自定义请求头) `:157+7`
- [调试技巧](#调试技巧) `:164+21`
  - [查看详细请求信息](#查看详细请求信息) `:166+6`
  - [启用访问日志](#启用访问日志) `:172+8`
  - [禁用访问日志](#禁用访问日志) `:180+5`

<!--TOC-->

uds-proxy 提供简洁的 RESTful API 接口。

## 端点概览

| 端点      | 方法 | 说明     |
| --------- | ---- | -------- |
| `/`       | GET  | 服务信息 |
| `/health` | GET  | 健康检查 |
| `/proxy`  | ALL  | 代理请求 |

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

| 参数     | 类型   | 必需   | 说明                     |
| -------- | ------ | ------ | ------------------------ |
| `path`   | string | **是** | Unix socket 文件路径     |
| `url`    | string | 否     | 目标 URL 路径，默认 `/`  |
| `method` | string | 否     | 覆盖 HTTP 方法           |
| `*`      | any    | 否     | 其他参数将转发到目标服务 |

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

作为纯网关代理，网关级错误只返回状态码，**无响应体**：

| 状态码      | 说明                    | 响应体         |
| ----------- | ----------------------- | -------------- |
| 2xx/4xx/5xx | 透传目标服务响应        | 目标服务响应体 |
| 400         | 缺少 `path` 参数        | 无             |
| 502         | Socket 不存在或连接失败 | 无             |
| 504         | 请求超时                | 无             |

> **设计原则**：调用方通过状态码区分是目标服务响应还是网关错误。网关错误无 body，目标服务响应原样透传。

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
