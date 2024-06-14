package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Request structure for the JSON input
type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

// Response structure for the JSON output
type Response struct {
	FactorialA int `json:"a_factorial"`
	FactorialB int `json:"b_factorial"`
}

// Middleware to check if a and b exist and are non-negative integers
func validate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Incorrect input"}`, http.StatusBadRequest)
			return
		}

		if req.A < 0 || req.B < 0 {
			http.Error(w, `{"error":"Incorrect input"}`, http.StatusBadRequest)
			return
		}

		// Attach the validated request to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "request", req)
		r = r.WithContext(ctx)

		next(w, r, ps)
	}
}

// Function to calculate factorial
func factorial(n int, resultChan chan int) {
	if n == 0 {
		resultChan <- 1
		return
	}

	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	resultChan <- result
}

// Handler for the /calculate endpoint
func calculate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Retrieve the validated request from the context
	req, ok := r.Context().Value("request").(Request)
	if !ok {
		http.Error(w, `{"error":"Incorrect input"}`, http.StatusBadRequest)
		return
	}

	resultChanA := make(chan int)
	resultChanB := make(chan int)

	go factorial(req.A, resultChanA)
	go factorial(req.B, resultChanB)

	resp := Response{
		FactorialA: <-resultChanA,
		FactorialB: <-resultChanB,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	router := httprouter.New()
	router.POST("/calculate", validate(calculate))

	http.ListenAndServe(":8989", router)
}
