# Conflux (C++ Edition)

## 功能

- YAML 路由加载（`configs/routes.yaml`）
- 前缀路由匹配（按 `priority` 和前缀长度排序）
- 反向代理转发（未命中显式路由时走网关匹配逻辑）
- 负载均衡：`round_robin` / `random` / `first`
- 内存滑动窗口限流（按来源 IP）
- 管理接口：`GET /admin/routes`（支持 `X-Admin-Token`）
- 健康接口：`GET /health`

## 本地构建

```bash
cmake -S . -B build
cmake --build build -j
```

## 运行

```bash
./build/conflux
```

默认监听 `:8080`。

## 环境变量（摘要）

| 变量 | 含义 | 默认 |
|------|------|------|
| `NEXUS_ROUTES_FILE` | 路由 YAML 路径 | `configs/routes.yaml` |
| `NEXUS_LB` | `round_robin` / `random` / `first` | `round_robin` |
| `NEXUS_ADMIN_PREFIX` | Admin 路由前缀 | `/admin` |
| `NEXUS_ADMIN_TOKEN` | Admin 接口令牌（空则不校验） | 空 |
| `NEXUS_RATELIMIT_ENABLE` | 是否启用限流 | `true` |
| `NEXUS_RATELIMIT_RPS` | 每秒请求额度基数 | `100` |
| `NEXUS_RATELIMIT_WINDOW_SEC` | 限流窗口秒数 | `10` |

## Docker Compose

```bash
docker compose up --build
```

- 网关：<http://localhost:8080>
- mock-backend：用于演示上游转发

## Mock Backend（单独运行）

```bash
cmake -S docker/mock-backend -B docker/mock-backend/build
cmake --build docker/mock-backend/build -j
./docker/mock-backend/build/mock-backend
```
