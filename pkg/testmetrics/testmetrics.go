package testmetrics

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

/**
*
* Variable to store custom metrics - counter
*
 */

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events Test",
	})
)

/**
*
* Variable to store custom metrics - gauge
*
 */

var (
	gaugeTest = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "testing_gauge_for_metrics",
		Help: "Testing decrementing gauge",
	})
)

func RecordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			gaugeTest.SetToCurrentTime()
			time.Sleep(2 * time.Second)
			log.Printf("Recording custom metrics")
		}
	}()
}
