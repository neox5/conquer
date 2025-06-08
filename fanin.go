package conquer

import (
	"context"
	"sync"
)

type FanIn struct {
	ctx     context.Context
	cancel  context.CancelFunc
	sources []<-chan any
	stream  chan any
	started bool
	wg      sync.WaitGroup
	mu      sync.Mutex
}

func NewFanIn(parent context.Context) *FanIn {
	ctx, cancel := context.WithCancel(parent)
	return &FanIn{
		ctx:    ctx,
		cancel: cancel,
		stream: make(chan any),
	}
}

func (f *FanIn) Add(ch <-chan any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.started {
		panic("cannot add sources after Start() has been called")
	}

	f.sources = append(f.sources, ch)
}

func (f *FanIn) Stream() <-chan any {
	return f.stream
}

func (f *FanIn) Start() {
	f.mu.Lock()
	if f.started {
		f.mu.Unlock()
		panic("Start() can only be called once")
	}
	f.started = true
	sources := f.sources
	f.mu.Unlock()

	for _, src := range sources {
		f.wg.Add(1)
		go func(ch <-chan any) {
			defer f.wg.Done()
			for {
				select {
				case <-f.ctx.Done():
					return
				case val, ok := <-ch:
					if !ok {
						return
					}
					select {
					case f.stream <- val:
					case <-f.ctx.Done():
						return
					}
				}
			}
		}(src)
	}
}

func (f *FanIn) Stop() {
	f.cancel()
	f.wg.Wait()
	close(f.stream)
}
