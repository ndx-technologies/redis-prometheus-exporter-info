package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

//go:embed metrics.yaml
var defaultMetricsConfig []byte

type Metric struct {
	Help string `yaml:"help"`
	Type string `yaml:"type"`
}

func PrintMetric(w io.Writer, name string, config Metric, value any, ts time.Time) {
	fmt.Fprintf(w, "# HELP %s %s\n", name, config.Help)
	fmt.Fprintf(w, "# TYPE %s %s\n", name, config.Type)
	fmt.Fprintf(w, "%s %v %d\n\n", name, value, ts.UnixMilli())
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

	metricConfig := make(map[string]Metric)
	if err := yaml.Unmarshal(defaultMetricsConfig, &metricConfig); err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&config)

	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	http.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		info, err := redisClient.InfoMap(r.Context()).Result()
		if err != nil {
			fmt.Println(err.Error())
		}
		ts := time.Now()
		info[""] = map[string]string{"up": "1"}
		for group, metrics := range info {
			for metric, value := range metrics {
				metric = strings.ToLower(group) + "_" + strings.ToLower(metric)
				config, ok := metricConfig[metric]
				if !ok {
					continue
				}
				var v any
				if vInt, err := strconv.Atoi(value); err == nil {
					v = vInt
				}
				if vFloat, err := strconv.ParseFloat(value, 64); err == nil {
					v = vFloat
				}
				if v == nil {
					continue
				}
				PrintMetric(w, "redis_"+metric, config, v, ts)
			}
		}
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}
