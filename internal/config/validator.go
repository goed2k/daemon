package config

import (
	"fmt"
	"net"
	"strings"
)

// Validate 校验配置合法性。
func Validate(c *Config) error {
	if c == nil {
		return fmt.Errorf("配置为空")
	}
	ApplyDefaults(c)
	if strings.TrimSpace(c.RPC.AuthToken) == "" {
		return fmt.Errorf("rpc.auth_token 不能为空")
	}
	if strings.TrimSpace(c.RPC.Listen) == "" {
		return fmt.Errorf("rpc.listen 不能为空")
	}
	if !c.RPC.AllowRemote {
		host, _, err := net.SplitHostPort(c.RPC.Listen)
		if err != nil {
			return fmt.Errorf("rpc.listen 无效: %w", err)
		}
		if host == "0.0.0.0" || host == "::" {
			return fmt.Errorf("allow_remote=false 时不允许监听 0.0.0.0 或::")
		}
	}
	if c.Engine.ListenPort <= 0 || c.Engine.UDPPort <= 0 {
		return fmt.Errorf("engine 端口无效")
	}
	if c.Engine.EnableDHTv6 && c.Engine.UDPPortV6 <= 0 {
		return fmt.Errorf("enable_dht_v6=true 时 engine.udp_port_v6 必须大于 0")
	}
	if strings.TrimSpace(c.Engine.DefaultDownloadDir) == "" {
		return fmt.Errorf("engine.default_download_dir 不能为空")
	}
	if c.State.Enabled && strings.TrimSpace(c.State.Path) == "" {
		return fmt.Errorf("state.enabled=true 时state.path 不能为空")
	}
	return nil
}
