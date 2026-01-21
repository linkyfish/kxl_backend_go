# kxl_backend_go（Golang + Echo）

本目录提供企业官网系统的 Golang（Echo + GORM）后端实现，目标是与现有 Rust / PHP 后端保持 API 路径与响应格式兼容，并支持 SSR 官网页面渲染。

## 目录结构

- `cmd/api/`：服务入口（同时提供 API + SSR）
- `internal/`：业务代码（handler/service/model/middleware 等）
- `pkg/`：可复用的基础组件（db/redis/session）
- `templates/`：SSR 模板（Pongo2）
- `static/`：静态资源（CSS/JS/图片）
- `uploads/`：上传文件目录（运行时生成）

## 本地开发

1) 安装依赖

```bash
cd kxl_backend_go
go mod download
```

若网络访问 `proxy.golang.org` 不稳定，可使用：

```bash
GOPROXY=direct go mod download
```

2) 配置

- 复制并按需修改 `.env.example`（可选）
- 或编辑 `config/config.yaml`

3) 运行

```bash
make run
# 或
go run ./cmd/api
```

默认监听：`0.0.0.0:8787`

## 构建

```bash
make build
./kxl-api
```

## 主要路由

- 健康检查：`GET /health`，`GET /ready`
- 公开 API：`/api/v1/*`
- 管理 API：`/api/admin/*`
- 上传 API：`POST /api/upload/image`，`POST /api/upload/video`
- SSR 官网：`/`、`/projects`、`/cases`、`/articles`、`/about`、`/contact`、`/search`、`/login`、`/register`
- 静态资源：`/static/*`（来自 `static/`）
- 上传文件：`/uploads/*`（来自 `uploads/`）

## Docker

```bash
cd kxl_backend_go
docker build -t kxl-backend-go:latest .
docker run --rm -p 8787:8787 \
  -e DB_HOST=host.docker.internal -e DB_PORT=5432 -e DB_DATABASE=kxl -e DB_USERNAME=postgres -e DB_PASSWORD=postgres \
  -e REDIS_HOST=host.docker.internal -e REDIS_PORT=6379 \
  kxl-backend-go:latest
```

