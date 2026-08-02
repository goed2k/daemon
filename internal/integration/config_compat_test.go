package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goed2k/daemon/internal/config"
)

// legacyConfigJSON 模拟 core 同步前旧版配置文件（无 IPv6 / 混淆等新字段）。
const legacyConfigJSON = `{
  "rpc": {
    "listen": "127.0.0.1:18080",
    "allow_remote": false,
    "auth_token": "legacy-token",
    "read_timeout_seconds": 15,
    "write_timeout_seconds": 15
  },
  "engine": {
    "listen_port": 4661,
    "udp_port": 4662,
    "enable_dht": true,
    "enable_upnp": true,
    "peer_connection_timeout": 30,
    "reconnect_to_server": true,
    "max_connections_per_second": 10,
    "session_connections_limit": 20,
    "upload_slots": 3,
    "max_upload_rate_kb": 0,
    "default_download_dir": "./data/downloads"
  },
  "bootstrap": {
    "server_addresses": [],
    "server_met_urls": [],
    "nodes_dat_urls": [],
    "kad_nodes": []
  },
  "state": {
    "enabled": false,
    "path": "./data/state/client-state.json",
    "load_on_start": false,
    "save_on_exit": false,
    "auto_save_interval_seconds": 30
  },
  "logging": {
    "level": "info",
    "path": "./data/logs/goed2kd.log"
  }
}`

func TestLegacyConfig_LoadAndValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")
	if err := os.WriteFile(path, []byte(legacyConfigJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Validate(c); err != nil {
		t.Fatalf("validate legacy config: %v", err)
	}
	if c.Engine.UDPPortV6 != 4672 {
		t.Fatalf("UDPPortV6 = %d, want default 4672 after migration", c.Engine.UDPPortV6)
	}
	if c.Engine.EnableDHTv6 {
		t.Fatal("legacy config should default enable_dht_v6 to false")
	}
	if c.Engine.MaxHttpSources != 4 {
		t.Fatalf("MaxHttpSources = %d, want default 4", c.Engine.MaxHttpSources)
	}
	if !c.Engine.EnableWebDownload {
		t.Fatal("legacy config should migrate enable_web_download to true")
	}
	if !c.Engine.PartialKadPublish {
		t.Fatal("legacy config should migrate partial_kad_publish to true")
	}
	if c.Version != config.CurrentConfigVersion {
		t.Fatalf("Version = %d, want %d", c.Version, config.CurrentConfigVersion)
	}
}

func TestExampleConfigs_Valid(t *testing.T) {
	root := findRepoRoot(t)
	for _, name := range []string{"configs/config.example.json", "configs/config.docker.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, name)
			c, err := config.LoadFromFile(path)
			if err != nil {
				t.Fatalf("load %s: %v", name, err)
			}
			if err := config.Validate(c); err != nil {
				t.Fatalf("validate %s: %v", name, err)
			}
		})
	}
}

func TestDefaultConfig_Valid(t *testing.T) {
	c := config.Default()
	if err := config.Validate(c); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
