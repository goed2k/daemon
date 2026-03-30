package model

import "time"

// EventEnvelope WebSocket 与内部桥接共用的事件外壳。
type EventEnvelope struct {
	Type string         `json:"type"`
	At   time.Time      `json:"at"`
	Data map[string]any `json:"data"`
}
