package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vad/elasticsearch_exporter/collectors"
	"github.com/vad/elasticsearch_exporter/parser"
)

type Metrics []parser.Metric

var (
	es           = flag.String("es", "http://localhost:9200", "ES URL")
	bind         = flag.String("bind", ":9092", "Address to bind to")
	timeInterval = flag.Int("time", 5, "Time interval between scrape runs, in seconds")
	username     = flag.String("username", "", "Username when XPack security is enabled")
	password     = flag.String("password", "", "Password for the user when XPack security is enabled")
	enableSiren  = flag.Bool("enable-siren", false, "Enable Siren Federate Plugin scraping")

	up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "es_up",
		Help: "Current status of ES",
	})
	sirenUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "es_up",
		Help: "Current status of Siren Federate",
	})

	client = &http.Client{
		Timeout: time.Second * 10,
	}
)

func init() {
	flag.Parse()
	log.Println(fmt.Sprintf("Scraping %s every %d seconds", *es, *timeInterval))

	if *timeInterval < 1 {
		log.Fatal("Time interval must be >= 1")
	}
}

func queryEs(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if len(*username) > 0 && len(*password) > 0 {
		req.SetBasicAuth(*username, *password)
	}

	return client.Do(req)
}

func scrape(ns string, upMetric prometheus.Gauge, metrics []parser.Metric) {
	resp, err := queryEs(ns)
	if err != nil {
		upMetric.Set(0)
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	v, err := parser.NewNodeStatsJson(resp.Body)
	if err != nil {
		upMetric.Set(0)
		log.Println("Error decoding ES JSON:", err.Error())
		return
	}

	for nodeName, jobject := range v.Nodes {
		var object interface{}
		err := json.Unmarshal(*jobject, &object)
		if err != nil {
			log.Println("Error decoding JSON for node", nodeName, ":", err.Error())
			continue
		}
		for _, metric := range metrics {
			err := metric.Observe(object)
			if err != nil {
				log.Println("Error observing metric from '", metric.String(), "' ", err.Error())
			}
		}
	}

	upMetric.Set(1)
}

func scrapeForever(endpoint string, up prometheus.Gauge, metrics []parser.Metric) {
	scrapeUrl := strings.TrimRight(*es, "/") + endpoint
	t := time.NewTicker(time.Duration(*timeInterval) * time.Second)
	for range t.C {
		scrape(scrapeUrl, up, metrics)
	}
}

func main() {
	// plain es
	nm := []*parser.NodeMetric{
		parser.NewRawMetric("indices.recovery.throttle_time_in_millis"),
	}

	nodeMetrics := make([]parser.Metric, len(nm))
	prometheus.MustRegister(collectors.NewClusterHealthCollector(*es, *username, *password))
	for i, metric := range nm {
		prometheus.MustRegister(metric.Gauge)
		nodeMetrics[i] = metric
	}
	go scrapeForever("/_nodes/stats", up, nodeMetrics)

	// siren
	if *enableSiren {
		m := parser.NewSirenMemoryMetric()
		prometheus.MustRegister(m.Peak)
		prometheus.MustRegister(m.Limit)
		sirenMetrics := []parser.Metric{m}

		go scrapeForever("/_siren/nodes/stats", sirenUp, sirenMetrics)
	}

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Listen on address", *bind)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
