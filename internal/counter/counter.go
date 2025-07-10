package counter

import (
	"sync"
	"time"
)

// RollingWindow represents a rolling window counter for tracking events over time.
type RollingWindow struct {
	mu            sync.Mutex
	buckets       []bucket
	bucketSize    time.Duration
	numBuckets    int
	windowSize    time.Duration
	lastRotation  time.Time
	totalRequests uint64
	totalFailures uint64
}

// bucket represents a time bucket in the rolling window.
type bucket struct {
	requests uint64
	failures uint64
}

// NewRollingWindow creates a new RollingWindow with the specified window size and number of buckets.
func NewRollingWindow(windowSize time.Duration, numBuckets int) *RollingWindow {
	if numBuckets <= 0 {
		numBuckets = 10 // Default to 10 buckets
	}
	
	bucketSize := windowSize / time.Duration(numBuckets)
	if bucketSize < time.Millisecond {
		bucketSize = time.Millisecond
	}
	
	buckets := make([]bucket, numBuckets)
	
	return &RollingWindow{
		buckets:      buckets,
		bucketSize:   bucketSize,
		numBuckets:   numBuckets,
		windowSize:   windowSize,
		lastRotation: time.Now(),
	}
}

// rotate rotates the buckets if necessary based on the current time.
func (rw *RollingWindow) rotate() {
	now := time.Now()
	elapsed := now.Sub(rw.lastRotation)
	
	if elapsed < rw.bucketSize {
		return
	}
	
	// Calculate how many buckets to rotate
	bucketsToRotate := int(elapsed / rw.bucketSize)
	if bucketsToRotate > rw.numBuckets {
		bucketsToRotate = rw.numBuckets
	}
	
	// Rotate the buckets
	for i := 0; i < bucketsToRotate; i++ {
		// Remove the oldest bucket's counts from the totals
		oldestBucket := (i + 1) % rw.numBuckets
		rw.totalRequests -= rw.buckets[oldestBucket].requests
		rw.totalFailures -= rw.buckets[oldestBucket].failures
		
		// Reset the bucket
		rw.buckets[oldestBucket].requests = 0
		rw.buckets[oldestBucket].failures = 0
	}
	
	// Update the last rotation time
	rw.lastRotation = now.Add(-elapsed % rw.bucketSize)
}

// IncrementSuccess increments the success counter.
func (rw *RollingWindow) IncrementSuccess() {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	
	rw.rotate()
	
	currentBucket := 0 // Always use the current bucket (index 0)
	rw.buckets[currentBucket].requests++
	rw.totalRequests++
}

// IncrementFailure increments the failure counter.
func (rw *RollingWindow) IncrementFailure() {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	
	rw.rotate()
	
	currentBucket := 0 // Always use the current bucket (index 0)
	rw.buckets[currentBucket].requests++
	rw.buckets[currentBucket].failures++
	rw.totalRequests++
	rw.totalFailures++
}

// Counts returns the total number of requests and failures in the window.
func (rw *RollingWindow) Counts() (requests, failures uint64) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	
	rw.rotate()
	
	return rw.totalRequests, rw.totalFailures
}

// Reset resets all counters to zero.
func (rw *RollingWindow) Reset() {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	
	for i := range rw.buckets {
		rw.buckets[i].requests = 0
		rw.buckets[i].failures = 0
	}
	
	rw.totalRequests = 0
	rw.totalFailures = 0
	rw.lastRotation = time.Now()
}

// ConsecutiveCounter tracks consecutive successes or failures.
type ConsecutiveCounter struct {
	mu                 sync.Mutex
	consecutiveSuccess uint64
	consecutiveFailure uint64
	totalSuccess       uint64
	totalFailure       uint64
}

// NewConsecutiveCounter creates a new ConsecutiveCounter.
func NewConsecutiveCounter() *ConsecutiveCounter {
	return &ConsecutiveCounter{}
}

// IncrementSuccess increments the success counter and resets the failure counter.
func (cc *ConsecutiveCounter) IncrementSuccess() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	cc.consecutiveSuccess++
	cc.consecutiveFailure = 0
	cc.totalSuccess++
}

// IncrementFailure increments the failure counter and resets the success counter.
func (cc *ConsecutiveCounter) IncrementFailure() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	cc.consecutiveFailure++
	cc.consecutiveSuccess = 0
	cc.totalFailure++
}

// ConsecutiveSuccesses returns the number of consecutive successes.
func (cc *ConsecutiveCounter) ConsecutiveSuccesses() uint64 {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	return cc.consecutiveSuccess
}

// ConsecutiveFailures returns the number of consecutive failures.
func (cc *ConsecutiveCounter) ConsecutiveFailures() uint64 {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	return cc.consecutiveFailure
}

// Totals returns the total number of successes and failures.
func (cc *ConsecutiveCounter) Totals() (successes, failures uint64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	return cc.totalSuccess, cc.totalFailure
}

// Reset resets all counters to zero.
func (cc *ConsecutiveCounter) Reset() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	cc.consecutiveSuccess = 0
	cc.consecutiveFailure = 0
	cc.totalSuccess = 0
	cc.totalFailure = 0
}
