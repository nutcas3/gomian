package counter

import (
	"testing"
	"time"
)

func TestConsecutiveCounter(t *testing.T) {
	cc := NewConsecutiveCounter()
	
	// Test initial state
	if cc.ConsecutiveSuccesses() != 0 || cc.ConsecutiveFailures() != 0 {
		t.Errorf("Initial state should be 0 for both counters, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	// Test incrementing success
	cc.IncrementSuccess()
	if cc.ConsecutiveSuccesses() != 1 || cc.ConsecutiveFailures() != 0 {
		t.Errorf("After one success, should have 1 success and 0 failures, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	// Test incrementing success again
	cc.IncrementSuccess()
	if cc.ConsecutiveSuccesses() != 2 || cc.ConsecutiveFailures() != 0 {
		t.Errorf("After two successes, should have 2 successes and 0 failures, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	// Test incrementing failure resets success counter
	cc.IncrementFailure()
	if cc.ConsecutiveSuccesses() != 0 || cc.ConsecutiveFailures() != 1 {
		t.Errorf("After failure, should have 0 successes and 1 failure, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	// Test incrementing failure again
	cc.IncrementFailure()
	if cc.ConsecutiveSuccesses() != 0 || cc.ConsecutiveFailures() != 2 {
		t.Errorf("After two failures, should have 0 successes and 2 failures, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	// Test incrementing success resets failure counter
	cc.IncrementSuccess()
	if cc.ConsecutiveSuccesses() != 1 || cc.ConsecutiveFailures() != 0 {
		t.Errorf("After success, should have 1 success and 0 failures, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	// Test totals
	successes, failures := cc.Totals()
	if successes != 3 || failures != 2 {
		t.Errorf("Totals should be 3 successes and 2 failures, got %d successes and %d failures",
			successes, failures)
	}
	
	// Test reset
	cc.Reset()
	if cc.ConsecutiveSuccesses() != 0 || cc.ConsecutiveFailures() != 0 {
		t.Errorf("After reset, should have 0 successes and 0 failures, got %d successes and %d failures",
			cc.ConsecutiveSuccesses(), cc.ConsecutiveFailures())
	}
	
	successes, failures = cc.Totals()
	if successes != 0 || failures != 0 {
		t.Errorf("After reset, totals should be 0 successes and 0 failures, got %d successes and %d failures",
			successes, failures)
	}
}

func TestRollingWindow(t *testing.T) {
	// Create a rolling window with a 100ms window size and 10 buckets
	rw := NewRollingWindow(100*time.Millisecond, 10)
	
	// Test initial state
	requests, failures := rw.Counts()
	if requests != 0 || failures != 0 {
		t.Errorf("Initial state should be 0 for both counters, got %d requests and %d failures",
			requests, failures)
	}
	
	// Test incrementing success
	rw.IncrementSuccess()
	requests, failures = rw.Counts()
	if requests != 1 || failures != 0 {
		t.Errorf("After one success, should have 1 request and 0 failures, got %d requests and %d failures",
			requests, failures)
	}
	
	// Test incrementing failure
	rw.IncrementFailure()
	requests, failures = rw.Counts()
	if requests != 2 || failures != 1 {
		t.Errorf("After one success and one failure, should have 2 requests and 1 failure, got %d requests and %d failures",
			requests, failures)
	}
	
	// Test reset
	rw.Reset()
	requests, failures = rw.Counts()
	if requests != 0 || failures != 0 {
		t.Errorf("After reset, should have 0 requests and 0 failures, got %d requests and %d failures",
			requests, failures)
	}
}

func TestRollingWindowRotation(t *testing.T) {
	// Create a rolling window with a 100ms window size and 2 buckets
	rw := NewRollingWindow(100*time.Millisecond, 2)
	
	// Add some events
	rw.IncrementSuccess()
	rw.IncrementSuccess()
	rw.IncrementFailure()
	
	// Check initial counts
	requests, failures := rw.Counts()
	if requests != 3 || failures != 1 {
		t.Errorf("Initial counts should be 3 requests and 1 failure, got %d requests and %d failures",
			requests, failures)
	}
	
	// Wait for more than one bucket duration but less than the full window
	time.Sleep(60 * time.Millisecond)
	
	// Add more events
	rw.IncrementSuccess()
	rw.IncrementFailure()
	
	// Check counts after rotation
	requests, failures = rw.Counts()
	if requests != 5 || failures != 2 {
		t.Errorf("After partial rotation, should have 5 requests and 2 failures, got %d requests and %d failures",
			requests, failures)
	}
	
	// Wait for the full window to expire
	time.Sleep(110 * time.Millisecond)
	
	// Check counts after full window expiration
	requests, failures = rw.Counts()
	if requests != 0 || failures != 0 {
		t.Errorf("After full window expiration, should have 0 requests and 0 failures, got %d requests and %d failures",
			requests, failures)
	}
}
