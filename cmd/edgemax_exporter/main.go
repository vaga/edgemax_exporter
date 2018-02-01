package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vaga/edgemax_exporter"
	"github.com/vaga/edgemax_exporter/edgemax"
)

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9132", "host:port for EdgeMAX exporter")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "URL path for surfacing collected metrics")

		address  = flag.String("edgemax.address", "", "address of EdgeMAX Controller API")
		username = flag.String("edgemax.username", "", "username for authentication against EdgeMAX Controller API")
		password = flag.String("edgemax.password", "", "password for authentication against EdgeMAX Controller API")
		insecure = flag.Bool("edgemax.insecure", false, "[optional] do not verify TLS certificate for EdgeMAX Controller API (warning: please use carefully)")
		timeout  = flag.Duration("edgemax.timeout", 5*time.Second, "[optional] timeout for EdgeMAX Controller API requests")
	)
	flag.Parse()

	if *address == "" {
		log.Fatalln("address of EdgeMAX Controller API must be specified with '-edgemax.address' flag")
	}
	if *username == "" {
		log.Fatal("username to authenticate to EdgeMAX Controller API must be specified with '-edgemax.username' flag")
	}
	if *password == "" {
		log.Fatal("password to authenticate to EdgeMAX Controller API must be specified with '-edgemax.password' flag")
	}

	c, err := edgemax.NewClient(*address, newHTTPClient(*timeout, *insecure))
	if err != nil {
		log.Fatalf("cannot create EdgeMAX Controller client: %v", err)
	}
	if err := c.Login(*username, *password); err != nil {
		log.Fatalln("failed to authenticate to EdgeMAX Controller: %v", err)
	}

	e, done, err := edgemax_exporter.New(c)
	if err != nil {
		log.Fatalln("cannot create EdgeMAX Controller exporter: %v", err)
	}
	defer done()

	prometheus.MustRegister(e)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Printf("Starting EdgeMAX exporter on %q for device at %q", *listenAddress, *address)

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Printf("cannot start EdgeMAX exporter: %s", err)
		return
	}
}

func newHTTPClient(timeout time.Duration, insecure bool) *http.Client {
	c := &http.Client{Timeout: timeout}

	if insecure {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	return c
}
