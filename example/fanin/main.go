package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neox5/conquer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickInterval := 500 * time.Millisecond
	inputInterval := 2 * time.Second

	// Signal-based shutdown (Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n[signal] shutting down")
		cancel()
	}()

	fan := conquer.NewFanIn(ctx)

	tick := make(chan any)
	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()
		eventNum := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				eventNum++
				tick <- fmt.Sprintf("tick")
			}
		}
	}()
	fan.Add("anim", tick)

	input := make(chan any)
	go func() {
		eventNum := 0
		time.Sleep(inputInterval)
		eventNum++
		input <- "keypress: up"
		time.Sleep(inputInterval)
		eventNum++
		input <- "keypress: down"
	}()
	fan.Add("input", input)

	eventCounts := map[string]int{}
	fan.Run(func(e conquer.Event) {
		eventCounts[e.Source]++
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		eventNum := eventCounts[e.Source]
		var interval string
		switch e.Source {
		case "anim":
			interval = tickInterval.String()
		case "input":
			interval = inputInterval.String()
		default:
			interval = "unknown"
		}
		fmt.Printf("%s | %-20s | %v\n", timestamp, fmt.Sprintf("%s_#%02d [%s]", e.Source, eventNum, interval), e.Payload)
	})

	<-ctx.Done()
	fan.Stop()

	totalEvents := 0
	for _, count := range eventCounts {
		totalEvents += count
	}
	duration := time.Since(time.Now().Add(-tickInterval * time.Duration(totalEvents))).Truncate(time.Millisecond)
	fmt.Printf("\n[summary] runtime: %v | total events: %d\n", duration, totalEvents)
}
