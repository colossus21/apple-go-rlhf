//go:build 2a
// +build 2a

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Observer represents an entity that can observe price changes.
type Observer interface {
	Update(price float64)
}

// Observable is a subject that can be observed.
type Observable struct {
	observers []Observer
	mu        sync.Mutex // Mutex for thread safety
}

// Attach registers an observer with the observable.
func (o *Observable) Attach(observer Observer) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.observers = append(o.observers, observer)
}

// Notify notifies all observers of a new price update.
func (o *Observable) Notify(price float64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, observer := range o.observers {
		observer.Update(price)
	}
}

// BayesianEstimator implements the Observer interface.
type BayesianEstimator struct {
	alpha, beta float64
	mean        float64
	count       int
	observable  *Observable
}

// NewBayesianEstimator creates a new Bayesian estimator with an observable.
func NewBayesianEstimator(alpha, beta float64, observable *Observable) *BayesianEstimator {
	estimator := &BayesianEstimator{
		alpha:      alpha,
		beta:       beta,
		observable: observable,
	}
	observable.Attach(estimator) // Register the estimator as an observer
	return estimator
}

func (e *BayesianEstimator) Update(price float64) {
	e.mean = (e.alpha*e.mean + price) / (e.alpha + 1)
	e.count++
	e.alpha++
}

func (e *BayesianEstimator) ProbabilityOfHigherMean(threshold float64) float64 {
	return 1.0 - math.GammaInc(e.alpha, (threshold-e.mean)*e.count)/(math.Gamma(e.alpha)*math.Gamma(e.beta+e.count))
}

func (e *BayesianEstimator) GenerateSignal(threshold float64, confidence float64) int {
	prob := e.ProbabilityOfHigherMean(threshold)
	if prob > confidence {
		return 1
	} else if prob < 1-confidence {
		return -1
	}
	return 0
}

func main() {
	// Create an observable market data source
	observable := &Observable{}

	// Initialize the Bayesian estimator with non-informative prior parameters
	estimator := NewBayesianEstimator(1, 1, observable)

	// Hypothesis testing parameters
	threshold := 100.0
	confidence := 0.95

	// Simulate real-time market data updates using a goroutine
	go func() {
		rand.Seed(time.Now().UnixNano())
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			price := rand.NormFloat64()*10 + 105
			observable.Notify(price) // Notify all observers of the new price
		}
	}()

	fmt.Println("Trading system started...")
	fmt.Println("Press Ctrl+C to exit")

	select {} // Wait for interrupt
}
