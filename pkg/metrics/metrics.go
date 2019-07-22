package metrics

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Error to be  returned when marketplace is unable to reach quay.io.
type OpsrcError string

const (
	// Default values which are to be used as labels
	// for prometheus metrics.
	ERROR      = "opsrc_error"
	OPSRC_NAME = "opsrc_name"

	// Path, Port and Host where custom metrics would be exposed.
	METRICSPATH = "/metrics"
	METRICSPORT = "8383"
	METRICSHOST = "0.0.0.0"

	// Error message to be returned when quay.io is down.
	UNREACHABLE_ERROR OpsrcError = "quay_unavailabe"

	// Error message to be returned when host address is wrong.
	NO_HOST_EXSISTS_ERROR OpsrcError = "no_host_exsists"
)

// Metric variables that report error based on the operator-source
// which fails to reconcile.
var (
	operatorSourceError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "operator_source_failure_count",
			Help: "Monotonic count of times operator-source could not download packages",
		},
		[]string{OPSRC_NAME, ERROR},
	)
	// List of custom metrics which are to be collected.
	metricsList = []prometheus.Collector{
		operatorSourceError,
	}
)

// registerMetrics registers the metrics with prometheus.
func registerMetrics() error {
	for _, metric := range metricsList {
		err := prometheus.Register(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

// RecordMetrics starts the server, records the custom metrics and exposes
// the metrics at the specified endpoint.
func RecordMetrics() {
	// Register metrics for the operator with the prometheus.
	registerMetrics()

	// Start the server and expose the registered metrics.
	http.Handle(METRICSPATH, promhttp.Handler())
	port := ":" + METRICSPORT
	go http.ListenAndServe(port, nil)
}

// counterForOpsrcError returns the number of times error was encountered while
// contacting quay for operstor-sources.
func counterForOpsrcError(opsrc, opsrcError string) prometheus.Counter {
	return operatorSourceError.WithLabelValues(opsrc, opsrcError)
}

// updateMetrics takes the name of the operator source which fails to reconcile, and
// increments the error count of the respective operator metric.
func UpdateMetrics(opsrc, opsrcError string) {
	counterForOpsrcError(opsrc, opsrcError).Inc()
}

// ErrorMessage takes the error which appears while Listing the packages,
// parses it, and returns the error message which is to be published.
func ErrorMessage(err string) OpsrcError {
	if strings.Contains(err, "unknown error (status 404)") {
		return UNREACHABLE_ERROR
	}
	if strings.Contains(err, "no such host") {
		return NO_HOST_EXSISTS_ERROR
	}
	return "Error occured while listing packages. Marketplace unable to contact the given quay endpoint"
}

// OpsrcErrorToString converts the error message of OpsrcError type to string.
func (operatorError OpsrcError) OpsrcErrorToString() string {
	return string(operatorError)
}
