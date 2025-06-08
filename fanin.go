package conquer

import (
	"context"
	"sync"
)

// Event represents a generic message coming from any source.
type Event struct {
	Source  string
	Payload any
}

// FanIn coordinates multiple asynchronous event sources into a single handler.
type FanIn struct {
	ctx     context.Context
	cancel  context.CancelFunc
	sources []<-chan Event
	out     chan Event
	wg      sync.WaitGroup
	mu      sync.Mutex
}

// NewFanIn creates a new FanIn with its own cancelable context.
func NewFanIn(parent context.Context) *FanIn {
	ctx, cancel := context.WithCancel(parent)
	return &FanIn{
		ctx:    ctx,
		cancel: cancel,
		out:    make(chan Event),
	}
}

// Add adds a new source channel to the fan-in. Safe to call before Run().
func (f *FanIn) Add(source string, ch <-chan any) {
	wrapped := make(chan Event)
	f.mu.Lock()
	f.sources = append(f.sources, wrapped)
	f.mu.Unlock()

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			select {
			case <-f.ctx.Done():
				return
			case val, ok := <-ch:
				if !ok {
					return
				}
				wrapped <- Event{Source: source, Payload: val}
			}
		}
	}()
}

// Run starts the fan-in and delivers all events to the provided handler.
func (f *FanIn) Run(handler func(Event)) {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for _, src := range f.sources {
			f.wg.Add(1)
			go func(ch <-chan Event) {
				defer f.wg.Done()
				for {
					select {
					case <-f.ctx.Done():
						return
					case ev, ok := <-ch:
						if !ok {
							return
						}
						handler(ev)
					}
				}
			}(src)
		}
	}()
}

// Stop cancels the fan-in and waits for all goroutines to exit.
func (f *FanIn) Stop() {
	f.cancel()
	f.wg.Wait()
}
