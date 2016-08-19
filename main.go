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
	es   = flag.String("es", "http://localhost:9200", "ES URL")
	bind = flag.String("bind", ":9092", "Address to bind to")

	up = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "es_up",
		Help: "Current status of ES",
	})

	MemoryPoolYoungUsedBytes    = pool_gauge_vec("young", "used")
	MemoryPoolOldUsedBytes      = pool_gauge_vec("old", "used")
	MemoryPoolSurvivorUsedBytes = pool_gauge_vec("survivor", "used")
	MemoryPoolYoungMaxBytes     = pool_gauge_vec("young", "max")
	MemoryPoolOldMaxBytes       = pool_gauge_vec("old", "max")
	MemoryPoolSurvivorMaxBytes  = pool_gauge_vec("survivor", "max")

	MemoryHeapUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "es_memory_heap_used_bytes",
			Help: "Current heap in bytes",
		},
		[]string{"node"},
	)
	MemoryHeapMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "es_memory_heap_max_bytes",
			Help: "Max heap in bytes",
		},
		[]string{"node"},
	)

	GcYoungCount = gc_count_gauge_vec("young")
	GcOldCount   = gc_count_gauge_vec("old")
	GcYoungTime  = gc_time_gauge_vec("young")
	GcOldTime    = gc_time_gauge_vec("old")
)

func pool_gauge_vec(pool, t string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("es_memory_pool_%s_%s_bytes", pool, t),
			Help: fmt.Sprintf("%s memory of pool %s", t, pool),
		},
		[]string{"node"},
	)
}

func gc_count_gauge_vec(pool string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("es_gc_%s_collection_count", pool),
			Help: fmt.Sprintf("Number of collections of %s GC", pool),
		},
		[]string{"node"},
	)
}

func gc_time_gauge_vec(pool string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("es_gc_%s_collection_time_ms", pool),
			Help: fmt.Sprintf("Time of collections of %s GC", pool),
		},
		[]string{"node"},
	)
}

func init() {
	prometheus.MustRegister(up)

	prometheus.MustRegister(MemoryPoolYoungUsedBytes)
	prometheus.MustRegister(MemoryPoolOldUsedBytes)
	prometheus.MustRegister(MemoryPoolSurvivorUsedBytes)
	prometheus.MustRegister(MemoryPoolYoungMaxBytes)
	prometheus.MustRegister(MemoryPoolOldMaxBytes)
	prometheus.MustRegister(MemoryPoolSurvivorMaxBytes)

	prometheus.MustRegister(MemoryHeapUsed)
	prometheus.MustRegister(MemoryHeapMax)

	prometheus.MustRegister(GcYoungCount)
	prometheus.MustRegister(GcOldCount)
	prometheus.MustRegister(GcYoungTime)
	prometheus.MustRegister(GcOldTime)
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

	for _, n := range v.Nodes {
		h := n.Host

		pools := n.Jvm.Mem.Pools
		MemoryPoolYoungUsedBytes.WithLabelValues(h).Set(pools.Young.UsedInBytes)
		MemoryPoolYoungMaxBytes.WithLabelValues(h).Set(pools.Young.MaxInBytes)
		MemoryPoolOldUsedBytes.WithLabelValues(h).Set(pools.Old.UsedInBytes)
		MemoryPoolOldMaxBytes.WithLabelValues(h).Set(pools.Old.MaxInBytes)
		MemoryPoolSurvivorUsedBytes.WithLabelValues(h).Set(pools.Survivor.UsedInBytes)
		MemoryPoolSurvivorMaxBytes.WithLabelValues(h).Set(pools.Survivor.MaxInBytes)

		MemoryHeapUsed.WithLabelValues(h).Set(n.Jvm.Mem.HeapUsedInBytes)
		MemoryHeapMax.WithLabelValues(h).Set(n.Jvm.Mem.HeapMaxInBytes)

		GcYoungCount.WithLabelValues(h).Set(n.Jvm.Gc.Collectors.Young.CollectionCount)
		GcOldCount.WithLabelValues(h).Set(n.Jvm.Gc.Collectors.Old.CollectionCount)
		GcYoungTime.WithLabelValues(h).Set(n.Jvm.Gc.Collectors.Young.CollectionTimeInMillis)
		GcOldTime.WithLabelValues(h).Set(n.Jvm.Gc.Collectors.Old.CollectionTimeInMillis)
	}

	up.Set(1)
}

func scrapeForever() {
	ns := strings.TrimRight(*es, "/") + "/_nodes/stats"
	for {
		scrape(ns)

		time.Sleep(5 * time.Second)
	}
}

func main() {
	flag.Parse()

	go scrapeForever()

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Listen on address", *bind)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
