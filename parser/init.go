package parser

import (
	"encoding/json"
	"fmt"
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

type LabelType int

const (
	LabelHost LabelType = iota
	LabelNodeId
)

type Metric struct {
	Path  string
	Label LabelType
	Gauge *prometheus.GaugeVec
}

func NewMetric(name string, desc string, path string, label LabelType) *Metric {
	return &Metric{
		Path:  path,
		Label: label,
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
	if metric.Label == LabelNodeId {
		label = fmt.Sprintf("%s-%s", label, nodeName)
	}
	metric.Gauge.WithLabelValues(label).Set(value)
	return nil

}
