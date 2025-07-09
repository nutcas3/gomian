# Gomian - Go Circuit Breaker Library

A robust and configurable circuit breaker library for Go, designed to enhance the resilience of your applications by preventing cascading failures in distributed systems. Protects your services from over-stressing unhealthy dependencies and provides graceful degradation during outages.

## Table of Contents

1.  [Introduction](https://www.google.com/search?q=%23introduction)
2.  [Core Concepts](https://www.google.com/search?q=%23core-concepts)
      * [States](https://www.google.com/search?q=%23states)
      * [State Transitions](https://www.google.com/search?q=%23state-transitions)
      * [Failure Counting](https://www.google.com/search?q=%23failure-counting)
      * [Concurrency](https://www.google.com/search?q=%23concurrency)
      * [Error Handling](https://www.google.com/search?q=%23error-handling)
      * [Fallback Mechanisms](https://www.google.com/search?q=%23fallback-mechanisms)
3.  [Features](https://www.google.com/search?q=%23features)
4.  [Installation](https://www.google.com/search?q=%23installation)
5.  [Usage](https://www.google.com/search?q=%23usage)
      * [Basic Usage](https://www.google.com/search?q=%23basic-usage)
      * [Configuration](https://www.google.com/search?q=%23configuration)
      * [Monitoring & Callbacks](https://www.google.com/search?q=%23monitoring--callbacks)
      * [Integrating with Context](https://www.google.com/search?q=%23integrating-with-context)
      * [Fallback Example](https://www.google.com/search?q=%23fallback-example)
6.  [Advanced Topics](https://www.google.com/search?q=%23advanced-topics)
      * [Choosing Failure Thresholds](https://www.google.com/search?q=%23choosing-failure-thresholds)
      * [Bulkheading](https://www.google.com/search?q=%23bulkheading)
      * [Distributed Considerations](https://www.google.com/search?q=%23distributed-considerations)
7.  [Contributing](https://www.google.com/search?q=%23contributing)
8.  [License](https://www.google.com/search?q=%23license)

## 1\. Introduction

In modern microservice architectures, an application often relies on numerous external services (databases, APIs, message queues). When one of these dependencies experiences an outage or performance degradation, it can quickly lead to a "cascading failure" where your application becomes overwhelmed trying to reach the unhealthy service, eventually failing itself.

A circuit breaker acts as a protective shield. Inspired by electrical circuit breakers, it monitors calls to a dependency and, if too many failures occur, "opens" the circuit to prevent further calls from being made. This provides the unhealthy service time to recover and prevents your application from consuming valuable resources on futile attempts, allowing it to gracefully degrade or serve cached data.

## 2\. Core Concepts

### States

A circuit breaker typically operates in one of three states:

  * **`Closed`**: The default state. Requests are allowed to pass through to the protected service. The circuit breaker monitors for failures.
  * **`Open`**: When the number of failures exceeds a predefined threshold, the circuit trips to this state. All subsequent requests are immediately short-circuited (blocked) without attempting to reach the protected service. After a configured `timeout` period, it transitions to `Half-Open`.
  * **`Half-Open`**: A transitory state. After the `Open` timeout, a single "test" request is allowed to pass through to the protected service. If this test request succeeds, the circuit transitions back to `Closed`. If it fails, the circuit immediately returns to `Open` for another timeout period.

### State Transitions

The transitions between states are driven by the health of the protected service:

  * **`Closed` to `Open`**: Triggered when the defined `FailureThreshold` (e.g., consecutive failures, or a high failure rate within a rolling window) is met.
  * **`Open` to `Half-Open`**: Occurs automatically after the configured `Timeout` duration has elapsed since the circuit first opened.
  * **`Half-Open` to `Closed`**: If the single "test" request (and potentially subsequent requests, based on `SuccessThreshold`) succeeds.
  * **`Half-Open` to `Open`**: If the single "test" request (or any request while in `Half-Open` and not yet `Closed`) fails.

### Failure Counting

The mechanism for detecting failures is crucial. This library supports:

  * **Consecutive Failures:** The circuit trips after N consecutive failed requests. Simple, but can be overly sensitive to transient issues.
  * **Failure Rate (Rolling Window):** A more robust approach that tracks the percentage of failures over a defined number of requests within a sliding time window. This requires a `MinimumRequestVolume` to avoid tripping on very few requests.

### Concurrency

Given Go's concurrency model, the circuit breaker's internal state (current state, failure counts, timers) is fully protected by `sync.Mutex` to ensure thread-safe operations across multiple goroutines concurrently interacting with the breaker.

### Error Handling

Not all errors should trip the circuit. This library provides mechanisms to:

  * **Define "Failure":** Users explicitly return an error to indicate a failure to the circuit breaker.
  * **Ignored Errors:** Configure specific error types (or provide a predicate function) that should *not* count towards the failure threshold. Common examples include `context.Canceled` or `context.DeadlineExceeded` errors, which often indicate client-side issues rather than server-side problems.

### Fallback Mechanisms

When the circuit is `Open`, requests are short-circuited. The circuit breaker provides a mechanism to execute a "fallback" function instead of calling the protected service. This allows your application to:

  * Return a default or cached value.
  * Serve a pre-defined error message.
  * Redirect to a degraded service.
  * Log the event without impacting the user experience.

## 3\. Features

  * **Configurable Failure Thresholds:**
      * **Consecutive Failures:** Set the number of consecutive failures to trip the circuit.
      * **Failure Rate Threshold:** Define the percentage of failures over a `RollingWindow` to trip the circuit.
      * **Minimum Request Volume:** Specify the minimum number of requests required within a `RollingWindow` before failure rate calculation begins.
  * **Configurable Success Threshold (for Half-Open):** Define how many consecutive successful requests are needed in the `Half-Open` state to transition back to `Closed`.
  * **Configurable Timeout:** Set the duration the circuit remains in the `Open` state before attempting a `Half-Open` test.
  * **Reset Timeout for Closed State:** Optionally reset the internal failure counter after a period of no failures in the `Closed` state.
  * **Ignored Errors:** Specify a list of error types or a custom function to determine which errors should not count towards tripping the circuit.
  * **Event Callbacks/Listeners:** Register functions to be called on state changes (e.g., `OnStateChange`, `OnTrip`, `OnReset`) for logging, metrics, and alerting.
  * **Context-Aware Operations:** Integrates seamlessly with `context.Context` for request cancellation and timeouts.
  * **Metrics Integration Hooks:** Provides hooks to export internal state and counters to your monitoring system (Prometheus, Grafana, etc.).
  * **Graceful Shutdown:** Designed to allow for proper shutdown of internal goroutines and timers.

## 4\. Installation

```bash
go get github.com/nutcase/gomian
```

## 5\. Usage

### Basic Usage

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nutcase/gomian"
)

func callExternalService() error {
	// Simulate calling an external service that might fail
	resp, err := http.Get("http://localhost:8081/health") // Replace with actual service endpoint
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("service returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	// Configure the circuit breaker
	settings := gomian.Settings{
		Name:                "MyServiceBreaker",
		FailureThreshold:     gomian.NewFailureRateThreshold(0.6, 10), // 60% failure rate over 10 requests
		HalfOpenSuccessThreshold: 3,                               // 3 consecutive successes to go from Half-Open to Closed
		Timeout:              5 * time.Second,                 // Stay Open for 5 seconds
		RollingWindow:        10 * time.Second,                // Calculate failure rate over 10 seconds
		MinimumRequestVolume: 5                           // Need at least 5 requests in window to calculate rate
	}

	// Create a new circuit breaker
	breaker := gomian.NewCircuitBreaker(settings)

	// Register a callback for state changes
	breaker.OnStateChange(func(name string, oldState, newState gomian.State) {
		log.Printf("Circuit Breaker '%s' changed state: %s -> %s\n", name, oldState, newState)
	})

	// Simulate repeated calls to the service
	for i := 0; i < 20; i++ {
		// Use the circuit breaker to protect the call
		err := breaker.Execute(func() error {
			log.Println("Attempting to call external service...")
			return callExternalService()
		})

		if err != nil {
			if errors.Is(err, gomian.ErrCircuitOpen) {
				log.Printf("Circuit is OPEN! Request %d rejected.\n", i)
			} else {
				log.Printf("Request %d failed with error: %v\n", i, err)
			}
		} else {
			log.Printf("Request %d succeeded.\n", i)
		}

		time.Sleep(500 * time.Millisecond) // Simulate some delay between requests
	}

	log.Println("Simulation finished.")
}
```

### Configuration

The `gomian.Settings` struct allows fine-grained control:

```go
type Settings struct {
	Name                    string                 // A unique name for this breaker (useful for metrics/logging)
	FailureThreshold        FailureThresholdType   // How to determine when to trip the circuit (e.g., ConsecutiveFailuresThreshold, FailureRateThreshold)
	HalfOpenSuccessThreshold uint64                // Number of consecutive successes required to close from Half-Open
	Timeout                 time.Duration          // Duration the circuit stays Open
	RollingWindow           time.Duration          // Time window for failure rate calculation
	MinimumRequestVolume    uint64                 // Min requests in window before failure rate applies
	ResetTimeout            time.Duration          // (Optional) Reset failure counter in Closed state after this duration of no failures
	IsFailure               func(error) bool       // Custom function to determine if an error counts as a failure
	IgnoredErrors           []error                // List of errors to explicitly ignore (won't count as failures)
}
```

**FailureThresholdType:**

  * `gomian.NewConsecutiveFailuresThreshold(n uint64)`: Trips after `n` consecutive failures.
  * `gomian.NewFailureRateThreshold(rate float64, samples uint64)`: Trips if `rate` (e.g., 0.6 for 60%) is exceeded over `samples` requests within the `RollingWindow`.

### Monitoring & Callbacks

Register functions to react to circuit breaker events:

```go
breaker.OnStateChange(func(name string, oldState, newState gomian.State) {
	log.Printf("Circuit Breaker '%s' changed state: %s -> %s\n", name, oldState, newState)
	// Here you could send metrics to Prometheus, log to an external system, etc.
})

breaker.OnTrip(func(name string, err error) {
	log.Printf("Circuit Breaker '%s' tripped! Reason: %v\n", name, err)
})

breaker.OnReset(func(name string) {
	log.Printf("Circuit Breaker '%s' reset to Closed.\n", name)
})
```

### Integrating with Context

The `Execute` method takes a `context.Context` to allow for cancellation and deadlines to propagate:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

err := breaker.ExecuteContext(ctx, func(ctx context.Context) error {
	// Your service call that respects ctx
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com/api", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err // If context timeout/cancel happens, this error is propagated
	}
	defer resp.Body.Close()
	return nil
})
```

### Fallback Example

Provide a fallback function to handle requests when the circuit is `Open`:

```go
func main() {
    // ... breaker setup ...

    for i := 0; i < 20; i++ {
        var result string
        err := breaker.ExecuteWithFallback(
            func() error {
                log.Println("Attempting to get data from external service...")
                // Simulate service call
                if i % 3 == 0 { // Simulate intermittent failures
                    return errors.New("simulated service error")
                }
                result = "Real Data"
                return nil
            }, 
            func(cbErr error) error {
                log.Printf("Circuit Open or primary call failed. Falling back! Breaker error: %v\n", cbErr)
                // This is your fallback logic
                result = "Cached Data"
                return nil
            }
        )

        if err != nil {
            log.Printf("Request %d failed with unhandled error: %v\n", i, err)
        } else {
            log.Printf("Request %d result: %s\n", i, result)
        }
        time.Sleep(500 * time.Millisecond)
    }
}
```

## 6\. Advanced Topics

### Choosing Failure Thresholds

  * **Consecutive Failures:** Good for services with extremely low tolerance for any failure, but can be too aggressive.
  * **Failure Rate Threshold:** Generally preferred. Requires careful tuning of `rate`, `samples`, and `RollingWindow`. Start with a reasonable `RollingWindow` (e.g., 5-10 seconds) and `MinimumRequestVolume` (e.g., 5-10 requests). Adjust the `rate` (e.g., 50-70%) based on your service's expected error rate. Monitor these values in production.

### Bulkheading

While this library implements the circuit breaker pattern, consider combining it with **bulkheading** strategies. Bulkheading isolates resource pools (e.g., goroutine pools, separate database connections) for different types of dependencies. This ensures that a failing circuit breaker for one service doesn't starve resources needed by other healthy services.

### Distributed Considerations

This circuit breaker operates locally within a single application instance. In a distributed system:

  * **Multiple Instances:** Each instance of your application will have its own circuit breaker for a given dependency. This is often desired, as it allows individual instances to adapt to their local network conditions or resource availability.
  * **Global Circuit Breaking:** For more complex scenarios requiring a "global" view of a service's health across all instances, you might need a centralized control plane or service mesh (like Istio, Linkerd) that can integrate distributed circuit breaking. This library is not designed for global circuit breaking out of the box.

## 7\. Contributing

Contributions are welcome\! Please see `CONTRIBUTING.md` (if applicable) for guidelines.

## 8\. License

This library is released under the MIT License.

## 9\. Project Structure

```
gomian/
├── circuitbreaker.go    # Main circuit breaker implementation
├── settings.go         # Configuration settings
├── state.go            # Circuit breaker state definitions
├── errors.go           # Custom error types
├── callbacks.go        # Event callback system
├── internal/
│   ├── counter/        # Failure/success counting implementations
│   │   ├── counter.go
│   │   └── counter_test.go
│   └── state_machine/  # State transition logic
│       └── state_machine.go
└── examples/
    ├── basic/          # Basic usage example
    │   └── main.go
    └── fallback/       # Example with fallback handling
        └── main.go
```