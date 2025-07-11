package gomian

import (
	"errors"
	"testing"
)

func TestErrCircuitOpen(t *testing.T) {
	// Test that ErrCircuitOpen is defined
	if ErrCircuitOpen == nil {
		t.Error("ErrCircuitOpen should not be nil")
	}
}

func TestCircuitError(t *testing.T) {
	// Test creating a new circuit error
	name := "TestBreaker"
	message := "test error"
	err := &CircuitError{
		Name: name,
		Err:  errors.New(message),
	}
	
	// Check error fields
	circuitErr := err
	if circuitErr.Name != name {
		t.Errorf("CircuitError.Name should be %s, got %s", name, circuitErr.Name)
	}
	
	// Check error message
	expected := "circuit breaker 'TestBreaker': test error"
	if err.Error() != expected {
		t.Errorf("Error message should be '%s', got '%s'", expected, err.Error())
	}
}

func TestCircuitErrorWithWrappedError(t *testing.T) {
	// Test creating a circuit error with a wrapped error
	name := "TestBreaker"
	wrappedErr := errors.New("wrapped error")
	err := &CircuitError{
		Name: name,
		Err:  wrappedErr,
	}
	
	// Check error fields
	circuitErr := err
	if circuitErr.Name != name {
		t.Errorf("CircuitError.Name should be %s, got %s", name, circuitErr.Name)
	}
	
	// Check error message
	expected := "circuit breaker 'TestBreaker': wrapped error"
	if err.Error() != expected {
		t.Errorf("Error message should be '%s', got '%s'", expected, err.Error())
	}
	
	// Check wrapped error
	if !errors.Is(err, wrappedErr) {
		t.Error("CircuitError should wrap the provided error")
	}
}

func TestIsCircuitOpen(t *testing.T) {
	// Test with nil error
	if IsCircuitOpen(nil) {
		t.Error("IsCircuitOpen should return false for nil error")
	}
	
	// Test with non-circuit error
	if IsCircuitOpen(errors.New("regular error")) {
		t.Error("IsCircuitOpen should return false for non-circuit errors")
	}
	
	// Test with circuit open error
	if !IsCircuitOpen(ErrCircuitOpen) {
		t.Error("IsCircuitOpen should return true for ErrCircuitOpen")
	}
	
	// Test with wrapped circuit open error
	circuitErr := &CircuitError{
		Name: "test",
		Err:  ErrCircuitOpen,
	}
	if !IsCircuitOpen(circuitErr) {
		t.Error("IsCircuitOpen should return true for wrapped ErrCircuitOpen")
	}
	
	// Test with other error
	otherErr := errors.New("other error")
	if IsCircuitOpen(otherErr) {
		t.Error("IsCircuitOpen should return false for other errors")
	}
}

func TestUnwrap(t *testing.T) {
	// Test unwrapping a CircuitError with a wrapped error
	wrappedErr := errors.New("wrapped error")
	err := &CircuitError{
		Name: "TestBreaker",
		Err:  wrappedErr,
	}
	
	// Check that errors.Unwrap returns the wrapped error
	if errors.Unwrap(err) != wrappedErr {
		t.Error("errors.Unwrap should return the wrapped error")
	}
	
	// Test unwrapping a CircuitError with a nil error
	err = &CircuitError{
		Name: "TestBreaker",
		Err:  nil,
	}
	if errors.Unwrap(err) != nil {
		t.Error("errors.Unwrap should return nil for CircuitError with nil error")
	}
}
