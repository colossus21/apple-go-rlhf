I'll provide an implementation using goroutines and channels to achieve non-blocking state transitions with clear demonstration of both slow and fast scenarios.

```go
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

type StateManager struct {
    state     State
    eventChan chan string
    stateChan chan State
    quit      chan struct{}
    mu        sync.RWMutex
}

func NewStateManager() *StateManager {
    return &StateManager{
        state:     StateIdle,
        eventChan: make(chan string, 10),
        stateChan: make(chan State, 1),
        quit:      make(chan struct{}),
    }
}

func (sm *StateManager) getState() State {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    return sm.state
}

func (sm *StateManager) setState(newState State) {
    sm.mu.Lock()
    sm.state = newState
    sm.mu.Unlock()
    select {
    case sm.stateChan <- newState:
    default:
    }
}

func (sm *StateManager) handleEvents() {
    for {
        select {
        case event := <-sm.eventChan:
            start := time.Now()
            sm.handleStateTransition(event)
            fmt.Printf("Transition time: %v\n", time.Since(start))
        case <-sm.quit:
            return
        }
    }
}

func (sm *StateManager) handleStateTransition(event string) {
    currentState := sm.getState()
    switch currentState {
    case StateIdle:
        if event == "start" {
            // Fast transition
            sm.setState(StateProcessing)
        }
    case StateProcessing:
        if event == "complete" {
            // Slow transition - simulate heavy processing
            time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
            sm.setState(StateCompleted)
        }
    case StateCompleted:
        if event == "reset" {
            // Fast transition
            sm.setState(StateIdle)
        }
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())
    sm := NewStateManager()
    go sm.handleEvents()

    // Monitor state changes in separate goroutine
    go func() {
        for state := range sm.stateChan {
            fmt.Printf("State changed to: %v\n", state)
        }
    }()

    fmt.Println("Testing fast transitions...")
    for i := 0; i < 3; i++ {
        sm.eventChan <- "start"
        sm.eventChan <- "reset"
    }

    fmt.Println("\nTesting slow transitions...")
    for i := 0; i < 3; i++ {
        sm.eventChan <- "start"
        sm.eventChan <- "complete"
        sm.eventChan <- "reset"
        time.Sleep(50 * time.Millisecond)
    }

    time.Sleep(2 * time.Second)
    close(sm.quit)
    close(sm.stateChan)
}
```

This implementation:
1. Uses goroutines for non-blocking state transitions.
2. Employs buffered channels to prevent blocking on state updates.
3. Demonstrates both fast (immediate) and slow (simulated processing) transitions.
4. Uses `RWMutex` for thread-safe state access.
5. Provides timing measurements for transitions.
6. Includes a separate goroutine for state monitoring.

The state changes are non-blocking thanks to the goroutine handling events, and we can measure the actual transition times to see the difference between fast and slow scenarios.