package model

// API 业务错误码（与实现文档一致）。
const (
	CodeOK                   = "OK"
	CodeBadRequest           = "BAD_REQUEST"
	CodeUnauthorized         = "UNAUTHORIZED"
	CodeForbidden            = "FORBIDDEN"
	CodeNotFound             = "NOT_FOUND"
	CodeInternalError        = "INTERNAL_ERROR"
	CodeEngineNotRunning     = "ENGINE_NOT_RUNNING"
	CodeEngineAlreadyRunning = "ENGINE_ALREADY_RUNNING"
	CodeInvalidHash          = "INVALID_HASH"
	CodeInvalidED2KLink      = "INVALID_ED2K_LINK"
	CodeTransferNotFound     = "TRANSFER_NOT_FOUND"
	CodeSharedFileNotFound   = "SHARED_FILE_NOT_FOUND"
	CodeSearchNotRunning     = "SEARCH_NOT_RUNNING"
	CodeSearchAlreadyRunning = "SEARCH_ALREADY_RUNNING"
	CodeConfigInvalid        = "CONFIG_INVALID"
	CodeStateStoreError      = "STATE_STORE_ERROR"
)
