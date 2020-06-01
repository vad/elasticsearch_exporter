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

```
# HELP es_gc_old_collection_count Number of collections of old GC
# TYPE es_gc_old_collection_count gauge
es_gc_old_collection_count{node="Bobby Drake"} 2
# HELP es_gc_old_collection_time_ms Time of collections of old GC
# TYPE es_gc_old_collection_time_ms gauge
es_gc_old_collection_time_ms{node="Bobby Drake"} 205
# HELP es_gc_young_collection_count Number of collections of young GC
# TYPE es_gc_young_collection_count gauge
es_gc_young_collection_count{node="Bobby Drake"} 5653
# HELP es_gc_young_collection_time_ms Time of collections of young GC
# TYPE es_gc_young_collection_time_ms gauge
es_gc_young_collection_time_ms{node="Bobby Drake"} 263169
# HELP es_indices_docs_count indices.docs.count
# TYPE es_indices_docs_count gauge
es_indices_docs_count{node="Bobby Drake"} 4.26795913e+08
# HELP es_indices_docs_deleted indices.docs.deleted
# TYPE es_indices_docs_deleted gauge
es_indices_docs_deleted{node="Bobby Drake"} 59332
# HELP es_indices_fielddata_memory_size_in_bytes indices.fielddata.memory_size_in_bytes
# TYPE es_indices_fielddata_memory_size_in_bytes gauge
es_indices_fielddata_memory_size_in_bytes{node="Bobby Drake"} 3.570304316e+09
# HELP es_indices_indexing_delete_time_in_millis indices.indexing.delete_time_in_millis
# TYPE es_indices_indexing_delete_time_in_millis gauge
es_indices_indexing_delete_time_in_millis{node="Bobby Drake"} 0
# HELP es_indices_indexing_delete_total indices.indexing.delete_total
# TYPE es_indices_indexing_delete_total gauge
es_indices_indexing_delete_total{node="Bobby Drake"} 0
# HELP es_indices_indexing_index_time_in_millis indices.indexing.index_time_in_millis
# TYPE es_indices_indexing_index_time_in_millis gauge
es_indices_indexing_index_time_in_millis{node="Bobby Drake"} 118569
# HELP es_indices_indexing_index_total indices.indexing.index_total
# TYPE es_indices_indexing_index_total gauge
es_indices_indexing_index_total{node="Bobby Drake"} 533691
# HELP es_indices_merges_total indices.merges.total
# TYPE es_indices_merges_total gauge
es_indices_merges_total{node="Bobby Drake"} 24
# HELP es_indices_merges_total_docs indices.merges.total_docs
# TYPE es_indices_merges_total_docs gauge
es_indices_merges_total_docs{node="Bobby Drake"} 1.0736352e+07
# HELP es_indices_merges_total_size_in_bytes indices.merges.total_size_in_bytes
# TYPE es_indices_merges_total_size_in_bytes gauge
es_indices_merges_total_size_in_bytes{node="Bobby Drake"} 1.1073799804e+10
# HELP es_indices_merges_total_throttled_time_in_millis indices.merges.total_throttled_time_in_millis
# TYPE es_indices_merges_total_throttled_time_in_millis gauge
es_indices_merges_total_throttled_time_in_millis{node="Bobby Drake"} 413282
# HELP es_indices_merges_total_time_in_millis indices.merges.total_time_in_millis
# TYPE es_indices_merges_total_time_in_millis gauge
es_indices_merges_total_time_in_millis{node="Bobby Drake"} 1.120623e+06
# HELP es_indices_query_cache_cache_count indices.query_cache.cache_count
# TYPE es_indices_query_cache_cache_count gauge
es_indices_query_cache_cache_count{node="Bobby Drake"} 7953
# HELP es_indices_query_cache_cache_size indices.query_cache.cache_size
# TYPE es_indices_query_cache_cache_size gauge
es_indices_query_cache_cache_size{node="Bobby Drake"} 7807
# HELP es_indices_query_cache_evictions indices.query_cache.evictions
# TYPE es_indices_query_cache_evictions gauge
es_indices_query_cache_evictions{node="Bobby Drake"} 146
# HELP es_indices_query_cache_hit_count indices.query_cache.hit_count
# TYPE es_indices_query_cache_hit_count gauge
es_indices_query_cache_hit_count{node="Bobby Drake"} 235137
# HELP es_indices_query_cache_memory_size_in_bytes indices.query_cache.memory_size_in_bytes
# TYPE es_indices_query_cache_memory_size_in_bytes gauge
es_indices_query_cache_memory_size_in_bytes{node="Bobby Drake"} 4.43459883e+08
# HELP es_indices_query_cache_miss_count indices.query_cache.miss_count
# TYPE es_indices_query_cache_miss_count gauge
es_indices_query_cache_miss_count{node="Bobby Drake"} 240254
# HELP es_indices_query_cache_total_count indices.query_cache.total_count
# TYPE es_indices_query_cache_total_count gauge
es_indices_query_cache_total_count{node="Bobby Drake"} 475391
# HELP es_indices_recovery_throttle_time_in_millis indices.recovery.throttle_time_in_millis
# TYPE es_indices_recovery_throttle_time_in_millis gauge
es_indices_recovery_throttle_time_in_millis{node="Bobby Drake"} 842912
# HELP es_indices_request_cache_evictions indices.request_cache.evictions
# TYPE es_indices_request_cache_evictions gauge
es_indices_request_cache_evictions{node="Bobby Drake"} 0
# HELP es_indices_request_cache_hit_count indices.request_cache.hit_count
# TYPE es_indices_request_cache_hit_count gauge
es_indices_request_cache_hit_count{node="Bobby Drake"} 3992
# HELP es_indices_request_cache_memory_size_in_bytes indices.request_cache.memory_size_in_bytes
# TYPE es_indices_request_cache_memory_size_in_bytes gauge
es_indices_request_cache_memory_size_in_bytes{node="Bobby Drake"} 5.423273e+06
# HELP es_indices_request_cache_miss_count indices.request_cache.miss_count
# TYPE es_indices_request_cache_miss_count gauge
es_indices_request_cache_miss_count{node="Bobby Drake"} 775
# HELP es_indices_search_fetch_time_in_millis indices.search.fetch_time_in_millis
# TYPE es_indices_search_fetch_time_in_millis gauge
es_indices_search_fetch_time_in_millis{node="Bobby Drake"} 923771
# HELP es_indices_search_fetch_total indices.search.fetch_total
# TYPE es_indices_search_fetch_total gauge
es_indices_search_fetch_total{node="Bobby Drake"} 30349
# HELP es_indices_search_query_time_in_millis indices.search.query_time_in_millis
# TYPE es_indices_search_query_time_in_millis gauge
es_indices_search_query_time_in_millis{node="Bobby Drake"} 1.000684e+06
# HELP es_indices_search_query_total indices.search.query_total
# TYPE es_indices_search_query_total gauge
es_indices_search_query_total{node="Bobby Drake"} 53889
# HELP es_indices_search_scroll_time_in_millis indices.search.scroll_time_in_millis
# TYPE es_indices_search_scroll_time_in_millis gauge
es_indices_search_scroll_time_in_millis{node="Bobby Drake"} 1.12172909e+08
# HELP es_indices_search_scroll_total indices.search.scroll_total
# TYPE es_indices_search_scroll_total gauge
es_indices_search_scroll_total{node="Bobby Drake"} 185
# HELP es_indices_segments_count indices.segments.count
# TYPE es_indices_segments_count gauge
es_indices_segments_count{node="Bobby Drake"} 1243
# HELP es_indices_segments_doc_values_memory_in_bytes indices.segments.doc_values_memory_in_bytes
# TYPE es_indices_segments_doc_values_memory_in_bytes gauge
es_indices_segments_doc_values_memory_in_bytes{node="Bobby Drake"} 1.4657184e+07
# HELP es_indices_segments_index_writer_memory_in_bytes indices.segments.index_writer_memory_in_bytes
# TYPE es_indices_segments_index_writer_memory_in_bytes gauge
es_indices_segments_index_writer_memory_in_bytes{node="Bobby Drake"} 0
# HELP es_indices_segments_memory_in_bytes indices.segments.memory_in_bytes
# TYPE es_indices_segments_memory_in_bytes gauge
es_indices_segments_memory_in_bytes{node="Bobby Drake"} 5.47627844e+08
# HELP es_indices_segments_norms_memory_in_bytes indices.segments.norms_memory_in_bytes
# TYPE es_indices_segments_norms_memory_in_bytes gauge
es_indices_segments_norms_memory_in_bytes{node="Bobby Drake"} 368640
# HELP es_indices_segments_points_memory_in_bytes indices.segments.points_memory_in_bytes
# TYPE es_indices_segments_points_memory_in_bytes gauge
es_indices_segments_points_memory_in_bytes{node="Bobby Drake"} 1.3347019e+07
# HELP es_indices_segments_stored_fields_memory_in_bytes indices.segments.stored_fields_memory_in_bytes
# TYPE es_indices_segments_stored_fields_memory_in_bytes gauge
es_indices_segments_stored_fields_memory_in_bytes{node="Bobby Drake"} 7.8751384e+07
# HELP es_indices_segments_term_vectors_memory_in_bytes indices.segments.term_vectors_memory_in_bytes
# TYPE es_indices_segments_term_vectors_memory_in_bytes gauge
es_indices_segments_term_vectors_memory_in_bytes{node="Bobby Drake"} 7.62704e+06
# HELP es_indices_segments_terms_memory_in_bytes indices.segments.terms_memory_in_bytes
# TYPE es_indices_segments_terms_memory_in_bytes gauge
es_indices_segments_terms_memory_in_bytes{node="Bobby Drake"} 4.32876577e+08
# HELP es_indices_segments_version_map_memory_in_bytes indices.segments.version_map_memory_in_bytes
# TYPE es_indices_segments_version_map_memory_in_bytes gauge
es_indices_segments_version_map_memory_in_bytes{node="Bobby Drake"} 0
# HELP es_indices_warmer_total indices.warmer.total
# TYPE es_indices_warmer_total gauge
es_indices_warmer_total{node="Bobby Drake"} 1275
# HELP es_indices_warmer_total_time_in_millis indices.warmer.total_time_in_millis
# TYPE es_indices_warmer_total_time_in_millis gauge
es_indices_warmer_total_time_in_millis{node="Bobby Drake"} 57770
# HELP es_memory_heap_max_bytes max heap in bytes
# TYPE es_memory_heap_max_bytes gauge
es_memory_heap_max_bytes{node="Bobby Drake"} 1.6071262208e+10
# HELP es_memory_heap_used_bytes used heap in bytes
# TYPE es_memory_heap_used_bytes gauge
es_memory_heap_used_bytes{node="Bobby Drake"} 5.623007504e+09
# HELP es_memory_pool_old_max_bytes max memory of pool old
# TYPE es_memory_pool_old_max_bytes gauge
es_memory_pool_old_max_bytes{node="Bobby Drake"} 1.5757213696e+10
# HELP es_memory_pool_old_used_bytes used memory of pool old
# TYPE es_memory_pool_old_used_bytes gauge
es_memory_pool_old_used_bytes{node="Bobby Drake"} 5.371961944e+09
# HELP es_memory_pool_young_max_bytes max memory of pool young
# TYPE es_memory_pool_young_max_bytes gauge
es_memory_pool_young_max_bytes{node="Bobby Drake"} 2.7918336e+08
# HELP es_memory_pool_young_used_bytes used memory of pool young
# TYPE es_memory_pool_young_used_bytes gauge
es_memory_pool_young_used_bytes{node="Bobby Drake"} 2.49621848e+08
```
