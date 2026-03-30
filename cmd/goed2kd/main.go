package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/chenjia404/goed2kd/internal/app"
	"github.com/chenjia404/goed2kd/internal/config"
)

func main() {
	cfgPath := flag.String("config", "data/config/config.json", "配置文件路径")
	flag.Parse()

	cfg, err := config.EnsureFile(*cfgPath)
	if err != nil {
		slog.Error("加载配置失败", "path", *cfgPath, "err", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		slog.Error("配置无效", "err", err)
		os.Exit(1)
	}

	log, closeLog, err := app.NewLogger(&cfg.Logging)
	if err != nil {
		slog.Error("初始化日志失败", "err", err)
		os.Exit(1)
	}
	defer func() { _ = closeLog() }()

	d := app.NewDaemon(log, *cfgPath, cfg)
	if err := d.Run(); err != nil {
		log.Error("运行失败", "err", err)
		os.Exit(1)
	}
}
