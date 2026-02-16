package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/models"
)

const RetentionMS = 60 * 60 * 24 * 14 * 1000 // 14 days in ms

type MetricDaoRedisTimeseries struct {
	RedisDao
}

func NewMetricTimeseriesDao(base RedisDao) *MetricDaoRedisTimeseries {
	return &MetricDaoRedisTimeseries{RedisDao: base}
}

func unixMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func (d *MetricDaoRedisTimeseries) Insert(ctx context.Context, reading models.MeterReading) error {
	return d.InsertWithPipeline(ctx, reading, nil)
}

func (d *MetricDaoRedisTimeseries) InsertWithPipeline(ctx context.Context, reading models.MeterReading, pipe goredis.Pipeliner) error {
	execute := false
	if pipe == nil {
		pipe = d.Client.Pipeline()
		execute = true
	}

	t := reading.TimestampTime()
	d.insertMetric(ctx, reading.SiteID, reading.WHGenerated, models.WHGenerated, t, pipe)
	d.insertMetric(ctx, reading.SiteID, reading.WHUsed, models.WHUsed, t, pipe)
	d.insertMetric(ctx, reading.SiteID, reading.TempC, models.TempCelsius, t, pipe)

	if execute {
		_, err := pipe.Exec(ctx)
		return err
	}
	return nil
}

func (d *MetricDaoRedisTimeseries) insertMetric(ctx context.Context, siteID int, value float64, unit models.MetricUnit, t time.Time, pipe goredis.Pipeliner) {
	key := d.KeySchema.TimeseriesKey(siteID, unit)
	timeMs := unixMilliseconds(t)
	pipe.Do(ctx, "TS.ADD", key, timeMs, value, "RETENTION", RetentionMS)
}

func (d *MetricDaoRedisTimeseries) GetRecent(ctx context.Context, siteID int, unit models.MetricUnit, t time.Time, limit int) ([]models.Measurement, error) {
	key := d.KeySchema.TimeseriesKey(siteID, unit)
	timeMs := unixMilliseconds(t)
	initialTimestamp := timeMs - int64(limit*60)*1000

	result, err := d.Client.Do(ctx, "TS.RANGE", key, initialTimestamp, timeMs).Result()
	if err != nil {
		return nil, err
	}

	pairs, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected TS.RANGE result type: %T", result)
	}

	measurements := make([]models.Measurement, 0, len(pairs))
	count := 0
	for _, pair := range pairs {
		if count >= limit {
			break
		}
		p, ok := pair.([]interface{})
		if !ok || len(p) != 2 {
			continue
		}
		ts, err := toInt64(p[0])
		if err != nil {
			continue
		}
		val, err := toFloat(p[1])
		if err != nil {
			continue
		}
		measurements = append(measurements, models.Measurement{
			SiteID:     siteID,
			MetricUnit: unit,
			Timestamp:  float64(ts) / 1000.0,
			Value:      val,
		})
		count++
	}

	return measurements, nil
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	case float64:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	case int64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
