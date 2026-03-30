package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/chenjia404/goed2kd/internal/config"
	"github.com/chenjia404/goed2kd/internal/engine"
	"github.com/chenjia404/goed2kd/internal/model"
	httpapi "github.com/chenjia404/goed2kd/internal/rpc/http"
	eventws "github.com/chenjia404/goed2kd/internal/rpc/ws"
	"github.com/chenjia404/goed2kd/internal/service"
	"github.com/chenjia404/goed2kd/internal/store"
)

// Daemon 聚合配置、引擎、HTTP 与事件桥接。
type Daemon struct {
	Log        *slog.Logger
	ConfigPath string
	Cfg        *config.Config
	Store      *store.AppConfigStore
	Engine     *engine.Engine
	Hub        *eventws.Hub

	Sys      *service.SystemService
	Net      *service.NetworkService
	Transfer *service.TransferService
	Search   *service.SearchService

	httpServer *http.Server
	bridgeStop context.CancelFunc
	wg         sync.WaitGroup
}

// NewDaemon 从已加载配置构造（不负责读文件）。
func NewDaemon(log *slog.Logger, configPath string, cfg *config.Config) *Daemon {
	st := store.NewAppConfigStore(cfg)
	eng := engine.NewEngine(log, st)
	return &Daemon{
		Log:        log,
		ConfigPath: configPath,
		Cfg:        cfg,
		Store:      st,
		Engine:     eng,
		Hub:        eventws.NewHub(),
		Sys:        service.NewSystemService(eng, st),
		Net:        service.NewNetworkService(eng),
		Transfer:   service.NewTransferService(eng),
		Search:     service.NewSearchService(eng),
	}
}

// Run 阻塞运行直至收到 SIGINT/SIGTERM。
func (d *Daemon) Run() error {
	ingress := make(chan model.EventEnvelope, 256)
	bridgeCtx, cancelBridge := context.WithCancel(context.Background())
	d.bridgeStop = cancelBridge

	d.wg.Add(3)
	go func() {
		defer d.wg.Done()
		d.Engine.WatchClientStatus(bridgeCtx, ingress)
	}()
	go func() {
		defer d.wg.Done()
		d.Engine.WatchTransferProgress(bridgeCtx, ingress)
	}()
	go func() {
		defer d.wg.Done()
		d.Hub.RunIngress(bridgeCtx, ingress)
	}()

	if err := d.Engine.Start(context.Background()); err != nil {
		if d.Log != nil {
			d.Log.Warn("引擎启动失败，可稍后通过 API 重试", "err", err)
		}
	}

	srv := &httpapi.Server{
		Log:                d.Log,
		ConfigPath:         d.ConfigPath,
		ConfigStore:        d.Store,
		Engine:             d.Engine,
		Hub:                d.Hub,
		Sys:                d.Sys,
		Net:                d.Net,
		Transfer:           d.Transfer,
		Search:             d.Search,
		AuthToken:          d.Cfg.RPC.AuthToken,
		ReadTimeoutSeconds: d.Cfg.RPC.ReadTimeoutSeconds,
	}
	h := httpapi.NewRouter(srv)
	wto := d.Cfg.RPC.WriteTimeoutSeconds
	if wto <= 0 {
		wto = 15
	}
	rto := d.Cfg.RPC.ReadTimeoutSeconds
	if rto <= 0 {
		rto = 15
	}
	d.httpServer = &http.Server{
		Addr:         d.Cfg.RPC.Listen,
		Handler:      h,
		ReadTimeout:  time.Duration(rto) * time.Second,
		WriteTimeout: time.Duration(wto) * time.Second,
	}

	go func() {
		if d.Log != nil {
			d.Log.Info("HTTP 监听", "addr", d.Cfg.RPC.Listen)
		}
		if err := d.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if d.Log != nil {
				d.Log.Error("HTTP 服务异常退出", "err", err)
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	return d.shutdown()
}

func (d *Daemon) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if d.httpServer != nil {
		_ = d.httpServer.Shutdown(ctx)
	}
	if d.bridgeStop != nil {
		d.bridgeStop()
	}
	d.wg.Wait()

	if d.Cfg.State.Enabled && d.Cfg.State.SaveOnExit && d.Engine.IsRunning() {
		if err := d.Engine.SaveState(ctx); err != nil && d.Log != nil {
			d.Log.Warn("退出时保存状态失败", "err", err)
		}
	}
	_ = d.Engine.Stop(ctx)
	if d.Log != nil {
		d.Log.Info("守护进程已退出")
	}
	return nil
}
