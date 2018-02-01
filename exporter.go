package edgemax_exporter

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vaga/edgemax_exporter/edgemax"
)

// Exporter is a Prometheus exporter for Ubiquiti UniFi Controller API
// metrics. It wraps all UniFi metrics collectors and provides a single global
// exporter which can serve metrics. It also ensures that the collection
// is done in a thread-safe manner, the necessary requirement stated by
// Prometheus. It implements the prometheus.Collector interface in order to
// register with Prometheus.
type Exporter struct {
	mu         sync.Mutex
	collectors []prometheus.Collector
}

// Verify that the Exporter implements the prometheus.Collector interface.
var _ prometheus.Collector = &Exporter{}

// namespace is the top-level namespace for this UniFi exporter.
const namespace = "edgemax"

// New creates a new Exporter which collects metrics from one or mote sites.
func New(client *edgemax.Client) (*Exporter, func(), error) {

	systemCh := make(chan edgemax.SystemStat)
	dpiCh := make(chan edgemax.DPIStat)
	ifacesCh := make(chan edgemax.InterfacesStat)

	done, err := client.Stats(systemCh, dpiCh, ifacesCh)
	if err != nil {
		return nil, nil, err
	}

	return &Exporter{
		collectors: []prometheus.Collector{
			newSystemCollector(systemCh),
			newDPICollector(dpiCh),
			newInterfacesCollector(ifacesCh),
		},
	}, done, nil
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range e.collectors {
		c.Describe(ch)
	}
}

// Collect sends the collected metrics from each of the collectors to
// prometheus. Collect could be called several times concurrently
// and thus its run is protected by a single mutex.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, c := range e.collectors {
		c.Collect(ch)
	}
}
