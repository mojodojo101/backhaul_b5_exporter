package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	
	"github.com/mojodojo101/backhaul_b5_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const version string = "0.1.1"

var (
	showVersion   = flag.Bool("version", false, "Print version information.")
	listenAddress = flag.String("listen-address", ":9332", "Address on which to expose metrics.")
	metricsPath   = flag.String("path", "/metrics", "Path under which to expose metrics.")
	snmpTargets   = flag.String("targets", "", "targets to scrape")
	snmpCommunity = flag.String("community", "", "SNMP community")
	configFile    = flag.String("config-file", "config.yml", "Path to config file")
	cfg           *config.Config
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: backhaul_b5_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	c, err := loadConfig()
	if err != nil {
		log.Fatalf("could not load config file. %v", err)
	}
	cfg = c

	if *snmpCommunity == "" {
		snmpCommunity = &cfg.Community
	}

	startServer()
}

func printVersion() {
	fmt.Println("backhaul_b5_exporter")
	fmt.Printf("Version: %s\n", version)
}

func loadConfig() (*config.Config, error) {
	log.Infoln("Loading config from", *configFile)
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	return config.Load(bytes.NewReader(b))
}

func startServer() {
	log.Infof("Starting Backhaul B5 exporter (Version: %s)\n", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Backhaul B5 Exporter (Version ` + version + `)</title></head>
			<body>
			<h1>Backhaul B5 Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/mojodojo101/bakhaul_b5_exporter">github.com/mojodojo101/bakhaul_b5_exporter</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()

	c := newBackhaulB5Collector(cfg)
	reg.MustRegister(c)

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}