# goed2kd

A daemon built on [goed2k/core](https://github.com/goed2k/core): without changing the underlying ED2K protocol implementation, it exposes **HTTP/JSON RPC** and **WebSocket** events for scripting, automation, and future Web UI integration.

## Feature overview

- **System**: health checks, runtime info, engine start/stop, state save/load, config read/write and hot reload (selected fields)
- **Network**: connect to servers, batch connect, load `server.met`, DHT status, enable DHT, load `nodes.dat`, add KAD bootstrap nodes
- **Downloads**: add ED2K links, list/detail, pause/resume/remove, peers, pieces
- **Search**: start search, current snapshot, stop, add downloads from current results (only one active search at a time, per underlying constraints)
- **Shared library**: list shared files, add/remove scan dirs, rescan dirs, import file, remove by hash (aligned with goed2k/core shared APIs)
- **Live events**: WebSocket pushes `client.status`, `transfer.progress`

## Architecture notes

- HTTP routing: [chi](https://github.com/go-chi/chi)
- Layers: `rpc` (HTTP/WS) → `service` → `engine` (sole wrapper around `goed2k.Client`) → `config` / `store`
- Handlers **must not** use `goed2k.Client` directly; all business logic goes through `engine` and `service`

## Requirements

- Go **1.25+** (same as the `go` directive in [go.mod](./go.mod))
- Dependencies: see [go.mod](./go.mod)

Go module path: `github.com/goed2k/daemon` (use this import prefix when importing as a library).

## Quick start

```bash
# Run in development (first run creates default config at -config path)
go run ./cmd/goed2kd -config data/config/config.json

# Build
go build -o bin/goed2kd ./cmd/goed2kd
./bin/goed2kd -config data/config/config.json
```

### Docker (Alpine multi-stage image)

Build locally:

```bash
docker build -t goed2kd .
docker run --rm -p 18080:18080 -p 4661:4661 -p 4662:4662/udp \
  -v goed2kd-data:/app/data goed2kd
```

On CI release (git tag), images are pushed to **GitHub Container Registry** and **Docker Hub**, for example:

```bash
docker pull ghcr.io/chenjia404/goed2kd:latest
docker pull chenjia404/goed2kd:latest
```

If GHCR is private, run `docker login ghcr.io` first (PAT needs `read:packages`). Same idea for private Docker Hub with `docker login`.

First start creates default config inside the data volume. To reach RPC inside the container from the host or another machine, set `rpc.listen` to `0.0.0.0:18080` and `rpc.allow_remote` to `true` in config (otherwise validation rejects listening on `0.0.0.0`). You can also mount a custom `config.json`: `-v /path/config.json:/app/data/config/config.json`.

Default bind is **`127.0.0.1:18080`**. Default token is `change-me`; **change it in production**.

### Health check (no token)

```bash
curl -s http://127.0.0.1:18080/api/v1/system/health
```

### Authenticated example

```bash
curl -s -H "Authorization: Bearer change-me" http://127.0.0.1:18080/api/v1/system/info
```

## Configuration

- Config file path is set via `-config`; in-repo defaults often use `data/config/config.json`.
- If the file is missing, the process **creates** a default config matching the structure of [configs/config.example.json](./configs/config.example.json).
- Main sections:
  - **`rpc`**: `listen`, `auth_token`, `allow_remote` (when `false`, binding to `0.0.0.0` / `::` is not allowed), read/write timeouts
  - **`engine`**: listen ports, UDP, DHT/UPnP, connection and upload limits, default download directory
  - **`bootstrap`**: servers used after startup, `server.met` / `nodes.dat` URLs, KAD bootstrap nodes
  - **`state`**: enabled, `path`, `load_on_start`, `save_on_exit`, auto-save interval (seconds)
  - **`logging`**: `level` (debug/info/warn/error), log file path

State files are read/written by **goed2k’s built-in mechanism**; the daemon handles paths, scheduling, and save on exit.

## Behavior

- **Startup**: load config → start HTTP → start event bridge → **try to start the engine automatically** (HTTP stays up on failure so you can retry with `POST /system/start` later).
- **Graceful shutdown** (SIGINT/SIGTERM): stop HTTP and event bridge; if `state.save_on_exit` and the engine is running, save state then stop the engine.
- **Auto-save**: when state is enabled and the interval is greater than 0, uses goed2k Client’s auto-save interval and internal loop (aligned with `state.auto_save_interval_seconds`).

## API and WebSocket

Full reference (paths, request/response bodies, error codes, WS protocol):

**[docs/API.md](./docs/API.md)**

## Repository layout (excerpt)

```text
cmd/goed2kd/          # entrypoint
internal/
  app/                # logging, daemon lifecycle
  config/             # config model and loading
  engine/             # goed2k wrapper
  model/              # DTOs, error codes, event models
  rpc/http/           # HTTP routes and handlers
  rpc/ws/             # WebSocket hub
  service/            # business orchestration
  store/              # config storage abstraction
configs/              # example config
docs/                 # docs (including API)
```

## License

If this repository does not specify a license separately, follow the LICENSE file at the repository root; the [goed2k/core](https://github.com/goed2k/core) dependency is governed by that project’s license.
