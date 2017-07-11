package noop

import (
	"sync"
	"time"
)

type Tracer struct{}

func (t *Tracer) Trace(name string, fn func(), wg *sync.WaitGroup) error {
	if wg == nil {
		fn()
	} else {
		wg.Add(1)

		go func() {
			defer wg.Done()

			fn()
		}()
	}

	return nil
}

func (t *Tracer) Timing(name string, duration time.Duration) error {
	return nil
}

func (t *Tracer) Count(bucket string, value int) error {
	return nil
}
