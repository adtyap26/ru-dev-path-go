package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/api"
	"redisolar-go/internal/config"
	redisdao "redisolar-go/internal/dao/redis"
	"redisolar-go/internal/keyschema"
)

func main() {
	cfg := config.Load()

	// Determine static dir relative to the binary or working directory
	staticDir := findStaticDir()

	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
	})

	ks := keyschema.New(cfg.RedisKeyPrefix)
	base := redisdao.NewRedisDao(client, ks)

	deps := api.Deps{
		SiteDao:         redisdao.NewSiteDao(base),
		SiteGeoDao:      redisdao.NewSiteGeoDao(base),
		CapacityDao:     redisdao.NewCapacityReportDao(base),
		MetricDao:       redisdao.NewMetricTimeseriesDao(base),
		FeedDao:         redisdao.NewFeedDao(base),
		MeterReadingDao: redisdao.NewMeterReadingDao(base),
		UseGeoSiteAPI:   cfg.UseGeoSiteAPI,
		StaticDir:       staticDir,
	}

	router := api.NewRouter(deps)

	addr := ":" + cfg.ServerPort
	log.Printf("Starting server on %s (geo=%v, prefix=%s, redis=%s:%s)",
		addr, cfg.UseGeoSiteAPI, cfg.RedisKeyPrefix, cfg.RedisHost, cfg.RedisPort)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

func findStaticDir() string {
	// Try relative to working directory first
	candidates := []string{
		"static",
		"../static",
		filepath.Join(os.Args[0], "..", "..", "static"),
	}

	// Also try relative to the executable
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "static"))
	}

	for _, dir := range candidates {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}

	// Default fallback
	return "static"
}
