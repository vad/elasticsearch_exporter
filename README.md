# elasticsearch_exporter
ElasticSearch cluster health and node stats exporter for Prometheus

## Usage

```bash
$ ./elasticsearch_exporter --es=https://your-es-url:port
```

### Options

 - `-es`: URL to ElasticSearch, default: `http://localhost:9200`
 - `-bind`: Address to listen on, default: `:9092`
 - `-time`: Scraping interval in seconds (applies to node stats), default `5`
 - `-username`: Set the username to be used for the request when XPack is enabled
 - `-password`: Used in conjuction with `username`, set its password
 - `-enable-siren`: Enable Siren Federate plugin scraping

## Metrics

The exporter collects metrics from `/_cluster/health` and from `/_nodes/stats`.

### Cluster health

```
# HELP es_active_shards_percent percentage of active shards
# TYPE es_active_shards_percent gauge
es_active_shards_percent 100
# HELP es_cluster_status status of the cluster
# TYPE es_cluster_status gauge
es_cluster_status{status="green"} 1
es_cluster_status{status="red"} 0
es_cluster_status{status="yellow"} 0
# HELP es_nodes_count number of nodes in the cluster
# TYPE es_nodes_count gauge
es_nodes_count 4
# HELP es_shards count of shards by status
# TYPE es_shards gauge
es_shards{status="active"} 348
es_shards{status="active_primary"} 178
es_shards{status="unassigned"} 0
# HELP up whether the last call to ES succeeded
# TYPE up gauge
es_up{endpoint="http://localhost:9200/_cluster/health"} 1
```


### Node stats

By default, this exporter exposes a minimal set of node level metrics. Currently only the recovery throttle time is collected.

```
# HELP es_indices_recovery_throttle_time_in_millis indices.recovery.throttle_time_in_millis
# TYPE es_indices_recovery_throttle_time_in_millis gauge
es_indices_recovery_throttle_time_in_millis{node="node1"} 0
```

If started with `-enable-siren`, additional Siren Federate metrics are exported:

```
# HELP es_siren_up Current status of Siren Federate
# TYPE es_siren_up gauge
es_siren_up 1
# HELP es_siren_federate_memory_peak Peak memory usage of Siren Federate off-heap storage
# TYPE es_siren_federate_memory_peak gauge
es_siren_federate_memory_peak{node="node1"} 0
# HELP es_siren_federate_memory_limit Memory limit of Siren Federate off-heap storage
# TYPE es_siren_federate_memory_limit gauge
es_siren_federate_memory_limit{node="node1"} 0
# HELP es_siren_license_valid Siren license validation status (1 if valid, 0 if invalid)
# TYPE es_siren_license_valid gauge
es_siren_license_valid 1
```
