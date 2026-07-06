# Conflux

Conflux 是一个基于 C++20 实现的轻量级 API 网关，提供路由匹配、反向代理、负载均衡、限流、管理接口、指标暴露和 Docker 部署能力。

## 功能

- YAML 路由配置加载
- 前缀路由匹配，支持 `priority` 优先级排序
- HTTP 反向代理，保留 query string，转发请求体和请求头
- 负载均衡策略：`round_robin`、`random`、`first`
- 内存滑动窗口限流，按来源 IP 统计
- 上游连接和读取超时控制
- 访问日志，记录方法、路径、路由、上游、状态码和耗时
- 管理接口：`GET /admin/routes`
- 健康检查接口：`GET /health`
- 指标接口：`GET /metrics`
- Docker Compose 一键启动网关和 mock backend

## 目录结构

```text
.
├── configs/                 # 路由配置
├── docker/mock-backend/      # 演示用上游服务
├── src/                     # 网关源码
├── tests/                   # C++ 测试
├── third_party/             # 第三方头文件
├── CMakeLists.txt
├── Dockerfile
└── docker-compose.yml
```

## 构建

```bash
cmake -S . -B build
cmake --build build -j
```

## 运行

```bash
./build/conflux
```

默认监听 `0.0.0.0:8080`，默认读取 `configs/routes.yaml`。

## 测试

```bash
ctest --test-dir build --output-on-failure
```

## Docker Compose

```bash
docker compose up --build
```

启动后：

- 网关：`http://localhost:8080`
- mock backend：由网关通过 Docker 网络访问

## 接口

```bash
curl http://localhost:8080/health
curl -H "X-Admin-Token: dev-admin-token" http://localhost:8080/admin/routes
curl http://localhost:8080/metrics
curl "http://localhost:8080/user/42?debug=true"
```

## 路由配置

```yaml
routes:
  - id: mock
    path_prefix: /user
    strip_prefix: false
    priority: 100
    targets:
      - http://127.0.0.1:18081
```

字段说明：

| 字段 | 说明 |
|------|------|
| `id` | 路由 ID |
| `path_prefix` | 请求路径前缀 |
| `strip_prefix` | 转发到上游时是否移除前缀 |
| `priority` | 路由优先级，数值越大越优先 |
| `targets` | 上游服务地址列表 |

## 环境变量

| 变量 | 含义 | 默认 |
|------|------|------|
| `NEXUS_ROUTES_FILE` | 路由 YAML 路径 | `configs/routes.yaml` |
| `NEXUS_LB` | 负载均衡策略 | `round_robin` |
| `NEXUS_ADMIN_PREFIX` | 管理接口前缀 | `/admin` |
| `NEXUS_ADMIN_TOKEN` | 管理接口令牌，空值表示不校验 | 空 |
| `NEXUS_RATELIMIT_ENABLE` | 是否启用限流 | `true` |
| `NEXUS_RATELIMIT_RPS` | 每秒请求额度基数 | `100` |
| `NEXUS_RATELIMIT_WINDOW_SEC` | 限流窗口秒数 | `10` |
| `NEXUS_PROXY_TIMEOUT_MS` | 上游连接和读取超时毫秒数 | `3000` |
