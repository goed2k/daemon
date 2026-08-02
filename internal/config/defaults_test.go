package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyDefaults_UDPPortV6(t *testing.T) {
	c := &Config{}
	ApplyDefaults(c)
	if c.Engine.UDPPortV6 != defaultUDPPortV6 {
		t.Fatalf("UDPPortV6 = %d, want %d", c.Engine.UDPPortV6, defaultUDPPortV6)
	}

	c.Engine.UDPPortV6 = 5000
	ApplyDefaults(c)
	if c.Engine.UDPPortV6 != 5000 {
		t.Fatalf("UDPPortV6 = %d, want 5000", c.Engine.UDPPortV6)
	}
}

func TestLoadFromFileAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	raw := `{
  "rpc": {"listen": "127.0.0.1:1", "auth_token": "t"},
  "engine": {"listen_port": 4661, "udp_port": 4662, "default_download_dir": "/tmp"},
  "bootstrap": {},
  "state": {},
  "logging": {}
}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Engine.UDPPortV6 != defaultUDPPortV6 {
		t.Fatalf("UDPPortV6 = %d, want %d", c.Engine.UDPPortV6, defaultUDPPortV6)
	}
}

func TestApplyDefaults_LegacyConfigMigration(t *testing.T) {
	c := &Config{
		Engine: EngineConfig{
			ListenPort:         4661,
			UDPPort:            4662,
			DefaultDownloadDir: "/tmp",
		},
	}
	ApplyDefaults(c)
	if !c.Engine.EnableWebDownload {
		t.Fatal("legacy config should migrate enable_web_download to true")
	}
	if !c.Engine.PartialKadPublish {
		t.Fatal("legacy config should migrate partial_kad_publish to true")
	}
	if c.Version != currentConfigVersion {
		t.Fatalf("Version = %d, want %d", c.Version, currentConfigVersion)
	}
	if c.Engine.MaxHttpSources != defaultMaxHttpSources {
		t.Fatalf("MaxHttpSources = %d, want %d", c.Engine.MaxHttpSources, defaultMaxHttpSources)
	}
}

func TestApplyDefaults_PreservesExplicitWebDownloadOff(t *testing.T) {
	c := &Config{
		Version: currentConfigVersion,
		Engine: EngineConfig{
			EnableWebDownload:       false,
			PartialKadPublish:       false,
			MaxHttpSources:          2,
			MaxConcurrentHttpBlocks: 1,
			HttpRequestTimeoutSec:   15,
			DefaultDownloadDir:      "/tmp",
		},
	}
	ApplyDefaults(c)
	if c.Engine.EnableWebDownload {
		t.Fatal("explicit enable_web_download=false should be preserved on v2 config")
	}
	if c.Engine.PartialKadPublish {
		t.Fatal("explicit partial_kad_publish=false should be preserved on v2 config")
	}
}
