package llmx

import (
	"fmt"
	"time"
)

// Error is the interface for llmx errors
type Error interface {
	error
	Code() string
	StatusCode() int
	Retryable() bool
}

// BaseError is the base error type
type BaseError struct {
	Message    string
	StatusCd   int
	ErrorCode  string
	IsRetry    bool
	Underlying error
}

func (e *BaseError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Underlying)
	}
	return e.Message
}

func (e *BaseError) Code() string {
	return e.ErrorCode
}

func (e *BaseError) StatusCode() int {
	return e.StatusCd
}

func (e *BaseError) Retryable() bool {
	return e.IsRetry
}

func (e *BaseError) Unwrap() error {
	return e.Underlying
}

// InvalidRequestError represents an invalid request
type InvalidRequestError struct {
	*BaseError
	Details map[string]interface{}
}

// NewInvalidRequestError creates a new invalid request error
func NewInvalidRequestError(message string, details map[string]interface{}) *InvalidRequestError {
	return &InvalidRequestError{
		BaseError: &BaseError{
			Message:   message,
			StatusCd:  400,
			ErrorCode: "invalid_request",
			IsRetry:   false,
		},
		Details: details,
	}
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	*BaseError
	RetryAfter time.Duration
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string, retryAfter time.Duration) *RateLimitError {
	return &RateLimitError{
		BaseError: &BaseError{
			Message:   message,
			StatusCd:  429,
			ErrorCode: "rate_limit",
			IsRetry:   true,
		},
		RetryAfter: retryAfter,
	}
}

// AuthenticationError represents an authentication error
type AuthenticationError struct {
	*BaseError
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		BaseError: &BaseError{
			Message:   message,
			StatusCd:  401,
			ErrorCode: "authentication",
			IsRetry:   false,
		},
	}
}

// NotFoundError represents a not found error
type NotFoundError struct {
	*BaseError
	Resource string
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, resource string) *NotFoundError {
	return &NotFoundError{
		BaseError: &BaseError{
			Message:   message,
			StatusCd:  404,
			ErrorCode: "not_found",
			IsRetry:   false,
		},
		Resource: resource,
	}
}

// InternalError represents an internal error
type InternalError struct {
	*BaseError
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{
		BaseError: &BaseError{
			Message:    message,
			StatusCd:   500,
			ErrorCode:  "internal",
			IsRetry:    true,
			Underlying: cause,
		},
	}
}

// ProviderError represents a provider-specific error
type ProviderError struct {
	*BaseError
	Provider string
}

// NewProviderError creates a new provider error
func NewProviderError(provider string, message string, statusCode int, cause error) *ProviderError {
	return &ProviderError{
		BaseError: &BaseError{
			Message:    message,
			StatusCd:   statusCode,
			ErrorCode:  "provider_error",
			IsRetry:    statusCode >= 500,
			Underlying: cause,
		},
		Provider: provider,
	}
}
