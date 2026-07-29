package version

// Version 由 goreleaser 构建时通过 -ldflags 注入；本地 go build 使用此默认值。
var Version = "0.2.0"
