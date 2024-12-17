package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

//go:embed metrics.yaml
var defaultMetricsConfig []byte

type Metric struct {
	Help string `yaml:"help"`
	Type string `yaml:"type"`
}

func PrintMetric(w io.Writer, name string, config Metric, labels map[string]string, value float64) {
	fmt.Fprintf(w, "# HELP %s %s\n", name, config.Help)
	fmt.Fprintf(w, "# TYPE %s %s\n", name, config.Type)
	fmt.Fprintf(w, "%s %.f\n\n", name+SprintLabels(labels), value)
}

func SprintLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, k+`="`+v+`"`)
	}

	return "{" + strings.Join(parts, ",") + "}"
}

func main() {
	config := redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

	addr := os.Getenv("SERVE")
	if addr == "" {
		addr = ":8070"
	}

	var indexesStr string
	flag.StringVar(&indexesStr, "indexes", "", "comma separated list of indexes for `FT.INFO`")
	flag.Parse()

	indexes := strings.Split(indexesStr, ",")

	metricConfig := make(map[string]Metric)
	if err := yaml.Unmarshal(defaultMetricsConfig, &metricConfig); err != nil {
		log.Fatal(err)
	}

	if len(indexes) > 0 {
		config.Protocol = 2
		config.UnstableResp3 = true
	}

	redisClient := redis.NewClient(&config)

	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	http.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		info, err := redisClient.InfoMap(r.Context()).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		PrintMetric(w, "redis_up", metricConfig["up"], nil, 1)

		for group, metrics := range info {
			for metric, value := range metrics {
				metric = strings.ToLower(group) + "_" + strings.ToLower(metric)

				config, ok := metricConfig[metric]
				if !ok {
					continue
				}
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					PrintMetric(w, "redis_"+metric, config, nil, v)
				}
			}
		}

		for _, index := range indexes {
			info, err := redisClient.FTInfo(r.Context(), index).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for metric, v := range FTInfoToMetrics(info) {
				config, ok := metricConfig["index_"+metric]
				if !ok {
					continue
				}
				PrintMetric(w, "redis_index_"+metric, config, map[string]string{"index": index}, v)
			}
		}
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}

func FTInfoToMetrics(v redis.FTInfoResult) map[string]float64 {
	return map[string]float64{
		"indexing_error_count":          float64(v.IndexErrors.IndexingFailures),
		"doc_count":                     float64(v.NumDocs),
		"doc_table_size_mb":             v.DocTableSizeMB,
		"record_count":                  float64(v.NumRecords),
		"term_count":                    float64(v.NumTerms),
		"use_count":                     float64(v.NumberOfUses),
		"cleaning":                      float64(v.Cleaning),
		"cursor_global_idle_count":      float64(v.CursorStats.GlobalIdle),
		"cursor_global_count":           float64(v.CursorStats.GlobalTotal),
		"cursor_index_capacity_count":   float64(v.CursorStats.IndexCapacity),
		"cursor_index_count":            float64(v.CursorStats.IndexTotal),
		"gc_collected_bytes":            float64(v.GCStats.BytesCollected),
		"gc_run_ms":                     float64(v.GCStats.TotalMsRun),
		"gc_cycles_count":               float64(v.GCStats.TotalCycles),
		"gc_last_run_time_ms":           float64(v.GCStats.LastRunTimeMs),
		"gc_numeric_trees_missed_count": float64(v.GCStats.GCNumericTreesMissed),
		"gc_blocks_denied_count":        float64(v.GCStats.GCBlocksDenied),
		"geo_shapes_size_mb":            v.GeoshapesSzMB,
		"hash_indexing_failures_count":  float64(v.HashIndexingFailures),
		"indexing":                      float64(v.Indexing),
		"inverted_size_mb":              float64(v.InvertedSzMB),
		"key_table_size_mb":             v.KeyTableSizeMB,
		"offset_vectors_size_mb":        v.OffsetVectorsSzMB,
		"percent_indexed":               v.PercentIndexed,
		"sortable_values_size_mb":       v.SortableValuesSizeMB,
		"tag_overhead_size_mb":          v.TagOverheadSzMB,
		"text_overhead_size_mb":         v.TextOverheadSzMB,
		"memory_size_mb":                v.TotalIndexMemorySzMB,
		"indexing_time_seconds":         float64(v.TotalIndexingTime),
		"inverted_index_block_count":    float64(v.TotalInvertedIndexBlocks),
		"vector_index_size_mb":          v.VectorIndexSzMB,
	}
}
