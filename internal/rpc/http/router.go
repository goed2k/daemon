package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/goed2k/daemon/internal/engine"
	eventws "github.com/goed2k/daemon/internal/rpc/ws"
	"github.com/goed2k/daemon/internal/service"
	"github.com/goed2k/daemon/internal/store"
	"github.com/go-chi/chi/v5"
)

// Server HTTP 依赖集合。
type Server struct {
	Log         *slog.Logger
	ConfigPath  string
	ConfigStore *store.AppConfigStore
	Engine      *engine.Engine
	Hub         *eventws.Hub

	Sys      *service.SystemService
	Net      *service.NetworkService
	Transfer *service.TransferService
	Search   *service.SearchService
	Shared   *service.SharedService

	AuthToken          string
	ReadTimeoutSeconds int
}

// NewRouter 构建 chi 路由树。
func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	readTO := time.Duration(s.ReadTimeoutSeconds) * time.Second
	if readTO <= 0 {
		readTO = 15 * time.Second
	}

	r.Use(RequestID())
	r.Use(WithRequestLog(s.Log))
	r.Use(RecoverJSON(s.Log))
	r.Use(RealIP())
	r.Use(CORS())
	r.Use(accessLog(s.Log))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(Timeout(readTO + 30*time.Second))
		r.Use(AuthToken(s.AuthToken))

		r.Get("/system/health", s.handleHealth)
		r.Get("/system/info", s.handleInfo)
		r.Post("/system/start", s.handleStart)
		r.Post("/system/stop", s.handleStop)
		r.Post("/system/save-state", s.handleSaveState)
		r.Post("/system/load-state", s.handleLoadState)
		r.Get("/system/config", s.handleGetConfig)
		r.Put("/system/config", s.handlePutConfig)

		r.Get("/network/servers", s.handleNetworkServers)
		r.Post("/network/servers/connect", s.handleNetworkConnect)
		r.Post("/network/servers/connect-batch", s.handleNetworkConnectBatch)
		r.Post("/network/servers/load-met", s.handleNetworkLoadMet)
		r.Get("/network/dht", s.handleNetworkDHT)
		r.Post("/network/dht/enable", s.handleNetworkDHTEnable)
		r.Post("/network/dht/load-nodes", s.handleNetworkDHTLoadNodes)
		r.Post("/network/dht/bootstrap-nodes", s.handleNetworkDHTBootstrap)

		r.Get("/transfers", s.handleTransfersList)
		r.Post("/transfers", s.handleTransfersAdd)
		r.Get("/transfers/{hash}", s.handleTransfersDetail)
		r.Post("/transfers/{hash}/pause", s.handleTransfersPause)
		r.Post("/transfers/{hash}/resume", s.handleTransfersResume)
		r.Delete("/transfers/{hash}", s.handleTransfersDelete)
		r.Get("/transfers/{hash}/peers", s.handleTransfersPeers)
		r.Get("/transfers/{hash}/pieces", s.handleTransfersPieces)

		r.Post("/searches", s.handleSearchesCreate)
		r.Get("/searches/current", s.handleSearchesCurrent)
		r.Post("/searches/current/stop", s.handleSearchesStop)
		r.Post("/searches/current/results/{hash}/download", s.handleSearchesResultDownload)

		r.Get("/shared/files", s.handleSharedFilesList)
		r.Get("/shared/dirs", s.handleSharedDirsList)
		r.Post("/shared/dirs", s.handleSharedDirsAdd)
		r.Post("/shared/dirs/remove", s.handleSharedDirsRemove)
		r.Post("/shared/dirs/rescan", s.handleSharedDirsRescan)
		r.Post("/shared/import", s.handleSharedImport)
		r.Delete("/shared/files/{hash}", s.handleSharedFileRemove)

		r.Get("/events/ws", s.Hub.ServeWS(s.Log, s.AuthToken))
	})

	return r
}

func accessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			if log != nil {
				log.Info("http", "method", r.Method, "path", r.URL.Path, "dur_ms", time.Since(start).Milliseconds())
			}
		})
	}
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
