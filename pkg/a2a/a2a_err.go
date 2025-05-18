package a2a

import (
	"fmt"
)

// ErrorCode represents A2A protocol error codes as defined in the JSON-RPC 2.0 specification
// with additional A2A-specific error codes.
type ErrorCode int

const (
	// Standard JSON-RPC 2.0 error codes
	ErrorParse          ErrorCode = -32700 // Invalid JSON was received
	ErrorInvalidRequest ErrorCode = -32600 // The JSON sent is not a valid Request object
	ErrorMethodNotFound ErrorCode = -32601 // The method does not exist / is not available
	ErrorInvalidParams  ErrorCode = -32602 // Invalid method parameter(s)
	ErrorInternal       ErrorCode = -32603 // Internal JSON-RPC error

	// A2A-specific error codes (range -32000 to -32099 are reserved for implementation-specific errors)
	
	// Task-related errors
	ErrorTaskNotFound                 ErrorCode = -32001 // The requested task was not found
	ErrorTaskCantCancel               ErrorCode = -32002 // The task cannot be canceled in its current state
	ErrorInvalidTaskState             ErrorCode = -32003 // The task is in an invalid state for the requested operation
	
	// Communication-related errors
	ErrorPushNotificationNotSupported ErrorCode = -32010 // Push notifications are not supported by this agent
	ErrorUnsupportedOperation         ErrorCode = -32011 // The requested operation is not supported
	ErrorIncompatibleContentType      ErrorCode = -32012 // The content type is not compatible with the operation
	
	// Authentication and authorization errors
	ErrorAuthenticationFailed ErrorCode = -32020 // Authentication failed
	ErrorPermissionDenied     ErrorCode = -32021 // Permission denied for the requested operation
	
	// Service-related errors
	ErrorRateLimitExceeded    ErrorCode = -32030 // Rate limit for requests has been exceeded
	ErrorServiceUnavailable   ErrorCode = -32031 // The service is currently unavailable
	ErrorTimeout              ErrorCode = -32032 // The operation timed out
)

// JSONRPCError represents an A2A protocol error following the JSON-RPC 2.0 specification.
// It contains an error code, a message describing the error, and optional additional data.
type JSONRPCError struct {
	// Code is the error code as defined by the JSON-RPC 2.0 specification or A2A-specific codes
	Code ErrorCode `json:"code"`
	
	// Message is a short description of the error
	Message string `json:"message"`
	
	// Data is optional additional information about the error
	Data map[string]any `json:"data,omitempty"`
}

// Error implements the error interface to make JSONRPCError usable as a standard Go error.
// It returns a formatted string containing the error code, message, and data.
func (e JSONRPCError) Error() string {
	return fmt.Sprintf("Micro-A2A error %d: %s | %v", e.Code, e.Message, e.Data)
}

// NewError creates a new A2A protocol error with the specified code, message, and optional data.
//
// Parameters:
//   - code: The error code as defined in the ErrorCode constants
//   - message: A short description of the error
//   - data: Optional additional information about the error (can be nil)
//
// Returns:
//   - A JSONRPCError instance that implements the error interface
func NewError(code ErrorCode, message string, data map[string]any) JSONRPCError {
	return JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
