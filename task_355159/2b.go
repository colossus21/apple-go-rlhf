//go:build 2b
// +build 2b

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
	successProfileFetches = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "success_profile_fetches",
		Help: "Number of successful profile fetches under 2 seconds",
	})

	timedOutProfileFetches = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "timed_out_profile_fetches",
		Help: "Number of timed-out profile fetches",
	})

	profileFetchDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "profile_fetch_duration_seconds",
		Help:    "Duration of profile fetches in seconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
	})
)

func init() {
	prometheus.MustRegister(successProfileFetches)
	prometheus.MustRegister(timedOutProfileFetches)
	prometheus.MustRegister(profileFetchDuration)
}

func fetchUserData(ctx context.Context, userID string) (string, string, error) {
	// Simulate random fetch time between 0 and 3 seconds
	fetchTime := time.Duration(rand.Intn(3)) * time.Second
	time.Sleep(fetchTime)

	// Replace this with actual user data lookup logic
	userName := fmt.Sprintf("User %s", userID)
	userEmail := fmt.Sprintf("%s@example.com", userID)
	return userName, userEmail, nil
}

func userDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		userID := "123" // Replace this with the actual way to get the user ID from the request

		var name, email string
		var err error

		start := time.Now()

		go func() {
			name, email, err = fetchUserData(ctx, userID)
		}()

		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				timedOutProfileFetches.Inc()
				profileFetchDuration.Observe(time.Since(start).Seconds())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.Write([]byte("User data fetch timed out."))
		default:
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			successProfileFetches.Inc()
			profileFetchDuration.Observe(time.Since(start).Seconds())
			fmt.Fprintf(w, "User: %s, Email: %s\n", name, email)
		}
	})
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.Handle("/user", userDataMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User details page."))
	})))

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8080", nil))
}
