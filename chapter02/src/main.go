package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		},
		[]string{"code", "method"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of all HTTP requests",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method"},
	)

	monteCarloIterations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "monte_carlo_iterations_total",
			Help: "Total number of Monte Carlo iterations performed",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(monteCarloIterations)
}

func main() {
	http.HandleFunc("/monte-carlo-pi", instrumentHandler(monteCarloPiHandler))
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func instrumentHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		httpRequestsTotal.WithLabelValues(strconv.Itoa(rw.statusCode), r.Method).Inc()
		httpRequestDuration.WithLabelValues(strconv.Itoa(rw.statusCode), r.Method).Observe(duration.Seconds())
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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

	monteCarloIterations.Add(float64(iterations))

	return 4 * float64(inside) / float64(iterations)
}