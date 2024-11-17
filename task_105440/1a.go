//go:build 1a
// +build 1a

package main

import (
	"fmt"
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
	state State
}

func (s *System) getState() State {
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

func main() {
	system := &System{state: StateIdle}

	fmt.Println("Initial State:", system.getState())

	// Polling-based state transitions
	for {
		event := getUserInput()
		system.transition(event)
		fmt.Println("Current State:", system.getState())

		// Simulate some work or delay in the Running state
		if system.getState() == StateRunning {
			fmt.Println("System Running...")
			time.Sleep(1 * time.Second)
		}

		if event == "exit" {
			break
		}
	}

	fmt.Println("System Exiting.")
}

func getUserInput() string {
	fmt.Print("Enter event (start, stop, exit): ")
	var input string
	fmt.Scanln(&input)
	return input
}
