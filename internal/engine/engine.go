package engine

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goed2k/core"
	"github.com/goed2k/daemon/internal/config"
	"github.com/goed2k/daemon/internal/model"
	"github.com/goed2k/daemon/internal/store"
)

// Engine 封装唯一 goed2k.Client，所有对内核的访问经此入口。
type Engine struct {
	mu      sync.RWMutex
	log     *slog.Logger
	cfg     *store.AppConfigStore
	client  *goed2k.Client
	running bool
	started time.Time

	runMu     sync.Mutex
	runCtx    context.Context
	runCancel context.CancelFunc
}

// NewEngine 构造引擎（此时未启动底层client）。
func NewEngine(log *slog.Logger, cfg *store.AppConfigStore) *Engine {
	return &Engine{log: log, cfg: cfg}
}

func (e *Engine) currentCfg() *config.Config {
	return e.cfg.Get()
}

func (e *Engine) settingsFromConfig() goed2k.Settings {
	c := e.currentCfg().Engine
	st := goed2k.NewSettings()
	st.ListenPort = c.ListenPort
	st.UDPPort = c.UDPPort
	st.EnableDHT = c.EnableDHT
	st.EnableUPnP = c.EnableUPnP
	st.PeerConnectionTimeout = c.PeerConnectionTimeout
	st.ReconnectToServer = c.ReconnectToServer
	st.MaxConnectionsPerSecond = c.MaxConnectionsPerSecond
	st.SessionConnectionsLimit = c.SessionConnectionsLimit
	st.UploadSlots = c.UploadSlots
	st.MaxUploadRateKB = c.MaxUploadRateKB
	if e.log != nil {
		st.Logger = e.log
	}
	return st
}

func (e *Engine) applyBootstrap(cli *goed2k.Client) {
	b := e.currentCfg().Bootstrap
	if len(b.ServerAddresses) > 0 {
		if err := cli.ConnectServers(b.ServerAddresses...); err != nil && e.log != nil {
			e.log.Warn("bootstrap connect servers", "err", err)
		}
	}
	for _, u := range b.ServerMetURLs {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		if err := cli.ConnectServerMet(u); err != nil && e.log != nil {
			e.log.Warn("bootstrap server.met", "url", u, "err", err)
		}
	}
	if e.currentCfg().Engine.EnableDHT || len(b.NodesDatURLs) > 0 || len(b.KadNodes) > 0 {
		if len(b.NodesDatURLs) > 0 {
			if err := cli.LoadDHTNodesDat(b.NodesDatURLs...); err != nil && e.log != nil {
				e.log.Warn("bootstrap nodes.dat", "err", err)
			}
		}
		if len(b.KadNodes) > 0 {
			if err := cli.AddDHTBootstrapNodes(b.KadNodes...); err != nil && e.log != nil {
				e.log.Warn("bootstrap kad nodes", "err", err)
			}
		}
	}
}

// Start 创建并启动goed2k.Client。
func (e *Engine) Start(ctx context.Context) error {
	_ = ctx
	e.runMu.Lock()
	defer e.runMu.Unlock()

	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return model.NewAppError(model.CodeEngineAlreadyRunning, "engine already running", nil)
	}
	st := e.settingsFromConfig()
	cli := goed2k.NewClient(st)
	sc := e.currentCfg().State
	if sc.Enabled && strings.TrimSpace(sc.Path) != "" {
		cli.SetStatePath(sc.Path)
		interval := time.Duration(sc.AutoSaveIntervalSeconds) * time.Second
		if interval > 0 {
			cli.SetAutoSaveInterval(interval)
		}
	}
	if err := cli.Start(); err != nil {
		e.mu.Unlock()
		return model.NewAppError(model.CodeInternalError, "engine start failed", err)
	}
	if sc.Enabled && sc.LoadOnStart {
		if err := cli.LoadState(""); err != nil {
			if e.log != nil {
				e.log.Warn("load state on start", "err", err)
			}
		}
	}
	e.applyBootstrap(cli)
	e.runCtx, e.runCancel = context.WithCancel(context.Background())
	e.client = cli
	e.running = true
	e.started = time.Now()
	e.mu.Unlock()
	if e.log != nil {
		e.log.Info("engine started")
	}
	return nil
}

// Stop 停止底层 client。
func (e *Engine) Stop(ctx context.Context) error {
	e.runMu.Lock()
	defer e.runMu.Unlock()

	e.mu.Lock()
	if e.runCancel != nil {
		e.runCancel()
	}
	cli := e.client
	e.client = nil
	e.running = false
	e.mu.Unlock()

	if cli == nil {
		return nil
	}
	err := cli.Stop()
	if e.log != nil {
		e.log.Info("engine stopped", "err", err)
	}
	_ = ctx
	return err
}

// IsRunning 是否已启动。
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Uptime 已运行时长。
func (e *Engine) Uptime() time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if !e.running {
		return 0
	}
	return time.Since(e.started)
}

func (e *Engine) requireClient() (*goed2k.Client, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if !e.running || e.client == nil {
		return nil, model.NewAppError(model.CodeEngineNotRunning, "engine not running", nil)
	}
	return e.client, nil
}

func (e *Engine) requireStateEnabled() error {
	if !e.currentCfg().State.Enabled {
		return model.NewAppError(model.CodeStateStoreError, "state persistence disabled", nil)
	}
	if strings.TrimSpace(e.currentCfg().State.Path) == "" {
		return model.NewAppError(model.CodeStateStoreError, "state path empty", nil)
	}
	return nil
}

// SaveState 立即保存状态。
func (e *Engine) SaveState(ctx context.Context) error {
	_ = ctx
	if err := e.requireStateEnabled(); err != nil {
		return err
	}
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if err := cli.SaveState(""); err != nil {
		return model.NewAppError(model.CodeStateStoreError, "save state failed", err)
	}
	return nil
}

// LoadState 立即加载状态。
func (e *Engine) LoadState(ctx context.Context) error {
	_ = ctx
	if err := e.requireStateEnabled(); err != nil {
		return err
	}
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if err := cli.LoadState(""); err != nil {
		return model.NewAppError(model.CodeStateStoreError, "load state failed", err)
	}
	return nil
}

// Info 系统信息快照。
func (e *Engine) Info(ctx context.Context) (*model.SystemInfo, error) {
	_ = ctx
	cfg := e.currentCfg()
	e.mu.RLock()
	run := e.running
	var up int64
	if run {
		up = int64(time.Since(e.started).Seconds())
	}
	e.mu.RUnlock()
	return &model.SystemInfo{
		DaemonVersion:      "0.1.0",
		EngineRunning:      run,
		UptimeSeconds:      up,
		RPCListen:          cfg.RPC.Listen,
		StatePath:          cfg.State.Path,
		DefaultDownloadDir: cfg.Engine.DefaultDownloadDir,
	}, nil
}

// Health 健康检查数据（rpc_available 用handler 填入）。
func (e *Engine) Health(ctx context.Context) (*model.HealthStatus, error) {
	_ = ctx
	cfg := e.currentCfg()
	ok := true
	if cfg.State.Enabled && strings.TrimSpace(cfg.State.Path) != "" {
		dir := filepath.Dir(cfg.State.Path)
		if _, err := filepath.Abs(dir); err != nil {
			ok = false
		}
	}
	e.mu.RLock()
	run := e.running
	e.mu.RUnlock()
	return &model.HealthStatus{
		DaemonRunning: true,
		EngineRunning: run,
		StateStoreOK:  ok,
		RPCAvailable:  true,
	}, nil
}

// ClientStatus 当前引擎状态DTO。
func (e *Engine) ClientStatus(ctx context.Context) (*model.ClientStatusDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		st := goed2k.ClientStatus{}
		dto := mapClientStatus(false, st, goed2k.DHTStatus{})
		return &dto, nil
	}
	ev := cli.Status()
	dto := mapClientStatus(true, ev, cli.DHTStatus())
	return &dto, nil
}

// Servers 列表。
func (e *Engine) Servers(ctx context.Context) ([]model.ServerDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	ss := cli.ServerStatuses()
	out := make([]model.ServerDTO, 0, len(ss))
	for _, s := range ss {
		out = append(out, mapServer(s))
	}
	return out, nil
}

// DHTStatus 当前 DHT。
func (e *Engine) DHTStatus(ctx context.Context) (*model.DHTStatusDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		z := mapDHT(goed2k.DHTStatus{})
		return &z, nil
	}
	d := cli.DHTStatus()
	dd := mapDHT(d)
	return &dd, nil
}

// KnownPeers 当前所有下载任务上的对端（全局已知客户端列表）。
func (e *Engine) KnownPeers(ctx context.Context) ([]model.ClientPeerEntryDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	st := cli.Status()
	out := make([]model.ClientPeerEntryDTO, 0, len(st.Peers))
	for _, p := range st.Peers {
		out = append(out, mapClientPeerEntry(p))
	}
	return out, nil
}

// ConnectServer 连接单个服务器。
func (e *Engine) ConnectServer(ctx context.Context, addr string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(addr) == "" {
		return model.NewAppError(model.CodeBadRequest, "address required", nil)
	}
	return cli.Connect(addr)
}

// ConnectServers 批量连接。
func (e *Engine) ConnectServers(ctx context.Context, addrs []string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return model.NewAppError(model.CodeBadRequest, "addresses required", nil)
	}
	return cli.ConnectServers(addrs...)
}

// LoadServerMetSources 从多个源加载 server.met 并尝试连接其中服务器。
func (e *Engine) LoadServerMetSources(ctx context.Context, sources []string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	for _, s := range sources {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if err := cli.ConnectServerMet(s); err != nil {
			return err
		}
	}
	return nil
}

// EnableDHT 运行时启用DHT（若尚未启动 tracker 则创建并监听）。
func (e *Engine) EnableDHT(ctx context.Context) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	tr := cli.EnableDHT()
	if tr != nil {
		if err := tr.Start(); err != nil {
			return model.NewAppError(model.CodeInternalError, "dht start failed", err)
		}
		cli.Session().SyncDHTListenPort()
	}
	return nil
}

// LoadDHTNodesSources 加载 nodes.dat（可多源）。
func (e *Engine) LoadDHTNodesSources(ctx context.Context, sources []string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return model.NewAppError(model.CodeBadRequest, "sources required", nil)
	}
	return cli.LoadDHTNodesDat(sources...)
}

// AddDHTBootstrapNodes 添加引导节点。
func (e *Engine) AddDHTBootstrapNodes(ctx context.Context, nodes []string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return model.NewAppError(model.CodeBadRequest, "nodes required", nil)
	}
	return cli.AddDHTBootstrapNodes(nodes...)
}

// AddTransferParams HTTP 层使用的添加任务参数。
type AddTransferParams struct {
	ED2KLink   string
	TargetDir  string
	TargetName string
	Paused     bool
}

// AddTransferByED2K 解析 ED2K 并添加下载任务。
func (e *Engine) AddTransferByED2K(ctx context.Context, p AddTransferParams) (*model.TransferDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	link, err := goed2k.ParseEMuleLink(strings.TrimSpace(p.ED2KLink))
	if err != nil {
		return nil, model.NewAppError(model.CodeInvalidED2KLink, "invalid ed2k link", err)
	}
	if link.Type != goed2k.LinkFile {
		return nil, model.NewAppError(model.CodeInvalidED2KLink, "not a file link", nil)
	}
	name := link.StringValue
	if strings.TrimSpace(p.TargetName) != "" {
		name = p.TargetName
	}
	dir := strings.TrimSpace(p.TargetDir)
	if dir == "" {
		dir = e.currentCfg().Engine.DefaultDownloadDir
	}
	synthetic := goed2k.FormatLink(name, link.NumberValue, link.Hash)
	_, targetPath, err := cli.AddLink(synthetic, dir)
	if err != nil {
		return nil, model.NewAppError(model.CodeBadRequest, "add transfer failed", err)
	}
	if p.Paused {
		_ = cli.PauseTransfer(link.Hash)
	}
	// 映射返回：从快照查找
	for _, ts := range cli.TransferSnapshots() {
		if ts.Hash.Compare(link.Hash) == 0 {
			t := mapTransfer(ts)
			t.FilePath = targetPath
			return &t, nil
		}
	}
	t := mapTransfer(goed2k.TransferSnapshot{
		Hash: link.Hash, FileName: name, FilePath: targetPath, Size: link.NumberValue,
		Status: goed2k.TransferStatus{Paused: p.Paused, State: goed2k.Downloading},
	})
	return &t, nil
}

// ListTransfers 返回任务列表（简单实现：全量；过滤由 service 做）。
func (e *Engine) ListTransfers(ctx context.Context) ([]model.TransferDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	snaps := cli.TransferSnapshots()
	out := make([]model.TransferDTO, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, mapTransfer(s))
	}
	return out, nil
}

// GetTransfer 单任务详情。
func (e *Engine) GetTransfer(ctx context.Context, hashHex string) (*model.TransferDetailDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	h, herr := parseHashParam(hashHex)
	if herr != nil {
		return nil, model.NewAppError(model.CodeInvalidHash, "invalid hash", herr)
	}
	for _, s := range cli.TransferSnapshots() {
		if s.Hash.Compare(h) == 0 {
			base := mapTransfer(s)
			peers := make([]model.PeerDTO, 0, len(s.Peers))
			for _, p := range s.Peers {
				peers = append(peers, mapPeer(p))
			}
			pieces := make([]model.PieceDTO, 0, len(s.Pieces))
			for _, pc := range s.Pieces {
				pieces = append(pieces, mapPiece(pc))
			}
			return &model.TransferDetailDTO{TransferDTO: base, Peers: peers, Pieces: pieces}, nil
		}
	}
	return nil, model.NewAppError(model.CodeTransferNotFound, "transfer not found", nil)
}

// PauseTransfer 暂停。
func (e *Engine) PauseTransfer(ctx context.Context, hashHex string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	h, herr := parseHashParam(hashHex)
	if herr != nil {
		return model.NewAppError(model.CodeInvalidHash, "invalid hash", herr)
	}
	if err := cli.PauseTransfer(h); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return model.NewAppError(model.CodeTransferNotFound, "transfer not found", err)
		}
		return err
	}
	return nil
}

// ResumeTransfer 恢复。
func (e *Engine) ResumeTransfer(ctx context.Context, hashHex string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	h, herr := parseHashParam(hashHex)
	if herr != nil {
		return model.NewAppError(model.CodeInvalidHash, "invalid hash", herr)
	}
	if err := cli.ResumeTransfer(h); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return model.NewAppError(model.CodeTransferNotFound, "transfer not found", err)
		}
		return err
	}
	return nil
}

// DeleteTransfer 删除任务。
func (e *Engine) DeleteTransfer(ctx context.Context, hashHex string, deleteFiles bool) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	h, herr := parseHashParam(hashHex)
	if herr != nil {
		return model.NewAppError(model.CodeInvalidHash, "invalid hash", herr)
	}
	if err := cli.RemoveTransfer(h, deleteFiles); err != nil {
		return err
	}
	return nil
}

// ListTransferPeers 仅peers。
func (e *Engine) ListTransferPeers(ctx context.Context, hashHex string) ([]model.PeerDTO, error) {
	d, err := e.GetTransfer(ctx, hashHex)
	if err != nil {
		return nil, err
	}
	return d.Peers, nil
}

// ListTransferPieces 仅pieces。
func (e *Engine) ListTransferPieces(ctx context.Context, hashHex string) ([]model.PieceDTO, error) {
	d, err := e.GetTransfer(ctx, hashHex)
	if err != nil {
		return nil, err
	}
	return d.Pieces, nil
}

// StartSearch 发起搜索（单活跃任务由goed2k 保证）。
func (e *Engine) StartSearch(ctx context.Context, p model.SearchParamsDTO) (*model.SearchDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	params := goed2k.SearchParams{
		Query:              strings.TrimSpace(p.Query),
		Scope:              ParseSearchScope(p.Scope),
		MinSize:            p.MinSize,
		MaxSize:            p.MaxSize,
		MinSources:         p.MinSources,
		MinCompleteSources: p.MinCompleteSources,
		FileType:           p.FileType,
		Extension:          p.Extension,
	}
	_, err = cli.StartSearch(params)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			return nil, model.NewAppError(model.CodeSearchAlreadyRunning, "search already running", err)
		}
		return nil, model.NewAppError(model.CodeBadRequest, "start search failed", err)
	}
	snap := mapSearchSnapshot(cli.SearchSnapshot())
	return &snap, nil
}

// CurrentSearch 当前搜索快照。
func (e *Engine) CurrentSearch(ctx context.Context) (*model.SearchDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	snap := mapSearchSnapshot(cli.SearchSnapshot())
	return &snap, nil
}

// StopSearch 停止当前搜索。
func (e *Engine) StopSearch(ctx context.Context) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	return cli.StopSearch()
}

// AddTransferFromSearchResult 从当前搜索结果添加下载。
func (e *Engine) AddTransferFromSearchResult(ctx context.Context, hashHex string, p AddTransferParams) (*model.TransferDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	want, herr := parseHashParam(hashHex)
	if herr != nil {
		return nil, model.NewAppError(model.CodeInvalidHash, "invalid hash", herr)
	}
	snap := cli.SearchSnapshot()
	var linkStr string
	for _, r := range snap.Results {
		if r.Hash.Compare(want) == 0 {
			linkStr = r.ED2KLink()
			break
		}
	}
	if linkStr == "" {
		return nil, model.NewAppError(model.CodeNotFound, "result not in current search", nil)
	}
	np := p
	np.ED2KLink = linkStr
	return e.AddTransferByED2K(ctx, np)
}

// SharedFiles 返回共享库文件列表。
func (e *Engine) SharedFiles(ctx context.Context) ([]model.SharedFileDTO, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	files := cli.SharedFiles()
	out := make([]model.SharedFileDTO, 0, len(files))
	for _, f := range files {
		out = append(out, mapSharedFile(f))
	}
	return out, nil
}

// ListSharedDirs 返回已注册的共享扫描目录。
func (e *Engine) ListSharedDirs(ctx context.Context) ([]string, error) {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	return cli.ListSharedDirs(), nil
}

// AddSharedDir 注册共享扫描目录。
func (e *Engine) AddSharedDir(ctx context.Context, path string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return model.NewAppError(model.CodeBadRequest, "path required", nil)
	}
	return cli.AddSharedDir(path)
}

// RemoveSharedDir 移除共享扫描目录。
func (e *Engine) RemoveSharedDir(ctx context.Context, path string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return model.NewAppError(model.CodeBadRequest, "path required", nil)
	}
	return cli.RemoveSharedDir(path)
}

// RescanSharedDirs 重新扫描已注册目录并导入文件。
func (e *Engine) RescanSharedDirs(ctx context.Context) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	return cli.RescanSharedDirs()
}

// ImportSharedFile 计算 ed2k 哈希并将文件加入共享库。
func (e *Engine) ImportSharedFile(ctx context.Context, path string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(path) == "" {
		return model.NewAppError(model.CodeBadRequest, "path required", nil)
	}
	return cli.ImportSharedFile(path)
}

// RemoveSharedFile 从共享库按 hash 移除。
func (e *Engine) RemoveSharedFile(ctx context.Context, hashHex string) error {
	_ = ctx
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	h, herr := parseHashParam(hashHex)
	if herr != nil {
		return model.NewAppError(model.CodeInvalidHash, "invalid hash", herr)
	}
	if !cli.RemoveSharedFile(h) {
		return model.NewAppError(model.CodeSharedFileNotFound, "shared file not found", nil)
	}
	return nil
}

// RunContext 返回与引擎运行期绑定的context（Stop 时取消）。
func (e *Engine) RunContext() context.Context {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.runCtx == nil {
		return context.Background()
	}
	return e.runCtx
}

// WatchClientStatus 将client.status 事件写入 sink，直到ctx 或引擎停止。
func (e *Engine) WatchClientStatus(ctx context.Context, sink chan<- model.EventEnvelope) {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		cli, run := e.snapshotClient()
		if !run || cli == nil {
			dto, _ := e.ClientStatus(ctx)
			e.pushEnvelope(sink, "client.status", map[string]any{"status": dto})
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				continue
			}
		}
		statusCh, unsub := cli.SubscribeStatusBuffered(32)
	inner:
		for {
			select {
			case <-ctx.Done():
				unsub()
				return
			case <-e.RunContext().Done():
				unsub()
				break inner
			case ev, ok := <-statusCh:
				if !ok {
					unsub()
					break inner
				}
				dto := mapClientStatus(true, ev.Status, ev.DHT)
				e.pushEnvelope(sink, "client.status", map[string]any{"status": dto})
			}
		}
	}
}

func (e *Engine) snapshotClient() (*goed2k.Client, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.client, e.running
}

func (e *Engine) pushEnvelope(sink chan<- model.EventEnvelope, typ string, data map[string]any) {
	env := model.EventEnvelope{Type: typ, At: time.Now().UTC(), Data: data}
	select {
	case sink <- env:
	default:
	}
}

// WatchTransferProgress 将transfer.progress 写入 sink。
func (e *Engine) WatchTransferProgress(ctx context.Context, sink chan<- model.EventEnvelope) {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		cli, run := func() (*goed2k.Client, bool) {
			e.mu.RLock()
			defer e.mu.RUnlock()
			return e.client, e.running
		}()
		if !run || cli == nil {
			e.pushEnvelope(sink, "transfer.progress", map[string]any{
				"progress": model.TransferProgressEventDTO{},
			})
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				continue
			}
		}
		progCh, unsub := cli.SubscribeTransferProgressBuffered(32)
	inner:
		for {
			select {
			case <-ctx.Done():
				unsub()
				return
			case <-e.RunContext().Done():
				unsub()
				break inner
			case ev, ok := <-progCh:
				if !ok {
					unsub()
					break inner
				}
				p := mapProgressEvent(ev)
				e.pushEnvelope(sink, "transfer.progress", map[string]any{"progress": p})
			}
		}
	}
}

// ApplyConfigPatch 热更新与引擎相关的配置项（不自动重启 client）。
func (e *Engine) ApplyConfigPatch(cfg *config.Config) {
	_ = cfg
	// 第一版：由上层替换AppConfigStore；引擎已在下次Start 使用新配置。
}
