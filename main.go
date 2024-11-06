package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

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

	redisClient := redis.NewClient(&config)

	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	http.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		info, err := redisClient.InfoMap(r.Context()).Result()
		if err != nil {
			fmt.Println(err.Error())
		}
		for group, metrics := range info {
			for metric, value := range metrics {
				if strings.HasSuffix(metric, "_human") {
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
				fmt.Fprintf(w, "redis_%s_%s %v\n", strings.ToLower(group), strings.ToLower(metric), v)
			}
		}
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}
