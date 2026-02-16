package redis

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/models"
	"redisolar-go/internal/scripts"
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

	reportingTime := time.Now().UTC().Format(time.RFC3339)
	pipe.HSet(ctx, key, models.SiteStatsLastReportingTime, reportingTime)
	pipe.HIncrBy(ctx, key, models.SiteStatsCount, 1)
	pipe.Expire(ctx, key, time.Duration(WeekSeconds)*time.Second)

	scripts.UpdateIfGreater(ctx, pipe, key, models.SiteStatsMaxWH, reading.WHGenerated)
	scripts.UpdateIfLess(ctx, pipe, key, models.SiteStatsMinWH, reading.WHGenerated)
	scripts.UpdateIfGreater(ctx, pipe, key, models.SiteStatsMaxCapacity, reading.CurrentCapacity())

	if execute {
		_, err := pipe.Exec(ctx)
		return err
	}
	return nil
}
