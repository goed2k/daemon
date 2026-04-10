package service

import (
	"context"

	"github.com/goed2k/daemon/internal/engine"
	"github.com/goed2k/daemon/internal/model"
)

// NetworkService 网络相关编排。
type NetworkService struct {
	eng *engine.Engine
}

// NewNetworkService 构造。
func NewNetworkService(eng *engine.Engine) *NetworkService {
	return &NetworkService{eng: eng}
}

// Servers 列表。
func (s *NetworkService) Servers(ctx context.Context) ([]model.ServerDTO, error) {
	return s.eng.Servers(ctx)
}

// DHT 状态。
func (s *NetworkService) DHT(ctx context.Context) (*model.DHTStatusDTO, error) {
	return s.eng.DHTStatus(ctx)
}

// KnownPeers 全局已知客户端（各任务上的对端）。
func (s *NetworkService) KnownPeers(ctx context.Context) ([]model.ClientPeerEntryDTO, error) {
	return s.eng.KnownPeers(ctx)
}

// Connect 单服务器。
func (s *NetworkService) Connect(ctx context.Context, addr string) error {
	return s.eng.ConnectServer(ctx, addr)
}

// ConnectBatch 批量连接。
func (s *NetworkService) ConnectBatch(ctx context.Context, addrs []string) error {
	return s.eng.ConnectServers(ctx, addrs)
}

// LoadServerMet 从多源加载。
func (s *NetworkService) LoadServerMet(ctx context.Context, sources []string) error {
	return s.eng.LoadServerMetSources(ctx, sources)
}

// EnableDHT 启用 DHT。
func (s *NetworkService) EnableDHT(ctx context.Context) error {
	return s.eng.EnableDHT(ctx)
}

// LoadNodes 加载 nodes.dat。
func (s *NetworkService) LoadNodes(ctx context.Context, sources []string) error {
	return s.eng.LoadDHTNodesSources(ctx, sources)
}

// BootstrapNodes 添加引导节点。
func (s *NetworkService) BootstrapNodes(ctx context.Context, nodes []string) error {
	return s.eng.AddDHTBootstrapNodes(ctx, nodes)
}
