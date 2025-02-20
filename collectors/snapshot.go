package collectors

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SnapshotCollector struct {
	esURL             string
	repo              string
	username          string
	password          string
	snapshotSuccess   *prometheus.Desc
	snapshotTimestamp *prometheus.Desc
	snapshotDuration  *prometheus.Desc
}

func NewSnapshotCollector(esURL, repo, username, password string) *SnapshotCollector {
	return &SnapshotCollector{
		esURL:    esURL,
		repo:     repo,
		username: username,
		password: password,
		snapshotSuccess: prometheus.NewDesc(
			"es_snapshot_last_successful",
			"Indicates whether the last snapshot was successful (1 for success, 0 for failure)",
			nil, nil,
		),
		snapshotTimestamp: prometheus.NewDesc(
			"es_snapshot_last_start_timestamp",
			"Unix timestamp of the last snapshot start time",
			nil, nil,
		),
		snapshotDuration: prometheus.NewDesc(
			"es_snapshot_last_duration_seconds",
			"Duration in seconds of the last snapshot",
			nil, nil,
		),
	}
}

func (sc *SnapshotCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.snapshotSuccess
	ch <- sc.snapshotTimestamp
	ch <- sc.snapshotDuration
}

func (sc *SnapshotCollector) Collect(ch chan<- prometheus.Metric) {
	snapshotURL := fmt.Sprintf("%s/_snapshot/%s/_all", sc.esURL, sc.repo)
	req, err := http.NewRequest("GET", snapshotURL, nil)
	if err != nil {
		log.Printf("Error creating snapshot request: %v", err)
		ch <- prometheus.MustNewConstMetric(sc.snapshotSuccess, prometheus.GaugeValue, 0)
		return
	}
	if sc.username != "" && sc.password != "" {
		req.SetBasicAuth(sc.username, sc.password)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching snapshot status: %v", err)
		ch <- prometheus.MustNewConstMetric(sc.snapshotSuccess, prometheus.GaugeValue, 0)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Snapshots []struct {
			State     string `json:"state"`
			StartTime int64  `json:"start_time_in_millis"`
			Duration  int64  `json:"duration_in_millis"`
		} `json:"snapshots"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Error decoding snapshot JSON: %v", err)
		ch <- prometheus.MustNewConstMetric(sc.snapshotSuccess, prometheus.GaugeValue, 0)
		return
	}

	if len(data.Snapshots) == 0 {
		log.Printf("No snapshots found in repository %s", sc.repo)
		ch <- prometheus.MustNewConstMetric(sc.snapshotSuccess, prometheus.GaugeValue, 0)
		return
	}

	lastSnapshot := data.Snapshots[len(data.Snapshots)-1]

	startTime := float64(lastSnapshot.StartTime / 1000.0)

	duration := float64(lastSnapshot.Duration / 1000.0)

	success := 0.0
	if lastSnapshot.State == "SUCCESS" {
		success = 1.0
	}

	ch <- prometheus.MustNewConstMetric(sc.snapshotSuccess, prometheus.GaugeValue, success)
	ch <- prometheus.MustNewConstMetric(sc.snapshotTimestamp, prometheus.GaugeValue, startTime)
	ch <- prometheus.MustNewConstMetric(sc.snapshotDuration, prometheus.GaugeValue, duration)
}
