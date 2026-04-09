package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/goed2k/daemon/internal/model"
)

// APIResponse 统一 HTTP JSON 响应。
type APIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteSuccess 写入成功响应。
func WriteSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(APIResponse{Code: model.CodeOK, Data: data})
}

// WriteError 写入错误响应（自动映射AppError）。
func WriteError(w http.ResponseWriter, log *slog.Logger, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var ae *model.AppError
	if errors.As(err, &ae) {
		status := codeToHTTPStatus(ae.Code)
		w.WriteHeader(status)
		msg := ae.Message
		if msg == "" && ae.Err != nil {
			msg = ae.Err.Error()
		}
		if log != nil {
			log.Warn("api error", "code", ae.Code, "msg", msg, "err", ae.Err)
		}
		_ = json.NewEncoder(w).Encode(APIResponse{Code: ae.Code, Message: msg})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	if log != nil {
		log.Error("internal error", "err", err)
	}
	_ = json.NewEncoder(w).Encode(APIResponse{
		Code:    model.CodeInternalError,
		Message: "internal error",
	})
}

func codeToHTTPStatus(code string) int {
	switch code {
	case model.CodeBadRequest, model.CodeInvalidHash, model.CodeInvalidED2KLink, model.CodeConfigInvalid:
		return http.StatusBadRequest
	case model.CodeUnauthorized:
		return http.StatusUnauthorized
	case model.CodeForbidden:
		return http.StatusForbidden
	case model.CodeNotFound, model.CodeTransferNotFound, model.CodeSharedFileNotFound, model.CodeSearchNotRunning:
		return http.StatusNotFound
	case model.CodeEngineNotRunning:
		return http.StatusServiceUnavailable
	case model.CodeEngineAlreadyRunning, model.CodeSearchAlreadyRunning, model.CodeStateStoreError:
		return http.StatusConflict
	case model.CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
