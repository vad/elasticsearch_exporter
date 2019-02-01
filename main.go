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

var (
	es           = flag.String("es", "http://localhost:9200", "ES URL")
	bind         = flag.String("bind", ":9092", "Address to bind to")
	timeInterval = flag.Int("time", 5, "Time interval between scrape runs, in seconds")

	up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "es_up",
		Help: "Current status of ES",
	})

	nodeMetrics []*parser.NodeMetric
)

func init() {
	flag.Parse()
	log.Println(fmt.Sprintf("Scraping %s every %d seconds", *es, *timeInterval))

	nodeMetrics = []*parser.NodeMetric{
		parser.NewGcPoolCountMetric("young"),
		parser.NewGcPoolCountMetric("old"),
		parser.NewGcPoolTimeMetric("young"),
		parser.NewGcPoolTimeMetric("old"),
		parser.NewMemPoolMetric("young", "used"),
		parser.NewMemPoolMetric("old", "used"),
		parser.NewMemPoolMetric("young", "max"),
		parser.NewMemPoolMetric("old", "max"),
		parser.NewHeapMetric("max"),
		parser.NewHeapMetric("used"),
		parser.NewRawMetric("indices.merges.total"),
		parser.NewRawMetric("indices.merges.total_time_in_millis"),
		parser.NewRawMetric("indices.merges.total_docs"),
		parser.NewRawMetric("indices.merges.total_size_in_bytes"),
		parser.NewRawMetric("indices.merges.total_throttled_time_in_millis"),
		parser.NewRawMetric("indices.warmer.total"),
		parser.NewRawMetric("indices.warmer.total_time_in_millis"),
		parser.NewRawMetric("indices.fielddata.memory_size_in_bytes"),
		parser.NewRawMetric("indices.segments.count"),
		parser.NewRawMetric("indices.segments.memory_in_bytes"),
		parser.NewRawMetric("indices.segments.terms_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.stored_fields_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.term_vectors_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.norms_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.points_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.doc_values_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.index_writer_memory_in_bytes"),
		parser.NewRawMetric("indices.segments.version_map_memory_in_bytes"),
		parser.NewRawMetric("indices.request_cache.memory_size_in_bytes"),
		parser.NewRawMetric("indices.request_cache.evictions"),
		parser.NewRawMetric("indices.request_cache.hit_count"),
		parser.NewRawMetric("indices.request_cache.miss_count"),
		parser.NewRawMetric("indices.docs.count"),
		parser.NewRawMetric("indices.docs.deleted"),
		parser.NewRawMetric("indices.query_cache.memory_size_in_bytes"),
		parser.NewRawMetric("indices.query_cache.total_count"),
		parser.NewRawMetric("indices.query_cache.hit_count"),
		parser.NewRawMetric("indices.query_cache.miss_count"),
		parser.NewRawMetric("indices.query_cache.cache_size"),
		parser.NewRawMetric("indices.query_cache.cache_count"),
		parser.NewRawMetric("indices.query_cache.evictions"),
		parser.NewRawMetric("indices.recovery.throttle_time_in_millis"),
	}
	nodeMetrics = append(nodeMetrics, parser.NewTotalAndMillisMetrics("indices.search.fetch")...)
	nodeMetrics = append(nodeMetrics, parser.NewTotalAndMillisMetrics("indices.search.query")...)
	nodeMetrics = append(nodeMetrics, parser.NewTotalAndMillisMetrics("indices.search.scroll")...)
	nodeMetrics = append(nodeMetrics, parser.NewTotalAndMillisMetrics("indices.indexing.index")...)
	nodeMetrics = append(nodeMetrics, parser.NewTotalAndMillisMetrics("indices.indexing.delete")...)

	prometheus.MustRegister(collectors.NewClusterHealthCollector(*es))
	for _, metric := range nodeMetrics {
		prometheus.MustRegister(metric.Gauge)
	}
}

func scrape(ns string) {
	resp, err := http.Get(ns)

	if err != nil {
		up.Set(0)
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	v, err := parser.NewNodeStatsJson(resp.Body)

	if err != nil {
		up.Set(0)
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
		for _, metric := range nodeMetrics {
			err := metric.Observe(object)
			if err != nil {
				log.Println("Error observing metric from '", metric.Path, "' ", err.Error())
			}
		}
	}

	up.Set(1)
}

func scrapeForever() {
	ns := strings.TrimRight(*es, "/") + "/_nodes/stats"
	t := time.NewTicker(time.Duration(*timeInterval) * time.Second)
	for range t.C {
		scrape(ns)
	}
}

func main() {
	if *timeInterval < 1 {
		log.Fatal("Time interval must be >= 1")
	}

	go scrapeForever()

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Listen on address", *bind)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
