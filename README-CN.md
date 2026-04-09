# goed2kd

基于 [goed2k/core](https://github.com/goed2k/core) 的守护进程：在**不改动底层 ED2K 协议实现**的前提下，提供 **HTTP/JSON RPC** 与 **WebSocket** 事件，便于脚本、自动化与后续 Web UI 接入。

## 功能概览

- **系统**：健康检查、运行信息、引擎启停、状态保存/加载、配置读写与热更新（部分字段）
- **网络**：连接服务器、批量连接、加载 `server.met`、DHT 状态、启用 DHT、加载 `nodes.dat`、添加 KAD 引导节点
- **下载任务**：添加 ED2K、列表/详情、暂停/恢复/删除、Peers、Pieces
- **搜索**：发起搜索、当前快照、停止、从当前结果添加下载（同一时间仅一个活跃搜索，由底层约束）
- **共享库**：列出共享文件、注册/移除扫描目录、重扫目录、导入文件、按 hash 移除（与 goed2k/core 共享 API 对齐）
- **实时事件**：WebSocket 推送 `client.status`、`transfer.progress`

## 架构要点

- HTTP 路由：[chi](https://github.com/go-chi/chi)
- 分层：`rpc`（HTTP/WS）→ `service` → `engine`（唯一封装 `goed2k.Client`）→ `config` / `store`
- Handler **不得**直接使用 `goed2k.Client`，业务一律经 `engine` 与 `service`

## 环境要求

- Go **1.25+**（与 [go.mod](./go.mod) 中 `go` 指令一致）
- 依赖见 [go.mod](./go.mod)

Go 模块路径：`github.com/chenjia404/goed2kd`（作为库引用时使用该 import 前缀）。

## 快速开始

```bash
# 开发运行（首次会在 -config 路径生成默认配置文件）
go run ./cmd/goed2kd -config data/config/config.json

# 编译
go build -o bin/goed2kd ./cmd/goed2kd
./bin/goed2kd -config data/config/config.json
```

### Docker（Alpine 多阶段镜像）

本地构建：

```bash
docker build -t goed2kd .
docker run --rm -p 18080:18080 -p 4661:4661 -p 4662:4662/udp \
  -v goed2kd-data:/app/data goed2kd
```

CI 发布（打 tag）时镜像同时推送到 **GitHub Container Registry** 与 **Docker Hub**，例如：

```bash
docker pull ghcr.io/chenjia404/goed2kd:latest
docker pull chenjia404/goed2kd:latest
```

GHCR 为 private 时需先 `docker login ghcr.io`（PAT 含 `read:packages`）。Docker Hub 私有同理使用 `docker login`。

首次启动会在数据卷内生成默认配置。若要从宿主机或其它机器访问容器内 RPC，请在配置中将 `rpc.listen` 改为 `0.0.0.0:18080` 且将 `rpc.allow_remote` 设为 `true`（否则校验会拒绝监听 `0.0.0.0`）。也可挂载自定义 `config.json`：`-v /path/config.json:/app/data/config/config.json`。

默认监听 **`127.0.0.1:18080`**。默认 token 为 `change-me`，**生产环境务必修改**。

### 健康检查（无需 Token）

```bash
curl -s http://127.0.0.1:18080/api/v1/system/health
```

### 带鉴权示例

```bash
curl -s -H "Authorization: Bearer change-me" http://127.0.0.1:18080/api/v1/system/info
```

## 配置说明

- 配置文件路径由启动参数 `-config` 指定，默认工程内常用为 `data/config/config.json`。
- 若文件不存在，进程会**自动创建**一份与 [configs/config.example.json](./configs/config.example.json) 结构一致的默认配置。
- 主要段落：
  - **`rpc`**：`listen`、`auth_token`、`allow_remote`（为 `false` 时不允许监听 `0.0.0.0` / `::`）、读写超时
  - **`engine`**：监听端口、UDP、DHT/UPnP、连接与上传相关参数、默认下载目录
  - **`bootstrap`**：启动后引导用的服务器地址、`server.met` / `nodes.dat` URL、KAD 引导节点
  - **`state`**：是否启用、`path`、`load_on_start`、`save_on_exit`、自动保存间隔（秒）
  - **`logging`**：`level`（debug/info/warn/error）、日志文件路径

状态文件由 **goed2k 自带机制**读写；守护进程负责路径、调度与退出时保存。

## 行为说明

- **启动流程**：加载配置 → 启动 HTTP → 启动事件桥接 → **自动尝试启动引擎**（失败时仍可提供 HTTP，便于稍后 `POST /system/start` 重试）。
- **优雅退出**（SIGINT/SIGTERM）：关闭 HTTP 与事件桥接；若 `state.save_on_exit` 且引擎在运行，则保存状态后停止引擎。
- **自动保存**：在启用 state 且配置间隔大于 0 时，通过 `goed2k` Client 的自动保存间隔与内部循环落盘（与 `state.auto_save_interval_seconds` 对齐）。

## API 与 WebSocket

完整接口说明（路径、请求体、响应字段、错误码、WS 协议）见：

**[docs/API.md](./docs/API.md)**

## 仓库结构（节选）

```text
cmd/goed2kd/          # 入口
internal/
  app/                # 日志、Daemon 生命周期
  config/             # 配置模型与加载
  engine/             # goed2k 封装
  model/              # DTO、错误码、事件模型
  rpc/http/           # HTTP 路由与 Handler
  rpc/ws/             # WebSocket Hub
  service/            # 业务编排
  store/              # 配置存储抽象
configs/              # 配置样例
docs/                 # 文档（含 API）
```

## 许可证

若本仓库未单独声明许可证，请以仓库根目录 LICENSE 为准；依赖库 [goed2k/core](https://github.com/goed2k/core) 以该项目许可证为准。
