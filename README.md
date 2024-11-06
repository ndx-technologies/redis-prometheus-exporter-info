Redis Prometheus Exporter `INFO`

* 50 lines of code, no dependencies (except `redis/v9`)
* exports `INFO` command numeric values

```bash
redis_replication_master_repl_offset 0
redis_replication_repl_backlog_first_byte_offset 0
redis_replication_master_replid2 0
redis_replication_second_repl_offset -1
redis_replication_repl_backlog_active 0
redis_replication_repl_backlog_size 1.048576e+06
redis_replication_repl_backlog_histlen 0
redis_replication_connected_slaves 0
redis_cluster_cluster_enabled 0
redis_memory_allocator_frag_bytes 6.963e+06
redis_memory_rss_overhead_ratio 2.23
...
```

## References

* https://github.com/oliver006/redis_exporter
* https://redis.io/docs/latest/integrate/prometheus-with-redis-enterprise/prometheus-metrics-definitions/
* https://github.com/prometheus/docs/blob/main/content/docs/instrumenting/exposition_formats.md
