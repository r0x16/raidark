package auth

import "fmt"

// CasdoorError represents an error in Casdoor operations
type CasdoorError struct {
	Message string
	Cause   error
}

// Error implements the error interface
func (e *CasdoorError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("casdoor error: %s, caused by: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("casdoor error: %s", e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *CasdoorError) Unwrap() error {
	return e.Cause
}

// newCasdoorError creates a new CasdoorError with a message
func newCasdoorError(message string) *CasdoorError {
	return &CasdoorError{
		Message: message,
	}
}

// newCasdoorErrorWithCause creates a new CasdoorError with a message and cause
func newCasdoorErrorWithCause(message string, cause error) *CasdoorError {
	return &CasdoorError{
		Message: message,
		Cause:   cause,
	}
}
