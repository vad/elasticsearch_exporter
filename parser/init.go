package parser

import (
	"encoding/json"
	"io"
)

type JvmPoolStats struct {
	UsedInBytes float64 `json:"used_in_bytes"`
	MaxInBytes  float64 `json:"max_in_bytes"`
}

type JvmPoolsStats struct {
	Young    JvmPoolStats `json:"young"`
	Old      JvmPoolStats `json:"old"`
	Survivor JvmPoolStats `json:"survivor"`
}

type JvmMemStats struct {
	HeapUsedInBytes float64       `json:"heap_used_in_bytes"`
	HeapMaxInBytes  float64       `json:"heap_max_in_bytes"`
	Pools           JvmPoolsStats `json:"pools"`
}

type JvmCollectorStats struct {
	CollectionCount        float64 `json:"collection_count"`
	CollectionTimeInMillis float64 `json:"collection_time_in_millis"`
}

type JvmGcCollectorsStats struct {
	Young JvmCollectorStats `json:"young"`
	Old   JvmCollectorStats `json:"old"`
}

type JvmGcStats struct {
	Collectors JvmGcCollectorsStats `json:"collectors"`
}

type JvmStats struct {
	Mem JvmMemStats `json:"mem"`
	Gc  JvmGcStats  `json:"gc"`
}

type NodeStats struct {
	Host string   `json:"host"`
	Jvm  JvmStats `json:"jvm"`
}

type NodeStatsJson struct {
	Nodes map[string]NodeStats `json:"nodes"`
}

func NewNodeStatsJson(r io.Reader) (*NodeStatsJson, error) {
	d := json.NewDecoder(r)

	v := &NodeStatsJson{}
	err := d.Decode(v)
	return v, err
}
