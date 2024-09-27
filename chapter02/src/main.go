package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/monte-carlo-pi", monteCarloPiHandler)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func monteCarloPiHandler(w http.ResponseWriter, r *http.Request) {
	iterations := r.URL.Query().Get("iterations")
	if iterations == "" {
		http.Error(w, "Please provide the number of iterations as a query parameter", http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(iterations)
	if err != nil {
		http.Error(w, "Invalid number of iterations", http.StatusBadRequest)
		return
	}

	pi := calculatePi(n)
	fmt.Fprintf(w, "Estimated value of Pi: %f", pi)
}

func calculatePi(iterations int) float64 {
	rand.Seed(time.Now().UnixNano())
	inside := 0

	for i := 0; i < iterations; i++ {
		x := rand.Float64()
		y := rand.Float64()

		if x*x+y*y <= 1 {
			inside++
		}
	}

	return 4 * float64(inside) / float64(iterations)
}