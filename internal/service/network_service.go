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

// DHTv6 状态。
func (s *NetworkService) DHTv6(ctx context.Context) (*model.KADV6StatusDTO, error) {
	return s.eng.DHTv6Status(ctx)
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

// EnableDHTv6 启用 IPv6 KAD/DHT。
func (s *NetworkService) EnableDHTv6(ctx context.Context) error {
	return s.eng.EnableDHTv6(ctx)
}

// LoadNodes 加载 nodes.dat。
func (s *NetworkService) LoadNodes(ctx context.Context, sources []string) error {
	return s.eng.LoadDHTNodesSources(ctx, sources)
}

// BootstrapNodes 添加引导节点。
func (s *NetworkService) BootstrapNodes(ctx context.Context, nodes []string) error {
	return s.eng.AddDHTBootstrapNodes(ctx, nodes)
}

// LoadNodes6 加载 nodes6.dat。
func (s *NetworkService) LoadNodes6(ctx context.Context, sources []string) error {
	return s.eng.LoadDHTv6NodesSources(ctx, sources)
}

// BootstrapNodes6 添加 IPv6 引导节点。
func (s *NetworkService) BootstrapNodes6(ctx context.Context, nodes []string) error {
	return s.eng.AddDHTv6BootstrapNodes(ctx, nodes)
}

// LoadIPFilter 加载 ipfilter.dat。
func (s *NetworkService) LoadIPFilter(ctx context.Context, path string) error {
	return s.eng.LoadIPFilter(ctx, path)
}
