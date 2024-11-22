package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Listen(address string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(address, mux)
}

var (
	httpRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP request to external API",
		},
		[]string{"status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "get_rates_duration_seconds",
			Help:    "Duration of GetRates execution",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"step"},
	)

	dbOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	requestsProcessedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requests_processed_total",
			Help: "Tolal number of processed requests",
		},
	)

	requestTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of requests received by the service",
		},
	)

	dbOperationsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestTotal, requestDuration, dbOperationsTotal,
		requestsProcessedTotal, requestTotal, dbOperationsDuration)
}

func StatusRequestToGarantex(status string) {
	httpRequestTotal.WithLabelValues(status).Inc()
}

func TimeRequestToGarantex(step string, duration float64) {
	requestDuration.WithLabelValues(step).Observe(duration)
}

func StatusRequestToDB(operation, status string) {
	dbOperationsTotal.WithLabelValues(operation, status).Inc()
}

func CountSuccessRequestToService() {
	requestsProcessedTotal.Inc()
}

func CountRequestToService() {
	requestTotal.Inc()
}

func TimeRequestToDB(operation string, duration float64) {
	dbOperationsDuration.WithLabelValues(operation).Observe(duration)
}
