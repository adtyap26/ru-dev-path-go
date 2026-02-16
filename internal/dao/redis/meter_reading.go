package redis

import (
	"context"

	"redisolar-go/internal/models"
)

type MeterReadingDaoRedis struct {
	RedisDao
	metricDao   *MetricDaoRedisTimeseries
	capacityDao *CapacityReportDaoRedis
	feedDao     *FeedDaoRedis
	// Uncomment for Challenge #3
	// statsDao *SiteStatsDaoRedis
}

func NewMeterReadingDao(base RedisDao) *MeterReadingDaoRedis {
	return &MeterReadingDaoRedis{
		RedisDao:    base,
		metricDao:   NewMetricTimeseriesDao(base),
		capacityDao: NewCapacityReportDao(base),
		feedDao:     NewFeedDao(base),
		// Uncomment for Challenge #3
		// statsDao: NewSiteStatsDao(base),
	}
}

func (d *MeterReadingDaoRedis) Add(ctx context.Context, reading models.MeterReading) error {
	return d.AddWithPipeline(ctx, reading, nil)
}

func (d *MeterReadingDaoRedis) AddWithPipeline(ctx context.Context, reading models.MeterReading, pipe interface{}) error {
	if err := d.metricDao.Insert(ctx, reading); err != nil {
		return err
	}
	if err := d.capacityDao.Update(ctx, reading); err != nil {
		return err
	}
	if err := d.feedDao.Insert(ctx, reading); err != nil {
		return err
	}
	// Uncomment for Challenge #3
	// if err := d.statsDao.Update(ctx, reading); err != nil {
	// 	return err
	// }
	return nil
}
