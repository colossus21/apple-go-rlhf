//go:build 1b
// +build 1b

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func fetchUserData(ctx context.Context, userID string) (string, string, error) {
	// Simulate fetching user data from an external source
	time.Sleep(2 * time.Second)

	// Replace this with actual user data lookup logic
	userName := fmt.Sprintf("User %s", userID)
	userEmail := fmt.Sprintf("%s@example.com", userID)
	return userName, userEmail, nil
}

func userDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		userID := "123" // Replace this with the actual way to get the user ID from the request

		var name, email string
		var err error

		go func() {
			name, email, err = fetchUserData(ctx, userID)
		}()

		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.Write([]byte("User data fetch timed out."))
		default:
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}
