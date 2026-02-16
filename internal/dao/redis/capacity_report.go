package redis

import (
	"context"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/models"
)

type CapacityReportDaoRedis struct {
	RedisDao
}

func NewCapacityReportDao(base RedisDao) *CapacityReportDaoRedis {
	return &CapacityReportDaoRedis{RedisDao: base}
}

func (d *CapacityReportDaoRedis) Update(ctx context.Context, reading models.MeterReading) error {
	return d.UpdateWithClient(ctx, reading, d.Client)
}

func (d *CapacityReportDaoRedis) UpdateWithClient(ctx context.Context, reading models.MeterReading, client goredis.Cmdable) error {
	key := d.KeySchema.CapacityRankingKey()
	return client.ZAdd(ctx, key, goredis.Z{
		Score:  reading.CurrentCapacity(),
		Member: strconv.Itoa(reading.SiteID),
	}).Err()
}

func (d *CapacityReportDaoRedis) GetReport(ctx context.Context, limit int) (models.CapacityReport, error) {
	key := d.KeySchema.CapacityRankingKey()
	pipe := d.Client.Pipeline()

	lowCmd := pipe.ZRangeWithScores(ctx, key, 0, int64(limit-1))
	highCmd := pipe.ZRevRangeWithScores(ctx, key, 0, int64(limit-1))
	_, err := pipe.Exec(ctx)
	if err != nil {
		return models.CapacityReport{}, err
	}

	lowResults, _ := lowCmd.Result()
	highResults, _ := highCmd.Result()

	lowest := make([]models.SiteCapacityTuple, 0, len(lowResults))
	for _, z := range lowResults {
		siteID, _ := strconv.Atoi(z.Member.(string))
		lowest = append(lowest, models.SiteCapacityTuple{
			SiteID:   siteID,
			Capacity: z.Score,
		})
	}

	highest := make([]models.SiteCapacityTuple, 0, len(highResults))
	for _, z := range highResults {
		siteID, _ := strconv.Atoi(z.Member.(string))
		highest = append(highest, models.SiteCapacityTuple{
			SiteID:   siteID,
			Capacity: z.Score,
		})
	}

	return models.CapacityReport{
		HighestCapacity: highest,
		LowestCapacity:  lowest,
	}, nil
}

func (d *CapacityReportDaoRedis) GetRank(ctx context.Context, siteID int) (int64, error) {
	key := d.KeySchema.CapacityRankingKey()
	return d.Client.ZRevRank(ctx, key, strconv.Itoa(siteID)).Result()
}
