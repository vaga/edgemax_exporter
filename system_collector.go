package edgemax_exporter

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vaga/edgemax_exporter/edgemax"
)

// A systemCollector is a Prometheus collector for metrics regarding Ubiquiti
// UniFi devices.
type systemCollector struct {
	cpuPercent    prometheus.Gauge
	uptimeSeconds prometheus.Gauge
	memoryPercent prometheus.Gauge
}

// Verify that the Exporter implements the prometheus.Collector interface.
var _ prometheus.Collector = &systemCollector{}

// newSystemCollector creates a new systemCollector which collects metrics for
// a specified site.
func newSystemCollector(ch <-chan edgemax.SystemStat) *systemCollector {
	const subsystem = "system"

	c := &systemCollector{
		cpuPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cpu_percent",
			Help:      "System CPU usage percentage",
		}),
		uptimeSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "uptime_seconds",
			Help:      "System uptime in seconds",
		}),
		memoryPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_percent",
			Help:      "System memory usage percentage",
		}),
	}

	go c.collect(ch)

	return c
}

// collect begins a metrics collection task for all metrics related to UniFi
// devices.
func (c *systemCollector) collect(ch <-chan edgemax.SystemStat) {
	for s := range ch {
		cpu, _ := strconv.Atoi(s.CPU)
		c.cpuPercent.Set(float64(cpu))

		uptime, _ := strconv.Atoi(s.Uptime)
		c.uptimeSeconds.Set(float64(uptime))

		mem, _ := strconv.Atoi(s.Mem)
		c.memoryPercent.Set(float64(mem))
	}
}

// metrics contains a list of metrics which are collected each time
// the exporter is scraped. This list must be kept in sync with the metrics
// in systemCollector.
func (c *systemCollector) metrics() []prometheus.Metric {
	return []prometheus.Metric{
		c.cpuPercent,
		c.uptimeSeconds,
		c.memoryPercent,
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *systemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics() {
		ch <- m.Desc()
	}
}

// Collect sends the metric values for each metric pertaining to the global
// cluster usage over to the provided prometheus Metric channel.
func (c *systemCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.metrics() {
		ch <- m
	}
}
