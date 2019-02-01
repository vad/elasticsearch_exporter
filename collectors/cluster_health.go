package collectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jmespath/go-jmespath"
	"github.com/prometheus/client_golang/prometheus"
)

var collectorAlreadyRegisteredError = errors.New("can't add metrics to an already registered collector")

func identity(i interface{}) (r float64, err error) {
	r, ok := i.(float64)
	if !ok {
		err = errors.New("not a float value")
	}
	return
}

type metricTransform func(interface{}) (float64, error)

type clusterHealthMetric struct {
	path       string
	jsonPath   *jmespath.JMESPath
	descriptor *prometheus.Desc
	transform  metricTransform
}

func newClusterHealthMetric(path string, desc *prometheus.Desc, transform metricTransform) *clusterHealthMetric {
	jmesPath := jmespath.MustCompile(path)
	if transform == nil {
		transform = identity
	}
	return &clusterHealthMetric{path, jmesPath, desc, transform}
}

type ClusterHealthCollector struct {
	url          string
	metrics      []*clusterHealthMetric
	upMetric     *clusterHealthMetric
	isRegistered bool
}

func NewClusterHealthCollector(clusterUrl string) *ClusterHealthCollector {
	c := &ClusterHealthCollector{
		url:          strings.TrimRight(clusterUrl, "/") + "/_cluster/health",
		metrics:      []*clusterHealthMetric{},
		isRegistered: false,
	}

	c.upMetric = &clusterHealthMetric{
		path:       "",
		descriptor: prometheus.NewDesc("es_up", "whether the last call to ES succeeded", nil, prometheus.Labels{"endpoint": c.url}),
	}
	c.metrics = append(c.metrics, c.upMetric)

	c.mustAddStatusMetrics()
	c.MustAddMetric(
		"number_of_nodes",
		prometheus.NewDesc("es_nodes_count", "number of nodes in the cluster", nil, nil),
		nil)
	c.mustAddShardMetric("active_primary")
	c.mustAddShardMetric("active")
	c.mustAddShardMetric("unassigned")
	c.MustAddMetric(
		"active_shards_percent_as_number",
		prometheus.NewDesc("es_active_shards_percent", "percentage of active shards", nil, nil),
		nil)
	return c
}

func (c *ClusterHealthCollector) AddMetric(path string, desc *prometheus.Desc, transform metricTransform) error {
	if c.isRegistered {
		return collectorAlreadyRegisteredError
	}
	c.metrics = append(c.metrics, newClusterHealthMetric(path, desc, transform))
	return nil
}

func (c *ClusterHealthCollector) MustAddMetric(path string, desc *prometheus.Desc, transform metricTransform) {
	err := c.AddMetric(path, desc, transform)
	if err != nil {
		panic(err)
	}
}

func statusTransf(status string) metricTransform {
	return func(i interface{}) (r float64, err error) {
		s, ok := i.(string)
		if !ok {
			err = errors.New("not a string")
			return
		}
		if s == status {
			r = 1
		} else {
			r = 0
		}
		return
	}
}

func (c *ClusterHealthCollector) mustAddStatusMetrics() {
	for _, status := range []string{"green", "yellow", "red"} {
		c.MustAddMetric(
			"status",
			prometheus.NewDesc(
				"es_cluster_status", "status of the cluster", nil, prometheus.Labels{"status": status}),
			statusTransf(status))
	}
}

func (c *ClusterHealthCollector) mustAddShardMetric(shardStatus string) {
	c.MustAddMetric(
		fmt.Sprintf("%s_shards", shardStatus),
		prometheus.NewDesc("es_shards", "count of shards by status", nil, prometheus.Labels{"status": shardStatus}),
		nil)
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

	metrics <- c.newUpMetric(1)

	for _, m := range c.metrics {
		if m.jsonPath == nil {
			continue
		}
		jresult, err := m.jsonPath.Search(jobj)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		value, err := m.transform(jresult)
		if err != nil {
			log.Println(fmt.Sprintf("transform failed for %s: %s", m.path, err.Error()))
		}
		metrics <- prometheus.MustNewConstMetric(m.descriptor, prometheus.GaugeValue, value)
	}
}

func (c ClusterHealthCollector) newUpMetric(value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(c.upMetric.descriptor, prometheus.GaugeValue, value)
}
