package server

import "fmt"

// ErrorCode classifies MCP tool errors for structured error handling
type ErrorCode string

const (
	// ErrInvalidInput indicates invalid or malformed input parameters
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound ErrorCode = "NOT_FOUND"
	// ErrDatabaseError indicates a database operation failed
	ErrDatabaseError ErrorCode = "DATABASE_ERROR"
	// ErrInternalError indicates an unexpected internal error
	ErrInternalError ErrorCode = "INTERNAL_ERROR"
)

// ToolError represents a structured tool error with code, message, and optional details
type ToolError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *ToolError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewInvalidInputError creates an error for invalid input parameters
func NewInvalidInputError(msg string) *ToolError {
	return &ToolError{Code: ErrInvalidInput, Message: msg}
}

// NewInvalidInputErrorWithDetails creates an error for invalid input with additional details
func NewInvalidInputErrorWithDetails(msg, details string) *ToolError {
	return &ToolError{Code: ErrInvalidInput, Message: msg, Details: details}
}

// NewNotFoundError creates an error for a missing resource
func NewNotFoundError(resource string) *ToolError {
	return &ToolError{Code: ErrNotFound, Message: fmt.Sprintf("%s not found", resource)}
}

// NewNotFoundErrorWithID creates an error for a missing resource with its identifier
func NewNotFoundErrorWithID(resource string, id interface{}) *ToolError {
	return &ToolError{
		Code:    ErrNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: fmt.Sprintf("id=%v", id),
	}
}

// NewDatabaseError creates an error for database operation failures
func NewDatabaseError(err error) *ToolError {
	return &ToolError{
		Code:    ErrDatabaseError,
		Message: "Database operation failed",
		Details: err.Error(),
	}
}

// NewDatabaseErrorWithContext creates a database error with additional context
func NewDatabaseErrorWithContext(operation string, err error) *ToolError {
	return &ToolError{
		Code:    ErrDatabaseError,
		Message: fmt.Sprintf("Database %s failed", operation),
		Details: err.Error(),
	}
}

// NewInternalError creates an error for unexpected internal failures
func NewInternalError(msg string) *ToolError {
	return &ToolError{Code: ErrInternalError, Message: msg}
}

// NewInternalErrorWithCause creates an internal error wrapping another error
func NewInternalErrorWithCause(msg string, err error) *ToolError {
	return &ToolError{
		Code:    ErrInternalError,
		Message: msg,
		Details: err.Error(),
	}
}
