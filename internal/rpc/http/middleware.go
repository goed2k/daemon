package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/chenjia404/goed2kd/internal/model"
)

type ctxKey int

const (
	ctxKeyLog ctxKey = iota
)

// WithRequestLog 将带 request_id 的logger 放入 context。
func WithRequestLog(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := middleware.GetReqID(r.Context())
			l := base
			if rid != "" {
				l = base.With("request_id", rid)
			}
			ctx := context.WithValue(r.Context(), ctxKeyLog, l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestLog 从 context 取logger。
func RequestLog(r *http.Request) *slog.Logger {
	if v := r.Context().Value(ctxKeyLog); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}

// RecoverJSON panic 时返回统一 JSON。
func RecoverJSON(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					l := RequestLog(r)
					if log != nil {
						l = log
					}
					l.Error("panic", "recover", rec)
					WriteError(w, l, model.NewAppError(model.CodeInternalError, "internal error", nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// AuthToken Bearer 或X-Auth-Token 校验（health 除外）。
func AuthToken(token string) func(http.Handler) http.Handler {
	tok := strings.TrimSpace(token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v1/system/health" {
				next.ServeHTTP(w, r)
				return
			}
			got := extractToken(r)
			if got == "" || got != tok {
				WriteError(w, RequestLog(r), model.NewAppError(model.CodeUnauthorized, "unauthorized", nil))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	if v := r.Header.Get("X-Auth-Token"); v != "" {
		return strings.TrimSpace(v)
	}
	return r.URL.Query().Get("token")
}

// Timeout 包装超时。
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(d)
}

// RequestID 透传 chi 中间件别名。
func RequestID() func(http.Handler) http.Handler {
	return middleware.RequestID
}

// RealIP 可选。
func RealIP() func(http.Handler) http.Handler {
	return middleware.RealIP
}
