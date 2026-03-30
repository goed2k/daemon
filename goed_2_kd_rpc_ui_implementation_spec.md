# goed2kd 实现文档（基于 goed2k 的 RPC 优先架构）

## 1. 文档目标

本文档定义一个基于 `goed2k` 核心库构建的新项目：先实现后台守护进程与 RPC 接口，再实现 Web UI。

目标不是重写 ED2K/eMule 下载内核，而是在现有 `goed2k` 能力之上，构建一个可管理、可扩展、可接 UI 的产品化外壳。

本文档面向 Cursor/Codex 直接执行，要求输出可运行、可维护、可扩展的 Go 项目。

---

## 2. 项目定位

新项目暂定名：`goed2kd`

组件划分：

- `goed2k`：底层下载引擎库（现有开源项目）
- `goed2kd`：后台守护进程（本次先实现）
- `RPC API`：HTTP/JSON + WebSocket（本次先实现）
- `Web UI`：后续实现，直接消费 RPC API
- `CLI`：可选，后续可复用 RPC API

设计原则：

1. 不修改或尽量少修改 `goed2k` 内核
2. 所有对外能力通过 RPC 暴露
3. UI 不直接调用内核，只调用 RPC
4. 下载状态和断点续传优先保证稳定性
5. 第一阶段先做本机使用，默认只监听 `127.0.0.1`
6. 第一阶段优先本地单实例，不做分布式

---

## 3. 第一阶段范围（必须实现）

### 3.1 系统能力

- 启动 `goed2k` client
- 停止 `goed2k` client
- 读取与更新 daemon 配置
- 健康检查
- 获取系统运行状态
- 手动保存状态
- 手动加载状态
- 支持自动保存状态

### 3.2 网络能力

- 连接单个 server
- 批量连接 servers
- 加载 server.met
- 开启 DHT/KAD
- 加载 nodes.dat
- 添加 bootstrap 节点
- 查询 server 状态
- 查询 DHT 状态

### 3.3 下载任务能力

- 添加 ED2K 下载任务
- 支持指定下载路径
- 查询任务列表
- 查询单任务详情
- 暂停任务
- 恢复任务
- 删除任务
- 查询任务 peers
- 查询任务 pieces
- 查询任务速度、进度、ETA、状态

### 3.4 搜索能力

- 发起搜索
- 获取当前搜索快照
- 停止当前搜索
- 从搜索结果中添加下载任务

### 3.5 实时事件

- WebSocket 推送整体状态变化
- WebSocket 推送下载任务进度变化

---

## 4. 第二阶段范围（暂不实现，但结构需预留）

- Web UI
- 任务分类/标签
- 批量操作
- 搜索结果事件推送
- 下载完成通知
- 日志页
- 用户认证体系升级
- 远程访问与 TLS

---

## 5. 非目标

本阶段不做以下内容：

- 不重写 ED2K/KAD 协议栈
- 不做多用户系统
- 不做数据库持久化
- 不做下载任务标签系统
- 不做桌面客户端
- 不做 P2P 文件分享 UI
- 不做复杂权限模型
- 不做插件系统

---

## 6. 技术选型

### 6.1 后端语言与依赖

使用 Go。

推荐：

- HTTP Router: `chi`
- JSON: 标准库 `encoding/json`
- Config: 标准库 + 自定义加载
- Logging: `log/slog`
- WebSocket: `github.com/gorilla/websocket` 或 `nhooyr.io/websocket`
- UUID: 可选，不强制

说明：

- 优先减少依赖
- 核心逻辑尽量使用标准库
- 所有业务都通过 service 层封装

---

## 7. 项目目录结构

```text
cmd/
  goed2kd/
    main.go

internal/
  app/
    daemon.go
    lifecycle.go

  config/
    config.go
    loader.go
    validator.go

  core/
    engine.go
    mapper.go
    errors.go

  service/
    system_service.go
    network_service.go
    transfer_service.go
    search_service.go
    event_service.go

  rpc/
    http/
      router.go
      middleware.go
      response.go
      handlers_system.go
      handlers_network.go
      handlers_transfer.go
      handlers_search.go
      dto.go
    ws/
      hub.go
      client.go
      events.go

  store/
    app_config_store.go
    runtime_store.go

  model/
    dto_system.go
    dto_network.go
    dto_transfer.go
    dto_search.go
    event.go

  util/
    time.go
    path.go
    parse.go

configs/
  config.example.json

data/
  config/
  state/
  logs/

web/
  (预留，后续实现)
```

要求：

- HTTP handler 不能直接操作 `goed2k.Client`
- 所有对 `goed2k` 的调用统一从 `core/engine.go` 进入
- service 层负责参数校验、业务编排、错误归一化
- rpc 层只做协议转换

---

## 8. 核心架构设计

### 8.1 分层关系

```text
HTTP / WebSocket
      ↓
Service Layer
      ↓
Engine Wrapper
      ↓
goed2k.Client
```

### 8.2 各层职责

#### RPC 层

负责：

- HTTP 路由
- 请求解析
- 参数校验（基础格式）
- 统一响应格式
- WebSocket 连接管理
- DTO 转换

不负责：

- 业务逻辑
- 状态持久化编排
- 直接调用 `goed2k.Client`

#### Service 层

负责：

- 业务流程编排
- 参数语义校验
- 错误归类
- 调用 engine
- 调用配置/存储模块
- 构建返回 DTO

#### Engine 层

负责：

- 持有唯一 `*goed2k.Client`
- 启动/停止底层 client
- 暴露线程安全的调用方法
- 适配底层 `goed2k` API
- 对订阅事件进行统一转发

#### Store/Config 层

负责：

- daemon 自身配置读写
- state 文件路径管理
- 自动保存调度参数

---

## 9. 运行模型

项目运行后：

1. 加载 daemon 配置
2. 初始化 engine
3. 根据配置构建 `goed2k.Settings`
4. 创建并启动 `goed2k.Client`
5. 若启用自动恢复，则加载 state
6. 启动 HTTP 服务
7. 启动 WebSocket hub
8. 启动自动保存协程
9. 启动状态订阅桥接协程

关闭时：

1. 停止接收新请求
2. 可选保存一次 state
3. 停止底层 client
4. 优雅退出

---

## 10. 配置设计

### 10.1 配置文件路径

默认：

```text
data/config/config.json
```

### 10.2 配置结构

```json
{
  "rpc": {
    "listen": "127.0.0.1:18080",
    "allow_remote": false,
    "auth_token": "change-me",
    "read_timeout_seconds": 15,
    "write_timeout_seconds": 15
  },
  "engine": {
    "listen_port": 4661,
    "udp_port": 4662,
    "enable_dht": true,
    "enable_upnp": true,
    "peer_connection_timeout": 30,
    "reconnect_to_server": true,
    "max_connections_per_second": 10,
    "session_connections_limit": 20,
    "upload_slots": 3,
    "max_upload_rate_kb": 0,
    "default_download_dir": "./data/downloads"
  },
  "bootstrap": {
    "server_addresses": [
      "45.82.80.155:5687"
    ],
    "server_met_urls": [
      "http://upd.emule-security.org/server.met"
    ],
    "nodes_dat_urls": [
      "https://upd.emule-security.org/nodes.dat"
    ],
    "kad_nodes": []
  },
  "state": {
    "enabled": true,
    "path": "./data/state/client-state.json",
    "load_on_start": true,
    "save_on_exit": true,
    "auto_save_interval_seconds": 30
  },
  "logging": {
    "level": "info",
    "path": "./data/logs/goed2kd.log"
  }
}
```

### 10.3 配置要求

- 提供默认配置生成能力
- 配置缺失时自动生成默认文件
- 启动时校验配置合法性
- `allow_remote=false` 时，listen 不能是 `0.0.0.0`
- `auth_token` 不能为空

---

## 11. Engine 设计

定义一个 `Engine` 结构体，对 `goed2k.Client` 做包装。

### 11.1 Engine 应具备的职责

- 初始化 `goed2k.Settings`
- 创建并持有 `*goed2k.Client`
- 暴露任务/搜索/网络相关方法
- 管理启动状态
- 暴露状态快照
- 暴露订阅桥接
- 保存/加载 state

### 11.2 推荐接口

```go
type Engine interface {
    Start() error
    Stop(ctx context.Context) error
    IsRunning() bool

    Info(ctx context.Context) (SystemInfo, error)
    Health(ctx context.Context) (HealthStatus, error)

    SaveState(ctx context.Context) error
    LoadState(ctx context.Context) error

    ConnectServer(ctx context.Context, addr string) error
    ConnectServers(ctx context.Context, addrs []string) error
    LoadServerMet(ctx context.Context, sources []string) error

    EnableDHT(ctx context.Context) error
    LoadDHTNodesDat(ctx context.Context, sources []string) error
    AddDHTBootstrapNodes(ctx context.Context, nodes []string) error

    Status(ctx context.Context) (ClientStatusDTO, error)
    Servers(ctx context.Context) ([]ServerDTO, error)
    DHTStatus(ctx context.Context) (DHTStatusDTO, error)

    AddTransferByED2K(ctx context.Context, req AddTransferRequest) (TransferDTO, error)
    ListTransfers(ctx context.Context) ([]TransferDTO, error)
    GetTransfer(ctx context.Context, hash string) (TransferDetailDTO, error)
    PauseTransfer(ctx context.Context, hash string) error
    ResumeTransfer(ctx context.Context, hash string) error
    DeleteTransfer(ctx context.Context, hash string, deleteFiles bool) error

    StartSearch(ctx context.Context, req StartSearchRequest) (SearchDTO, error)
    CurrentSearch(ctx context.Context) (SearchDTO, error)
    StopSearch(ctx context.Context) error
    AddTransferFromSearch(ctx context.Context, hash string, targetPath string) (TransferDTO, error)

    SubscribeStatus(ctx context.Context) (<-chan EventEnvelope, func(), error)
    SubscribeTransferProgress(ctx context.Context) (<-chan EventEnvelope, func(), error)
}
```

说明：

- 不要求一字不差，但结构要类似
- 对外只暴露字符串 hash，不向 RPC 泄漏底层复杂类型
- 所有返回对象都应为项目自己的 DTO，而不是 `goed2k` 原始结构体

---

## 12. DTO 设计

### 12.1 统一响应格式

成功：

```json
{
  "code": "OK",
  "message": "",
  "data": {}
}
```

失败：

```json
{
  "code": "TRANSFER_NOT_FOUND",
  "message": "transfer not found"
}
```

### 12.2 通用错误码

必须定义统一错误码：

- `OK`
- `BAD_REQUEST`
- `UNAUTHORIZED`
- `FORBIDDEN`
- `NOT_FOUND`
- `INTERNAL_ERROR`
- `ENGINE_NOT_RUNNING`
- `ENGINE_ALREADY_RUNNING`
- `INVALID_HASH`
- `INVALID_ED2K_LINK`
- `TRANSFER_NOT_FOUND`
- `SEARCH_NOT_RUNNING`
- `SEARCH_ALREADY_RUNNING`
- `CONFIG_INVALID`
- `STATE_STORE_ERROR`

### 12.3 Transfer DTO

```json
{
  "hash": "0123456789ABCDEF0123456789ABCDEF",
  "file_name": "example.iso",
  "file_path": "./data/downloads/example.iso",
  "size": 1234567890,
  "create_time": 1710000000,
  "state": "DOWNLOADING",
  "paused": false,
  "download_rate": 1024,
  "upload_rate": 64,
  "total_done": 734003200,
  "total_received": 734003200,
  "total_wanted": 1234567890,
  "eta": 3600,
  "num_peers": 12,
  "active_peers": 4,
  "downloading_pieces": 3,
  "progress": 0.59,
  "ed2k_link": "ed2k://|file|example.iso|1234567890|...|/"
}
```

### 12.4 Transfer Detail DTO

在基础 DTO 上增加：

- `peers`
- `pieces`

### 12.5 Server DTO

```json
{
  "identifier": "45.82.80.155:5687",
  "address": "45.82.80.155:5687",
  "configured": true,
  "connected": true,
  "handshake_completed": true,
  "primary": true,
  "disconnecting": false,
  "client_id": 123456,
  "id_class": "HIGH_ID",
  "download_rate": 0,
  "upload_rate": 0,
  "milliseconds_since_last_receive": 2500
}
```

### 12.6 Search DTO

```json
{
  "id": 1,
  "state": "RUNNING",
  "params": {
    "query": "ubuntu iso",
    "scope": "all",
    "min_size": 0,
    "max_size": 0,
    "min_sources": 0,
    "min_complete_sources": 0,
    "file_type": "",
    "extension": ""
  },
  "results": [
    {
      "hash": "...",
      "file_name": "ubuntu.iso",
      "file_size": 1234567890,
      "sources": 100,
      "complete_sources": 50,
      "media_bitrate": 0,
      "media_length": 0,
      "media_codec": "",
      "extension": "iso",
      "file_type": "CDImage",
      "source": "server|kad",
      "ed2k_link": "ed2k://|file|..."
    }
  ],
  "updated_at": 1710000000,
  "started_at": 1710000000,
  "server_busy": true,
  "dht_busy": true,
  "kad_keyword": "ubuntu",
  "error": ""
}
```

### 12.7 System DTO

包含：

- daemon version
- engine running
- config summary
- state path
- default download dir
- rpc listen address
- uptime

---

## 13. HTTP API 设计

Base path:

```text
/api/v1
```

### 13.1 系统接口

#### GET /api/v1/system/info

返回 daemon 与 engine 基本信息。

#### GET /api/v1/system/health

返回健康状态：

- daemon 是否运行
- engine 是否运行
- state store 是否可用
- rpc 服务是否可用

#### POST /api/v1/system/start

启动 engine。

#### POST /api/v1/system/stop

停止 engine。

#### POST /api/v1/system/save-state

立即保存 state。

#### POST /api/v1/system/load-state

立即加载 state。

#### GET /api/v1/system/config

返回当前配置。

#### PUT /api/v1/system/config

更新配置。

规则：

- 仅允许更新支持热更的字段
- 若涉及 engine 重建，可要求 stop/start 后生效
- 第一版至少支持修改 bootstrap 和 state 自动保存配置

---

### 13.2 网络接口

#### GET /api/v1/network/servers

返回 server 列表。

#### POST /api/v1/network/servers/connect

请求体：

```json
{
  "address": "45.82.80.155:5687"
}
```

#### POST /api/v1/network/servers/connect-batch

请求体：

```json
{
  "addresses": [
    "45.82.80.155:5687",
    "176.123.5.89:4725"
  ]
}
```

#### POST /api/v1/network/servers/load-met

请求体：

```json
{
  "sources": [
    "http://upd.emule-security.org/server.met"
  ]
}
```

#### GET /api/v1/network/dht

返回 DHT 状态。

#### POST /api/v1/network/dht/enable

开启 DHT。

#### POST /api/v1/network/dht/load-nodes

请求体：

```json
{
  "sources": [
    "https://upd.emule-security.org/nodes.dat"
  ]
}
```

#### POST /api/v1/network/dht/bootstrap-nodes

请求体：

```json
{
  "nodes": [
    "1.2.3.4:4661"
  ]
}
```

---

### 13.3 下载任务接口

#### GET /api/v1/transfers

返回任务列表。

支持 query 参数：

- `state`
- `paused`
- `limit`
- `offset`
- `sort`

第一版即使内部不分页，也要预留参数结构。

#### POST /api/v1/transfers

请求体：

```json
{
  "ed2k_link": "ed2k://|file|example.iso|1234567890|...|/",
  "target_dir": "./data/downloads",
  "target_name": "",
  "paused": false
}
```

要求：

- 支持解析 ed2k link
- 自动计算最终文件路径
- `target_name` 为空则使用链接内文件名
- 若目录不存在则自动创建

#### GET /api/v1/transfers/{hash}

返回单任务详情。

#### POST /api/v1/transfers/{hash}/pause

暂停任务。

#### POST /api/v1/transfers/{hash}/resume

恢复任务。

#### DELETE /api/v1/transfers/{hash}

支持 query 参数：

- `delete_files=true|false`

默认不删除文件。

#### GET /api/v1/transfers/{hash}/peers

返回 peers 列表。

#### GET /api/v1/transfers/{hash}/pieces

返回 pieces 快照。

---

### 13.4 搜索接口

#### POST /api/v1/searches

请求体：

```json
{
  "query": "ubuntu iso",
  "scope": "all",
  "min_size": 0,
  "max_size": 0,
  "min_sources": 0,
  "min_complete_sources": 0,
  "file_type": "",
  "extension": ""
}
```

规则：

- 第一阶段只允许同一时间一个 active search
- 若已有运行中的 search，返回 `SEARCH_ALREADY_RUNNING`

#### GET /api/v1/searches/current

返回当前搜索快照。

#### POST /api/v1/searches/current/stop

停止当前搜索。

#### POST /api/v1/searches/current/results/{hash}/download

请求体：

```json
{
  "target_dir": "./data/downloads",
  "target_name": "",
  "paused": false
}
```

规则：

- 从 current search results 中找到 hash 对应结果
- 用搜索结果生成 ed2k link
- 添加为下载任务

---

## 14. WebSocket 设计

### 14.1 地址

```text
GET /api/v1/events/ws
```

### 14.2 鉴权

支持以下任一方式：

- `Authorization: Bearer <token>`
- query 参数 `?token=...`

### 14.3 事件格式

统一 envelope：

```json
{
  "type": "transfer.progress",
  "at": "2026-03-31T10:00:00Z",
  "data": {}
}
```

### 14.4 必须支持的事件类型

#### client.status

推送整体状态快照，适合概览页。

#### transfer.progress

推送任务进度变化，适合任务列表与详情页。

### 14.5 心跳

必须支持 ping/pong 或服务端定时发送 heartbeat。

推荐：

- 每 20 秒发送一次心跳
- 60 秒内无响应则断开

### 14.6 Hub 设计要求

- 多客户端订阅
- 慢客户端不能阻塞全局
- 每个连接有发送缓冲区
- 缓冲区满时可断开慢客户端
- 连接断开要清理资源

---

## 15. 认证与安全

### 15.1 第一阶段最小安全模型

- 默认只监听 `127.0.0.1`
- 所有 HTTP API 除 `health` 外都需要 token
- WebSocket 需要 token
- token 从配置文件读取

### 15.2 Middleware 要求

需要实现：

- Recover middleware
- Request ID middleware
- Logger middleware
- Auth middleware
- Timeout middleware

### 15.3 敏感操作审计日志

记录以下操作：

- system start/stop
- save/load state
- connect server
- load server.met
- load nodes.dat
- add transfer
- delete transfer
- update config

---

## 16. 状态持久化设计

### 16.1 原则

下载引擎状态直接复用 `goed2k` 的 state store 机制。

应用层只负责：

- 决定 state 文件路径
- 调用 save/load
- 自动保存调度

### 16.2 自动保存

需要一个后台协程：

- 按配置间隔触发 `SaveState`
- 保存失败要记录日志
- 不应使服务崩溃

### 16.3 启动恢复

若 `state.load_on_start = true`，则：

1. 启动 engine
2. 调用 load state
3. 记录恢复结果

### 16.4 退出保存

若 `state.save_on_exit = true`，则在优雅退出时保存一次

---

## 17. 错误处理规范

### 17.1 原则

- service 层统一归类错误
- handler 层不直接返回底层原始错误文本给前端
- 日志记录完整错误
- API 返回用户可理解错误

### 17.2 错误映射

例如：

- 底层 hash 解析失败 -> `INVALID_HASH`
- 找不到 transfer -> `TRANSFER_NOT_FOUND`
- engine 未启动 -> `ENGINE_NOT_RUNNING`
- state 保存失败 -> `STATE_STORE_ERROR`

---

## 18. 日志设计

### 18.1 输出位置

默认输出到：

```text
./data/logs/goed2kd.log
```

### 18.2 日志内容

必须记录：

- 服务启动/停止
- 配置加载结果
- engine 启动/停止结果
- bootstrap 操作结果
- 每次 API 请求概要
- WebSocket 连接建立与断开
- 状态保存/恢复结果
- 关键错误堆栈或上下文

### 18.3 日志等级

支持：

- debug
- info
- warn
- error

---

## 19. 删除任务语义

需要明确删除操作分两种：

### 19.1 仅删除任务

- 从 engine 中移除任务
- 不删除磁盘文件

### 19.2 删除任务并删除文件

- 从 engine 中移除任务
- 删除目标文件
- 若有临时文件，也一并删除

第一阶段若底层不方便彻底删除文件，可先实现“移除任务 + 删除主文件”，但要在代码中留 TODO 注释并保证接口语义稳定。

---

## 20. 搜索模型约束

第一阶段限制：

- 只支持一个当前搜索
- 新搜索启动前必须先停止旧搜索，或返回错误
- 搜索结果仅保存在内存
- 搜索详情通过轮询 `GET /searches/current` 获取

原因：

- 降低复杂度
- 先与 `goed2k` 当前公开能力对齐
- 后续可扩展为多搜索任务管理器

---

## 21. 前端友好性要求

即使本阶段不实现 UI，也必须保证 RPC 对 UI 友好。

要求：

- DTO 字段命名稳定
- 时间统一为 Unix 秒或 RFC3339，不能混乱
- 枚举值统一大写字符串
- 返回体包含足够展示用信息，不让前端自行拼很多逻辑
- `progress` 直接返回 0~1 浮点数
- `ed2k_link` 直接返回，便于复制与分享

---

## 22. 代码实现要求

### 22.1 代码风格

- 使用 Go 1.22+ 语法风格
- 清晰分层
- 避免巨大 God object
- 避免 handler 中写业务逻辑
- 避免循环依赖
- 所有导出类型与方法写注释

### 22.2 并发要求

- Engine 内部访问底层 client 时要考虑线程安全
- WebSocket hub 要避免广播阻塞
- 自动保存和事件桥接不能互相阻塞
- 使用 context 控制 goroutine 生命周期

### 22.3 测试要求

至少添加以下测试：

- 配置加载与校验
- DTO mapper
- 错误映射
- HTTP handler 基本接口测试
- Auth middleware 测试
- transfer service 参数校验测试

若无法完整做集成测试，可先通过 mock engine 做 handler/service 测试。

---

## 23. 实现顺序（严格按此顺序）

### Phase A：项目骨架

1. 初始化 Go module
2. 建立目录结构
3. 实现 config 读写与默认值
4. 实现 logger
5. 实现统一 response/error 模型

### Phase B：engine 封装

1. 创建 `Engine` 封装
2. 实现 start/stop/status
3. 实现 save/load state
4. 实现 transfers 基础方法
5. 实现 search 基础方法
6. 实现 server / DHT 基础方法

### Phase C：service 层

1. SystemService
2. NetworkService
3. TransferService
4. SearchService

### Phase D：HTTP API

1. router
2. middleware
3. system handlers
4. network handlers
5. transfer handlers
6. search handlers

### Phase E：WebSocket

1. hub
2. 状态桥接
3. 任务进度桥接
4. 心跳
5. 认证

### Phase F：自动保存与优雅退出

1. auto-save goroutine
2. graceful shutdown
3. startup load state
4. exit save state

### Phase G：测试与修整

1. handler tests
2. middleware tests
3. config tests
4. 文档补充
5. config.example.json

---

## 24. 验收标准

### 24.1 基础运行

- 启动后自动生成默认配置
- daemon 可正常监听 HTTP 端口
- `/api/v1/system/health` 可返回健康状态

### 24.2 网络能力

- 可通过 API 连接 server
- 可通过 API 加载 server.met
- 可通过 API 启用 DHT 并加载 nodes.dat

### 24.3 下载能力

- 可通过 API 添加 ED2K 下载
- 可查询任务列表
- 可暂停/恢复任务
- 可查看任务详情、peers、pieces

### 24.4 搜索能力

- 可发起搜索
- 可查询搜索快照
- 可停止搜索
- 可从搜索结果添加下载

### 24.5 事件能力

- WebSocket 可收到 `client.status`
- WebSocket 可收到 `transfer.progress`
- 多客户端连接互不影响

### 24.6 持久化

- 手动保存 state 成功
- 重启后可加载已保存任务
- 自动保存可按配置执行

### 24.7 安全

- 未携带 token 的请求被拒绝
- WebSocket 未认证不能连接
- 默认不允许远程监听

---

## 25. 后续 Web UI 预留建议

后续 Web UI 建议路由：

- `/dashboard`
- `/transfers`
- `/transfers/:hash`
- `/search`
- `/network`
- `/settings`

前端状态模型：

- 初始页面通过 HTTP 拉全量
- 后续通过 WebSocket 订阅增量
- 搜索页先采用轮询 current search

---

## 26. Cursor 实现要求

请 Cursor 按以下要求执行：

1. 不要重写 `goed2k` 底层协议逻辑
2. 先完成后端 daemon 与 RPC
3. 代码必须分层，不能把全部逻辑塞进 `main.go`
4. 先保证 API 可用与稳定，再考虑额外优化
5. 所有接口都返回统一 JSON 结构
6. WebSocket 仅实现 `client.status` 与 `transfer.progress`
7. 搜索第一版用单 active search 模型
8. 先完成可运行版本，再补测试与细节

---

## 27. 额外工程要求

### 27.1 README

需要生成 README，包含：

- 项目简介
- 功能列表
- 配置说明
- 启动方式
- API 概览
- WebSocket 简述

### 27.2 配置样例

需要生成：

```text
configs/config.example.json
```

### 27.3 启动命令

支持：

```bash
go run ./cmd/goed2kd
```

后续可补：

```bash
go build -o bin/goed2kd ./cmd/goed2kd
```

---

## 28. 最终交付物

Cursor 第一轮实现后，至少要产出：

1. 可运行的 `goed2kd`
2. 完整 HTTP API
3. 可工作的 WebSocket 推送
4. 默认配置文件
5. 自动保存/恢复 state
6. README
7. 基础测试

---

## 29. 本次实现优先级总结

最高优先级：

1. daemon 跑起来
2. transfers API 跑通
3. status/progress WS 跑通
4. state save/load 跑通

中优先级：

5. search API
6. network bootstrap API
7. config update API

低优先级：

8. 更复杂的过滤/分页
9. 更细日志查询
10. 高级管理能力

---

## 30. 给 Cursor 的最终执行指令

请基于本实现文档，直接开始实现 `goed2kd` 项目。

要求：

- 使用 Go + chi
- 封装 `goed2k` 为 engine 层
- 实现 HTTP/JSON API 与 WebSocket
- 先做后端，不做前端
- 优先保证架构清晰、接口稳定、代码可运行
- 所有关键代码补充注释
- 每完成一个 phase，都保持项目可编译运行

如果底层 `goed2k` 某些删除/搜索/路径能力存在边界，请以最小侵入方式兼容，并在代码中写清楚 TODO 与限制，不要阻塞整体交付。

