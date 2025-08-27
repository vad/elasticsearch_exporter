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

type Metric interface {
	Observe(interface{}) error
	String() string
}

type NodeMetric struct {
	Path  string
	Gauge *prometheus.GaugeVec
}

func NewNodeMetric(name string, desc string, path string) *NodeMetric {
	gv := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: desc,
		},
		[]string{"node"},
	)

	return &NodeMetric{
		Path:  path,
		Gauge: gv,
	}
}

func (m *NodeMetric) Observe(object interface{}) error {
	jresult, err := jmespath.Search(m.Path, object)
	if err != nil {
		return err
	}
	value, ok := jresult.(float64)
	if !ok {
		return fmt.Errorf("the value of %s is not a float", m.Path)
	}

	jlabel, err := jmespath.Search("host", object)
	if err != nil {
		return err
	}
	label, ok := jlabel.(string)
	if !ok {
		return errors.New("host label is not a string")
	}
	m.Gauge.WithLabelValues(label).Set(value)
	return nil
}

func (m *NodeMetric) String() string {
	return m.Path
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

func NewThreadPoolMetrics(pool string) []*NodeMetric {
	obs := []string{"threads", "queue", "active", "rejected", "largest", "completed"}
	out := make([]*NodeMetric, 0, len(obs))

	for _, s := range obs {
		out = append(out,
			NewNodeMetric(
				fmt.Sprintf("thread_pool_%s_%s", pool, s),
				"See thread_pool ES doc",
				fmt.Sprintf("thread_pool.%s.%s", pool, s),
			),
		)
	}
	return out
}

func NewTotalAndMillisMetrics(m string) []*NodeMetric {
	out := make([]*NodeMetric, 2)
	out[0] = NewRawMetric(m + "_total")
	out[1] = NewRawMetric(m + "_time_in_millis")
	return out
}

// siren metrics

type SirenMemoryMetric struct {
	Peak  *prometheus.GaugeVec
	Limit *prometheus.GaugeVec
}

func NewSirenMemoryMetric() *SirenMemoryMetric {
	peak := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "es_siren_federate_memory_peak",
			Help: "Peak memory usage of Siren Federate off-heap storage",
		},
		[]string{"node"},
	)
	limit := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "es_siren_federate_memory_limit",
			Help: "Memory limit of Siren Federate off-heap storage",
		},
		[]string{"node"},
	)

	return &SirenMemoryMetric{
		Peak:  peak,
		Limit: limit,
	}
}

func (m SirenMemoryMetric) Observe(object interface{}) error {
	jresult, err := jmespath.Search("memory.root_allocator_dump_peak_in_bytes", object)
	if err != nil {
		return err
	}
	val, ok := jresult.(float64)
	if !ok {
		return errors.New("Cannot find root allocator dump peak")
	}
	jlabel, err := jmespath.Search("host", object)
	if err != nil {
		return err
	}
	label, ok := jlabel.(string)
	if !ok {
		return errors.New("host label is not a string")
	}
	m.Peak.WithLabelValues(label).Set(val)

	jresult, err = jmespath.Search("memory.root_allocator_dump_limit_in_bytes", object)
	if err != nil {
		return err
	}
	val, ok = jresult.(float64)
	if !ok {
		return errors.New("Cannot find root allocator dump limit")
	}
	m.Limit.WithLabelValues(label).Set(val)

	return nil
}

func (m SirenMemoryMetric) String() string {
	return "Siren Federate Metrics"
}

type SirenLicenseMetric struct {
	Valid prometheus.Gauge
}

func NewSirenLicenseMetric() *SirenLicenseMetric {
	valid := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "siren_license_valid",
			Help: "Siren license validation status (1 if valid, 0 if invalid)",
		},
	)

	return &SirenLicenseMetric{
		Valid: valid,
	}
}

func (m SirenLicenseMetric) Observe(object interface{}) error {
	jresult, err := jmespath.Search("license_validation.is_valid", object)
	if err != nil {
		return err
	}
	val, ok := jresult.(bool)
	if !ok {
		return errors.New("cannot find license_validation.is_valid boolean value")
	}

	if val {
		m.Valid.Set(1)
	} else {
		m.Valid.Set(0)
	}

	return nil
}

func (m SirenLicenseMetric) String() string {
	return "Siren License Validation"
}
