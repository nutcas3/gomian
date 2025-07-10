package gomian

import (
	"context"
	"sync"
	"time"

	"github.com/nutcase/gomian/internal/counter"
	"github.com/nutcase/gomian/internal/state_machine"
)

// convertState converts a state_machine.State to a gomian.State
func convertState(state state_machine.State) State {
	switch state {
	case state_machine.Open:
		return Open
	case state_machine.HalfOpen:
		return HalfOpen
	case state_machine.Closed:
		return Closed
	default:
		return Closed
	}
}

// CircuitBreaker is the main struct that implements the circuit breaker pattern.
type CircuitBreaker struct {
	name           string
	settings       Settings
	stateMachine   *state_machine.StateMachine
	rollingWindow  *counter.RollingWindow
	consecutiveCounter *counter.ConsecutiveCounter
	callbacks      *Callbacks
	mu             sync.Mutex
	timer          *time.Timer
	timerMu        sync.Mutex
	resetTimer     *time.Timer
	resetTimerMu   sync.Mutex
}

// Metrics represents the current metrics of a circuit breaker.
type Metrics struct {
	Name                string
	State               State
	TotalRequests       uint64
	TotalFailures       uint64
	ConsecutiveFailures uint64
	ConsecutiveSuccesses uint64
	LastStateChange     time.Time
	TimeInState         time.Duration
}

// NewCircuitBreaker creates a new CircuitBreaker with the provided settings.
func NewCircuitBreaker(settings Settings) *CircuitBreaker {
	if settings.Name == "" {
		settings.Name = "default"
	}

	cb := &CircuitBreaker{
		name:     settings.Name,
		settings: settings,
		callbacks: NewCallbacks(),
		consecutiveCounter: counter.NewConsecutiveCounter(),
	}

	// Initialize the rolling window if needed
	if _, ok := settings.FailureThreshold.(FailureRateThreshold); ok {
		cb.rollingWindow = counter.NewRollingWindow(settings.RollingWindow, 10)
	}

	// Initialize the state machine
	cb.stateMachine = state_machine.NewStateMachine(func(from, to state_machine.State) {
		// Convert state_machine.State to gomian.State
		fromState := convertState(from)
		toState := convertState(to)
		cb.callbacks.NotifyStateChange(cb.name, fromState, toState)
		
		// Handle specific state transitions
		if from == state_machine.Closed && to == state_machine.Open {
			cb.callbacks.NotifyTrip(cb.name, nil)
		} else if (from == state_machine.Open || from == state_machine.HalfOpen) && to == state_machine.Closed {
			cb.callbacks.NotifyReset(cb.name)
		}
		
		// Set up timers based on state
		if to == state_machine.Open {
			cb.startOpenStateTimer()
		} else if to == state_machine.Closed && cb.settings.ResetTimeout > 0 {
			cb.startResetTimer()
		}
	})

	// Start the reset timer if configured
	if cb.settings.ResetTimeout > 0 {
		cb.startResetTimer()
	}

	return cb
}

// startOpenStateTimer starts a timer that will transition the circuit from Open to HalfOpen
// after the configured timeout period.
func (cb *CircuitBreaker) startOpenStateTimer() {
	cb.timerMu.Lock()
	defer cb.timerMu.Unlock()

	// Cancel any existing timer
	if cb.timer != nil {
		cb.timer.Stop()
	}

	cb.timer = time.AfterFunc(cb.settings.Timeout, func() {
		cb.stateMachine.TransitionToHalfOpen()
	})
}

// startResetTimer starts a timer that will reset the failure counters if no failures
// occur within the configured reset timeout period.
func (cb *CircuitBreaker) startResetTimer() {
	cb.resetTimerMu.Lock()
	defer cb.resetTimerMu.Unlock()

	// Cancel any existing timer
	if cb.resetTimer != nil {
		cb.resetTimer.Stop()
	}

	cb.resetTimer = time.AfterFunc(cb.settings.ResetTimeout, func() {
		cb.mu.Lock()
		defer cb.mu.Unlock()

		// Only reset if we're still in the Closed state
		if cb.stateMachine.IsClosed() {
			cb.consecutiveCounter.Reset()
			if cb.rollingWindow != nil {
				cb.rollingWindow.Reset()
			}
		}
	})
}

// Execute executes the given function if the circuit is closed or half-open.
// If the circuit is open, it returns ErrCircuitOpen without executing the function.
func (cb *CircuitBreaker) Execute(op func() error) error {
	return cb.ExecuteContext(context.Background(), func(ctx context.Context) error {
		return op()
	})
}

// ExecuteContext executes the given function with context if the circuit is closed or half-open.
// If the circuit is open, it returns ErrCircuitOpen without executing the function.
func (cb *CircuitBreaker) ExecuteContext(ctx context.Context, op func(context.Context) error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	state := cb.stateMachine.State()

	// If the circuit is open, reject the request
	if state == state_machine.Open {
		cb.callbacks.NotifyRejection(cb.name)
		return ErrCircuitOpen
	}

	// If the circuit is half-open, only allow one request at a time
	if state == state_machine.HalfOpen {
		cb.mu.Lock()
		defer cb.mu.Unlock()
	}

	// Execute the operation
	err := op(ctx)

	// Record the result
	if err != nil {
		if cb.isFailure(err) {
			cb.recordFailure(err)
		}
		return err
	}

	cb.recordSuccess()
	return nil
}

// ExecuteWithFallback executes the given function if the circuit is closed or half-open.
// If the circuit is open or if the function fails, it executes the fallback function.
func (cb *CircuitBreaker) ExecuteWithFallback(op func() error, fallback func(error) error) error {
	err := cb.Execute(op)
	if err != nil {
		return fallback(err)
	}
	return nil
}

// ExecuteWithFallbackContext executes the given function with context if the circuit is closed or half-open.
// If the circuit is open or if the function fails, it executes the fallback function.
func (cb *CircuitBreaker) ExecuteWithFallbackContext(ctx context.Context, op func(context.Context) error, fallback func(context.Context, error) error) error {
	err := cb.ExecuteContext(ctx, op)
	if err != nil {
		return fallback(ctx, err)
	}
	return nil
}

// isFailure determines if an error should be considered a failure.
func (cb *CircuitBreaker) isFailure(err error) bool {
	// If a custom IsFailure function is provided, use it
	if cb.settings.IsFailure != nil {
		return cb.settings.IsFailure(err)
	}

	// Check if the error is in the ignored errors list
	for _, ignoredErr := range cb.settings.IgnoredErrors {
		if err == ignoredErr {
			return false
		}
	}

	// By default, any non-nil error is a failure
	return err != nil
}

// recordSuccess records a successful request and updates the circuit state if necessary.
func (cb *CircuitBreaker) recordSuccess() {
	cb.callbacks.NotifySuccess(cb.name)

	// Update counters
	cb.consecutiveCounter.IncrementSuccess()
	if cb.rollingWindow != nil {
		cb.rollingWindow.IncrementSuccess()
	}

	// If we're in the half-open state and have reached the success threshold,
	// transition to closed
	if cb.stateMachine.IsHalfOpen() && 
	   cb.consecutiveCounter.ConsecutiveSuccesses() >= cb.settings.SuccessThreshold {
		cb.stateMachine.TransitionToClosed()
		
		// Reset counters
		cb.consecutiveCounter.Reset()
		if cb.rollingWindow != nil {
			cb.rollingWindow.Reset()
		}
		
		// Start the reset timer if configured
		if cb.settings.ResetTimeout > 0 {
			cb.startResetTimer()
		}
	}
}

// recordFailure records a failed request and updates the circuit state if necessary.
func (cb *CircuitBreaker) recordFailure(err error) {
	cb.callbacks.NotifyFailure(cb.name, err)

	// Update counters
	cb.consecutiveCounter.IncrementFailure()
	if cb.rollingWindow != nil {
		cb.rollingWindow.IncrementFailure()
	}

	// If we're in the half-open state, any failure should trip the circuit
	if cb.stateMachine.IsHalfOpen() {
		cb.stateMachine.TransitionToOpen()
		return
	}

	// If we're in the closed state, check if we should trip the circuit
	if cb.stateMachine.IsClosed() {
		shouldTrip := false

		// Check if we should trip based on the failure threshold type
		switch threshold := cb.settings.FailureThreshold.(type) {
		case ConsecutiveFailuresThreshold:
			shouldTrip = cb.consecutiveCounter.ConsecutiveFailures() >= threshold.Threshold
		case FailureRateThreshold:
			if cb.rollingWindow != nil {
				requests, failures := cb.rollingWindow.Counts()
				if requests >= cb.settings.MinimumRequestVolume {
					shouldTrip = threshold.ShouldTrip(failures, 0, requests, cb.settings.RollingWindow)
				}
			}
		}

		if shouldTrip {
			cb.stateMachine.TransitionToOpen()
			cb.callbacks.NotifyTrip(cb.name, err)
		}
	}
}

// OnStateChange registers a callback for state changes.
func (cb *CircuitBreaker) OnStateChange(callback StateChangeCallback) {
	cb.callbacks.AddOnStateChange(callback)
}

// OnTrip registers a callback for when the circuit trips.
func (cb *CircuitBreaker) OnTrip(callback TripCallback) {
	cb.callbacks.AddOnTrip(callback)
}

// OnReset registers a callback for when the circuit resets.
func (cb *CircuitBreaker) OnReset(callback ResetCallback) {
	cb.callbacks.AddOnReset(callback)
}

// OnSuccess registers a callback for successful requests.
func (cb *CircuitBreaker) OnSuccess(callback SuccessCallback) {
	cb.callbacks.AddOnSuccess(callback)
}

// OnFailure registers a callback for failed requests.
func (cb *CircuitBreaker) OnFailure(callback FailureCallback) {
	cb.callbacks.AddOnFailure(callback)
}

// OnRejection registers a callback for rejected requests.
func (cb *CircuitBreaker) OnRejection(callback RejectionCallback) {
	cb.callbacks.AddOnRejection(callback)
}

// Name returns the name of the circuit breaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	return convertState(cb.stateMachine.State())
}

// GetMetrics returns the current metrics of the circuit breaker.
func (cb *CircuitBreaker) GetMetrics() Metrics {
	var totalRequests, totalFailures uint64
	
	if cb.rollingWindow != nil {
		totalRequests, totalFailures = cb.rollingWindow.Counts()
	} else {
		totalRequests, totalFailures = cb.consecutiveCounter.Totals()
	}
	
	return Metrics{
		Name:                cb.name,
		State:               convertState(cb.stateMachine.State()),
		TotalRequests:       totalRequests,
		TotalFailures:       totalFailures,
		ConsecutiveFailures: cb.consecutiveCounter.ConsecutiveFailures(),
		ConsecutiveSuccesses: cb.consecutiveCounter.ConsecutiveSuccesses(),
		LastStateChange:     cb.stateMachine.LastStateChange(),
		TimeInState:         cb.stateMachine.TimeInState(),
	}
}

// Close stops all timers and releases resources.
func (cb *CircuitBreaker) Close() {
	cb.timerMu.Lock()
	if cb.timer != nil {
		cb.timer.Stop()
		cb.timer = nil
	}
	cb.timerMu.Unlock()

	cb.resetTimerMu.Lock()
	if cb.resetTimer != nil {
		cb.resetTimer.Stop()
		cb.resetTimer = nil
	}
	cb.resetTimerMu.Unlock()
}
