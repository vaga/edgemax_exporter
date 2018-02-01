package edgemax_exporter

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vaga/edgemax_exporter/edgemax"
)

// dpiCollector is a Prometheus collector for metrics regarding EdgeMAX
// deep packet inspection statistics.
type dpiCollector struct {
	receivedBytes    *prometheus.GaugeVec
	transmittedBytes *prometheus.GaugeVec
}

// Verify that the Exporter implements the prometheus.Collector interface.
var _ prometheus.Collector = &dpiCollector{}

// newDPICollector creates a new dpiCollector.
func newDPICollector(ch <-chan edgemax.DPIStat) *dpiCollector {
	const subsystem = "dpi"
	labels := []string{"client_ip", "category", "type"}

	c := &dpiCollector{
		receivedBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "received_bytes",
				Help:      "Number of bytes received by devices (client download)",
			},
			labels,
		),
		transmittedBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "transmitted_bytes",
				Help:      "Number of bytes transmitted by devices (client upload)",
			},
			labels,
		),
	}

	go c.collect(ch)

	return c
}

// collect begins a metrics collection task for all metrics related to UniFi
// devices.
func (c *dpiCollector) collect(ch <-chan edgemax.DPIStat) {
	for s := range ch {
		for ip, a := range s {
			for tc, stat := range a {
				t := strings.Split(tc, "|")
				labels := []string{ip, t[1], t[0]}

				rxBytes, _ := strconv.Atoi(stat.RXBytes)
				c.receivedBytes.WithLabelValues(labels...).Set(float64(rxBytes))

				txBytes, _ := strconv.Atoi(stat.TXBytes)
				c.transmittedBytes.WithLabelValues(labels...).Set(float64(txBytes))
			}
		}
	}
}

// collectors contains a list of collectors which are collected each time
// the exporter is scraped. This list must be kept in sync with the collectors
// in dpiCollector.
func (c *dpiCollector) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		c.receivedBytes,
		c.transmittedBytes,
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *dpiCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.collectors() {
		m.Describe(ch)
	}
}

// Collect sends the metric values for each metric pertaining to deep packet
// inspection to the provided prometheus Metric channel.
func (c *dpiCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.collectors() {
		m.Collect(ch)
	}
}
