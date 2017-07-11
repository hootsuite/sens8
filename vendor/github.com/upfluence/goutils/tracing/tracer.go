package tracing

import (
	"sync"
	"time"
)

type Tracer interface {
	// if the third parameter is nil the closure will be executed synchronous
	// otherwise asynchronous
	Trace(string, func(), *sync.WaitGroup) error

	Count(string, int) error
	Timing(string, time.Duration) error
}
