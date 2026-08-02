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
	migrateLegacyConfig(c)
}

// migrateLegacyConfig 将 v0.1.3 之前的配置文件迁移到当前语义。
// 旧版 JSON 不含 enable_web_download / partial_kad_publish 时，Go 零值为 false，
// 与 goed2k.Settings 默认 true 不一致，需在首次加载时补全。
// 若需显式关闭上述开关，请在配置中设置 config_version 为 CurrentConfigVersion。
func migrateLegacyConfig(c *Config) {
	if c.Version >= currentConfigVersion {
		return
	}
	if !c.Engine.EnableWebDownload {
		c.Engine.EnableWebDownload = true
	}
	if !c.Engine.PartialKadPublish {
		c.Engine.PartialKadPublish = true
	}
	c.Version = currentConfigVersion
}
