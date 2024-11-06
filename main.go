package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

func main() {
	var config redis.Options
	var addr string

	flag.StringVar(&config.Network, "network", "tcp", "redis network")
	flag.StringVar(&config.Addr, "addr", "0.0.0.0:6379", "redis host:port address")
	flag.StringVar(&config.Username, "username", "", "redis username")
	flag.StringVar(&config.Password, "password", "", "redis password")
	flag.IntVar(&config.DB, "db", 0, "redis db")
	flag.StringVar(&addr, "serve", ":8070", "exporter address")
	flag.Parse()

	if v := os.Getenv("REDIS_ADDR"); v != "" {
		config.Addr = v
	}
	if v := os.Getenv("REDIS_USERNAME"); v != "" {
		config.Username = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		config.Password = v
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
