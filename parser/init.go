package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jmespath/go-jmespath"
	"github.com/prometheus/client_golang/prometheus"
)

type NodeStatsJson struct {
	Nodes map[string]*json.RawMessage `json:"nodes"`
}

func NewNodeStatsJson(r io.Reader) (*NodeStatsJson, error) {
	d := json.NewDecoder(r)

	v := &NodeStatsJson{}
	err := d.Decode(v)
	return v, err
}

type NodeMetric struct {
	Path  string
	Gauge *prometheus.GaugeVec
}

func NewNodeMetric(name string, desc string, path string) *NodeMetric {
	return &NodeMetric{
		Path: path,
		Gauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name,
				Help: desc,
			},
			[]string{"node"}),
	}
}

func (metric NodeMetric) Observe(object interface{}) error {
	jresult, err := jmespath.Search(metric.Path, object)
	if err != nil {
		return err
	}
	value, ok := jresult.(float64)
	if !ok {
		return errors.New(fmt.Sprintf("the value of %s is not a float", metric.Path))
	}
	
	jlabel, err := jmespath.Search("host", object)
	if err != nil {
		return err
	}
	label, ok := jlabel.(string)
	if !ok {
		return errors.New("host label is not a string")
	}
	metric.Gauge.WithLabelValues(label).Set(value)
	return nil

}

func NewHeapMetric(t string) *NodeMetric {
	return NewNodeMetric(
		fmt.Sprintf("es_memory_heap_%s_bytes", t),
		fmt.Sprintf("%s heap in bytes", t),
		fmt.Sprintf("jvm.mem.heap_%s_in_bytes", t),
	)
}

func NewMemPoolMetric(pool, t string) *NodeMetric {
	return NewNodeMetric(
		fmt.Sprintf("es_memory_pool_%s_%s_bytes", pool, t),
		fmt.Sprintf("%s memory of pool %s", t, pool),
		fmt.Sprintf("jvm.mem.pools.%s.%s_in_bytes", pool, t),
	)
}

func NewGcPoolTimeMetric(pool string) *NodeMetric {
	return NewNodeMetric(
		fmt.Sprintf("es_gc_%s_collection_time_ms", pool),
		fmt.Sprintf("Time of collections of %s GC", pool),
		fmt.Sprintf("jvm.gc.collectors.%s.collection_time_in_millis", pool),
	)
}

func NewGcPoolCountMetric(pool string) *NodeMetric {
	return NewNodeMetric(
		fmt.Sprintf("es_gc_%s_collection_count", pool),
		fmt.Sprintf("Number of collections of %s GC", pool),
		fmt.Sprintf("jvm.gc.collectors.%s.collection_count", pool),
	)
}

func NewRawMetric(op string) *NodeMetric {
	o := strings.Replace(op, ".", "_", -1)
	return NewNodeMetric(fmt.Sprintf("es_%s", o), op, op)
}

func NewTotalAndMillisMetrics(m string) []*NodeMetric {
	var out []*NodeMetric

	out = make([]*NodeMetric, 2)
	out[0] = NewRawMetric(m + "_total")
	out[1] = NewRawMetric(m + "_time_in_millis")
	return out
}

