# 编程语言集成

本文档展示如何在不同编程语言中使用 uds-proxy。

## Python

### 使用 requests 库

```python
import requests

PROXY_BASE = "http://localhost:8080/proxy"
DOCKER_SOCK = "/var/run/docker.sock"

def docker_request(url, method="GET", **kwargs):
    """发送 Docker API 请求"""
    params = {
        "path": DOCKER_SOCK,
        "url": url
    }
    return requests.request(method, PROXY_BASE, params=params, **kwargs)

# 获取版本
resp = docker_request("/version")
print(resp.json())

# 列出容器
resp = docker_request("/containers/json", params={"all": "true"})
containers = resp.json()
for c in containers:
    print(f"{c['Names'][0]}: {c['State']}")

# 创建容器
resp = docker_request(
    "/containers/create",
    method="POST",
    params={"name": "my-nginx"},
    json={"Image": "nginx:alpine"}
)
print(resp.json())
```

### 异步客户端 (aiohttp)

```python
import aiohttp
import asyncio

async def async_docker_client():
    PROXY_BASE = "http://localhost:8080/proxy"
    DOCKER_SOCK = "/var/run/docker.sock"

    async with aiohttp.ClientSession() as session:
        # 获取版本
        params = {"path": DOCKER_SOCK, "url": "/version"}
        async with session.get(PROXY_BASE, params=params) as resp:
            version = await resp.json()
            print(f"Docker Version: {version['Version']}")

        # 并发获取多个容器详情
        params = {"path": DOCKER_SOCK, "url": "/containers/json"}
        async with session.get(PROXY_BASE, params=params) as resp:
            containers = await resp.json()

        tasks = []
        for c in containers[:5]:  # 限制前5个
            params = {"path": DOCKER_SOCK, "url": f"/containers/{c['Id']}/json"}
            tasks.append(session.get(PROXY_BASE, params=params))

        responses = await asyncio.gather(*tasks)
        for resp in responses:
            data = await resp.json()
            print(f"Container: {data['Name']}")

asyncio.run(async_docker_client())
```

## Go

### 使用标准库

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

const (
    proxyBase  = "http://localhost:8080/proxy"
    dockerSock = "/var/run/docker.sock"
)

type DockerVersion struct {
    Version    string `json:"Version"`
    APIVersion string `json:"ApiVersion"`
}

func dockerRequest(targetURL string) ([]byte, error) {
    params := url.Values{}
    params.Set("path", dockerSock)
    params.Set("url", targetURL)

    resp, err := http.Get(proxyBase + "?" + params.Encode())
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

func main() {
    // 获取版本
    data, err := dockerRequest("/version")
    if err != nil {
        panic(err)
    }

    var version DockerVersion
    json.Unmarshal(data, &version)
    fmt.Printf("Docker Version: %s (API: %s)\n", version.Version, version.APIVersion)

    // 列出容器
    data, err = dockerRequest("/containers/json?all=true")
    if err != nil {
        panic(err)
    }

    var containers []map[string]interface{}
    json.Unmarshal(data, &containers)

    for _, c := range containers {
        names := c["Names"].([]interface{})
        state := c["State"].(string)
        fmt.Printf("%s: %s\n", names[0], state)
    }
}
```

### 封装客户端

```go
package udsproxy

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

type Client struct {
    BaseURL    string
    HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
    return &Client{
        BaseURL:    baseURL,
        HTTPClient: &http.Client{},
    }
}

func (c *Client) Request(socketPath, targetURL, method string, body interface{}) ([]byte, error) {
    params := url.Values{}
    params.Set("path", socketPath)
    params.Set("url", targetURL)

    reqURL := c.BaseURL + "/proxy?" + params.Encode()

    var bodyReader io.Reader
    if body != nil {
        data, _ := json.Marshal(body)
        bodyReader = bytes.NewReader(data)
    }

    req, err := http.NewRequest(method, reqURL, bodyReader)
    if err != nil {
        return nil, err
    }

    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("request failed: %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}
```

## JavaScript / Node.js

### 使用 fetch

```javascript
const PROXY_BASE = 'http://localhost:8080/proxy';
const DOCKER_SOCK = '/var/run/docker.sock';

async function dockerRequest(url, options = {}) {
  const params = new URLSearchParams({
    path: DOCKER_SOCK,
    url: url
  });

  const response = await fetch(`${PROXY_BASE}?${params}`, options);
  return response.json();
}

// 获取版本
const version = await dockerRequest('/version');
console.log(`Docker Version: ${version.Version}`);

// 列出容器
const containers = await dockerRequest('/containers/json?all=true');
containers.forEach(c => {
  console.log(`${c.Names[0]}: ${c.State}`);
});

// 创建容器
const newContainer = await dockerRequest('/containers/create?name=my-nginx', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ Image: 'nginx:alpine' })
});
console.log(`Created: ${newContainer.Id}`);
```

### 使用 axios

```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'http://localhost:8080'
});

async function listContainers() {
  const { data } = await client.get('/proxy', {
    params: {
      path: '/var/run/docker.sock',
      url: '/containers/json',
      all: true
    }
  });
  return data;
}

async function createContainer(name, image) {
  const { data } = await client.post('/proxy',
    { Image: image },
    {
      params: {
        path: '/var/run/docker.sock',
        url: '/containers/create',
        name: name
      }
    }
  );
  return data;
}
```

## Shell / Bash

### 基础封装

```bash
#!/bin/bash

PROXY_BASE="http://localhost:8080/proxy"
DOCKER_SOCK="/var/run/docker.sock"

docker_api() {
    local url=$1
    local method=${2:-GET}
    local data=$3

    if [ -n "$data" ]; then
        curl -s -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$PROXY_BASE?path=$DOCKER_SOCK&url=$url"
    else
        curl -s -X "$method" \
            "$PROXY_BASE?path=$DOCKER_SOCK&url=$url"
    fi
}

# 使用示例
echo "=== Docker Version ==="
docker_api "/version" | jq .

echo "=== Containers ==="
docker_api "/containers/json?all=true" | jq -r '.[] | "\(.Names[0]): \(.State)"'

echo "=== Creating Container ==="
docker_api "/containers/create?name=test-nginx" "POST" '{"Image":"nginx:alpine"}' | jq .
```

### 实用函数库

```bash
#!/bin/bash

# uds-proxy 客户端函数库

UDS_PROXY_URL="${UDS_PROXY_URL:-http://localhost:8080}"

uds_get() {
    local socket=$1
    local url=$2
    curl -s "$UDS_PROXY_URL/proxy?path=$socket&url=$url"
}

uds_post() {
    local socket=$1
    local url=$2
    local data=$3
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$UDS_PROXY_URL/proxy?path=$socket&url=$url"
}

uds_delete() {
    local socket=$1
    local url=$2
    curl -s -X DELETE "$UDS_PROXY_URL/proxy?path=$socket&url=$url"
}

# 导出函数
export -f uds_get uds_post uds_delete
```

## Rust

```rust
use reqwest;
use serde::{Deserialize, Serialize};

const PROXY_BASE: &str = "http://localhost:8080/proxy";
const DOCKER_SOCK: &str = "/var/run/docker.sock";

#[derive(Debug, Deserialize)]
struct DockerVersion {
    #[serde(rename = "Version")]
    version: String,
    #[serde(rename = "ApiVersion")]
    api_version: String,
}

async fn docker_get<T: for<'de> Deserialize<'de>>(url: &str) -> Result<T, reqwest::Error> {
    let client = reqwest::Client::new();
    let resp = client.get(PROXY_BASE)
        .query(&[("path", DOCKER_SOCK), ("url", url)])
        .send()
        .await?
        .json::<T>()
        .await?;
    Ok(resp)
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let version: DockerVersion = docker_get("/version").await?;
    println!("Docker Version: {} (API: {})", version.version, version.api_version);
    Ok(())
}
```

## 最佳实践

### 1. 错误处理

```python
def safe_docker_request(url, method="GET", **kwargs):
    try:
        resp = docker_request(url, method, **kwargs)
        resp.raise_for_status()
        return resp.json()
    except requests.exceptions.ConnectionError:
        print("无法连接到代理服务")
        return None
    except requests.exceptions.HTTPError as e:
        print(f"请求失败: {e.response.status_code}")
        return None
```

### 2. 连接池复用

```go
// 复用 HTTP 客户端
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 3. 超时控制

```javascript
const controller = new AbortController();
const timeout = setTimeout(() => controller.abort(), 30000);

try {
  const response = await fetch(url, { signal: controller.signal });
} finally {
  clearTimeout(timeout);
}
```
