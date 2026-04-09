package engine

import (
	"strings"

	"github.com/goed2k/core"
	"github.com/goed2k/core/protocol"

	"github.com/chenjia404/goed2kd/internal/model"
)

func mapServer(s goed2k.ServerSnapshot) model.ServerDTO {
	return model.ServerDTO{
		Identifier:                   s.Identifier,
		Address:                      s.Address,
		Configured:                   s.Configured,
		Connected:                    s.Connected,
		HandshakeCompleted:           s.HandshakeCompleted,
		Primary:                      s.Primary,
		Disconnecting:                s.Disconnecting,
		ClientID:                     s.ClientID,
		IDClass:                      s.IDClass(),
		DownloadRate:                 s.DownloadRate,
		UploadRate:                   s.UploadRate,
		MillisecondsSinceLastReceive: s.MillisecondsSinceLastReceive,
	}
}

func mapDHT(d goed2k.DHTStatus) model.DHTStatusDTO {
	return model.DHTStatusDTO{
		Bootstrapped:      d.Bootstrapped,
		Firewalled:        d.Firewalled,
		LiveNodes:         d.LiveNodes,
		ReplacementNodes:  d.ReplacementNodes,
		RouterNodes:       d.RouterNodes,
		RunningTraversals: d.RunningTraversals,
		KnownNodes:        d.KnownNodes,
		InitialBootstrap:  d.InitialBootstrap,
		ListenPort:        d.ListenPort,
		StoragePoint:      d.StoragePoint,
	}
}

func mapTransfer(s goed2k.TransferSnapshot) model.TransferDTO {
	st := s.Status
	prog := float64(0)
	if st.TotalWanted > 0 {
		prog = float64(st.TotalDone) / float64(st.TotalWanted)
		if prog > 1 {
			prog = 1
		}
	}
	return model.TransferDTO{
		Hash:              s.Hash.String(),
		FileName:          s.FileName,
		FilePath:          s.FilePath,
		Size:              s.Size,
		CreateTime:        s.CreateTime,
		State:             string(st.State),
		Paused:            st.Paused,
		DownloadRate:      st.DownloadRate,
		UploadRate:        st.UploadRate,
		TotalDone:         st.TotalDone,
		TotalReceived:     st.TotalReceived,
		TotalWanted:       st.TotalWanted,
		ETA:               st.ETA,
		NumPeers:          st.NumPeers,
		ActivePeers:       s.ActivePeers,
		DownloadingPieces: st.DownloadingPieces,
		Progress:          prog,
		ED2KLink:          s.ED2KLink(),
	}
}

func mapPeer(p goed2k.PeerInfo) model.PeerDTO {
	return model.PeerDTO{
		Endpoint:             p.Endpoint.String(),
		DownloadSpeed:        p.DownloadSpeed,
		PayloadDownloadSpeed: p.PayloadDownloadSpeed,
		UploadSpeed:          p.UploadSpeed,
		Source:               p.SourceString(),
		ModName:              p.ModName,
		FailCount:            p.FailCount,
	}
}

func mapPiece(p goed2k.PieceSnapshot) model.PieceDTO {
	return model.PieceDTO{
		Index:         p.Index,
		State:         string(p.State),
		TotalBytes:    p.TotalBytes,
		DoneBytes:     p.DoneBytes,
		ReceivedBytes: p.ReceivedBytes,
		BlocksTotal:   p.BlocksTotal,
		BlocksDone:    p.BlocksDone,
		BlocksWriting: p.BlocksWriting,
		BlocksPending: p.BlocksPending,
	}
}

func mapClientStatus(engineRunning bool, st goed2k.ClientStatus, dht goed2k.DHTStatus) model.ClientStatusDTO {
	servers := make([]model.ServerDTO, 0, len(st.Servers))
	for _, s := range st.Servers {
		servers = append(servers, mapServer(s))
	}
	transfers := make([]model.TransferDTO, 0, len(st.Transfers))
	for _, t := range st.Transfers {
		transfers = append(transfers, mapTransfer(t))
	}
	return model.ClientStatusDTO{
		EngineRunning: engineRunning,
		Servers:       servers,
		Transfers:     transfers,
		DHT:           mapDHT(dht),
		Totals: map[string]any{
			"total_done":     st.TotalDone,
			"total_received": st.TotalReceived,
			"total_wanted":   st.TotalWanted,
			"upload":         st.Upload,
			"download_rate":  st.DownloadRate,
			"upload_rate":    st.UploadRate,
		},
	}
}

func mapSearchSnapshot(snap goed2k.SearchSnapshot) model.SearchDTO {
	p := snap.Params
	dto := model.SearchDTO{
		ID:         snap.ID,
		State:      string(snap.State),
		UpdatedAt:  snap.UpdatedAt,
		StartedAt:  snap.StartedAt,
		ServerBusy: snap.ServerBusy,
		DHTBusy:    snap.DHTBusy,
		KadKeyword: snap.KadKeyword,
		Error:      snap.Error,
		Params: model.SearchParamsDTO{
			Query:              p.Query,
			Scope:              searchScopeToString(p.Scope),
			MinSize:            p.MinSize,
			MaxSize:            p.MaxSize,
			MinSources:         p.MinSources,
			MinCompleteSources: p.MinCompleteSources,
			FileType:           p.FileType,
			Extension:          p.Extension,
		},
	}
	dto.Results = make([]model.SearchResultDTO, 0, len(snap.Results))
	for _, r := range snap.Results {
		dto.Results = append(dto.Results, mapSearchResult(r))
	}
	if snap.ID == 0 && snap.State == "" {
		dto.State = "IDLE"
	}
	return dto
}

func mapSearchResult(r goed2k.SearchResult) model.SearchResultDTO {
	src := ""
	switch {
	case r.Source&(goed2k.SearchResultServer|goed2k.SearchResultKAD) == (goed2k.SearchResultServer | goed2k.SearchResultKAD):
		src = "server|kad"
	case r.Source&goed2k.SearchResultServer != 0:
		src = "server"
	case r.Source&goed2k.SearchResultKAD != 0:
		src = "kad"
	default:
		src = "unknown"
	}
	return model.SearchResultDTO{
		Hash:            r.Hash.String(),
		FileName:        r.FileName,
		FileSize:        r.FileSize,
		Sources:         r.Sources,
		CompleteSources: r.CompleteSources,
		MediaBitrate:    r.MediaBitrate,
		MediaLength:     r.MediaLength,
		MediaCodec:      r.MediaCodec,
		Extension:       r.Extension,
		FileType:        r.FileType,
		Source:          src,
		ED2KLink:        r.ED2KLink(),
	}
}

func searchScopeToString(s goed2k.SearchScope) string {
	if s == goed2k.SearchScopeAll {
		return "all"
	}
	if s == goed2k.SearchScopeServer {
		return "server"
	}
	if s == goed2k.SearchScopeDHT {
		return "dht"
	}
	return "all"
}

// ParseSearchScope 将API 字符串解析为 goed2k 搜索范围。
func ParseSearchScope(s string) goed2k.SearchScope {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "server":
		return goed2k.SearchScopeServer
	case "dht", "kad":
		return goed2k.SearchScopeDHT
	case "all", "":
		return goed2k.SearchScopeAll
	default:
		return goed2k.SearchScopeAll
	}
}

func parseHashParam(hexHash string) (protocol.Hash, error) {
	h, err := protocol.HashFromString(strings.TrimSpace(hexHash))
	if err != nil {
		return protocol.Invalid, err
	}
	return h, nil
}

func sharedOriginString(o goed2k.SharedOrigin) string {
	switch o {
	case goed2k.SharedOriginDownloaded:
		return "downloaded"
	case goed2k.SharedOriginImported:
		return "imported"
	default:
		return "unknown"
	}
}

func mapSharedFile(f *goed2k.SharedFile) model.SharedFileDTO {
	if f == nil {
		return model.SharedFileDTO{}
	}
	return model.SharedFileDTO{
		Hash:       f.Hash.String(),
		FileSize:   f.FileSize,
		Path:       f.Path,
		Name:       f.FileLabel(),
		Origin:     sharedOriginString(f.Origin),
		Completed:  f.Completed,
		CanUpload:  f.CanUpload(),
		LastHashAt: f.LastHashAt,
	}
}

func mapProgressEvent(ev goed2k.TransferProgressEvent) model.TransferProgressEventDTO {
	out := make([]model.TransferDTO, 0, len(ev.Transfers))
	for _, p := range ev.Transfers {
		st := goed2k.TransferStatus{
			Paused:            p.Paused,
			NumPeers:          p.NumPeers,
			DownloadingPieces: p.DownloadingPieces,
			TotalDone:         p.TotalDone,
			TotalReceived:     p.TotalReceived,
			TotalWanted:       p.TotalWanted,
			State:             p.State,
		}
		prog := float64(0)
		if st.TotalWanted > 0 {
			prog = float64(st.TotalDone) / float64(st.TotalWanted)
			if prog > 1 {
				prog = 1
			}
		}
		out = append(out, model.TransferDTO{
			Hash:              p.Hash.String(),
			FileName:          p.FileName,
			FilePath:          p.FilePath,
			State:             string(p.State),
			Paused:            p.Paused,
			TotalDone:         p.TotalDone,
			TotalReceived:     p.TotalReceived,
			TotalWanted:       p.TotalWanted,
			NumPeers:          p.NumPeers,
			ActivePeers:       p.ActivePeers,
			DownloadingPieces: p.DownloadingPieces,
			Progress:          prog,
		})
	}
	return model.TransferProgressEventDTO{Transfers: out}
}
