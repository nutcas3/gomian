package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nutcase/gomian"
)

// simulateExternalService simulates a call to an external service that might fail
func simulateExternalService() error {
	// Simulate some latency
	time.Sleep(100 * time.Millisecond)

	// Randomly fail 60% of the time
	if rand.Float64() < 0.6 {
		return errors.New("service unavailable")
	}
	return nil
}

func main() {

	// Configure the circuit breaker
	settings := gomian.Settings{
		Name:                "ExampleBreaker",
		FailureThreshold:    gomian.ConsecutiveFailures(3),
		SuccessThreshold:    2,
		Timeout:             5 * time.Second,
		RollingWindow:       10 * time.Second,
		MinimumRequestVolume: 5,
	}

	// Create a new circuit breaker
	breaker := gomian.NewCircuitBreaker(settings)

	// Register callbacks for state changes
	breaker.OnStateChange(func(name string, from, to gomian.State) {
		log.Printf("Circuit '%s' state changed: %s -> %s\n", name, from, to)
	})

	breaker.OnTrip(func(name string, err error) {
		log.Printf("Circuit '%s' tripped! The service appears to be down.\n", name)
	})

	breaker.OnReset(func(name string) {
		log.Printf("Circuit '%s' reset! The service appears to be healthy again.\n", name)
	})

	// Simulate a series of requests
	for i := 1; i <= 20; i++ {
		fmt.Printf("\nRequest %d:\n", i)
		
		// Execute the request through the circuit breaker
		err := breaker.Execute(func() error {
			fmt.Printf("  Calling external service... ")
			err := simulateExternalService()
			if err != nil {
				fmt.Println("FAILED!")
				return err
			}
			fmt.Println("SUCCESS!")
			return nil
		})

		// Handle the result
		if err != nil {
			if gomian.IsCircuitOpen(err) {
				fmt.Println("  Circuit is OPEN! Request was rejected without calling the service.")
			} else {
				fmt.Printf("  Request failed: %v\n", err)
			}
		} else {
			fmt.Println("  Request succeeded!")
		}

		// Display current metrics
		metrics := breaker.GetMetrics()
		fmt.Printf("  Current state: %s, Consecutive failures: %d, Consecutive successes: %d\n", 
			metrics.State, metrics.ConsecutiveFailures, metrics.ConsecutiveSuccesses)

		// Wait a bit before the next request
		time.Sleep(500 * time.Millisecond)
	}
}
