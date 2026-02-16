package keyschema

import (
	"fmt"
	"time"

	"redisolar-go/internal/models"
)

const DefaultKeyPrefix = "ru102py-test"

type KeySchema struct {
	Prefix string
}

func New(prefix string) *KeySchema {
	return &KeySchema{Prefix: prefix}
}

func (ks *KeySchema) prefixed(key string) string {
	return fmt.Sprintf("%s:%s", ks.Prefix, key)
}

// SiteHashKey returns the key for a site's hash: sites:info:[site_id]
func (ks *KeySchema) SiteHashKey(siteID int) string {
	return ks.prefixed(fmt.Sprintf("sites:info:%d", siteID))
}

// SiteIDsKey returns the key for the set of all site IDs: sites:ids
func (ks *KeySchema) SiteIDsKey() string {
	return ks.prefixed("sites:ids")
}

// SiteGeoKey returns the key for the geo index: sites:geo
func (ks *KeySchema) SiteGeoKey() string {
	return ks.prefixed("sites:geo")
}

// SiteStatsKey returns the key for site stats: sites:stats:[day]:[site_id]
func (ks *KeySchema) SiteStatsKey(siteID int, day time.Time) string {
	return ks.prefixed(fmt.Sprintf("sites:stats:%s:%d", day.Format("2006-01-02"), siteID))
}

// CapacityRankingKey returns the key for capacity rankings: sites:capacity:ranking
func (ks *KeySchema) CapacityRankingKey() string {
	return ks.prefixed("sites:capacity:ranking")
}

// DayMetricKey returns the key for a day's metrics: metric:[unit]:[day]:[site_id]
func (ks *KeySchema) DayMetricKey(siteID int, unit models.MetricUnit, t time.Time) string {
	return ks.prefixed(fmt.Sprintf("metric:%s:%s:%d", string(unit), t.Format("2006-01-02"), siteID))
}

// GlobalFeedKey returns the key for the global feed stream: sites:feed
func (ks *KeySchema) GlobalFeedKey() string {
	return ks.prefixed("sites:feed")
}

// FeedKey returns the key for a site's feed stream: sites:feed:[site_id]
func (ks *KeySchema) FeedKey(siteID int) string {
	return ks.prefixed(fmt.Sprintf("sites:feed:%d", siteID))
}

// FixedRateLimiterKey returns the key for a fixed-window rate limiter.
func (ks *KeySchema) FixedRateLimiterKey(name string, minuteBlock int, maxHits int) string {
	return ks.prefixed(fmt.Sprintf("limiter:%s:%d:%d", name, minuteBlock, maxHits))
}

// SlidingWindowRateLimiterKey returns the key for a sliding-window rate limiter.
func (ks *KeySchema) SlidingWindowRateLimiterKey(name string, windowSizeMs int, maxHits int) string {
	return ks.prefixed(fmt.Sprintf("limiter:%s:%d:%d", name, windowSizeMs, maxHits))
}

// TimeseriesKey returns the key for a timeseries: sites:ts:[site_id]:[unit]
func (ks *KeySchema) TimeseriesKey(siteID int, unit models.MetricUnit) string {
	return ks.prefixed(fmt.Sprintf("sites:ts:%d:%s", siteID, string(unit)))
}
