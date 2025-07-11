package state_machine

import (
	"testing"
	"time"
)

func TestNewStateMachine(t *testing.T) {
	// Test creating a new state machine
	callback := func(from, to State) {

	}
	
	sm := NewStateMachine(callback)
	

	if sm.State() != Closed {
		t.Errorf("Initial state should be Closed, got %v", sm.State())
	}
}

func TestStateTransitions(t *testing.T) {
	var transitions []struct {
		from State
		to   State
	}
	
	callback := func(from, to State) {
		transitions = append(transitions, struct {
			from State
			to   State
		}{from, to})
	}
	
	sm := NewStateMachine(callback)
	
	// Test Closed to Open transition
	sm.TransitionToOpen()
	if sm.State() != Open {
		t.Errorf("State should be Open, got %v", sm.State())
	}
	if len(transitions) != 1 || transitions[0].from != Closed || transitions[0].to != Open {
		t.Errorf("Transition should be from Closed to Open, got %v to %v", 
			transitions[0].from, transitions[0].to)
	}
	
	sm.TransitionToHalfOpen()
	if sm.State() != HalfOpen {
		t.Errorf("State should be HalfOpen, got %v", sm.State())
	}
	if len(transitions) != 2 || transitions[1].from != Open || transitions[1].to != HalfOpen {
		t.Errorf("Transition should be from Open to HalfOpen, got %v to %v", 
			transitions[1].from, transitions[1].to)
	}
	
	sm.TransitionToClosed()
	if sm.State() != Closed {
		t.Errorf("State should be Closed, got %v", sm.State())
	}
	if len(transitions) != 3 || transitions[2].from != HalfOpen || transitions[2].to != Closed {
		t.Errorf("Transition should be from HalfOpen to Closed, got %v to %v", 
			transitions[2].from, transitions[2].to)
	}
	
	// Test HalfOpen to Open transition
	sm = NewStateMachine(callback)
	transitions = nil
	sm.TransitionToHalfOpen() // First go to HalfOpen
	sm.TransitionToOpen()     // Then to Open
	if sm.State() != Open {
		t.Errorf("State should be Open, got %v", sm.State())
	}
	if len(transitions) != 2 || transitions[1].from != HalfOpen || transitions[1].to != Open {
		t.Errorf("Transition should be from HalfOpen to Open, got %v to %v", 
			transitions[1].from, transitions[1].to)
	}
}

func TestNoTransitionForSameState(t *testing.T) {
	callbackCalled := false
	callback := func(from, to State) {
		callbackCalled = true
	}
	
	sm := NewStateMachine(callback)
	
	// Test no transition for Closed to Closed
	callbackCalled = false
	sm.TransitionToClosed()
	if callbackCalled {
		t.Error("Callback should not be called for transition to same state")
	}
	
	// Test no transition for Open to Open
	sm.TransitionToOpen()
	callbackCalled = false
	sm.TransitionToOpen()
	if callbackCalled {
		t.Error("Callback should not be called for transition to same state")
	}
	
	// Test no transition for HalfOpen to HalfOpen
	sm.TransitionToHalfOpen()
	callbackCalled = false
	sm.TransitionToHalfOpen()
	if callbackCalled {
		t.Error("Callback should not be called for transition to same state")
	}
}

func TestStateCheckers(t *testing.T) {
	sm := NewStateMachine(nil)
	
	// Test Closed state
	if !sm.IsClosed() {
		t.Error("IsClosed should return true for Closed state")
	}
	if sm.IsOpen() {
		t.Error("IsOpen should return false for Closed state")
	}
	if sm.IsHalfOpen() {
		t.Error("IsHalfOpen should return false for Closed state")
	}
	
	// Test Open state
	sm.TransitionToOpen()
	if sm.IsClosed() {
		t.Error("IsClosed should return false for Open state")
	}
	if !sm.IsOpen() {
		t.Error("IsOpen should return true for Open state")
	}
	if sm.IsHalfOpen() {
		t.Error("IsHalfOpen should return false for Open state")
	}
	
	// Test HalfOpen state
	sm.TransitionToHalfOpen()
	if sm.IsClosed() {
		t.Error("IsClosed should return false for HalfOpen state")
	}
	if sm.IsOpen() {
		t.Error("IsOpen should return false for HalfOpen state")
	}
	if !sm.IsHalfOpen() {
		t.Error("IsHalfOpen should return true for HalfOpen state")
	}
}

func TestLastStateChange(t *testing.T) {
	sm := NewStateMachine(nil)
	initialTime := sm.LastStateChange()
	
	time.Sleep(10 * time.Millisecond)
	
	// Transition to Open
	sm.TransitionToOpen()
	openTime := sm.LastStateChange()
	
	// Check that time was updated
	if !openTime.After(initialTime) {
		t.Error("LastStateChange should be updated after state transition")
	}
	
	// Sleep again
	time.Sleep(10 * time.Millisecond)
	
	// Transition to HalfOpen
	sm.TransitionToHalfOpen()
	halfOpenTime := sm.LastStateChange()
	
	// Check that time was updated again
	if !halfOpenTime.After(openTime) {
		t.Error("LastStateChange should be updated after state transition")
	}
}

func TestTimeInState(t *testing.T) {
	sm := NewStateMachine(nil)
	
	// Sleep a bit
	time.Sleep(10 * time.Millisecond)
	
	// Check that time in state is positive
	if sm.TimeInState() <= 0 {
		t.Error("TimeInState should be positive")
	}
}
