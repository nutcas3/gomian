package gomian

import (
	"errors"
	"fmt"
)

var (
	// ErrCircuitOpen is returned when a request is rejected because the circuit is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitError represents an error that occurred within the circuit breaker.
type CircuitError struct {
	Name string
	Err  error
}

// Error returns a string representation of the CircuitError.
func (e *CircuitError) Error() string {
	return fmt.Sprintf("circuit breaker '%s': %v", e.Name, e.Err)
}

// Unwrap returns the underlying error.
func (e *CircuitError) Unwrap() error {
	return e.Err
}

// IsCircuitOpen checks if the error is or wraps an ErrCircuitOpen error.
func IsCircuitOpen(err error) bool {
	return errors.Is(err, ErrCircuitOpen)
}
