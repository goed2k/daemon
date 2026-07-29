# Changelog

本文件记录 goed2kd 的版本变更。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [0.2.0] - 2026-07-29

### 新增

- 同步 [goed2k/core](https://github.com/goed2k/core) 至 `v0.0.3-0.20260729185258-04a253aeb7a7`，支持 KADV6（IPv6 DHT）、下载限速、协议混淆、Secure Ident 等配置映射
- HTTP API：`GET/POST /api/v1/network/dht-v6/*`（IPv6 KAD/DHT 状态、启用、加载 nodes6.dat、引导节点）
- HTTP API：`POST /api/v1/transfers/{hash}/priority`（下载任务优先级 P0–P4）
- `ClientStatusDTO` 增加 `dht_v6`；`TransferDTO` 增加 `download_priority` / `download_priority_label`
- `ServerDTO` 增加 `aux_port`；搜索结果增加 `note` 字段
- 配置项：`enable_dht_v6`、`udp_port_v6`、`max_download_rate_kb`、混淆/SecIdent 相关字段及 `nodes6_dat_urls` / `kad_v6_nodes`
- GitHub Actions CI：push/PR 自动执行 build、vet、race 测试
- 集成测试：`internal/integration/` 覆盖 API 与配置向后兼容性

### 修复

- 旧配置缺 `udp_port_v6` 时自动补全为 `4672`（`ApplyDefaults`）
- 添加下载时保留 ED2K 链接中的 AICH/分片哈希扩展段
- 运行时启用 DHT/DHTv6 后刷新 UPnP 端口映射

### 变更

- 构建时通过 goreleaser `-ldflags` 注入 `internal/version.Version`，`GET /system/info` 的 `daemon_version` 与发布标签一致

[0.2.0]: https://github.com/goed2k/daemon/releases/tag/v0.2.0
