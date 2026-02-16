package redis

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/models"
	// Uncomment for Challenge #3
	// "redisolar-go/internal/scripts"
)

const WeekSeconds = 60 * 60 * 24 * 7

type SiteStatsDaoRedis struct {
	RedisDao
}

func NewSiteStatsDao(base RedisDao) *SiteStatsDaoRedis {
	return &SiteStatsDaoRedis{RedisDao: base}
}

func (d *SiteStatsDaoRedis) FindByID(ctx context.Context, siteID int, day time.Time) (models.SiteStats, error) {
	if day.IsZero() {
		day = time.Now()
	}
	key := d.KeySchema.SiteStatsKey(siteID, day)
	fields, err := d.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return models.SiteStats{}, err
	}
	if len(fields) == 0 {
		return models.SiteStats{}, nil
	}

	count, _ := strconv.ParseInt(fields[models.SiteStatsCount], 10, 64)
	maxWH, _ := strconv.ParseFloat(fields[models.SiteStatsMaxWH], 64)
	minWH, _ := strconv.ParseFloat(fields[models.SiteStatsMinWH], 64)
	maxCap, _ := strconv.ParseFloat(fields[models.SiteStatsMaxCapacity], 64)

	return models.SiteStats{
		LastReportingTime: fields[models.SiteStatsLastReportingTime],
		MeterReadingCount: count,
		MaxWHGenerated:    maxWH,
		MinWHGenerated:    minWH,
		MaxCapacity:       maxCap,
	}, nil
}

func (d *SiteStatsDaoRedis) Update(ctx context.Context, reading models.MeterReading) error {
	return d.UpdateWithPipeline(ctx, reading, nil)
}

func (d *SiteStatsDaoRedis) UpdateWithPipeline(ctx context.Context, reading models.MeterReading, pipe goredis.Pipeliner) error {
	t := reading.TimestampTime()
	key := d.KeySchema.SiteStatsKey(reading.SiteID, t)

	execute := false
	if pipe == nil {
		pipe = d.Client.Pipeline()
		execute = true
	}

	// START Challenge #3
	// Use a pipeline to update site stats efficiently with the following steps:
	//
	// 1. Set the last reporting time using pipe.HSet() with models.SiteStatsLastReportingTime
	//    (use time.Now().UTC().Format(time.RFC3339) for the value)
	// 2. Increment the count using pipe.HIncrBy() with models.SiteStatsCount
	// 3. Set the key expiration using pipe.Expire() with WeekSeconds
	// 4. Use Lua scripts for atomic compare-and-update operations:
	//    - scripts.UpdateIfGreater() for models.SiteStatsMaxWH with reading.WHGenerated
	//    - scripts.UpdateIfLess() for models.SiteStatsMinWH with reading.WHGenerated
	//    - scripts.UpdateIfGreater() for models.SiteStatsMaxCapacity with reading.CurrentCapacity()
	//
	// Hint: The scripts package provides UpdateIfGreater(ctx, pipe, key, field, value)
	//        and UpdateIfLess(ctx, pipe, key, field, value)
	_ = key     // TODO: remove after implementing
	_ = reading // TODO: remove after implementing
	// END Challenge #3

	if execute {
		_, err := pipe.Exec(ctx)
		return err
	}
	return nil
}
