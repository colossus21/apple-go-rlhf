package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type State int

const (
	StateIdle State = iota
	StateProcessing
	StateCompleted
)

type Event struct {
	name     string
	response chan error
}

type StateManager struct {
	state     State
	eventChan chan Event
	stateChan chan State
	done      chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
}

func NewStateManager() *StateManager {
	return &StateManager{
		state:     StateIdle,
		eventChan: make(chan Event, 10),
		stateChan: make(chan State, 1),
		done:      make(chan struct{}),
	}
}

func (sm *StateManager) getState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

func (sm *StateManager) setState(newState State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.state != newState {
		sm.state = newState
		select {
		case sm.stateChan <- newState:
		default:
		}
	}
}

func (sm *StateManager) SendEvent(name string) error {
	resp := make(chan error, 1)
	select {
	case sm.eventChan <- Event{name: name, response: resp}:
		return <-resp
	case <-sm.done:
		return fmt.Errorf("state manager is shutting down")
	}
}

func (sm *StateManager) Start() {
	sm.wg.Add(2)
	go sm.handleEvents()
	go sm.monitorState()
}

func (sm *StateManager) Stop() {
	close(sm.done)
	sm.wg.Wait()
	close(sm.stateChan)
}

func (sm *StateManager) handleEvents() {
	defer sm.wg.Done()
	for {
		select {
		case event := <-sm.eventChan:
			start := time.Now()
			err := sm.handleStateTransition(event.name)
			event.response <- err
			fmt.Printf("Transition '%s' took: %v\n", event.name, time.Since(start))
		case <-sm.done:
			return
		}
	}
}

func (sm *StateManager) handleStateTransition(event string) error {
	currentState := sm.getState()
	switch currentState {
	case StateIdle:
		if event == "start" {
			sm.setState(StateProcessing)
			return nil
		}
	case StateProcessing:
		if event == "complete" {
			// Simulate heavy processing
			workDuration := time.Duration(100+rand.Intn(400)) * time.Millisecond
			time.Sleep(workDuration)
			sm.setState(StateCompleted)
			return nil
		}
	case StateCompleted:
		if event == "reset" {
			sm.setState(StateIdle)
			return nil
		}
	}
	return fmt.Errorf("invalid transition: state=%v, event=%s", currentState, event)
}

func (sm *StateManager) monitorState() {
	defer sm.wg.Done()
	for {
		select {
		case state := <-sm.stateChan:
			fmt.Printf("State changed to: %v\n", state)
		case <-sm.done:
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	sm := NewStateManager()
	sm.Start()
	defer sm.Stop()

	fmt.Println("Testing fast transitions...")
	for i := 0; i < 3; i++ {
		if err := sm.SendEvent("start"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		if err := sm.SendEvent("reset"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Println("\nTesting slow transitions...")
	for i := 0; i < 3; i++ {
		if err := sm.SendEvent("start"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		if err := sm.SendEvent("complete"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		if err := sm.SendEvent("reset"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	time.Sleep(2 * time.Second)
}
