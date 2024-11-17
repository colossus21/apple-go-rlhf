//go:build 2a
// +build 2a

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// State represents the different states of the system
type State int

const (
	StateIdle State = iota
	StateRunning
	StateStopped
)

// System encapsulates the state and state transitions
type System struct {
	state     State
	eventChan chan string
	quit      chan struct{}
	mu        sync.Mutex
}

func (s *System) getState() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Transition to the next state based on the current state and event
func (s *System) transition(event string) {
	switch s.state {
	case StateIdle:
		switch event {
		case "start":
			s.state = StateRunning
		default:
			fmt.Println("Invalid event in Idle state")
		}
	case StateRunning:
		switch event {
		case "stop":
			s.state = StateStopped
		default:
			fmt.Println("Invalid event in Running state")
		}
	case StateStopped:
		switch event {
		case "start":
			s.state = StateRunning
		default:
			fmt.Println("Invalid event in Stopped state")
		}
	}
}

func (s *System) handleEvents() {
	for {
		select {
		case event := <-s.eventChan:
			s.transition(event)
			fmt.Println("Current State:", s.getState())
			// Simulate different work durations based on the state
			switch s.getState() {
			case StateRunning:
				// Slow transition scenario in Running state
				s.simulateWork(time.Duration(rand.Intn(2)+1) * time.Second)
			default:
				// Fast transition scenario in other states
				s.simulateWork(10 * time.Millisecond)
			}
		case <-s.quit:
			return
		}
	}
}

func (s *System) simulateWork(duration time.Duration) {
	fmt.Println("System Working...")
	time.Sleep(duration)
}

func main() {
	system := &System{
		state:     StateIdle,
		eventChan: make(chan string),
		quit:      make(chan struct{}),
	}

	go system.handleEvents()

	fmt.Println("Initial State:", system.getState())

	sendEvents(system.eventChan)

	// Wait for the system to complete before exiting
	time.Sleep(2 * time.Second)
	close(system.quit)
}

func sendEvents(eventChan chan string) {
	events := []string{"start", "slow", "stop", "start", "fast", "exit"}
	for _, event := range events {
		eventChan <- event
		if event != "fast" {
			time.Sleep(500 * time.Millisecond)
		}
	}
}
