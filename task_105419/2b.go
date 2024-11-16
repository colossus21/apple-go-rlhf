//go:build 2b
// +build 2b

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Observer represents an observer of market data.
type Observer interface {
	Update(price float64)
}

// BayesianEstimator is an observer that uses Bayesian methods for market analysis.
type BayesianEstimator struct {
	mu, sigma   float64 // Prior parameters of the Normal distribution
	alpha, beta float64 // Prior parameters of the Beta distribution
	mean        float64 // Estimated mean price
	count       int     // Number of observed prices
	mux         sync.Mutex
}

func NewBayesianEstimator() *BayesianEstimator {
	return &BayesianEstimator{
		mu:    100.0,
		sigma: 10.0,
		alpha: 1.0,
		beta:  1.0,
	}
}

// Update updates the estimator with a new observed price.
func (e *BayesianEstimator) Update(price float64) {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.count++
	e.mean = (e.alpha*e.mean + price) / (e.alpha + 1)
	e.alpha++
}

// ProbabilityOfHigherMean calculates the probability that the true mean price is higher than the given threshold.
func (e *BayesianEstimator) ProbabilityOfHigherMean(threshold float64) float64 {
	return 1.0 - math.GammaInc(e.alpha, (threshold-e.mean)*e.count)/(math.Gamma(e.alpha)*math.Gamma(e.beta+e.count))
}

// GenerateSignal generates a trading signal.
func (e *BayesianEstimator) GenerateSignal(threshold float64, confidence float64) int {
	prob := e.ProbabilityOfHigherMean(threshold)
	if prob > confidence {
		return 1 // Buy signal
	} else if prob < 1-confidence {
		return -1 // Sell signal
	}
	return 0 // No signal
}

// DataSource provides real-time market data updates to observers.
type DataSource struct {
	observers    []Observer
	tickInterval time.Duration
	mux          sync.RWMutex
}

func NewDataSource(tickInterval time.Duration) *DataSource {
	return &DataSource{
		tickInterval: tickInterval,
	}
}

// Attach adds an observer to the data source.
func (ds *DataSource) Attach(observer Observer) {
	ds.mux.Lock()
	defer ds.mux.Unlock()
	ds.observers = append(ds.observers, observer)
}

// Start generates random price data and notifies observers at the specified tick interval.
func (ds *DataSource) Start() {
	rand.Seed(time.Now().UnixNano())
	ticker := time.NewTicker(ds.tickInterval)
	defer ticker.Stop()

	for range ticker.C {
		price := rand.NormFloat64()*10 + 105
		ds.mux.RLock()
		for _, observer := range ds.observers {
			observer.Update(price)
		}
		ds.mux.RUnlock()
	}
}

func main() {
	ds := NewDataSource(time.Second)
	estimator := NewBayesianEstimator()
	ds.Attach(estimator)

	threshold := 100.0
	confidence := 0.95

	fmt.Println("Starting data source...")
	go ds.Start()

	time.Sleep(10 * time.Second) // Wait for some data updates

	fmt.Printf("Final estimated mean: %.2f\n", estimator.mean)
	fmt.Printf("Signal: %d\n", estimator.GenerateSignal(threshold, confidence))
	fmt.Printf("Probability of mean > %.2f: %.2f\n", threshold, estimator.ProbabilityOfHigherMean(threshold))
}
