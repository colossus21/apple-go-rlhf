//go:build 2b
// +build 2b

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
				// Slow transition scenario (random time between 0.1 and 2 seconds)
				workDuration := time.Duration(rand.Intn(19)) * 10 * time.Millisecond
				fmt.Println("System Running... Work duration:", workDuration)
				time.Sleep(workDuration)
			default:
				// Fast transition scenario (instantaneous)
				fmt.Println("System in", s.getState(), "state.")
			}
		case <-s.quit:
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	system := &System{
		state:     StateIdle,
		eventChan: make(chan string),
		quit:      make(chan struct{}),
	}

	go system.handleEvents()

	fmt.Println("Initial State:", system.getState())

	// Test slow transition
	time.Sleep(2 * time.Second)
	system.eventChan <- "start"
	system.eventChan <- "stop"

	fmt.Println("\nTesting fast transitions:")
	// Test fast transitions
	for i := 0; i < 5; i++ {
		system.eventChan <- "start"
		time.Sleep(50 * time.Millisecond)
		system.eventChan <- "stop"
	}

	system.quit <- struct{}{}
	<-system.quit

	fmt.Println("System Exiting.")
}
