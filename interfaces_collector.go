package edgemax_exporter

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vaga/edgemax_exporter/edgemax"
)

// A interfacesCollector is a Prometheus collector for metrics regarding Ubiquiti
// UniFi devices.
type interfacesCollector struct {
	receivedBytes    *prometheus.GaugeVec
	transmittedBytes *prometheus.GaugeVec
}

// Verify that the Exporter implements the prometheus.Collector interface.
var _ prometheus.Collector = &interfacesCollector{}

// newInterfacesCollector creates a new interfacesCollector which collects metrics for
// a specified site.
func newInterfacesCollector(ch <-chan edgemax.InterfacesStat) *interfacesCollector {
	const subsystem = "interfaces"
	labels := []string{"name", "mac"}

	c := &interfacesCollector{
		receivedBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "received_bytes",
				Help:      "Number of bytes received by interfaces, partitioned by network interface",
			},
			labels,
		),
		transmittedBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "transmitted_bytes",
				Help:      "Number of bytes transmitted by interfaces, partitioned by network interface",
			},
			labels,
		),
	}

	go c.collect(ch)

	return c
}

// collect begins a metrics collection task for all metrics related to UniFi
// devices.
func (c *interfacesCollector) collect(ch <-chan edgemax.InterfacesStat) {
	for s := range ch {
		for name, iface := range s {
			labels := []string{name, iface.Mac}

			rxBytes, _ := strconv.Atoi(iface.Stats.RXBytes)
			c.receivedBytes.WithLabelValues(labels...).Set(float64(rxBytes))

			txBytes, _ := strconv.Atoi(iface.Stats.TXBytes)
			c.transmittedBytes.WithLabelValues(labels...).Set(float64(txBytes))
		}
	}
}

// collectors contains a list of collectors which are collected each time
// the exporter is scraped. This list must be kept in sync with the collectors
// in InterfacesCollector.
func (c *interfacesCollector) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		c.receivedBytes,
		c.transmittedBytes,
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *interfacesCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.collectors() {
		m.Describe(ch)
	}
}

// Collect sends the metric values for each metric pertaining to the global
// cluster usage over to the provided prometheus Metric channel.
func (c *interfacesCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.collectors() {
		m.Collect(ch)
	}
}
