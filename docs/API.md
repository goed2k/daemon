# goed2kd HTTP / WebSocket API

本文描述当前实现的 RPC 接口。基础路径前缀为 **`/api/v1`**。

## 通用约定

### Content-Type

- 请求与响应 JSON 均为 `application/json; charset=utf-8`（成功与业务错误均返回 JSON）。

### 统一响应结构

**成功：**

```json
{
  "code": "OK",
  "data": {}
}
```

**失败：**

```json
{
  "code": "ERROR_CODE",
  "message": "人类可读说明"
}
```

`message` 在成功时通常省略。部分成功响应的 `data` 可能为 `null` 或省略，以实际接口为准。

### HTTP 状态码与业务码

业务错误通过响应体中的 `code` 表达；HTTP 状态码用于粗分类：

| 业务码 | 典型 HTTP 状态 |
|--------|----------------|
| `BAD_REQUEST`, `INVALID_HASH`, `INVALID_ED2K_LINK`, `CONFIG_INVALID` | 400 |
| `UNAUTHORIZED` | 401 |
| `FORBIDDEN` | 403 |
| `NOT_FOUND`, `TRANSFER_NOT_FOUND`, `SHARED_FILE_NOT_FOUND`, `SEARCH_NOT_RUNNING` | 404 |
| `ENGINE_NOT_RUNNING` | 503 |
| `ENGINE_ALREADY_RUNNING`, `SEARCH_ALREADY_RUNNING`, `STATE_STORE_ERROR` | 409 |
| `INTERNAL_ERROR` | 500 |
| 其他未单独映射的 `AppError` | 400 |

### 鉴权

- **`GET /api/v1/system/health`**：**不需要** Token。
- **其余 HTTP 接口**（含 WebSocket 握手）：需携带与配置文件 `rpc.auth_token` 一致的凭证，任选一种：
  - 请求头：`Authorization: Bearer <token>`
  - 请求头：`X-Auth-Token: <token>`
  - 查询参数：`?token=<token>`（便于浏览器 WebSocket）

### 请求体大小

- JSON 请求体最大约 **1 MiB**；未知字段默认拒绝（`DisallowUnknownFields`）。

---

## 错误码列表

| code | 说明 |
|------|------|
| `OK` | 成功 |
| `BAD_REQUEST` | 参数或请求不合法 |
| `UNAUTHORIZED` | 未鉴权或 Token 错误 |
| `FORBIDDEN` | 禁止访问 |
| `NOT_FOUND` | 资源不存在（泛型） |
| `INTERNAL_ERROR` | 未归类的服务端错误 |
| `ENGINE_NOT_RUNNING` | 引擎未启动 |
| `ENGINE_ALREADY_RUNNING` | 引擎已在运行 |
| `INVALID_HASH` | 任务 hash 格式无效（32 位十六进制） |
| `INVALID_ED2K_LINK` | ED2K 链接无效或非文件链 |
| `TRANSFER_NOT_FOUND` | 找不到对应任务 |
| `SHARED_FILE_NOT_FOUND` | 共享库中不存在该 hash |
| `SEARCH_NOT_RUNNING` | 当前无运行中搜索（预留/部分场景） |
| `SEARCH_ALREADY_RUNNING` | 已有活跃搜索，需先停止 |
| `CONFIG_INVALID` | 配置校验失败 |
| `STATE_STORE_ERROR` | 状态未启用、路径无效或 save/load 失败 |

---

## 系统 `/system`

### GET `/system/health`

健康检查，**无需 Token**。

**响应 `data` 示例：**

```json
{
  "daemon_running": true,
  "engine_running": true,
  "state_store_ok": true,
  "rpc_available": true
}
```

### GET `/system/info`

守护进程与引擎概要。

**响应 `data` 字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `daemon_version` | string | 守护进程版本号 |
| `engine_running` | bool | 引擎是否运行中 |
| `uptime_seconds` | int64 | 引擎已运行秒数（未运行为 0） |
| `rpc_listen` | string | 配置中的 RPC 监听地址 |
| `state_path` | string | 配置中的状态文件路径 |
| `default_download_dir` | string | 默认下载目录 |

### POST `/system/start`

启动引擎（幂等：已运行则 `ENGINE_ALREADY_RUNNING`）。

**响应 `data`：** `{ "started": true }`

### POST `/system/stop`

停止引擎。

**响应 `data`：** `{ "stopped": true }`

### POST `/system/save-state`

立即保存状态（需启用 `state` 且引擎运行）。

**响应 `data`：** `{ "saved": true }`

### POST `/system/load-state`

立即加载状态（需启用 `state` 且引擎运行）。

**响应 `data`：** `{ "loaded": true }`

### GET `/system/config`

返回当前完整配置 JSON（与磁盘结构一致，**含 `auth_token`**，仅限可信环境）。

### PUT `/system/config`

热更新**允许字段**并写回配置文件。请求体只需包含要更新的块；未出现的块保持不变。

**请求体（均可选，按需组合）：**

```json
{
  "bootstrap": {
    "server_addresses": ["host:port"],
    "server_met_urls": ["http://..."],
    "nodes_dat_urls": ["https://..."],
    "kad_nodes": ["ip:port"]
  },
  "state": {
    "enabled": true,
    "path": "./data/state/client-state.json",
    "load_on_start": true,
    "save_on_exit": true,
    "auto_save_interval_seconds": 30
  },
  "logging": {
    "level": "info"
  }
}
```

**响应 `data`：** 更新后的完整配置（与 GET 相同）。

---

## 网络 `/network`

### GET `/network/servers`

**响应 `data`：** `ServerDTO` 数组。

| 字段 | 类型 | 说明 |
|------|------|------|
| `identifier` | string | 连接标识 |
| `address` | string | 地址 |
| `name` | string | 服务器名（来自 `server.met`，若通过 `ConnectServerMet` 连接） |
| `description` | string | 描述（同上） |
| `configured` | bool | 是否已配置 |
| `connected` | bool | 是否连接 |
| `handshake_completed` | bool | 握手是否完成 |
| `primary` | bool | 是否主连接 |
| `disconnecting` | bool | 是否正在断开 |
| `client_id` | int32 | **本机**在该服务器分配的客户端 ID |
| `id_class` | string | 本机 ID 类型：`HIGH_ID` / `LOW_ID` / `UNKNOWN` |
| `tcp_flags` | int32 | 服务器 `IdChange` 中的 TCP 标志（原始值） |
| `reported_ip` | uint32 | 扩展 `IdChange` 中的 ReportedIP（若有） |
| `obfuscation_tcp_port` | uint32 | 扩展 `IdChange` 中的混淆监听端口；非 0 表示支持乱序加密通告 |
| `status_users` / `status_files` | int32 | TCP `Status`（0x34）包中的用户/文件数 |
| `udp_users` / `udp_files` | uint32 | UDP `GlobServStat` 响应 |
| `max_users` | uint32 | 最大用户数（UDP） |
| `soft_files_limit` / `hard_files_limit` | uint32 | 软性/硬性文件限制（UDP） |
| `udp_stats_valid` | bool | 是否已成功解析过 UDP 统计（需本机 UDP 可用） |
| `download_rate` | int | 下载速率（实现相关） |
| `upload_rate` | int | 上传速率 |
| `milliseconds_since_last_receive` | int64 | 距上次接收毫秒 |

### GET `/network/peers`

**响应 `data`：** `ClientPeerEntryDTO` 数组。列出当前所有下载任务上的**对端**（与底层 `ClientStatus.Peers` / 各任务 `PeerInfo` 一致），即全局「已知客户端」视图。

每条包含：

| 字段 | 类型 | 说明 |
|------|------|------|
| `transfer_hash` | string | 所属任务文件哈希（十六进制） |
| `file_name` | string | 任务文件名 |
| `file_path` | string | 本地路径 |
| `peer` | `PeerDTO` | 对端详情（与 `GET /transfers/{hash}/peers` 中单条结构相同） |

引擎未运行或尚无任务/对端时返回空数组 `[]`（若引擎未启动则可能返回 `503` / `ENGINE_NOT_RUNNING`，与现有接口一致）。

### POST `/network/servers/connect`

```json
{ "address": "host:port" }
```

**响应 `data`：** `{ "ok": true }`

### POST `/network/servers/connect-batch`

```json
{ "addresses": ["host:port", "..."] }
```

**响应 `data`：** `{ "ok": true }`

### POST `/network/servers/load-met`

从 URL 或本地路径加载 `server.met` 并连接其中服务器（行为由 goed2k 实现）。

```json
{ "sources": ["http://.../server.met"] }
```

**响应 `data`：** `{ "ok": true }`

### GET `/network/dht`

**响应 `data`：** `DHTStatusDTO`（引擎未运行时字段多为零值）。

### POST `/network/dht/enable`

运行时启用 DHT（若引擎已运行但未建 tracker，会创建并启动）。

**响应 `data`：** `{ "ok": true }`

### POST `/network/dht/load-nodes`

```json
{ "sources": ["https://.../nodes.dat", "/path/to/nodes.dat"] }
```

**响应 `data`：** `{ "ok": true }`

### POST `/network/dht/bootstrap-nodes`

```json
{ "nodes": ["1.2.3.4:4661"] }
```

**响应 `data`：** `{ "ok": true }`

---

## 下载任务 `/transfers`

路径参数 **`{hash}`** 为 **32 位十六进制** ED2K 文件哈希（大小写均可，与底层解析一致）。

### GET `/transfers`

查询参数（可选，第一版部分为预留/简单过滤）：

| 参数 | 说明 |
|------|------|
| `state` | 按任务状态字符串过滤（大小写不敏感） |
| `paused` | `true` / `1` 或 `false` / `0` |
| `limit` | 返回条数上限 |
| `offset` | 偏移 |
| `sort` | 预留 |

**响应 `data`：** `TransferDTO` 数组。

主要字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `hash` | string | 文件哈希 |
| `file_name` | string | 文件名 |
| `file_path` | string | 本地路径 |
| `size` | int64 | 字节 |
| `create_time` | int64 | Unix 时间戳（秒或毫秒，与 goed2k 一致） |
| `state` | string | 如 `DOWNLOADING`, `PAUSED`, `FINISHED` 等 |
| `paused` | bool | 是否暂停 |
| `download_rate` / `upload_rate` | int | 速率 |
| `total_done` / `total_received` / `total_wanted` | int64 | 进度相关 |
| `eta` | int64 | 预估剩余秒（底层语义） |
| `num_peers` / `active_peers` / `downloading_pieces` | int | 统计 |
| `progress` | float | 0~1 |
| `ed2k_link` | string | 完整 ED2K 链接 |

### POST `/transfers`

```json
{
  "ed2k_link": "ed2k://|file|name|size|HASH|/",
  "target_dir": "./data/downloads",
  "target_name": "",
  "paused": false
}
```

- `target_dir` 可省略，则用配置中 `default_download_dir`。
- `target_name` 为空则使用链接内文件名。

**响应 `data`：** `TransferDTO`。

### GET `/transfers/{hash}`

**响应 `data`：** `TransferDetailDTO`（在 `TransferDTO` 基础上增加 `peers`、`pieces`）。

### POST `/transfers/{hash}/pause` / `resume`

**响应 `data`：** `{ "ok": true }`

### DELETE `/transfers/{hash}`

查询参数：

| 参数 | 说明 |
|------|------|
| `delete_files` | `true` / `1` 表示同时删文件；默认仅移除任务 |

**响应 `data`：** `{ "ok": true }`

### GET `/transfers/{hash}/peers`

**响应 `data`：** `PeerDTO` 数组。对端 **Hello / HelloAnswer** 标签解析结果（昵称、客户端标识、`hello_misc1` / `hello_misc2` 等）由底层 `goed2k.PeerInfo` 提供，经本接口透出。

| 字段 | 类型 | 说明 |
|------|------|------|
| `endpoint` | string | `host:port` |
| `user_hash` | string | 对端用户 Hash（十六进制） |
| `nick_name` | string | Hello 昵称标签 |
| `connected` | bool | 当前是否仍连接 |
| `total_uploaded` / `total_downloaded` | uint64 | 与本用户累计上下传（积分，字节） |
| `download_speed` / `payload_download_speed` / `upload_speed` / `payload_upload_speed` | int | 速率（字节/秒） |
| `source` | string | 来源标记拼接（如 server、kad、resume 等） |
| `mod_name` | string | Mod 名称（Hello） |
| `version` | int | 客户端版本号（Hello 标签 0x11） |
| `mod_version` | int | Mod 复合版本原始值 |
| `str_mod_version` | string | Mod 版本可读串（如 `2.3.1`） |
| `hello_misc1` | int | Hello 标签 **0xFA**（Misc options）原始整型 |
| `hello_misc2` | int | Hello 标签 **0xFE**（Misc options 2）原始整型 |
| `fail_count` | int | 失败计数 |

### GET `/transfers/{hash}/pieces`

**响应 `data`：** `PieceDTO` 数组（`index`, `state`, 各字节/块计数）。

---

## 搜索 `/searches`

同一时间仅允许 **一个** 活跃搜索；重复发起返回 `SEARCH_ALREADY_RUNNING`。

### POST `/searches`

```json
{
  "query": "关键词",
  "scope": "all",
  "min_size": 0,
  "max_size": 0,
  "min_sources": 0,
  "min_complete_sources": 0,
  "file_type": "",
  "extension": ""
}
```

`scope`：`all`（默认）、`server`、`dht`（或 `kad`）。

**响应 `data`：** `SearchDTO`（含 `id`, `state`, `params`, `results`, `updated_at`, `started_at`, `server_busy`, `dht_busy`, `kad_keyword`, `error` 等）。

无活跃任务时 `GET /searches/current` 中 `state` 可能为 `IDLE`。

### GET `/searches/current`

**响应 `data`：** 当前 `SearchDTO`。

### POST `/searches/current/stop`

**响应 `data`：** `{ "ok": true }`

### POST `/searches/current/results/{hash}/download`

从**当前搜索结果**中取指定 hash 生成下载任务。

```json
{
  "target_dir": "./data/downloads",
  "target_name": "",
  "paused": false
}
```

**响应 `data`：** `TransferDTO`。

---

## 共享库 `/shared`

对应 [goed2k/core](https://github.com/goed2k/core) 的 `Client` 共享 API：`SharedFiles`、`AddSharedDir`、`RemoveSharedDir`、`ListSharedDirs`、`RescanSharedDirs`、`ImportSharedFile`、`RemoveSharedFile`。引擎未运行时返回 `ENGINE_NOT_RUNNING`。

### GET `/shared/files`

**响应 `data`：** `SharedFileDTO` 数组。

| 字段 | 类型 | 说明 |
|------|------|------|
| `hash` | string | ED2K 根哈希（十六进制） |
| `file_size` | int64 | 字节 |
| `path` | string | 本地路径 |
| `name` | string | 展示名（通常文件名） |
| `origin` | string | `downloaded`（下载完成入库）或 `imported`（本地导入） |
| `completed` | bool | 是否视为可共享完成态 |
| `can_upload` | bool | 是否可向其他 peer 上传 |
| `last_hash_at` | int64 | 最近哈希时间戳（内核语义） |

### GET `/shared/dirs`

**响应 `data`：** 字符串数组，已注册的扫描目录绝对路径。

### POST `/shared/dirs`

注册一个用于扫描的目录（须为已存在目录）。

```json
{ "path": "/path/to/dir" }
```

**响应 `data`：** `{ "ok": true }`

### POST `/shared/dirs/remove`

移除扫描目录（不删除磁盘文件）。

```json
{ "path": "/path/to/dir" }
```

**响应 `data`：** `{ "ok": true }`

### POST `/shared/dirs/rescan`

扫描已注册目录下的普通文件并尝试导入共享库（与内核 `RescanSharedDirs` 一致）。

**响应 `data`：** `{ "ok": true }`

### POST `/shared/import`

对单个文件计算 ED2K 元数据并加入共享库。

```json
{ "path": "/path/to/file.ext" }
```

**响应 `data`：** `{ "ok": true }`（若 hash 已存在，内核不覆盖，仍可能成功返回）

### DELETE `/shared/files/{hash}`

**路径参数：** `{hash}` 为 32 位十六进制 ED2K 哈希。

**响应 `data`：** `{ "ok": true }`

若 hash 不在共享库中，返回 `SHARED_FILE_NOT_FOUND`。

---

## WebSocket `/events/ws`

- **URL：** `GET /api/v1/events/ws`（与 HTTP 相同主机端口）。
- **鉴权：** 握手请求需带 Token（`Authorization: Bearer` / `X-Auth-Token` / `?token=`）。
- **消息格式：** 每条一条 **文本帧**，JSON 为事件外壳：

```json
{
  "type": "client.status",
  "at": "2026-03-31T10:00:00Z",
  "data": {}
}
```

- **`at`：** RFC3339 UTC 时间。
- **服务端心跳：** 约每 **20 秒** 发送 **WebSocket Ping**；客户端应响应 Pong（浏览器默认会处理）。

### 事件类型

| type | data 说明 |
|------|-----------|
| `client.status` | `data.status` 为引擎状态快照（`engine_running`, `servers`, `transfers`, `dht`, `totals` 等） |
| `transfer.progress` | `data.progress` 内含 `transfers` 数组（进度有变化的子集，字段与列表 DTO 对齐为主） |

引擎未运行时仍会周期性推送「空/停止」意义上的状态，便于前端统一处理。

### 广播语义

- 多客户端独立订阅；Hub 对慢客户端采用**有界缓冲**，满则**丢弃该条**而不阻塞其他连接。

---

## 版本与变更

API 随 `goed2kd` 版本迭代；本文以仓库当前实现为准。若与 [goed_2_kd_rpc_ui_implementation_spec.md](../goed_2_kd_rpc_ui_implementation_spec.md) 有出入，以**代码行为**为准。
