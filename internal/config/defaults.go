package config

const defaultUDPPortV6 = 4672

// ApplyDefaults 为旧版/缺字段配置补全与 goed2k.Settings 一致的默认值。
func ApplyDefaults(c *Config) {
	if c == nil {
		return
	}
	if c.Engine.UDPPortV6 <= 0 {
		c.Engine.UDPPortV6 = defaultUDPPortV6
	}
}
