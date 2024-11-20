//go:build 2ideal
// +build 2ideal

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type metrics struct {
	successfulFetches uint64
	timedOutFetches   uint64
	totalDuration     time.Duration
}

var m = &metrics{}

func fetchUserData(ctx context.Context, userID string) (string, string, error) {
	fetchTime := time.Duration(rand.Intn(3)) * time.Second
	time.Sleep(fetchTime)

	userName := fmt.Sprintf("User %s", userID)
	userEmail := fmt.Sprintf("%s@example.com", userID)
	return userName, userEmail, nil
}

func userDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		userID := "123"

		var name, email string
		var err error

		start := time.Now()

		name, email, err = fetchUserData(ctx, userID)
		duration := time.Since(start)

		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				atomic.AddUint64(&m.timedOutFetches, 1)
				atomic.AddInt64((*int64)(&m.totalDuration), int64(duration))
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.Write([]byte("User data fetch timed out."))
		default:
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			atomic.AddUint64(&m.successfulFetches, 1)
			atomic.AddInt64((*int64)(&m.totalDuration), int64(duration))
			fmt.Fprintf(w, "User: %s, Email: %s\n", name, email)
		}
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	successful := atomic.LoadUint64(&m.successfulFetches)
	timedOut := atomic.LoadUint64(&m.timedOutFetches)
	total := successful + timedOut
	duration := time.Duration(atomic.LoadInt64((*int64)(&m.totalDuration)))

	var avgDuration float64
	if total > 0 {
		avgDuration = float64(duration) / float64(total) / float64(time.Second)
	}

	fmt.Fprintf(w, "Successful fetches: %d\n", successful)
	fmt.Fprintf(w, "Timed out fetches: %d\n", timedOut)
	fmt.Fprintf(w, "Average duration: %.2f seconds\n", avgDuration)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.Handle("/user", userDataMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User details page."))
	})))

	http.HandleFunc("/metrics", metricsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
