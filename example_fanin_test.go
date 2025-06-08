package conquer_test

import (
	"context"
	"fmt"
	"time"

	"github.com/neox5/conquer"
)

// ExampleFanIn demonstrates how to use conquer.FanIn with multiple sources.
func ExampleFanIn() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fan := conquer.NewFanIn(ctx)

	// Simulated ticker source (animation)
	tick := make(chan any)
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				tick <- fmt.Sprintf("tick at %v", t)
			}
		}
	}()
	fan.Add("anim", tick)

	// Simulated input source
	input := make(chan any)
	go func() {
		time.Sleep(500 * time.Millisecond)
		input <- "keypress: up"
		time.Sleep(400 * time.Millisecond)
		input <- "keypress: down"
	}()
	fan.Add("input", input)

	// Run handler
	fan.Run(func(e conquer.Event) {
		fmt.Printf("[%s] %v\n", e.Source, e.Payload)
	})

	// Wait until context times out and shutdown all sources
	fan.Stop()

	// Output:
	// [anim] tick at ...
	// [input] keypress: up
	// [anim] tick at ...
	// [input] keypress: down
	// ...
}
