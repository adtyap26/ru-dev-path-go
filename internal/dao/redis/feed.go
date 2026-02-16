package redis

import (
	"context"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/models"
)

const (
	GlobalMaxFeedLength = 10000
	SiteMaxFeedLength   = 2440
)

type FeedDaoRedis struct {
	RedisDao
}

func NewFeedDao(base RedisDao) *FeedDaoRedis {
	return &FeedDaoRedis{RedisDao: base}
}

func (d *FeedDaoRedis) Insert(ctx context.Context, reading models.MeterReading) error {
	return d.InsertWithPipeline(ctx, reading, nil)
}

func (d *FeedDaoRedis) InsertWithPipeline(ctx context.Context, reading models.MeterReading, pipe goredis.Pipeliner) error {
	execute := false
	if pipe == nil {
		pipe = d.Client.Pipeline()
		execute = true
	}

	data := models.MeterReadingToStreamMap(reading)

	// START Challenge #6
	// Add the meter reading data to two Redis Streams using XADD:
	//
	// 1. The global feed stream (d.KeySchema.GlobalFeedKey()) with
	//    MaxLen: GlobalMaxFeedLength and Approx: true
	// 2. The site-specific feed stream (d.KeySchema.FeedKey(reading.SiteID)) with
	//    MaxLen: SiteMaxFeedLength and Approx: true
	//
	// Hint: Use pipe.XAdd(ctx, &goredis.XAddArgs{
	//     Stream: ..., MaxLen: ..., Approx: true, Values: data,
	// })
	_ = data // TODO: remove after implementing
	// END Challenge #6

	if execute {
		_, err := pipe.Exec(ctx)
		return err
	}
	return nil
}

func (d *FeedDaoRedis) GetRecentGlobal(ctx context.Context, limit int) ([]models.MeterReading, error) {
	return d.getRecent(ctx, d.KeySchema.GlobalFeedKey(), limit)
}

func (d *FeedDaoRedis) GetRecentForSite(ctx context.Context, siteID int, limit int) ([]models.MeterReading, error) {
	return d.getRecent(ctx, d.KeySchema.FeedKey(siteID), limit)
}

func (d *FeedDaoRedis) getRecent(ctx context.Context, key string, limit int) ([]models.MeterReading, error) {
	messages, err := d.Client.XRevRangeN(ctx, key, "+", "-", int64(limit)).Result()
	if err != nil {
		return nil, err
	}

	readings := make([]models.MeterReading, 0, len(messages))
	for _, msg := range messages {
		reading, err := streamMapToMeterReading(msg.Values)
		if err != nil {
			continue
		}
		readings = append(readings, reading)
	}
	return readings, nil
}

func streamMapToMeterReading(m map[string]interface{}) (models.MeterReading, error) {
	siteID, err := parseIntField(m, "site_id")
	if err != nil {
		return models.MeterReading{}, err
	}
	whUsed, err := parseFloatField(m, "wh_used")
	if err != nil {
		return models.MeterReading{}, err
	}
	whGenerated, err := parseFloatField(m, "wh_generated")
	if err != nil {
		return models.MeterReading{}, err
	}
	tempC, err := parseFloatField(m, "temp_c")
	if err != nil {
		return models.MeterReading{}, err
	}
	timestamp, err := parseFloatField(m, "timestamp")
	if err != nil {
		return models.MeterReading{}, err
	}

	return models.MeterReading{
		SiteID:      siteID,
		WHUsed:      whUsed,
		WHGenerated: whGenerated,
		TempC:       tempC,
		Timestamp:   timestamp,
	}, nil
}

func parseFloatField(m map[string]interface{}, field string) (float64, error) {
	v, ok := m[field]
	if !ok {
		return 0, nil
	}
	switch val := v.(type) {
	case string:
		return strconv.ParseFloat(val, 64)
	case float64:
		return val, nil
	default:
		return 0, nil
	}
}

func parseIntField(m map[string]interface{}, field string) (int, error) {
	v, ok := m[field]
	if !ok {
		return 0, nil
	}
	switch val := v.(type) {
	case string:
		return strconv.Atoi(val)
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	default:
		return 0, nil
	}
}
