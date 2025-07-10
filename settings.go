package gomian

import (
	"time"
)

// FailureThresholdType is an interface for different failure threshold strategies.
type FailureThresholdType interface {
	// ShouldTrip evaluates whether the circuit should trip based on the current failure data.
	ShouldTrip(failures, successes, total uint64, window time.Duration) bool
	// String returns a string representation of the threshold type.
	String() string
}

// ConsecutiveFailuresThreshold represents a threshold based on consecutive failures.
type ConsecutiveFailuresThreshold struct {
	Threshold uint64
}

// ShouldTrip returns true if the number of consecutive failures exceeds the threshold.
func (c ConsecutiveFailuresThreshold) ShouldTrip(failures, _, _ uint64, _ time.Duration) bool {
	return failures >= c.Threshold
}

// String returns a string representation of the ConsecutiveFailuresThreshold.
func (c ConsecutiveFailuresThreshold) String() string {
	return "ConsecutiveFailures"
}

// ConsecutiveFailures creates a new ConsecutiveFailuresThreshold.
func ConsecutiveFailures(threshold uint64) FailureThresholdType {
	return ConsecutiveFailuresThreshold{Threshold: threshold}
}

// FailureRateThreshold represents a threshold based on the failure rate within a rolling window.
type FailureRateThreshold struct {
	Rate    float64
	Samples uint64
}

// ShouldTrip returns true if the failure rate exceeds the threshold and the minimum sample count is met.
func (f FailureRateThreshold) ShouldTrip(failures, _, total uint64, _ time.Duration) bool {
	if total < f.Samples {
		return false
	}
	return float64(failures)/float64(total) >= f.Rate
}

// String returns a string representation of the FailureRateThreshold.
func (f FailureRateThreshold) String() string {
	return "FailureRate"
}

// NewFailureRateThreshold creates a new FailureRateThreshold.
func NewFailureRateThreshold(rate float64, samples uint64) FailureThresholdType {
	return FailureRateThreshold{Rate: rate, Samples: samples}
}

// Settings defines the configuration for a CircuitBreaker.
type Settings struct {
	// Name is a unique identifier for this circuit breaker.
	Name string

	// FailureThreshold defines the strategy for determining when to trip the circuit.
	FailureThreshold FailureThresholdType

	// SuccessThreshold is the number of consecutive successes required to close from Half-Open.
	SuccessThreshold uint64

	// Timeout is the duration the circuit stays Open before transitioning to Half-Open.
	Timeout time.Duration

	// RollingWindow is the time window for failure rate calculation.
	RollingWindow time.Duration

	// MinimumRequestVolume is the minimum number of requests required within the RollingWindow
	// before the failure rate calculation is applied.
	MinimumRequestVolume uint64

	// ResetTimeout is the duration after which the failure counter is reset in the Closed state
	// if no failures occur during that period.
	ResetTimeout time.Duration

	// IsFailure is a custom function to determine if an error counts as a failure.
	// If nil, any non-nil error is considered a failure.
	IsFailure func(error) bool

	// IgnoredErrors is a list of errors that should not count as failures.
	IgnoredErrors []error
}

// DefaultSettings returns a Settings struct with sensible default values.
func DefaultSettings() Settings {
	return Settings{
		Name:                "default",
		FailureThreshold:    ConsecutiveFailures(5),
		SuccessThreshold:    1,
		Timeout:             60 * time.Second,
		RollingWindow:       10 * time.Second,
		MinimumRequestVolume: 3,
		ResetTimeout:        0, // Disabled by default
		IsFailure:           nil, // Any non-nil error is a failure
		IgnoredErrors:       nil,
	}
}
