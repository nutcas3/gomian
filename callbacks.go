package gomian

// StateChangeCallback is a function that is called when the circuit breaker changes state.
type StateChangeCallback func(name string, from, to State)

// TripCallback is a function that is called when the circuit breaker trips from Closed to Open.
type TripCallback func(name string, err error)

// ResetCallback is a function that is called when the circuit breaker resets from Open/HalfOpen to Closed.
type ResetCallback func(name string)

// SuccessCallback is a function that is called when a request succeeds.
type SuccessCallback func(name string)

// FailureCallback is a function that is called when a request fails.
type FailureCallback func(name string, err error)

// RejectionCallback is a function that is called when a request is rejected due to the circuit being open.
type RejectionCallback func(name string)

// Callbacks holds all the callback functions for a circuit breaker.
type Callbacks struct {
	onStateChange []StateChangeCallback
	onTrip        []TripCallback
	onReset       []ResetCallback
	onSuccess     []SuccessCallback
	onFailure     []FailureCallback
	onRejection   []RejectionCallback
}

// NewCallbacks creates a new Callbacks instance.
func NewCallbacks() *Callbacks {
	return &Callbacks{
		onStateChange: make([]StateChangeCallback, 0),
		onTrip:        make([]TripCallback, 0),
		onReset:       make([]ResetCallback, 0),
		onSuccess:     make([]SuccessCallback, 0),
		onFailure:     make([]FailureCallback, 0),
		onRejection:   make([]RejectionCallback, 0),
	}
}

// AddOnStateChange adds a callback for state changes.
func (c *Callbacks) AddOnStateChange(cb StateChangeCallback) {
	c.onStateChange = append(c.onStateChange, cb)
}

// AddOnTrip adds a callback for when the circuit trips.
func (c *Callbacks) AddOnTrip(cb TripCallback) {
	c.onTrip = append(c.onTrip, cb)
}

// AddOnReset adds a callback for when the circuit resets.
func (c *Callbacks) AddOnReset(cb ResetCallback) {
	c.onReset = append(c.onReset, cb)
}

// AddOnSuccess adds a callback for successful requests.
func (c *Callbacks) AddOnSuccess(cb SuccessCallback) {
	c.onSuccess = append(c.onSuccess, cb)
}

// AddOnFailure adds a callback for failed requests.
func (c *Callbacks) AddOnFailure(cb FailureCallback) {
	c.onFailure = append(c.onFailure, cb)
}

// AddOnRejection adds a callback for rejected requests.
func (c *Callbacks) AddOnRejection(cb RejectionCallback) {
	c.onRejection = append(c.onRejection, cb)
}

// NotifyStateChange notifies all registered state change callbacks.
func (c *Callbacks) NotifyStateChange(name string, from, to State) {
	for _, cb := range c.onStateChange {
		cb(name, from, to)
	}
}

// NotifyTrip notifies all registered trip callbacks.
func (c *Callbacks) NotifyTrip(name string, err error) {
	for _, cb := range c.onTrip {
		cb(name, err)
	}
}

// NotifyReset notifies all registered reset callbacks.
func (c *Callbacks) NotifyReset(name string) {
	for _, cb := range c.onReset {
		cb(name)
	}
}

// NotifySuccess notifies all registered success callbacks.
func (c *Callbacks) NotifySuccess(name string) {
	for _, cb := range c.onSuccess {
		cb(name)
	}
}

// NotifyFailure notifies all registered failure callbacks.
func (c *Callbacks) NotifyFailure(name string, err error) {
	for _, cb := range c.onFailure {
		cb(name, err)
	}
}

// NotifyRejection notifies all registered rejection callbacks.
func (c *Callbacks) NotifyRejection(name string) {
	for _, cb := range c.onRejection {
		cb(name)
	}
}
