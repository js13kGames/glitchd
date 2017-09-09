package server

import "time"

type Runnable interface {
	// Starts tasks the underlying runnable to start processing.
	Start()
	// Stops the underlying runnable from further processing. If deadline is nil, the intent is a
	// forceful close. Otherwise attempts a graceful close.
	Stop(deadline *time.Time)
}

type Interface interface {
	Runnable
	// GetKind returns what kind of a server interface this is in human readable form.
	GetKind() string
}
