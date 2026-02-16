package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/models"
)

const (
	MaxMetricRetentionDays    = 30
	MaxDaysToReturn           = 7
	MetricsPerDay             = 60 * 24
	MetricExpirationSeconds   = 60 * 60 * 24 * MaxMetricRetentionDays + 1
)

type MetricDaoRedis struct {
	RedisDao
}

func NewMetricDao(base RedisDao) *MetricDaoRedis {
	return &MetricDaoRedis{RedisDao: base}
}

func (d *MetricDaoRedis) Insert(ctx context.Context, reading models.MeterReading) error {
	return d.InsertWithPipeline(ctx, reading, nil)
}

func (d *MetricDaoRedis) InsertWithPipeline(ctx context.Context, reading models.MeterReading, pipe goredis.Pipeliner) error {
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

func (d *MetricDaoRedis) insertMetric(ctx context.Context, siteID int, value float64, unit models.MetricUnit, t time.Time, pipe goredis.Pipeliner) {
	metricKey := d.KeySchema.DayMetricKey(siteID, unit, t)
	minuteOfDay := getDayMinute(t)

	// START Challenge #2
	// Insert a metric into a sorted set.
	//
	// The member should be a string in the format "%.2f:%d" (value:minuteOfDay).
	// The score should be the minuteOfDay (as float64).
	// Also set an expiration on the key using MetricExpirationSeconds.
	//
	// Hint: Use pipe.ZAdd() with goredis.Z{Score: ..., Member: ...}
	// Hint: Use pipe.Expire() with time.Duration(MetricExpirationSeconds)*time.Second
	// Hint: Use fmt.Sprintf("%.2f:%d", value, minuteOfDay) for the member string
	_ = metricKey    // TODO: remove after implementing
	_ = minuteOfDay  // TODO: remove after implementing
	// END Challenge #2
}

func (d *MetricDaoRedis) GetRecent(ctx context.Context, siteID int, unit models.MetricUnit, t time.Time, limit int) ([]models.Measurement, error) {
	if limit > MetricsPerDay*MaxMetricRetentionDays {
		return nil, fmt.Errorf("cannot request more than two weeks of minute-level data")
	}

	// Collect newest -> oldest, then reverse for oldest -> newest
	var collected []models.Measurement
	currentDate := t
	count := limit
	iterations := 0

	for count > 0 && iterations < MaxDaysToReturn {
		ms, err := d.getMeasurementsForDate(ctx, siteID, currentDate, unit, count)
		if err != nil {
			return nil, err
		}
		// Prepend (extendleft equivalent)
		collected = append(ms, collected...)
		count -= len(ms)
		currentDate = currentDate.AddDate(0, 0, -1)
		iterations++
	}

	return collected, nil
}

func (d *MetricDaoRedis) getMeasurementsForDate(ctx context.Context, siteID int, date time.Time, unit models.MetricUnit, count int) ([]models.Measurement, error) {
	key := d.KeySchema.DayMetricKey(siteID, unit, date)
	results, err := d.Client.ZRevRangeWithScores(ctx, key, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	measurements := make([]models.Measurement, 0, len(results))
	for _, z := range results {
		member := z.Member.(string)
		value, minuteOfDay, err := parseMeasurementMinute(member)
		if err != nil {
			continue
		}
		ts := getDateFromDayMinute(date, minuteOfDay)
		measurements = append(measurements, models.Measurement{
			SiteID:     siteID,
			MetricUnit: unit,
			Timestamp:  float64(ts.Unix()),
			Value:      value,
		})
	}
	return measurements, nil
}

func getDayMinute(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

func getDateFromDayMinute(date time.Time, dayMinute int) time.Time {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return start.Add(time.Duration(dayMinute) * time.Minute)
}

func parseMeasurementMinute(s string) (float64, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid measurement minute: %s", s)
	}
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return value, minute, nil
}
