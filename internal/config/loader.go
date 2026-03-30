package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// LoadFromFile 从 JSON 文件加载；文件不存在时返回 ErrNotExist。
func LoadFromFile(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveToFile 将配置写入JSON（原子替换）。
func SaveToFile(path string, c *Config) error {
	if c == nil {
		return errors.New("配置为空")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// EnsureFile 若文件不存在则写入默认配置并返回该默认配置。
func EnsureFile(path string) (*Config, error) {
	c, err := LoadFromFile(path)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	def := Default()
	if err := SaveToFile(path, def); err != nil {
		return nil, err
	}
	return def, nil
}
