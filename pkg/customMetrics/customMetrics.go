package customMetrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Registry-namespaces of operator sources.
	REDHAT_OPERATOR_NS    = "redhat-operators"
	COMMUNITY_OPERATOR_NS = "community-operators"
	CERTIFIED_OPERATOR_NS = "certified-operators"

	// Default values which are to be used as labels
	// for prometheus metrics.
	REGISTRY_NAMESPACE = "registry_ns"
	OPSRC_NAME         = "opsrc_name"

	// custom metrics port, path and host where the metrics would be
	// exposed.
	MetricsPath = "/metrics"
	MetricsPort = "8383"
	MetricsHost = "0.0.0.0"
)

// Metric variables that report the error based on the operator-source
// which fails to reconcile.
var (
	redhatErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redhat_operator_reconcile_failure_count",
			Help: "Monotonic count of times redhat-opsrc was unable to reconcile",
		},
		[]string{REGISTRY_NAMESPACE, OPSRC_NAME},
	)
	communityErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "community_operator_reconcile_failure_count",
			Help: "Monotonic count of times community-opsrc was unable to reconcile",
		},
		[]string{REGISTRY_NAMESPACE, OPSRC_NAME},
	)
	certifiedErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "certified_operator_reconcile_failure_count",
			Help: "Monotonic count of times certified-opsrc was unable to reconcile",
		},
		[]string{REGISTRY_NAMESPACE, OPSRC_NAME},
	)
	customOperatorErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "custom_operator_reconcile_failure_count",
			Help: "Monotonic count of any of the custom operator was unable to reconcile",
		},
		[]string{OPSRC_NAME},
	)

	// List of custom metrics which are to be collected.
	metricsList = []prometheus.Collector{
		redhatErrorCount,
		communityErrorCount,
		certifiedErrorCount,
		customOperatorErrorCount,
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
	http.Handle(MetricsPath, promhttp.Handler())

	port := ":" + (MetricsPort)
	go http.ListenAndServe(port, nil)
}

// CounterForRedHatCount returns the prometheus counter redhatErrorCount.
func counterForRedHatOpsrcError(opsrc string) prometheus.Counter {
	return redhatErrorCount.WithLabelValues(REDHAT_OPERATOR_NS, opsrc)
}

// CounterForRedHatCount returns the prometheus counter communityErrorCount.
func counterForCommunityOpsrcError(opsrc string) prometheus.Counter {
	return redhatErrorCount.WithLabelValues(COMMUNITY_OPERATOR_NS, opsrc)
}

// CounterForRedHatCount returns the prometheus counter certifiedErrorCount.
func counterForCertifiedOpsrcError(opsrc string) prometheus.Counter {
	return redhatErrorCount.WithLabelValues(CERTIFIED_OPERATOR_NS, opsrc)
}

// CounterForRedHatCount returns the prometheus counter customOperatorErrorCount.
func counterForCustomOpsrcError(opsrc string) prometheus.Counter {
	return redhatErrorCount.WithLabelValues("namespace", opsrc)
}

// updateMetrics takes the name of the operator source which fails to reconcile, and
// increments the error count of the respective operator metric.
func UpdateMetrics(opsrc string) {
	switch opsrcName := opsrc; opsrcName {
	case REDHAT_OPERATOR_NS:
		counterForRedHatOpsrcError(opsrcName).Inc()
	case COMMUNITY_OPERATOR_NS:
		counterForCommunityOpsrcError(opsrcName).Inc()
	case CERTIFIED_OPERATOR_NS:
		counterForCertifiedOpsrcError(opsrcName).Inc()
	default:
		counterForCustomOpsrcError(opsrcName).Inc()
	}
}
