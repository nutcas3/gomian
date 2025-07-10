package state_machine

import (
	"sync"
	"time"
)

// State represents the possible states of a circuit breaker.
type State int

const (
	// Closed is the default state where requests are allowed to pass through.
	Closed State = iota

	// Open is the state where all requests are rejected without attempting to reach the service.
	Open

	// HalfOpen is the state where a limited number of test requests are allowed to pass through.
	HalfOpen
)

// StateMachine manages the state transitions of a circuit breaker.
type StateMachine struct {
	mu             sync.Mutex
	state          State
	lastStateChange time.Time
	onStateChange  func(from, to State)
}

// NewStateMachine creates a new StateMachine with the initial state set to Closed.
func NewStateMachine(onStateChange func(from, to State)) *StateMachine {
	return &StateMachine{
		state:          Closed,
		lastStateChange: time.Now(),
		onStateChange:  onStateChange,
	}
}

// State returns the current state of the circuit breaker.
func (sm *StateMachine) State() State {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}

// LastStateChange returns the time of the last state change.
func (sm *StateMachine) LastStateChange() time.Time {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.lastStateChange
}

// TransitionToOpen transitions the circuit breaker to the Open state.
func (sm *StateMachine) TransitionToOpen() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.state == Open {
		return
	}

	oldState := sm.state
	sm.state = Open
	sm.lastStateChange = time.Now()

	if sm.onStateChange != nil {
		sm.onStateChange(oldState, Open)
	}
}

// TransitionToHalfOpen transitions the circuit breaker to the HalfOpen state.
func (sm *StateMachine) TransitionToHalfOpen() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.state == HalfOpen {
		return
	}

	oldState := sm.state
	sm.state = HalfOpen
	sm.lastStateChange = time.Now()

	if sm.onStateChange != nil {
		sm.onStateChange(oldState, HalfOpen)
	}
}

// TransitionToClosed transitions the circuit breaker to the Closed state.
func (sm *StateMachine) TransitionToClosed() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.state == Closed {
		return
	}

	oldState := sm.state
	sm.state = Closed
	sm.lastStateChange = time.Now()

	if sm.onStateChange != nil {
		sm.onStateChange(oldState, Closed)
	}
}

// IsOpen returns true if the circuit breaker is in the Open state.
func (sm *StateMachine) IsOpen() bool {
	return sm.State() == Open
}

// IsHalfOpen returns true if the circuit breaker is in the HalfOpen state.
func (sm *StateMachine) IsHalfOpen() bool {
	return sm.State() == HalfOpen
}

// IsClosed returns true if the circuit breaker is in the Closed state.
func (sm *StateMachine) IsClosed() bool {
	return sm.State() == Closed
}

// TimeInState returns the duration since the last state change.
func (sm *StateMachine) TimeInState() time.Duration {
	return time.Since(sm.LastStateChange())
}
