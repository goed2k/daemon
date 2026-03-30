package service

import (
	"context"
	"path/filepath"

	"github.com/chenjia404/goed2kd/internal/config"
	"github.com/chenjia404/goed2kd/internal/engine"
	"github.com/chenjia404/goed2kd/internal/model"
	"github.com/chenjia404/goed2kd/internal/store"
)

// SystemService 系统与配置相关编排。
type SystemService struct {
	eng *engine.Engine
	st  *store.AppConfigStore
}

// NewSystemService 构造。
func NewSystemService(eng *engine.Engine, st *store.AppConfigStore) *SystemService {
	return &SystemService{eng: eng, st: st}
}

// Info 系统信息。
func (s *SystemService) Info(ctx context.Context) (*model.SystemInfo, error) {
	return s.eng.Info(ctx)
}

// Health 健康检查。
func (s *SystemService) Health(ctx context.Context) (*model.HealthStatus, error) {
	return s.eng.Health(ctx)
}

// StartEngine 启动引擎。
func (s *SystemService) StartEngine(ctx context.Context) error {
	return s.eng.Start(ctx)
}

// StopEngine 停止引擎。
func (s *SystemService) StopEngine(ctx context.Context) error {
	return s.eng.Stop(ctx)
}

// SaveState 保存状态。
func (s *SystemService) SaveState(ctx context.Context) error {
	return s.eng.SaveState(ctx)
}

// LoadState 加载状态。
func (s *SystemService) LoadState(ctx context.Context) error {
	return s.eng.LoadState(ctx)
}

// GetConfig 返回当前配置（本地守护进程场景下完整返回）。
func (s *SystemService) GetConfig(ctx context.Context) *config.Config {
	_ = ctx
	return s.st.Get()
}

// ConfigSummary 对外摘要。
func (s *SystemService) ConfigSummary(ctx context.Context) *model.ConfigSummary {
	_ = ctx
	c := s.st.Get()
	return &model.ConfigSummary{
		RPCListen:              c.RPC.Listen,
		EngineListenPort:       c.Engine.ListenPort,
		EngineUDPPort:          c.Engine.UDPPort,
		EnableDHT:              c.Engine.EnableDHT,
		DefaultDownloadDir:     c.Engine.DefaultDownloadDir,
		StateEnabled:           c.State.Enabled,
		StatePath:              c.State.Path,
		AutoSaveIntervalSec:    c.State.AutoSaveIntervalSeconds,
		BootstrapServerCount:   len(c.Bootstrap.ServerAddresses),
		BootstrapServerMetURLs: len(c.Bootstrap.ServerMetURLs),
		BootstrapNodesDatURLs:  len(c.Bootstrap.NodesDatURLs),
	}
}

// UpdateConfigPatch 热更新允许字段并写回文件。
type UpdateConfigPatch struct {
	Bootstrap *config.BootstrapConfig `json:"bootstrap"`
	State     *config.StateConfig     `json:"state"`
	Logging   *struct {
		Level string `json:"level"`
	} `json:"logging"`
}

// UpdateConfig 合并配置并持久化。
func (s *SystemService) UpdateConfig(ctx context.Context, path string, patch UpdateConfigPatch) error {
	_ = ctx
	cur := s.st.Get()
	merged := *cur
	if patch.Bootstrap != nil {
		merged.Bootstrap = *patch.Bootstrap
	}
	if patch.State != nil {
		merged.State = *patch.State
	}
	if patch.Logging != nil && patch.Logging.Level != "" {
		merged.Logging.Level = patch.Logging.Level
	}
	if err := config.Validate(&merged); err != nil {
		return model.NewAppError(model.CodeConfigInvalid, err.Error(), err)
	}
	if err := config.SaveToFile(path, &merged); err != nil {
		return model.NewAppError(model.CodeInternalError, "save config failed", err)
	}
	s.st.Replace(&merged)
	return nil
}

// StateDirOK 状态目录是否可写（用于健康检查细化）。
func (s *SystemService) StateDirOK() bool {
	c := s.st.Get()
	if !c.State.Enabled || c.State.Path == "" {
		return true
	}
	dir := filepath.Dir(c.State.Path)
	return dir != ""
}
