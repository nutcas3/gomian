package gomian

import (
	"fmt"
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

// String returns a string representation of the State.
func (s State) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return fmt.Sprintf("Unknown State(%d)", s)
	}
}

// IsValidTransition checks if a transition from one state to another is valid.
func IsValidTransition(from, to State) bool {
	switch from {
	case Closed:
		return to == Open
	case Open:
		return to == HalfOpen
	case HalfOpen:
		return to == Closed || to == Open
	default:
		return false
	}
}
