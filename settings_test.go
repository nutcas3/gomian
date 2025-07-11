package gomian

import (
	"errors"
	"testing"
	"time"
)

func TestConsecutiveFailuresThreshold(t *testing.T) {
	threshold := ConsecutiveFailures(3).(ConsecutiveFailuresThreshold)
	
	// Test initial state
	if threshold.ShouldTrip(0, 0, 0, 0) {
		t.Error("Should not trip with 0 failures")
	}
	
	// Test below threshold
	if threshold.ShouldTrip(2, 0, 10, 0) {
		t.Error("Should not trip with failures below threshold")
	}
	
	// Test at threshold
	if !threshold.ShouldTrip(3, 0, 10, 0) {
		t.Error("Should trip with failures at threshold")
	}
	
	// Test above threshold
	if !threshold.ShouldTrip(4, 0, 10, 0) {
		t.Error("Should trip with failures above threshold")
	}
}

func TestFailureRateThreshold(t *testing.T) {
	// 50% failure rate threshold with minimum 10 requests
	threshold := NewFailureRateThreshold(0.5, 10)
	
	// Test initial state with no requests
	if threshold.ShouldTrip(0, 0, 0, 10*time.Second) {
		t.Error("Should not trip with 0 requests")
	}
	
	// Test below minimum request volume
	if threshold.ShouldTrip(4, 4, 8, 10*time.Second) {
		t.Error("Should not trip below minimum request volume")
	}
	
	// Test at minimum request volume but below failure rate
	if threshold.ShouldTrip(4, 6, 10, 10*time.Second) {
		t.Error("Should not trip at minimum request volume but below failure rate")
	}
	
	// Test at minimum request volume and at failure rate
	if !threshold.ShouldTrip(5, 5, 10, 10*time.Second) {
		t.Error("Should trip at minimum request volume and at failure rate")
	}
	
	// Test at minimum request volume and above failure rate
	if !threshold.ShouldTrip(6, 4, 10, 10*time.Second) {
		t.Error("Should trip at minimum request volume and above failure rate")
	}
	
	// Test with higher request volume
	if !threshold.ShouldTrip(10, 10, 20, 10*time.Second) {
		t.Error("Should trip with higher request volume and at failure rate")
	}
	
	if threshold.ShouldTrip(9, 11, 20, 10*time.Second) {
		t.Error("Should not trip with higher request volume but below failure rate")
	}
}

func TestSettings(t *testing.T) {
	// Test default settings
	settings := Settings{
		Name:             "TestBreaker",
		FailureThreshold: ConsecutiveFailures(5),
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
	}
	
	// Test name
	if settings.Name != "TestBreaker" {
		t.Errorf("Name should be 'TestBreaker', got %s", settings.Name)
	}
	
	// Test failure threshold
	threshold, ok := settings.FailureThreshold.(ConsecutiveFailuresThreshold)
	if !ok {
		t.Error("FailureThreshold should be ConsecutiveFailuresThreshold")
	}
	if threshold.Threshold != 5 {
		t.Errorf("Threshold should be 5, got %d", threshold.Threshold)
	}
	
	// Test success threshold
	if settings.SuccessThreshold != 2 {
		t.Errorf("SuccessThreshold should be 2, got %d", settings.SuccessThreshold)
	}
	
	// Test timeout
	if settings.Timeout != 1*time.Second {
		t.Errorf("Timeout should be 1s, got %v", settings.Timeout)
	}
}

func TestIsFailureFunction(t *testing.T) {
	settings := Settings{
		Name:             "TestBreaker",
		FailureThreshold: ConsecutiveFailures(3),
		IsFailure: func(err error) bool {
			return err != nil && err.Error() == "critical error"
		},
	}
	
	// Test with nil error
	if settings.IsFailure(nil) {
		t.Error("IsFailure should return false for nil error")
	}
	
	// Test with non-matching error
	if settings.IsFailure(errors.New("non-critical error")) {
		t.Error("IsFailure should return false for non-matching error")
	}
	
	// Test with matching error
	if !settings.IsFailure(errors.New("critical error")) {
		t.Error("IsFailure should return true for matching error")
	}
}

func TestIgnoredErrors(t *testing.T) {
	ignoredErr := errors.New("ignored error")
	settings := Settings{
		Name:             "TestBreaker",
		FailureThreshold: ConsecutiveFailures(3),
		IgnoredErrors:    []error{ignoredErr},
	}
	
	// Test with nil error
	if settings.IsFailure != nil {
		t.Skip("Custom IsFailure function is set, skipping default behavior test")
	}
	
	// Create a circuit breaker to test isFailure method
	cb := NewCircuitBreaker(settings)
	
	// Test with nil error
	if cb.isFailure(nil) {
		t.Error("isFailure should return false for nil error")
	}
	
	// Test with ignored error
	if cb.isFailure(ignoredErr) {
		t.Error("isFailure should return false for ignored error")
	}
	
	// Test with non-ignored error
	if !cb.isFailure(errors.New("non-ignored error")) {
		t.Error("isFailure should return true for non-ignored error")
	}
}
