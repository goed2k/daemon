package store

import (
	"sync"

	"github.com/goed2k/daemon/internal/config"
)

// AppConfigStore 持有当前守护进程配置指针（由上层在更新后替换）。
type AppConfigStore struct {
	mu  sync.RWMutex
	cfg *config.Config
}

// NewAppConfigStore 构造配置存储。
func NewAppConfigStore(c *config.Config) *AppConfigStore {
	return &AppConfigStore{cfg: c}
}

// Get 返回当前配置副本（避免调用方意外修改共享结构时可再深拷贝；此处返回指针由 daemon 保证串行更新）。
func (s *AppConfigStore) Get() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Replace 原子替换配置指针。
func (s *AppConfigStore) Replace(c *config.Config) {
	s.mu.Lock()
	s.cfg = c
	s.mu.Unlock()
}
