
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var requests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests handled by the application.",
	},
	[]string{"method", "path", "status"},
)

var latency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

func init() {
	prometheus.MustRegister(requests, latency)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func instrument(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next(rec, r)

		if rec.status == 0 {
			rec.status = http.StatusOK
		}

		requests.WithLabelValues(
			r.Method,
			path,
			strconv.Itoa(rec.status),
		).Inc()

		latency.WithLabelValues(
			r.Method,
			path,
		).Observe(time.Since(start).Seconds())
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /home", instrument("/home",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Welcome to my Go HTTP server!")
		},
	))

	// Keep the metrics endpoint separate from application request metrics.
	mux.Handle("GET /metrics", promhttp.Handler())

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
