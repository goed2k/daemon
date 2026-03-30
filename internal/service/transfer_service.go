package service

import (
	"context"
	"strings"

	"github.com/chenjia404/goed2kd/internal/engine"
	"github.com/chenjia404/goed2kd/internal/model"
)

// TransferService 下载任务编排。
type TransferService struct {
	eng *engine.Engine
}

// NewTransferService 构造。
func NewTransferService(eng *engine.Engine) *TransferService {
	return &TransferService{eng: eng}
}

// ListQuery 列表查询参数（第一版部分生效）。
type ListQuery struct {
	State  string
	Paused *bool
	Limit  int
	Offset int
	Sort   string
}

// List 任务列表。
func (s *TransferService) List(ctx context.Context, q ListQuery) ([]model.TransferDTO, error) {
	all, err := s.eng.ListTransfers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.TransferDTO, 0, len(all))
	for _, t := range all {
		if q.State != "" && !strings.EqualFold(t.State, q.State) {
			continue
		}
		if q.Paused != nil && t.Paused != *q.Paused {
			continue
		}
		out = append(out, t)
	}
	if q.Offset > 0 && q.Offset < len(out) {
		out = out[q.Offset:]
	} else if q.Offset >= len(out) {
		out = nil
	}
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	_ = q.Sort
	return out, nil
}

// Add 添加任务。
func (s *TransferService) Add(ctx context.Context, ed2k, targetDir, targetName string, paused bool) (*model.TransferDTO, error) {
	return s.eng.AddTransferByED2K(ctx, engine.AddTransferParams{
		ED2KLink: ed2k, TargetDir: targetDir, TargetName: targetName, Paused: paused,
	})
}

// Detail 详情。
func (s *TransferService) Detail(ctx context.Context, hash string) (*model.TransferDetailDTO, error) {
	return s.eng.GetTransfer(ctx, hash)
}

// Pause 暂停。
func (s *TransferService) Pause(ctx context.Context, hash string) error {
	return s.eng.PauseTransfer(ctx, hash)
}

// Resume 恢复。
func (s *TransferService) Resume(ctx context.Context, hash string) error {
	return s.eng.ResumeTransfer(ctx, hash)
}

// Delete 删除。
func (s *TransferService) Delete(ctx context.Context, hash string, deleteFiles bool) error {
	return s.eng.DeleteTransfer(ctx, hash, deleteFiles)
}

// Peers 对端列表。
func (s *TransferService) Peers(ctx context.Context, hash string) ([]model.PeerDTO, error) {
	return s.eng.ListTransferPeers(ctx, hash)
}

// Pieces 分片列表。
func (s *TransferService) Pieces(ctx context.Context, hash string) ([]model.PieceDTO, error) {
	return s.eng.ListTransferPieces(ctx, hash)
}
