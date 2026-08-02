package config

// Config 守护进程完整配置（与实现文档 JSON 对齐）。
type Config struct {
	RPC       RPCConfig       `json:"rpc"`
	Engine    EngineConfig    `json:"engine"`
	Bootstrap BootstrapConfig `json:"bootstrap"`
	State     StateConfig     `json:"state"`
	Logging   LoggingConfig   `json:"logging"`
}

// RPCConfig HTTP/WebSocket 服务参数。
type RPCConfig struct {
	Listen               string `json:"listen"`
	AllowRemote          bool   `json:"allow_remote"`
	AuthToken            string `json:"auth_token"`
	ReadTimeoutSeconds   int    `json:"read_timeout_seconds"`
	WriteTimeoutSeconds  int    `json:"write_timeout_seconds"`
}

// EngineConfig 映射到goed2k.Settings 的字段。
type EngineConfig struct {
	ListenPort              int    `json:"listen_port"`
	UDPPort                 int    `json:"udp_port"`
	UDPPortV6               int    `json:"udp_port_v6"`
	EnableDHT               bool   `json:"enable_dht"`
	EnableDHTv6             bool   `json:"enable_dht_v6"`
	EnableUPnP              bool   `json:"enable_upnp"`
	PeerConnectionTimeout   int    `json:"peer_connection_timeout"`
	ReconnectToServer       bool   `json:"reconnect_to_server"`
	MaxConnectionsPerSecond int    `json:"max_connections_per_second"`
	SessionConnectionsLimit int    `json:"session_connections_limit"`
	UploadSlots             int    `json:"upload_slots"`
	MaxUploadRateKB         int    `json:"max_upload_rate_kb"`
	MaxDownloadRateKB       int    `json:"max_download_rate_kb"`
	EnableCryptLayer        bool   `json:"enable_crypt_layer"`
	CryptLayerRequired      bool   `json:"crypt_layer_required"`
	ObfuscationTCPPort      int    `json:"obfuscation_tcp_port"`
	EnableSecIdent          bool   `json:"enable_sec_ident"`
	SecIdentRequired        bool   `json:"sec_ident_required"`
	CreditsOnlyVerified     bool   `json:"credits_only_verified"`
	IdentityKeyPath         string `json:"identity_key_path"`
	UseEmuleTempLayout      bool   `json:"use_emule_temp_layout"`
	PartialKadPublish       bool   `json:"partial_kad_publish"`
	PreallocateDiskSpace    bool   `json:"preallocate_disk_space"`
	UseSparseFiles          bool   `json:"use_sparse_files"`
	EnableWebDownload       bool   `json:"enable_web_download"`
	MaxHttpSources          int    `json:"max_http_sources"`
	MaxConcurrentHttpBlocks int    `json:"max_concurrent_http_blocks"`
	WebCacheDir             string `json:"web_cache_dir"`
	HttpRequestTimeoutSec   int    `json:"http_request_timeout_sec"`
	DefaultDownloadDir      string `json:"default_download_dir"`
}

// BootstrapConfig 启动时可选引导。
type BootstrapConfig struct {
	ServerAddresses []string `json:"server_addresses"`
	ServerMetURLs   []string `json:"server_met_urls"`
	NodesDatURLs    []string `json:"nodes_dat_urls"`
	KadNodes        []string `json:"kad_nodes"`
	Nodes6DatURLs   []string `json:"nodes6_dat_urls"`
	KadV6Nodes      []string `json:"kad_v6_nodes"`
	IPFilterPaths   []string `json:"ipfilter_paths"`
}

// StateConfig 状态持久化。
type StateConfig struct {
	Enabled                 bool   `json:"enabled"`
	Path                    string `json:"path"`
	LoadOnStart             bool   `json:"load_on_start"`
	SaveOnExit              bool   `json:"save_on_exit"`
	AutoSaveIntervalSeconds int    `json:"auto_save_interval_seconds"`
}

// LoggingConfig 日志。
type LoggingConfig struct {
	Level string `json:"level"`
	Path  string `json:"path"`
}

// Default 返回内建默认配置。
func Default() *Config {
	return &Config{
		RPC: RPCConfig{
			Listen:               "127.0.0.1:18080",
			AllowRemote:          false,
			AuthToken:            "change-me",
			ReadTimeoutSeconds:   15,
			WriteTimeoutSeconds:  15,
		},
		Engine: EngineConfig{
			ListenPort:              4661,
			UDPPort:                 4662,
			UDPPortV6:               4672,
			EnableDHT:               true,
			EnableDHTv6:             false,
			EnableUPnP:              true,
			PeerConnectionTimeout:   30,
			ReconnectToServer:       true,
			MaxConnectionsPerSecond: 10,
			SessionConnectionsLimit: 20,
			UploadSlots:             3,
			MaxUploadRateKB:         0,
			MaxDownloadRateKB:       0,
			PartialKadPublish:       true,
			EnableWebDownload:       true,
			MaxHttpSources:          4,
			MaxConcurrentHttpBlocks: 2,
			HttpRequestTimeoutSec:   30,
			DefaultDownloadDir:      "./data/downloads",
		},
		Bootstrap: BootstrapConfig{
			ServerAddresses: []string{"45.82.80.155:5687"},
			ServerMetURLs:   []string{"http://upd.emule-security.org/server.met"},
			NodesDatURLs:    []string{"https://upd.emule-security.org/nodes.dat"},
			KadNodes:        []string{},
			Nodes6DatURLs:   []string{},
			KadV6Nodes:      []string{},
			IPFilterPaths:   []string{},
		},
		State: StateConfig{
			Enabled:                 true,
			Path:                    "./data/state/client-state.json",
			LoadOnStart:             true,
			SaveOnExit:              true,
			AutoSaveIntervalSeconds: 30,
		},
		Logging: LoggingConfig{
			Level: "info",
			Path:  "./data/logs/goed2kd.log",
		},
	}
}
