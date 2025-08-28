package main

import (
    "fmt"
    "math/rand"
    "sync/atomic"
    "net/http"
    "strconv"
    "time"
    "log"
    "os"
    "os/signal"
    "context"
    "syscall"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    isReady int64 = 0  // 0 = not ready, 1 = ready
    isAlive int64 = 1  // 0 = dead, 1 = alive
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

    monteCarloLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "monte_carlo_latency_seconds",
            Help:    "Latency of Monte Carlo Pi calculations",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
    )

    applicationReadiness = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "application_readiness_status",
            Help: "Application readiness status (1 = ready, 0 = not ready)",
        },
    )
    applicationLiveness = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "application_liveness_status",  
            Help: "Application liveness status (1 = alive, 0 = dead)",
        },
    )
    gracefulShutdownCounter = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "application_graceful_shutdowns_total",
            Help: "Total number of graceful shutdown attempts",
        },
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(monteCarloLatency)
    prometheus.MustRegister(applicationReadiness)
    prometheus.MustRegister(applicationLiveness)
    prometheus.MustRegister(gracefulShutdownCounter)
}

func main() {
    // Simulate application startup (database connections, config loading, etc.)
    log.Println("Application starting up...")
    time.Sleep(5 * time.Second)
    
    // Mark application as ready after startup completes
    atomic.StoreInt64(&isReady, 1)
    applicationReadiness.Set(1)
    applicationLiveness.Set(1)
    log.Println("Application startup complete, ready to serve traffic")
        
    http.HandleFunc("/monte-carlo-pi", instrumentHandler(monteCarloPiHandler))    
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/ready", readinessHandler)
    http.Handle("/metrics", promhttp.Handler())
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: nil,
    }

    // Start server in a goroutine
    go func() {
        fmt.Println("Server is running on http://localhost:8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()

    handleGracefulShutdown(server)
}

func handleGracefulShutdown(server *http.Server) {
    // Set up signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Wait for termination signal
    sig := <-sigChan
    log.Printf("Received shutdown signal: %v", sig)
    gracefulShutdownCounter.Inc()

    // Begin graceful shutdown sequence
    log.Println("Starting graceful shutdown...")
    
    // Mark application as not ready (stops new traffic)
    atomic.StoreInt64(&isReady, 0)
    applicationReadiness.Set(0)
    log.Println("Application marked as not ready, stopped accepting new traffic")

    // Simulate application-specific cleanup work
    log.Println("Performing application cleanup operations...")
    time.Sleep(75 * time.Second)
    log.Println("Application cleanup operations completed")

    // Shutdown HTTP server gracefully
    log.Println("Shutting down HTTP server...")
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server shutdown error: %v", err)
        applicationLiveness.Set(0)
    } else {
        log.Println("Server shutdown completed successfully")
    }

    // Mark application as not alive
    applicationLiveness.Set(0)
    log.Println("Application shutdown complete")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    if atomic.LoadInt64(&isAlive) == 1 {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Not healthy"))
    }
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    if atomic.LoadInt64(&isReady) == 1 {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Ready"))
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Not ready"))
    }
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

    start := time.Now()
    pi := calculatePi(n)
    duration := time.Since(start)
    
    monteCarloLatency.Observe(duration.Seconds())

    fmt.Fprintf(w, "Estimated value of Pi: %f (calculated in %v)", pi, duration)
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