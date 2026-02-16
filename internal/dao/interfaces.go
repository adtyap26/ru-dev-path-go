package dao

import (
	"context"
	"time"

	"redisolar-go/internal/models"
)

type SiteDao interface {
	Insert(ctx context.Context, site models.Site) error
	InsertMany(ctx context.Context, sites ...models.Site) error
	FindByID(ctx context.Context, siteID int) (models.Site, error)
	FindAll(ctx context.Context) ([]models.Site, error)
}

type SiteGeoDao interface {
	SiteDao
	FindByGeo(ctx context.Context, query models.GeoQuery) ([]models.Site, error)
}

type SiteStatsDao interface {
	FindByID(ctx context.Context, siteID int, day time.Time) (models.SiteStats, error)
	Update(ctx context.Context, reading models.MeterReading) error
}

type CapacityDao interface {
	Update(ctx context.Context, reading models.MeterReading) error
	GetReport(ctx context.Context, limit int) (models.CapacityReport, error)
	GetRank(ctx context.Context, siteID int) (int64, error)
}

type MetricDao interface {
	Insert(ctx context.Context, reading models.MeterReading) error
	GetRecent(ctx context.Context, siteID int, unit models.MetricUnit, t time.Time, limit int) ([]models.Measurement, error)
}

type FeedDao interface {
	Insert(ctx context.Context, reading models.MeterReading) error
	GetRecentGlobal(ctx context.Context, limit int) ([]models.MeterReading, error)
	GetRecentForSite(ctx context.Context, siteID int, limit int) ([]models.MeterReading, error)
}

type MeterReadingDao interface {
	Add(ctx context.Context, reading models.MeterReading) error
}

type RateLimiterDao interface {
	Hit(ctx context.Context, name string) error
}
