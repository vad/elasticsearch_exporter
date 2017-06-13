package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

	metrics = []*parser.Metric{
		gcPoolCount("young"),
		gcPoolCount("old"),
		gcPoolTime("young"),
		gcPoolTime("old"),
		memPool("young", "used"),
		memPool("old", "used"),
		memPool("young", "max"),
		memPool("old", "max"),
		heap("max"),
		heap("used"),
	}
)

func heap(t string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_memory_heap_%s_bytes", t),
		fmt.Sprintf("%s heap in bytes", t),
		fmt.Sprintf("jvm.mem.heap_%s_in_bytes", t),
		parser.LabelHost,
	)
}

func memPool(pool, t string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_memory_pool_%s_%s_bytes", pool, t),
		fmt.Sprintf("%s memory of pool %s", t, pool),
		fmt.Sprintf("jvm.mem.pools.%s.%s_in_bytes", pool, t),
		parser.LabelHost,
	)

}

func gcPoolTime(pool string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_gc_%s_collection_time_ms", pool),
		fmt.Sprintf("Time of collections of %s GC", pool),
		fmt.Sprintf("jvm.gc.collectors.%s.collection_time_in_millis", pool),
		parser.LabelNodeId,
	)
}

func gcPoolCount(pool string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_gc_%s_collection_count", pool),
		fmt.Sprintf("Number of collections of %s GC", pool),
		fmt.Sprintf("jvm.gc.collectors.%s.collection_count", pool),
		parser.LabelNodeId,
	)
}

func init() {
	prometheus.MustRegister(up)
	for _, metric := range metrics {
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
		for _, metric := range metrics {
			err := metric.Observe(nodeName, jobject)
			if err != nil {
				log.Println("Error observing metric from '", metric.Path, "' ", err.Error())
			}
		}
	}

	up.Set(1)
}

func scrapeForever() {
	ns := strings.TrimRight(*es, "/") + "/_nodes/stats"
	for {
		scrape(ns)

		time.Sleep(time.Duration(*timeInterval) * time.Second)
	}
}

func main() {
	flag.Parse()
	if *timeInterval < 1 {
		log.Fatal("Time interval must be >= 1")
	}

	go scrapeForever()

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Listen on address", *bind)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
