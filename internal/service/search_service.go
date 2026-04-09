package service

import (
	"context"

	"github.com/goed2k/daemon/internal/engine"
	"github.com/goed2k/daemon/internal/model"
)

// SearchService 搜索编排（单活跃搜索由底层保证）。
type SearchService struct {
	eng *engine.Engine
}

// NewSearchService 构造。
func NewSearchService(eng *engine.Engine) *SearchService {
	return &SearchService{eng: eng}
}

// Start 发起搜索。
func (s *SearchService) Start(ctx context.Context, p model.SearchParamsDTO) (*model.SearchDTO, error) {
	return s.eng.StartSearch(ctx, p)
}

// Current 当前快照。
func (s *SearchService) Current(ctx context.Context) (*model.SearchDTO, error) {
	return s.eng.CurrentSearch(ctx)
}

// Stop 停止。
func (s *SearchService) Stop(ctx context.Context) error {
	return s.eng.StopSearch(ctx)
}

// ResultDownload 从当前结果添加下载。
func (s *SearchService) ResultDownload(ctx context.Context, hash, targetDir, targetName string, paused bool) (*model.TransferDTO, error) {
	return s.eng.AddTransferFromSearchResult(ctx, hash, engine.AddTransferParams{
		TargetDir: targetDir, TargetName: targetName, Paused: paused,
	})
}
