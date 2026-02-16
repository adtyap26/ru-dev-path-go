package redis

import (
	"context"
	"testing"
	"time"

	"redisolar-go/internal/models"
)

// Challenge #3
func TestSiteStats_Update(t *testing.T) {
	t.Skip("Remove for Challenge #3")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	statsDao := NewSiteStatsDao(base)
	ctx := context.Background()
	now := time.Now()

	reading1 := models.MeterReading{
		SiteID:      1,
		Timestamp:   float64(now.Unix()),
		TempC:       15.0,
		WHGenerated: 1.0,
		WHUsed:      0.0,
	}
	reading2 := models.MeterReading{
		SiteID:      1,
		Timestamp:   float64(now.Unix()),
		TempC:       15.0,
		WHGenerated: 2.0,
		WHUsed:      0.0,
	}

	if err := statsDao.Update(ctx, reading1); err != nil {
		t.Fatalf("Update reading1 failed: %v", err)
	}
	if err := statsDao.Update(ctx, reading2); err != nil {
		t.Fatalf("Update reading2 failed: %v", err)
	}

	stats, err := statsDao.FindByID(ctx, 1, now)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if stats.MaxWHGenerated != 2.0 {
		t.Fatalf("expected MaxWHGenerated=2.0, got %f", stats.MaxWHGenerated)
	}
	if stats.MinWHGenerated != 1.0 {
		t.Fatalf("expected MinWHGenerated=1.0, got %f", stats.MinWHGenerated)
	}
	if stats.MaxCapacity != 2.0 {
		t.Fatalf("expected MaxCapacity=2.0, got %f", stats.MaxCapacity)
	}
}
