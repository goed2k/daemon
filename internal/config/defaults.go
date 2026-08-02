package config

const (
	defaultUDPPortV6              = 4672
	defaultMaxHttpSources         = 4
	defaultMaxConcurrentHttpBlocks = 2
	defaultHttpRequestTimeoutSec  = 30
)

// ApplyDefaults 为旧版/缺字段配置补全与 goed2k.Settings 一致的默认值。
func ApplyDefaults(c *Config) {
	if c == nil {
		return
	}
	if c.Engine.UDPPortV6 <= 0 {
		c.Engine.UDPPortV6 = defaultUDPPortV6
	}
	if c.Engine.MaxHttpSources <= 0 {
		c.Engine.MaxHttpSources = defaultMaxHttpSources
	}
	if c.Engine.MaxConcurrentHttpBlocks <= 0 {
		c.Engine.MaxConcurrentHttpBlocks = defaultMaxConcurrentHttpBlocks
	}
	if c.Engine.HttpRequestTimeoutSec <= 0 {
		c.Engine.HttpRequestTimeoutSec = defaultHttpRequestTimeoutSec
	}
}
