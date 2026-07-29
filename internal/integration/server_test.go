package integration

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/goed2k/daemon/internal/config"
	"github.com/goed2k/daemon/internal/engine"
	httpapi "github.com/goed2k/daemon/internal/rpc/http"
	eventws "github.com/goed2k/daemon/internal/rpc/ws"
	"github.com/goed2k/daemon/internal/service"
	"github.com/goed2k/daemon/internal/store"
)

const testAuthToken = "ci-test-token"

type testHTTPServer struct {
	*httptest.Server
	Engine *engine.Engine
}

// newTestHTTPServer 构造带完整路由的 httptest 服务。
// startEngine=true 时会尝试启动 goed2k 引擎（使用高端口，DHT/UPnP 关闭）。
func newTestHTTPServer(t *testing.T, cfg *config.Config, startEngine bool) *testHTTPServer {
	t.Helper()
	if cfg == nil {
		cfg = testConfig(t)
	}
	st := store.NewAppConfigStore(cfg)
	eng := engine.NewEngine(slog.New(slog.NewTextHandler(io.Discard, nil)), st)
	if startEngine {
		if err := eng.Start(context.Background()); err != nil {
			t.Skipf("engine start unavailable in this environment: %v", err)
		}
		t.Cleanup(func() { _ = eng.Stop(context.Background()) })
	}
	srv := &httpapi.Server{
		Log:                slog.New(slog.NewTextHandler(io.Discard, nil)),
		ConfigPath:         filepath.Join(t.TempDir(), "config.json"),
		ConfigStore:        st,
		Engine:             eng,
		Hub:                eventws.NewHub(),
		Sys:                service.NewSystemService(eng, st),
		Net:                service.NewNetworkService(eng),
		Transfer:           service.NewTransferService(eng),
		Search:             service.NewSearchService(eng),
		Shared:             service.NewSharedService(eng),
		AuthToken:          cfg.RPC.AuthToken,
		ReadTimeoutSeconds: cfg.RPC.ReadTimeoutSeconds,
	}
	return &testHTTPServer{
		Server: httptest.NewServer(httpapi.NewRouter(srv)),
		Engine: eng,
	}
}

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.Default()
	cfg.RPC.AuthToken = testAuthToken
	cfg.RPC.Listen = "127.0.0.1:0"
	cfg.State.Enabled = false
	cfg.Engine.EnableDHT = false
	cfg.Engine.EnableDHTv6 = false
	cfg.Engine.EnableUPnP = false
	cfg.Engine.ListenPort = 36661
	cfg.Engine.UDPPort = 36662
	cfg.Engine.UDPPortV6 = 36672
	cfg.Bootstrap = config.BootstrapConfig{}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("test config invalid: %v", err)
	}
	return cfg
}

func authRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	return req
}
