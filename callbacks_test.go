package gomian

import (
	"errors"
	"testing"
)

func TestCallbacks(t *testing.T) {
	// Create a new callbacks manager
	cb := NewCallbacks()
	
	// Test initial state
	if len(cb.onStateChange) != 0 {
		t.Errorf("Initial state change callbacks should be empty, got %d", len(cb.onStateChange))
	}
	if len(cb.onTrip) != 0 {
		t.Errorf("Initial trip callbacks should be empty, got %d", len(cb.onTrip))
	}
	if len(cb.onReset) != 0 {
		t.Errorf("Initial reset callbacks should be empty, got %d", len(cb.onReset))
	}
	if len(cb.onSuccess) != 0 {
		t.Errorf("Initial success callbacks should be empty, got %d", len(cb.onSuccess))
	}
	if len(cb.onFailure) != 0 {
		t.Errorf("Initial failure callbacks should be empty, got %d", len(cb.onFailure))
	}
	if len(cb.onRejection) != 0 {
		t.Errorf("Initial rejection callbacks should be empty, got %d", len(cb.onRejection))
	}
}

func TestOnStateChange(t *testing.T) {
	cb := NewCallbacks()
	
	// Track callback invocations
	var called bool
	var calledName string
	var calledFrom, calledTo State
	
	// Register callback
	cb.AddOnStateChange(func(name string, from, to State) {
		called = true
		calledName = name
		calledFrom = from
		calledTo = to
	})
	
	// Check callback was registered
	if len(cb.onStateChange) != 1 {
		t.Errorf("Should have 1 state change callback, got %d", len(cb.onStateChange))
	}
	
	// Test callback invocation
	cb.NotifyStateChange("TestBreaker", Closed, Open)
	if !called {
		t.Error("State change callback should have been called")
	}
	if calledName != "TestBreaker" {
		t.Errorf("Callback name should be 'TestBreaker', got '%s'", calledName)
	}
	if calledFrom != Closed {
		t.Errorf("Callback from state should be Closed, got %v", calledFrom)
	}
	if calledTo != Open {
		t.Errorf("Callback to state should be Open, got %v", calledTo)
	}
	
	// Test multiple callbacks
	var secondCalled bool
	cb.AddOnStateChange(func(name string, from, to State) {
		secondCalled = true
	})
	
	// Check both callbacks are registered
	if len(cb.onStateChange) != 2 {
		t.Errorf("Should have 2 state change callbacks, got %d", len(cb.onStateChange))
	}
	
	// Test both callbacks are invoked
	called = false
	secondCalled = false
	cb.NotifyStateChange("TestBreaker", Open, HalfOpen)
	if !called || !secondCalled {
		t.Error("Both state change callbacks should have been called")
	}
}

func TestOnTrip(t *testing.T) {
	cb := NewCallbacks()
	
	// Track callback invocations
	var called bool
	var calledName string
	var calledErr error
	
	// Register callback
	cb.AddOnTrip(func(name string, err error) {
		called = true
		calledName = name
		calledErr = err
	})
	
	// Check callback was registered
	if len(cb.onTrip) != 1 {
		t.Errorf("Should have 1 trip callback, got %d", len(cb.onTrip))
	}
	
	// Test callback invocation
	testErr := errors.New("test error")
	cb.NotifyTrip("TestBreaker", testErr)
	if !called {
		t.Error("Trip callback should have been called")
	}
	if calledName != "TestBreaker" {
		t.Errorf("Callback name should be 'TestBreaker', got '%s'", calledName)
	}
	if calledErr != testErr {
		t.Errorf("Callback error should be 'test error', got '%v'", calledErr)
	}
}

func TestOnReset(t *testing.T) {
	cb := NewCallbacks()
	
	// Track callback invocations
	var called bool
	var calledName string
	
	// Register callback
	cb.AddOnReset(func(name string) {
		called = true
		calledName = name
	})
	
	// Check callback was registered
	if len(cb.onReset) != 1 {
		t.Errorf("Should have 1 reset callback, got %d", len(cb.onReset))
	}
	
	// Test callback invocation
	cb.NotifyReset("TestBreaker")
	if !called {
		t.Error("Reset callback should have been called")
	}
	if calledName != "TestBreaker" {
		t.Errorf("Callback name should be 'TestBreaker', got '%s'", calledName)
	}
}

func TestOnSuccess(t *testing.T) {
	cb := NewCallbacks()
	
	// Track callback invocations
	var called bool
	var calledName string
	
	// Register callback
	cb.AddOnSuccess(func(name string) {
		called = true
		calledName = name
	})
	
	// Check callback was registered
	if len(cb.onSuccess) != 1 {
		t.Errorf("Should have 1 success callback, got %d", len(cb.onSuccess))
	}
	
	// Test callback invocation
	cb.NotifySuccess("TestBreaker")
	if !called {
		t.Error("Success callback should have been called")
	}
	if calledName != "TestBreaker" {
		t.Errorf("Callback name should be 'TestBreaker', got '%s'", calledName)
	}
}

func TestOnFailure(t *testing.T) {
	cb := NewCallbacks()
	
	// Track callback invocations
	var called bool
	var calledName string
	var calledErr error
	
	// Register callback
	cb.AddOnFailure(func(name string, err error) {
		called = true
		calledName = name
		calledErr = err
	})
	
	// Check callback was registered
	if len(cb.onFailure) != 1 {
		t.Errorf("Should have 1 failure callback, got %d", len(cb.onFailure))
	}
	
	// Test callback invocation
	testErr := errors.New("test error")
	cb.NotifyFailure("TestBreaker", testErr)
	if !called {
		t.Error("Failure callback should have been called")
	}
	if calledName != "TestBreaker" {
		t.Errorf("Callback name should be 'TestBreaker', got '%s'", calledName)
	}
	if calledErr != testErr {
		t.Errorf("Callback error should be 'test error', got '%v'", calledErr)
	}
}

func TestOnRejection(t *testing.T) {
	cb := NewCallbacks()
	
	// Track callback invocations
	var called bool
	var calledName string
	
	// Register callback
	cb.AddOnRejection(func(name string) {
		called = true
		calledName = name
	})
	
	// Check callback was registered
	if len(cb.onRejection) != 1 {
		t.Errorf("Should have 1 rejection callback, got %d", len(cb.onRejection))
	}
	
	// Test callback invocation
	cb.NotifyRejection("TestBreaker")
	if !called {
		t.Error("Rejection callback should have been called")
	}
	if calledName != "TestBreaker" {
		t.Errorf("Callback name should be 'TestBreaker', got '%s'", calledName)
	}
}
