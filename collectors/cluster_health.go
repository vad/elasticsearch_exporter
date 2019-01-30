package collectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmespath/go-jmespath"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strings"
)

var CollectorAlreadyRegisteredError = errors.New("can't add metrics to an already registered collector")

type ClusterHealthMetric struct {
	path       string
	jsonPath   *jmespath.JMESPath
	descriptor *prometheus.Desc
}

func NewClusterHealthMetric(path string, desc *prometheus.Desc) *ClusterHealthMetric {
	jmesPath := jmespath.MustCompile(path)
	return &ClusterHealthMetric{path, jmesPath, desc}
}

type ClusterHealthCollector struct {
	url         string
	metrics      []*ClusterHealthMetric
	upMetric     *ClusterHealthMetric
	isRegistered bool
}

func NewClusterHealthCollector(clusterUrl string) *ClusterHealthCollector {
	upMetric := &ClusterHealthMetric{
		path: 		"",
		descriptor: prometheus.NewDesc("up", "status of the exporter", nil, nil),
	}
	return &ClusterHealthCollector{
		url:          strings.TrimRight(clusterUrl, "/") + "/_cluster/health",
		metrics:      []*ClusterHealthMetric{upMetric},
		upMetric:     upMetric,
		isRegistered: false,
	}
}

func (c ClusterHealthCollector) AddMetric(path string, desc *prometheus.Desc) error {
	if c.isRegistered {
		return CollectorAlreadyRegisteredError
	}
	c.metrics = append(c.metrics, NewClusterHealthMetric(path, desc))
	return nil
}

func (c ClusterHealthCollector) AddMetrics(metrics map[string]*prometheus.Desc) (errs []error) {
	for path, desc := range metrics {
		err := c.AddMetric(path, desc)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return
}

func (c ClusterHealthCollector) Describe(describers chan<- *prometheus.Desc) {
	c.isRegistered = true
	for _, m := range c.metrics {
		describers <- m.descriptor
	}
}

func (c ClusterHealthCollector) Collect(metrics chan<- prometheus.Metric) {
	resp, err := http.Get(c.url)
	if err != nil {
		metrics <- c.newUpMetric(0)
		log.Println(err.Error())
		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}()

	var jobj interface{}
	err = json.NewDecoder(resp.Body).Decode(&jobj)
	if err != nil {
		metrics <- c.newUpMetric(0)
		log.Println(err.Error())
		return
	}

	for _, m := range c.metrics {
		if m.jsonPath == nil {
			continue
		}
		jresult, err := m.jsonPath.Search(jobj)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		value, ok := jresult.(float64)
		if !ok {
			log.Println(fmt.Sprintf("the value of %s is not a float", m.path))
		}
		metrics <- prometheus.MustNewConstMetric(m.descriptor, prometheus.GaugeValue, value)
	}
}

func (c ClusterHealthCollector) newUpMetric(value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(c.upMetric.descriptor, prometheus.GaugeValue, value)
}
