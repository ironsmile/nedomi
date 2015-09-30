package types

import "time"

// Scheduler efficiently manages and executes callbacks at specified times.
type Scheduler interface {
	// AddEvent schedules the passed callback to be executed at the supplied time.
	AddEvent(key string, callback func(), in time.Duration)

	// Contains checks whether an event with the supplied key is scheduled.
	Contains(key string) bool

	//!TODO: expose the other methods of the current scheduler implementation.
}
