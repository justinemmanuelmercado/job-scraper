package errors

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeDatabase    ErrorType = "database"
	ErrorTypeScraper     ErrorType = "scraper"
	ErrorTypeNetwork     ErrorType = "network"
	ErrorTypeConfig      ErrorType = "config"
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeNotifier    ErrorType = "notifier"
	ErrorTypeInternal    ErrorType = "internal"
)

// AppError represents an application-specific error
type AppError struct {
	Type     ErrorType              `json:"type"`
	Message  string                 `json:"message"`
	Cause    error                  `json:"-"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Source   string                 `json:"source,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSource adds source information to the error
func (e *AppError) WithSource(source string) *AppError {
	e.Source = source
	return e
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// Database error constructors
func NewDatabaseError(message string) *AppError {
	return New(ErrorTypeDatabase, message)
}

func WrapDatabaseError(err error, message string) *AppError {
	return Wrap(err, ErrorTypeDatabase, message)
}

// Scraper error constructors
func NewScraperError(message string) *AppError {
	return New(ErrorTypeScraper, message)
}

func WrapScraperError(err error, message string) *AppError {
	return Wrap(err, ErrorTypeScraper, message)
}

// Network error constructors
func NewNetworkError(message string) *AppError {
	return New(ErrorTypeNetwork, message)
}

func WrapNetworkError(err error, message string) *AppError {
	return Wrap(err, ErrorTypeNetwork, message)
}

// Config error constructors
func NewConfigError(message string) *AppError {
	return New(ErrorTypeConfig, message)
}

func WrapConfigError(err error, message string) *AppError {
	return Wrap(err, ErrorTypeConfig, message)
}

// Notifier error constructors
func NewNotifierError(message string) *AppError {
	return New(ErrorTypeNotifier, message)
}

func WrapNotifierError(err error, message string) *AppError {
	return Wrap(err, ErrorTypeNotifier, message)
}