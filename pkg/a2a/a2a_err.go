package a2a

import (
	"fmt"
)

// ErrorCode represents A2A protocol error codes
type ErrorCode int

const (
	ErrorParse          ErrorCode = -32700
	ErrorInvalidRequest ErrorCode = -32600
	ErrorMethodNotFound ErrorCode = -32601
	ErrorInvalidParams  ErrorCode = -32602
	ErrorInternal       ErrorCode = -32603

	// -32000 to -32099	are Server errors	reserved for implementation specific error codes

	ErrorTaskNotFound                 ErrorCode = -32001
	ErrorTaskCantCancel               ErrorCode = -32002
	ErrorPushNotificationNotSupported ErrorCode = -32003
	ErrorUnsupportedOperation         ErrorCode = -32004
	ErrorIncompatibleContentType      ErrorCode = -32005

	// TODO(daniel) reassigne new values
	ErrorAuthenticationFailed ErrorCode = -32000
	ErrorPermissionDenied     ErrorCode = -32001
	ErrorInvalidTaskState     ErrorCode = -32002
	ErrorRateLimitExceeded    ErrorCode = -32004
	ErrorServiceUnavailable   ErrorCode = -32005
	ErrorTimeout              ErrorCode = -32006
)

// Error represents an A2A protocol error
type JSONRPCError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

// Error implements the error interface
func (e JSONRPCError) Error() string {
	return fmt.Sprintf("Micro-A2A error %d: %s | %s", e.Code, e.Message, e.Data)
}

// NewError creates a new A2A protocol error
func NewError(code ErrorCode, message string, data map[string]any) JSONRPCError {
	return JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
