package types

import "time"

// Scheduler efficiently manages and executes callbacks at specified times.
type Scheduler interface {
	// AddEvent schedules the passed callback to be executed at the supplied time.
	AddEvent(key ObjectIDHash, callback func(), in time.Duration)

	// Contains checks whether an event with the supplied key is scheduled.
	Contains(key ObjectIDHash) bool

	//!TODO: expose the other methods of the current scheduler implementation.
}
