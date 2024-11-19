//go:build 1a
// +build 1a

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const maxWaitTime = time.Second * 2

func getUserInfo(ctx context.Context) (name, email string, err error) {
	defer close(ctx.Done())

	select {
	case <-ctx.Done():
		err = fmt.Errorf("request timed out")
		return
	case <-time.After(time.Second):
		name = "John Doe"
		email = "johndoe@example.com"
	}
	return
}
func userInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
		defer cancel()
		go func() {
			name, email, _ := getUserInfo(ctx)
			fmt.Println("User Info:", name, email)
		}()

		next.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type userData struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		data := userData{}

		// Middleware has already fetched user information and it's printed here asynchronously
		// This is just to illustrate that the middleware works in the background.
		fmt.Fprintf(w, "Hello! Your request is being processed.\n")

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", userInfoMiddleware(nil)); err != nil {
		panic(err)
	}
}
