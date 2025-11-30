# 自定义服务示例

本文档展示如何使用 uds-proxy 代理自定义 Unix Socket 服务。

## 创建示例 Socket 服务

### Python 示例服务

创建一个简单的 Unix Socket HTTP 服务：

```python
#!/usr/bin/env python3
"""简单的 Unix Socket HTTP 服务示例"""

import socket
import os
import json
from datetime import datetime

SOCKET_PATH = '/tmp/demo-service.sock'

def handle_request(data):
    """处理 HTTP 请求"""
    lines = data.decode().split('\r\n')
    request_line = lines[0]
    method, path, _ = request_line.split(' ')

    # 路由处理
    if path == '/':
        response_body = json.dumps({
            'service': 'demo-service',
            'version': '1.0.0',
            'endpoints': ['/status', '/time', '/echo']
        })
    elif path == '/status':
        response_body = json.dumps({'status': 'running'})
    elif path == '/time':
        response_body = json.dumps({'time': datetime.now().isoformat()})
    elif path.startswith('/echo'):
        response_body = json.dumps({'method': method, 'path': path})
    else:
        return b'HTTP/1.1 404 Not Found\r\nContent-Type: application/json\r\n\r\n{"error": "not found"}'

    return f'HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {len(response_body)}\r\n\r\n{response_body}'.encode()

def main():
    # 清理旧的 socket 文件
    if os.path.exists(SOCKET_PATH):
        os.unlink(SOCKET_PATH)

    server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    server.bind(SOCKET_PATH)
    server.listen(5)

    # 设置权限
    os.chmod(SOCKET_PATH, 0o666)

    print(f'Server listening on {SOCKET_PATH}')

    try:
        while True:
            conn, _ = server.accept()
            data = conn.recv(4096)
            if data:
                response = handle_request(data)
                conn.sendall(response)
            conn.close()
    finally:
        server.close()
        os.unlink(SOCKET_PATH)

if __name__ == '__main__':
    main()
```

### Go 示例服务

```go
package main

import (
    "encoding/json"
    "log"
    "net"
    "net/http"
    "os"
    "time"
)

const socketPath = "/tmp/demo-service.sock"

func main() {
    // 清理旧的 socket 文件
    os.Remove(socketPath)

    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()

    // 设置权限
    os.Chmod(socketPath, 0666)

    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "service":   "demo-service",
            "version":   "1.0.0",
            "endpoints": []string{"/status", "/time", "/echo"},
        })
    })

    mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "running"})
    })

    mux.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{
            "time": time.Now().Format(time.RFC3339),
        })
    })

    log.Printf("Server listening on %s", socketPath)
    http.Serve(listener, mux)
}
```

## 通过代理访问服务

### 启动服务和代理

```bash
# 终端 1: 启动示例服务
python3 demo_service.py

# 终端 2: 启动 uds-proxy
./uds-proxy --port 8080
```

### 测试请求

```bash
# 获取服务信息
curl "http://localhost:8080/proxy?path=/tmp/demo-service.sock&url=/"

# 获取状态
curl "http://localhost:8080/proxy?path=/tmp/demo-service.sock&url=/status"

# 获取当前时间
curl "http://localhost:8080/proxy?path=/tmp/demo-service.sock&url=/time"

# Echo 测试
curl "http://localhost:8080/proxy?path=/tmp/demo-service.sock&url=/echo"
```

## 实际应用场景

### 场景 1：微服务通信

在容器化环境中，服务间通过 Unix Socket 通信可以获得更好的性能：

```bash
# 服务 A 通过代理访问服务 B
curl "http://localhost:8080/proxy?path=/var/run/service-b.sock&url=/api/data"
```

### 场景 2：PHP-FPM 状态监控

```bash
# 获取 PHP-FPM 状态
curl "http://localhost:8080/proxy?path=/var/run/php-fpm.sock&url=/status?full&html"
```

### 场景 3：MySQL Socket 管理

虽然 MySQL 协议不是 HTTP，但可以用于管理接口：

```bash
# 假设有 MySQL 管理 HTTP 接口
curl "http://localhost:8080/proxy?path=/var/run/mysql-admin.sock&url=/status"
```

### 场景 4：Nginx 状态页

```bash
# 如果 Nginx 配置了 Unix Socket 后端
curl "http://localhost:8080/proxy?path=/var/run/nginx-status.sock&url=/nginx_status"
```

## 安全注意事项

### Socket 文件权限

```bash
# 检查 socket 权限
ls -la /tmp/demo-service.sock

# 设置合适的权限
chmod 660 /tmp/demo-service.sock
chown root:app-group /tmp/demo-service.sock
```

### 访问控制

在生产环境中，建议：

1. 限制 uds-proxy 的访问来源
2. 使用反向代理添加认证
3. 限制可代理的 socket 路径（通过配置白名单）

## 调试技巧

### 检查 Socket 是否存在

```bash
test -S /tmp/demo-service.sock && echo "Socket exists" || echo "Socket not found"
```

### 直接测试 Socket

```bash
# 使用 socat 测试
echo -e "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n" | socat - UNIX-CONNECT:/tmp/demo-service.sock

# 使用 curl 直接访问（如果支持）
curl --unix-socket /tmp/demo-service.sock http://localhost/
```

### 查看代理日志

```bash
# 启动时启用详细日志
./uds-proxy --port 8080 2>&1 | tee proxy.log
```
