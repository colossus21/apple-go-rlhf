//go:build 1a
// +build 1a

package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// BayesianEstimator represents a Bayesian estimator for the mean price.
type BayesianEstimator struct {
	alpha, beta float64 // Prior parameters of the Beta distribution
	mean        float64 // Estimated mean price
	count       int     // Number of observed prices
}

// NewBayesianEstimator creates a new Bayesian estimator with initial prior parameters.
func NewBayesianEstimator(alpha, beta float64) *BayesianEstimator {
	return &BayesianEstimator{
		alpha: alpha,
		beta:  beta,
	}
}

// Update updates the estimator with a new observed price.
func (e *BayesianEstimator) Update(price float64) {
	e.mean = (e.alpha*e.mean + price) / (e.alpha + 1)
	e.count++
	e.alpha++
}

// ProbabilityOfHigherMean calculates the probability that the true mean price is higher than the given threshold.
func (e *BayesianEstimator) ProbabilityOfHigherMean(threshold float64) float64 {
	// Use the incomplete beta function to calculate the probability
	// In Go, this can be approximated using the regularized incomplete gamma function.
	return 1.0 - math.GammaInc(e.alpha, (threshold-e.mean)*e.count)/(math.Gamma(e.alpha)*math.Gamma(e.beta+e.count))
}

// GenerateSignal generates a trading signal based on the estimated mean price and a threshold.
// A signal of 1 indicates a buy signal, -1 indicates a sell signal, and 0 indicates no signal.
func (e *BayesianEstimator) GenerateSignal(threshold float64, confidence float64) int {
	prob := e.ProbabilityOfHigherMean(threshold)
	if prob > confidence {
		return 1 // Buy signal
	} else if prob < 1-confidence {
		return -1 // Sell signal
	}
	return 0 // No signal
}

func main() {
	// Initialize the Bayesian estimator with non-informative prior parameters
	estimator := NewBayesianEstimator(1, 1)

	// Hypothesis testing: Let's test the hypothesis that the true mean price is higher than 100
	threshold := 100.0
	confidence := 0.95 // 95% confidence level

	// Simulate price observations
	rand.Seed(time.Now().UnixNano())
	numObservations := 100
	for i := 0; i < numObservations; i++ {
		// Simulate a random price
		price := rand.NormFloat64()*10 + 105 // Mean of 105 with standard deviation of 10
		estimator.Update(price)

		// Print the estimated mean and signal after each observation
		fmt.Printf("Observation %d: Estimated mean: %.2f, Signal: %d\n", i+1, estimator.mean, estimator.GenerateSignal(threshold, confidence))
	}

	// Probability updating after all observations
	finalProb := estimator.ProbabilityOfHigherMean(threshold)
	fmt.Printf("\nFinal probability that the true mean price is higher than %.2f: %.2f\n", threshold, finalProb)
}
