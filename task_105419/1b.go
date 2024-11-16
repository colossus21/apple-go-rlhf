//go:build 1b
// +build 1b

package main

import (
	"fmt"
	"math"
)

type stock struct {
	name           string
	prices         []float64
	muPrior        float64
	sigmaPrior     float64
	muPosterior    float64
	sigmaPosterior float64
}

// Assumes normally distributed returns
func (s *stock) updateBayesianParameters(price float64) {
	n := float64(len(s.prices) + 1)
	sigmaSq := (s.sigmaPrior * s.sigmaPrior) / (len(s.prices))
	newMu := ((s.muPrior * len(s.prices)) + price) / n
	newSigmaSq := sigmaSq + ((price-s.muPrior)*(price-s.muPrior))/n - sigmaSq/n
	s.muPosterior = newMu
	s.sigmaPosterior = math.Sqrt(newSigmaSq)
	s.prices = append(s.prices, price)
}

func (s *stock) generateTradingSignal() string {
	// Simple signal generation: Buy if expected return is greater than zero
	expectedReturn := s.muPosterior
	if expectedReturn > 0 {
		return "Buy"
	} else if expectedReturn < 0 {
		return "Sell"
	} else {
		return "Hold"
	}
}

func main() {
	// Example stock prices
	prices := []float64{100.0, 101.5, 100.8, 102.0, 103.0, 102.5, 101.0}

	stock := &stock{
		name:       "Example Stock",
		prices:     prices,
		muPrior:    0.0,  // Uninformative prior mean
		sigmaPrior: 10.0, // Uninformative prior standard deviation
	}

	fmt.Printf("Initial Stock Parameters: Mu=%.2f, Sigma=%.2f\n", stock.muPrior, stock.sigmaPrior)

	for _, price := range prices {
		stock.updateBayesianParameters(price)
		signal := stock.generateTradingSignal()
		fmt.Printf("Today's Price: $%.2f, Mu=%.2f, Sigma=%.2f, Signal: %s\n", price, stock.muPosterior, stock.sigmaPosterior, signal)
	}
}
