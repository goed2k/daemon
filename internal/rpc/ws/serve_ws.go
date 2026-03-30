package eventws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func extractWSToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	if v := r.Header.Get("X-Auth-Token"); v != "" {
		return strings.TrimSpace(v)
	}
	return r.URL.Query().Get("token")
}

// ServeWS 返回 WebSocket 处理器：推送EventEnvelope JSON。
func (h *Hub) ServeWS(log *slog.Logger, authToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := extractWSToken(r)
		if tok != strings.TrimSpace(authToken) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			if log != nil {
				log.Warn("ws upgrade failed", "err", err)
			}
			return
		}
		if log != nil {
			log.Info("websocket connected", "remote", r.RemoteAddr)
		}

		clientCh, unsub := h.Subscribe()
		defer func() {
			unsub()
			_ = conn.Close()
			if log != nil {
				log.Info("websocket disconnected", "remote", r.RemoteAddr)
			}
		}()

		writeMu := make(chan struct{}, 1)
		writeMu <- struct{}{}

		pingPeriod := 20 * time.Second
		pongWait := 60 * time.Second
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(pongWait))
		})

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				<-writeMu
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					writeMu <- struct{}{}
					return
				}
				writeMu <- struct{}{}
			case ev, ok := <-clientCh:
				if !ok {
					return
				}
				b, err := json.Marshal(ev)
				if err != nil {
					continue
				}
				<-writeMu
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				err = conn.WriteMessage(websocket.TextMessage, b)
				writeMu <- struct{}{}
				if err != nil {
					return
				}
			}
		}
	}
}
