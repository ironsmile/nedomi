package types

import "time"

// Scheduler efficiently manages and executes callbacks at specified times.
type Scheduler interface {
	// AddEvent schedules the passed callback to be executed at the supplied time.
	AddEvent(key ObjectIDHash, callback ScheduledCallback, in time.Duration)

	// Contains checks whether an event with the supplied key is scheduled.
	Contains(key ObjectIDHash) bool

	// ChangeConfig change configs of the scheduler and start using them
	ChangeConfig(Logger)
}

// ScheduledCallback is the type of the function that Scheduler will callback when scheduled
type ScheduledCallback func(Logger)
