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

	metrics []*parser.Metric
)

func init() {
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
		raw("indices.merges.total"),
		raw("indices.merges.total_time_in_millis"),
		raw("indices.merges.total_docs"),
		raw("indices.merges.total_size_in_bytes"),
		raw("indices.merges.total_throttled_time_in_millis"),
		raw("indices.warmer.total"),
		raw("indices.warmer.total_time_in_millis"),
		raw("indices.fielddata.memory_size_in_bytes"),
		raw("indices.segments.count"),
		raw("indices.segments.memory_in_bytes"),
		raw("indices.segments.terms_memory_in_bytes"),
		raw("indices.segments.stored_fields_memory_in_bytes"),
		raw("indices.segments.term_vectors_memory_in_bytes"),
		raw("indices.segments.norms_memory_in_bytes"),
		raw("indices.segments.points_memory_in_bytes"),
		raw("indices.segments.doc_values_memory_in_bytes"),
		raw("indices.segments.index_writer_memory_in_bytes"),
		raw("indices.segments.version_map_memory_in_bytes"),
		raw("indices.request_cache.memory_size_in_bytes"),
		raw("indices.request_cache.evictions"),
		raw("indices.request_cache.hit_count"),
		raw("indices.request_cache.miss_count"),
		raw("indices.docs.count"),
		raw("indices.docs.deleted"),
		raw("indices.query_cache.memory_size_in_bytes"),
		raw("indices.query_cache.total_count"),
		raw("indices.query_cache.hit_count"),
		raw("indices.query_cache.miss_count"),
		raw("indices.query_cache.cache_size"),
		raw("indices.query_cache.cache_count"),
		raw("indices.query_cache.evictions"),
		raw("indices.recovery.throttle_time_in_millis"),
	}
	addToMetrics(totalAndMillis("indices.search.fetch"))
	addToMetrics(totalAndMillis("indices.search.query"))
	addToMetrics(totalAndMillis("indices.search.scroll"))
	addToMetrics(totalAndMillis("indices.indexing.index"))
	addToMetrics(totalAndMillis("indices.indexing.delete"))

	prometheus.MustRegister(up)
	for _, metric := range metrics {
		prometheus.MustRegister(metric.Gauge)
	}
}

func heap(t string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_memory_heap_%s_bytes", t),
		fmt.Sprintf("%s heap in bytes", t),
		fmt.Sprintf("jvm.mem.heap_%s_in_bytes", t),
	)
}

func memPool(pool, t string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_memory_pool_%s_%s_bytes", pool, t),
		fmt.Sprintf("%s memory of pool %s", t, pool),
		fmt.Sprintf("jvm.mem.pools.%s.%s_in_bytes", pool, t),
	)
}

func gcPoolTime(pool string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_gc_%s_collection_time_ms", pool),
		fmt.Sprintf("Time of collections of %s GC", pool),
		fmt.Sprintf("jvm.gc.collectors.%s.collection_time_in_millis", pool),
	)
}

func gcPoolCount(pool string) *parser.Metric {
	return parser.NewMetric(
		fmt.Sprintf("es_gc_%s_collection_count", pool),
		fmt.Sprintf("Number of collections of %s GC", pool),
		fmt.Sprintf("jvm.gc.collectors.%s.collection_count", pool),
	)
}

func raw(op string) *parser.Metric {
	o := strings.Replace(op, ".", "_", -1)
	return parser.NewMetric(fmt.Sprintf("es_%s", o), op, op)
}

func totalAndMillis(m string) []*parser.Metric {
	var out []*parser.Metric

	out = make([]*parser.Metric, 2)
	out[0] = raw(m + "_total")
	out[1] = raw(m + "_time_in_millis")
	return out
}

func addToMetrics(m []*parser.Metric) {
	metrics = append(metrics, m...)
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
