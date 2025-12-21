# Docker API 示例

<!--TOC-->

- [前置条件](#前置条件) `:43+6`
- [启动代理](#启动代理) `:49+7`
- [基础操作](#基础操作) `:56+34`
  - [获取 Docker 版本](#获取-docker-版本) `:58+20`
  - [获取系统信息](#获取系统信息) `:78+6`
  - [检查 Docker 连通性](#检查-docker-连通性) `:84+6`
- [容器操作](#容器操作) `:90+74`
  - [列出所有容器](#列出所有容器) `:92+13`
  - [创建容器](#创建容器) `:105+18`
  - [启动容器](#启动容器) `:123+7`
  - [停止容器](#停止容器) `:130+7`
  - [删除容器](#删除容器) `:137+11`
  - [查看容器日志](#查看容器日志) `:148+10`
  - [检查容器详情](#检查容器详情) `:158+6`
- [镜像操作](#镜像操作) `:164+28`
  - [列出镜像](#列出镜像) `:166+6`
  - [拉取镜像](#拉取镜像) `:172+7`
  - [删除镜像](#删除镜像) `:179+7`
  - [检查镜像详情](#检查镜像详情) `:186+6`
- [网络操作](#网络操作) `:192+24`
  - [列出网络](#列出网络) `:194+6`
  - [创建网络](#创建网络) `:200+9`
  - [删除网络](#删除网络) `:209+7`
- [数据卷操作](#数据卷操作) `:216+17`
  - [列出数据卷](#列出数据卷) `:218+6`
  - [创建数据卷](#创建数据卷) `:224+9`
- [错误处理](#错误处理) `:233+32`
  - [Socket 不存在](#socket-不存在) `:235+7`
  - [权限不足](#权限不足) `:242+10`
  - [判断错误类型](#判断错误类型) `:252+13`
- [实用脚本](#实用脚本) `:265+30`
  - [批量清理停止的容器](#批量清理停止的容器) `:267+15`
  - [监控容器状态](#监控容器状态) `:282+13`

<!--TOC-->

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
curl -w "%{http_code}" "http://localhost:8080/proxy?path=/var/run/nonexistent.sock&url=/version"
# 返回 502 (无响应体)
```

### 权限不足

如果返回 502（连接失败），检查 socket 权限：

```bash
ls -la /var/run/docker.sock
# 确保当前用户在 docker 组中
sudo usermod -aG docker $USER
```

### 判断错误类型

```bash
# 网关错误：状态码 502/504 且无响应体
# 目标服务错误：状态码透传，有响应体
status=$(curl -s -o /dev/null -w "%{http_code}" "$URL")
if [ "$status" = "502" ] || [ "$status" = "504" ]; then
    echo "网关错误"
else
    echo "目标服务响应: $status"
fi
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
