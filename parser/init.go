package parser

import (
	"encoding/json"
	"io"

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

type Metric struct {
	Path  string
	Gauge *prometheus.GaugeVec
}

func NewMetric(name string, desc string, path string) *Metric {
	return &Metric{
		Path: path,
		Gauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name,
				Help: desc,
			},
			[]string{"node"}),
	}
}

func (metric Metric) Observe(nodeName string, jobject *json.RawMessage) error {
	var result interface{}
	err := json.Unmarshal(*jobject, &result)
	if err != nil {
		return err
	}
	jresult, err := jmespath.Search(metric.Path, result)
	if err != nil {
		return err
	}
	value := jresult.(float64)

	jlabel, err := jmespath.Search("host", result)
	if err != nil {
		return err
	}
	label := jlabel.(string)
	metric.Gauge.WithLabelValues(label).Set(value)
	return nil

}
