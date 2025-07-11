package gomian

import (
	"testing"
)

func TestStateString(t *testing.T) {
	// Test string representation of states
	tests := []struct {
		state State
		want  string
	}{
		{Closed, "Closed"},
		{Open, "Open"},
		{HalfOpen, "HalfOpen"},
		{State(99), "Unknown State(99)"}, // Invalid state
	}
	
	for _, tt := range tests {
		got := tt.state.String()
		if got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestStateValues(t *testing.T) {
	// Test state values
	if Closed != 0 {
		t.Errorf("Closed should be 0, got %d", Closed)
	}
	
	if Open != 1 {
		t.Errorf("Open should be 1, got %d", Open)
	}
	
	if HalfOpen != 2 {
		t.Errorf("HalfOpen should be 2, got %d", HalfOpen)
	}
}

func TestStateTransitions(t *testing.T) {
	// Test valid state transitions
	validTransitions := map[State][]State{
		Closed:   {Open},
		Open:     {HalfOpen},
		HalfOpen: {Closed, Open},
	}
	
	for from, validTo := range validTransitions {
		for _, to := range validTo {
			if !IsValidTransition(from, to) {
				t.Errorf("Transition from %s to %s should be valid", from, to)
			}
		}
	}
	
	// Test invalid transitions
	invalidTransitions := []struct {
		from State
		to   State
	}{
		{Closed, HalfOpen},
		{Open, Closed},
		{HalfOpen, HalfOpen},
		{Closed, Closed},
		{Open, Open},
	}
	
	for _, tt := range invalidTransitions {
		if IsValidTransition(tt.from, tt.to) {
			t.Errorf("Transition from %s to %s should be invalid", tt.from, tt.to)
		}
	}
}

func TestIsValidTransition(t *testing.T) {
	// Test all possible transitions
	transitions := []struct {
		from  State
		to    State
		valid bool
	}{
		{Closed, Closed, false},
		{Closed, Open, true},
		{Closed, HalfOpen, false},
		{Open, Closed, false},
		{Open, Open, false},
		{Open, HalfOpen, true},
		{HalfOpen, Closed, true},
		{HalfOpen, Open, true},
		{HalfOpen, HalfOpen, false},
	}
	
	for _, tt := range transitions {
		got := IsValidTransition(tt.from, tt.to)
		if got != tt.valid {
			t.Errorf("IsValidTransition(%s, %s) = %v, want %v", 
				tt.from, tt.to, got, tt.valid)
		}
	}
}
