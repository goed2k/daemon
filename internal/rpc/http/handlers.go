package httpapi

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/chenjia404/goed2kd/internal/model"
	"github.com/chenjia404/goed2kd/internal/service"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	h, err := s.Sys.Health(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, h)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	info, err := s.Sys.Info(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, info)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Sys.StartEngine(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "system.start")
	WriteSuccess(w, map[string]any{"started": true})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Sys.StopEngine(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "system.stop")
	WriteSuccess(w, map[string]any{"stopped": true})
}

func (s *Server) handleSaveState(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Sys.SaveState(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "system.save_state")
	WriteSuccess(w, map[string]any{"saved": true})
}

func (s *Server) handleLoadState(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Sys.LoadState(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "system.load_state")
	WriteSuccess(w, map[string]any{"loaded": true})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, s.Sys.GetConfig(r.Context()))
}

func (s *Server) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body service.UpdateConfigPatch
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Sys.UpdateConfig(r.Context(), s.ConfigPath, body); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "system.update_config")
	WriteSuccess(w, s.Sys.GetConfig(r.Context()))
}

func (s *Server) handleNetworkServers(w http.ResponseWriter, r *http.Request) {
	list, err := s.Net.Servers(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, list)
}

type connectOne struct {
	Address string `json:"address"`
}

func (s *Server) handleNetworkConnect(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body connectOne
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Net.Connect(r.Context(), body.Address); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "network.connect", "address", body.Address)
	WriteSuccess(w, map[string]any{"ok": true})
}

type connectBatch struct {
	Addresses []string `json:"addresses"`
}

func (s *Server) handleNetworkConnectBatch(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body connectBatch
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Net.ConnectBatch(r.Context(), body.Addresses); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "network.connect_batch", "count", len(body.Addresses))
	WriteSuccess(w, map[string]any{"ok": true})
}

type sourcesBody struct {
	Sources []string `json:"sources"`
}

func (s *Server) handleNetworkLoadMet(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body sourcesBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Net.LoadServerMet(r.Context(), body.Sources); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "network.load_server_met", "count", len(body.Sources))
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleNetworkDHT(w http.ResponseWriter, r *http.Request) {
	st, err := s.Net.DHT(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, st)
}

func (s *Server) handleNetworkDHTEnable(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Net.EnableDHT(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "network.dht_enable")
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleNetworkDHTLoadNodes(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body sourcesBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Net.LoadNodes(r.Context(), body.Sources); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "network.load_nodes_dat", "count", len(body.Sources))
	WriteSuccess(w, map[string]any{"ok": true})
}

type nodesBody struct {
	Nodes []string `json:"nodes"`
}

func (s *Server) handleNetworkDHTBootstrap(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body nodesBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Net.BootstrapNodes(r.Context(), body.Nodes); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "network.dht_bootstrap", "count", len(body.Nodes))
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleTransfersList(w http.ResponseWriter, r *http.Request) {
	q := service.ListQuery{State: r.URL.Query().Get("state"), Sort: r.URL.Query().Get("sort")}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Offset = n
		}
	}
	if v := r.URL.Query().Get("paused"); v != "" {
		b := v == "true" || v == "1"
		q.Paused = &b
	}
	list, err := s.Transfer.List(r.Context(), q)
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, list)
}

type addTransferBody struct {
	ED2KLink   string `json:"ed2k_link"`
	TargetDir  string `json:"target_dir"`
	TargetName string `json:"target_name"`
	Paused     bool   `json:"paused"`
}

func (s *Server) handleTransfersAdd(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body addTransferBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	t, err := s.Transfer.Add(r.Context(), body.ED2KLink, body.TargetDir, body.TargetName, body.Paused)
	if err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "transfer.add", "hash", t.Hash)
	WriteSuccess(w, t)
}

func (s *Server) handleTransfersDetail(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	d, err := s.Transfer.Detail(r.Context(), hash)
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, d)
}

func (s *Server) handleTransfersPause(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	hash := chi.URLParam(r, "hash")
	if err := s.Transfer.Pause(r.Context(), hash); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "transfer.pause", "hash", hash)
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleTransfersResume(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	hash := chi.URLParam(r, "hash")
	if err := s.Transfer.Resume(r.Context(), hash); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "transfer.resume", "hash", hash)
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleTransfersDelete(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	hash := chi.URLParam(r, "hash")
	delFiles := r.URL.Query().Get("delete_files") == "true" || r.URL.Query().Get("delete_files") == "1"
	if err := s.Transfer.Delete(r.Context(), hash, delFiles); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "transfer.delete", "hash", hash, "delete_files", delFiles)
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleTransfersPeers(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	p, err := s.Transfer.Peers(r.Context(), hash)
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, p)
}

func (s *Server) handleTransfersPieces(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	p, err := s.Transfer.Pieces(r.Context(), hash)
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, p)
}

func (s *Server) handleSearchesCreate(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body model.SearchParamsDTO
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	sch, err := s.Search.Start(r.Context(), body)
	if err != nil {
		WriteError(w, log, err)
		return
	}
	WriteSuccess(w, sch)
}

func (s *Server) handleSearchesCurrent(w http.ResponseWriter, r *http.Request) {
	sch, err := s.Search.Current(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, sch)
}

func (s *Server) handleSearchesStop(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Search.Stop(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	WriteSuccess(w, map[string]any{"ok": true})
}

type resultDLBody struct {
	TargetDir  string `json:"target_dir"`
	TargetName string `json:"target_name"`
	Paused     bool   `json:"paused"`
}

func (s *Server) handleSearchesResultDownload(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	hash := chi.URLParam(r, "hash")
	var body resultDLBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	t, err := s.Search.ResultDownload(r.Context(), hash, body.TargetDir, body.TargetName, body.Paused)
	if err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "search.result_download", "hash", hash)
	WriteSuccess(w, t)
}

func (s *Server) handleSharedFilesList(w http.ResponseWriter, r *http.Request) {
	list, err := s.Shared.ListFiles(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, list)
}

func (s *Server) handleSharedDirsList(w http.ResponseWriter, r *http.Request) {
	list, err := s.Shared.ListDirs(r.Context())
	if err != nil {
		WriteError(w, RequestLog(r), err)
		return
	}
	WriteSuccess(w, list)
}

type sharedPathBody struct {
	Path string `json:"path"`
}

func (s *Server) handleSharedDirsAdd(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body sharedPathBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Shared.AddDir(r.Context(), body.Path); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "shared.dir_add", "path", body.Path)
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleSharedDirsRemove(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body sharedPathBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Shared.RemoveDir(r.Context(), body.Path); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "shared.dir_remove", "path", body.Path)
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleSharedDirsRescan(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	if err := s.Shared.RescanDirs(r.Context()); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "shared.dirs_rescan")
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleSharedImport(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	var body sharedPathBody
	if err := decodeJSON(r, &body); err != nil {
		WriteError(w, log, model.NewAppError(model.CodeBadRequest, "invalid json", err))
		return
	}
	if err := s.Shared.ImportFile(r.Context(), body.Path); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "shared.import", "path", body.Path)
	WriteSuccess(w, map[string]any{"ok": true})
}

func (s *Server) handleSharedFileRemove(w http.ResponseWriter, r *http.Request) {
	log := RequestLog(r)
	hash := chi.URLParam(r, "hash")
	if err := s.Shared.RemoveFile(r.Context(), hash); err != nil {
		WriteError(w, log, err)
		return
	}
	log.Info("audit", "action", "shared.file_remove", "hash", hash)
	WriteSuccess(w, map[string]any{"ok": true})
}
