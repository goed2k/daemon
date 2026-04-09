package service

import (
	"context"

	"github.com/chenjia404/goed2kd/internal/engine"
	"github.com/chenjia404/goed2kd/internal/model"
)

// SharedService 共享库编排（封装 goed2k/core Client 共享 API）。
type SharedService struct {
	eng *engine.Engine
}

// NewSharedService 构造。
func NewSharedService(eng *engine.Engine) *SharedService {
	return &SharedService{eng: eng}
}

// ListFiles 共享文件列表。
func (s *SharedService) ListFiles(ctx context.Context) ([]model.SharedFileDTO, error) {
	return s.eng.SharedFiles(ctx)
}

// ListDirs 已注册扫描目录。
func (s *SharedService) ListDirs(ctx context.Context) ([]string, error) {
	return s.eng.ListSharedDirs(ctx)
}

// AddDir 注册目录。
func (s *SharedService) AddDir(ctx context.Context, path string) error {
	return s.eng.AddSharedDir(ctx, path)
}

// RemoveDir 移除目录。
func (s *SharedService) RemoveDir(ctx context.Context, path string) error {
	return s.eng.RemoveSharedDir(ctx, path)
}

// RescanDirs 扫描目录并导入。
func (s *SharedService) RescanDirs(ctx context.Context) error {
	return s.eng.RescanSharedDirs(ctx)
}

// ImportFile 导入单个文件到共享库。
func (s *SharedService) ImportFile(ctx context.Context, path string) error {
	return s.eng.ImportSharedFile(ctx, path)
}

// RemoveFile 按 hash 移除。
func (s *SharedService) RemoveFile(ctx context.Context, hashHex string) error {
	return s.eng.RemoveSharedFile(ctx, hashHex)
}
