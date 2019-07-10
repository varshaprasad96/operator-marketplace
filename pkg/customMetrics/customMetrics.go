package customMetrics

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// example custom metric variables of Counter and Gauge types.
	opsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events Test",
	})
	gaugeTest = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "testing_gauge_for_metrics",
		Help: "Testing decrementing gauge",
	})

	// List of custom metrics which are to be collected.
	metricsList = []prometheus.Collector{
		opsProcessed,
		gaugeTest,
	}

	// custom metrics port and path where the metrics would be
	// exposed. This is the same port and path where the SDK
	// exposes metrics.
	metricsPath = "/metrics"
	metricsPort = ":8383"
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
	http.Handle(metricsPath, promhttp.Handler())
	go http.ListenAndServe(metricsPort, nil)

	// Start recording custom metrics.
	go func() {
		for {
			opsProcessed.Inc()
			gaugeTest.SetToCurrentTime()
			time.Sleep(2 * time.Second)
			log.Printf("Recording custom metrics")
		}
	}()
}
