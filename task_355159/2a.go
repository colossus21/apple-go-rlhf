//go:build 2a
// +build 2a

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	successfulFetchesUnder2s = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "successful_profile_fetches_under_2s",
		Help: "Number of successful profile fetches under 2 seconds.",
	})
	timedOutProfileFetches = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "timed_out_profile_fetches",
		Help: "Number of timed out profile fetches.",
	})
	profileFetchLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "profile_fetch_latency_seconds",
		Help:    "Latency of profile fetches in seconds.",
		Buckets: prometheus.LinearBuckets(0, 0.2, 11), // Buckets from 0 to 2 seconds in increments of 0.2 seconds
	})
)

func init() {
	prometheus.MustRegister(successfulFetchesUnder2s)
	prometheus.MustRegister(timedOutProfileFetches)
	prometheus.MustRegister(profileFetchLatency)
}

func fetchUserData(ctx context.Context, userID string) (string, string, error) {
	// Simulate fetching user data. Randomly determine whether it succeeds or times out
	delay := time.Duration(rand.Intn(2*1000)) * time.Millisecond
	time.Sleep(delay)

	if delay >= 2*time.Second {
		return "", "", fmt.Errorf("user data fetch timed out for user %s", userID)
	}

	// Replace this with actual user data lookup logic
	userName := fmt.Sprintf("User %s", userID)
	userEmail := fmt.Sprintf("%s@example.com", userID)
	return userName, userEmail, nil
}

func userDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		start := time.Now()
		userID := "123"

		var name, email string
		var err error

		go func() {
			name, email, err = fetchUserData(ctx, userID)
		}()

		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				timedOutProfileFetches.Inc()
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.Write([]byte("User data fetch timed out."))
		default:
			if err != nil {
				log.Println("Error fetching user data:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			latency := time.Since(start)
			profileFetchLatency.Observe(latency.Seconds())
			if latency.Seconds() < 2 {
				successfulFetchesUnder2s.Inc()
			}
			fmt.Fprintf(w, "User: %s, Email: %s\n", name, email)
		}
	})
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.Handle("/user", userDataMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User details page."))
	})))

	http.Handle("/metrics", metricsHandler())

	log.Fatal(http.ListenAndServe(":8080", nil))
}
