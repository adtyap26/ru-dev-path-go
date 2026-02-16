package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/config"
	"redisolar-go/internal/datagen"
	redisdao "redisolar-go/internal/dao/redis"
	"redisolar-go/internal/keyschema"
	"redisolar-go/internal/models"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	ks := keyschema.New(cfg.RedisKeyPrefix)
	base := redisdao.NewRedisDao(client, ks)
	siteDao := redisdao.NewSiteDao(base)
	siteGeoDao := redisdao.NewSiteGeoDao(base)

	// Read sites from fixtures
	filename := "fixtures/sites.json"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", filename, err)
	}

	var rawSites []json.RawMessage
	if err := json.Unmarshal(data, &rawSites); err != nil {
		log.Fatalf("Failed to parse sites JSON: %v", err)
	}

	sites := make([]models.Site, 0, len(rawSites))
	for _, raw := range rawSites {
		var s models.Site
		if err := json.Unmarshal(raw, &s); err != nil {
			log.Printf("Warning: skipping site: %v", err)
			continue
		}
		sites = append(sites, s)
	}

	// Load sites with pipeline
	pipe := client.Pipeline()
	fmt.Printf("Loading %d sites...\n", len(sites))
	for _, site := range sites {
		siteDao.InsertWithClient(ctx, site, pipe)
		siteGeoDao.InsertWithClient(ctx, site, pipe)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		log.Fatalf("Failed to load sites: %v", err)
	}
	fmt.Println("Sites loaded.")

	// Generate sample data
	fmt.Println("Generating sample metrics data (1 day)...")
	generator := datagen.NewSampleDataGenerator(client, sites, 1, ks)
	fmt.Printf("Total readings to generate: %d\n", generator.Size())

	p := client.Pipeline()
	count := generator.Generate(ctx, p)
	fmt.Printf("Generated %d readings, flushing pipeline...\n", count)

	fmt.Println("Data load complete!")
}
