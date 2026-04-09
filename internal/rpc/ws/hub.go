package eventws

import (
	"context"
	"sync"

	"github.com/goed2k/daemon/internal/model"
)

// Hub WebSocket 广播中心：慢客户端丢事件不阻塞全局。
type Hub struct {
	mu    sync.RWMutex
	subs  map[int]chan model.EventEnvelope
	next  int
	bufSz int
}

// NewHub 构造Hub。
func NewHub() *Hub {
	return &Hub{subs: make(map[int]chan model.EventEnvelope), bufSz: 32}
}

// Subscribe 注册订阅者，返回取消函数。
func (h *Hub) Subscribe() (<-chan model.EventEnvelope, func()) {
	ch := make(chan model.EventEnvelope, h.bufSz)
	h.mu.Lock()
	id := h.next
	h.next++
	h.subs[id] = ch
	h.mu.Unlock()
	cancel := func() {
		h.mu.Lock()
		if c, ok := h.subs[id]; ok {
			delete(h.subs, id)
			close(c)
		}
		h.mu.Unlock()
	}
	return ch, cancel
}

// Publish 向所有订阅者非阻塞发送。
func (h *Hub) Publish(ev model.EventEnvelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.subs {
		select {
		case ch <- ev:
		default:
			// 慢客户端丢弃本次事件
		}
	}
}

// RunIngress 从上游channel 消费并广播，直到 ctx 结束。
func (h *Hub) RunIngress(ctx context.Context, ch <-chan model.EventEnvelope) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			h.Publish(ev)
		}
	}
}
