// Package crawl4ai provides a Go SDK for Crawl4AI Cloud API
package crawl4ai

import "fmt"

// CloudError is the base error type for all API errors.
type CloudError struct {
	Message    string
	StatusCode int
	Response   map[string]interface{}
	Headers    map[string]string
}

func (e *CloudError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
	}
	return e.Message
}

// NewCloudError creates a new CloudError.
func NewCloudError(message string, statusCode int, response map[string]interface{}, headers map[string]string) *CloudError {
	if response == nil {
		response = make(map[string]interface{})
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	return &CloudError{
		Message:    message,
		StatusCode: statusCode,
		Response:   response,
		Headers:    headers,
	}
}

// AuthenticationError represents a 401 error.
type AuthenticationError struct {
	*CloudError
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(message string, response map[string]interface{}, headers map[string]string) *AuthenticationError {
	return &AuthenticationError{
		CloudError: NewCloudError(message, 401, response, headers),
	}
}

// RateLimitError represents a 429 rate limit error.
type RateLimitError struct {
	*CloudError
}

// NewRateLimitError creates a new RateLimitError.
func NewRateLimitError(message string, response map[string]interface{}, headers map[string]string) *RateLimitError {
	return &RateLimitError{
		CloudError: NewCloudError(message, 429, response, headers),
	}
}

// RetryAfter returns the seconds until rate limit resets.
func (e *RateLimitError) RetryAfter() int {
	if val, ok := e.Headers["x-ratelimit-reset"]; ok {
		var result int
		fmt.Sscanf(val, "%d", &result)
		return result
	}
	return 0
}

// QuotaExceededError represents a 429 quota exceeded error.
type QuotaExceededError struct {
	*CloudError
}

// NewQuotaExceededError creates a new QuotaExceededError.
func NewQuotaExceededError(message string, response map[string]interface{}, headers map[string]string) *QuotaExceededError {
	return &QuotaExceededError{
		CloudError: NewCloudError(message, 429, response, headers),
	}
}

// NotFoundError represents a 404 error.
type NotFoundError struct {
	*CloudError
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(message string, response map[string]interface{}, headers map[string]string) *NotFoundError {
	return &NotFoundError{
		CloudError: NewCloudError(message, 404, response, headers),
	}
}

// ValidationError represents a 400 error.
type ValidationError struct {
	*CloudError
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message string, response map[string]interface{}, headers map[string]string) *ValidationError {
	return &ValidationError{
		CloudError: NewCloudError(message, 400, response, headers),
	}
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	*CloudError
}

// NewTimeoutError creates a new TimeoutError.
func NewTimeoutError(message string) *TimeoutError {
	return &TimeoutError{
		CloudError: NewCloudError(message, 504, nil, nil),
	}
}

// ServerError represents a 500/503 error.
type ServerError struct {
	*CloudError
}

// NewServerError creates a new ServerError.
func NewServerError(message string, statusCode int, response map[string]interface{}, headers map[string]string) *ServerError {
	return &ServerError{
		CloudError: NewCloudError(message, statusCode, response, headers),
	}
}
