# Docker API 示例

通过 uds-proxy 访问 Docker Engine API 的完整示例。

## 前置条件

- Docker 已安装并运行
- uds-proxy 服务已启动
- 当前用户有权限访问 `/var/run/docker.sock`

## 启动代理

```bash
# 启动 uds-proxy
./uds-proxy --port 8080
```

## 基础操作

### 获取 Docker 版本

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/version"
```

响应示例：

```json
{
  "Version": "24.0.7",
  "ApiVersion": "1.43",
  "MinAPIVersion": "1.12",
  "GitCommit": "afdd53b",
  "GoVersion": "go1.20.10",
  "Os": "linux",
  "Arch": "amd64"
}
```

### 获取系统信息

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/info"
```

### 检查 Docker 连通性

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/_ping"
```

## 容器操作

### 列出所有容器

```bash
# 仅运行中的容器
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/json"

# 包含已停止的容器
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/json&all=true"

# 限制返回数量
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/json&limit=5"
```

### 创建容器

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "Image": "nginx:alpine",
    "Name": "my-nginx",
    "ExposedPorts": {"80/tcp": {}},
    "HostConfig": {
      "PortBindings": {
        "80/tcp": [{"HostPort": "8888"}]
      }
    }
  }' \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/create&name=my-nginx"
```

### 启动容器

```bash
curl -X POST \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx/start"
```

### 停止容器

```bash
curl -X POST \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx/stop"
```

### 删除容器

```bash
curl -X DELETE \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx"

# 强制删除运行中的容器
curl -X DELETE \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx&force=true"
```

### 查看容器日志

```bash
# 获取日志
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx/logs&stdout=true&stderr=true"

# 获取最近 100 行
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx/logs&stdout=true&tail=100"
```

### 检查容器详情

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/containers/my-nginx/json"
```

## 镜像操作

### 列出镜像

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/images/json"
```

### 拉取镜像

```bash
curl -X POST \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/images/create&fromImage=alpine&tag=latest"
```

### 删除镜像

```bash
curl -X DELETE \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/images/alpine:latest"
```

### 检查镜像详情

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/images/nginx:alpine/json"
```

## 网络操作

### 列出网络

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/networks"
```

### 创建网络

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"Name": "my-network", "Driver": "bridge"}' \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/networks/create"
```

### 删除网络

```bash
curl -X DELETE \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/networks/my-network"
```

## 数据卷操作

### 列出数据卷

```bash
curl "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/volumes"
```

### 创建数据卷

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"Name": "my-volume"}' \
  "http://localhost:8080/proxy?path=/var/run/docker.sock&url=/volumes/create"
```

## 错误处理

### Socket 不存在

```bash
curl "http://localhost:8080/proxy?path=/var/run/nonexistent.sock&url=/version"
# 返回 404: {"error": "socket file not found: /var/run/nonexistent.sock"}
```

### 权限不足

如果返回连接错误，检查 socket 权限：

```bash
ls -la /var/run/docker.sock
# 确保当前用户在 docker 组中
sudo usermod -aG docker $USER
```

## 实用脚本

### 批量清理停止的容器

```bash
#!/bin/bash
PROXY="http://localhost:8080/proxy?path=/var/run/docker.sock"

# 获取已停止的容器
containers=$(curl -s "$PROXY&url=/containers/json&all=true&filters={\"status\":[\"exited\"]}" | jq -r '.[].Id')

for id in $containers; do
  echo "Removing container: $id"
  curl -X DELETE "$PROXY&url=/containers/$id"
done
```

### 监控容器状态

```bash
#!/bin/bash
PROXY="http://localhost:8080/proxy?path=/var/run/docker.sock"

while true; do
  clear
  echo "=== Container Status ==="
  curl -s "$PROXY&url=/containers/json" | jq -r '.[] | "\(.Names[0]): \(.State) (\(.Status))"'
  sleep 5
done
```
