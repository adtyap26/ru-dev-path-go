package redis

import (
	"context"
	"testing"
	"time"

	"redisolar-go/internal/models"
)

func generateFeedReading(siteID int, ts time.Time) models.MeterReading {
	return models.MeterReading{
		SiteID:      siteID,
		Timestamp:   float64(ts.Unix()),
		TempC:       15.0,
		WHGenerated: 0.025,
		WHUsed:      0.015,
	}
}

// Challenge #6
func TestFeed_BasicInsertReturnsRecent(t *testing.T) {
	t.Skip("Remove for Challenge #6")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	feedDao := NewFeedDao(base)
	ctx := context.Background()

	now := time.Now()
	reading0 := generateFeedReading(1, now)
	reading1 := generateFeedReading(1, now.Add(-time.Minute))

	if err := feedDao.Insert(ctx, reading0); err != nil {
		t.Fatalf("Insert reading0 failed: %v", err)
	}
	if err := feedDao.Insert(ctx, reading1); err != nil {
		t.Fatalf("Insert reading1 failed: %v", err)
	}

	globalList, err := feedDao.GetRecentGlobal(ctx, 100)
	if err != nil {
		t.Fatalf("GetRecentGlobal failed: %v", err)
	}
	if len(globalList) != 2 {
		t.Fatalf("expected 2 global entries, got %d", len(globalList))
	}
	// XRevRange returns newest first, so reading1 (older) is first after reversal
	if globalList[0].Timestamp != reading1.Timestamp {
		t.Fatalf("expected globalList[0] timestamp %f, got %f", reading1.Timestamp, globalList[0].Timestamp)
	}
	if globalList[1].Timestamp != reading0.Timestamp {
		t.Fatalf("expected globalList[1] timestamp %f, got %f", reading0.Timestamp, globalList[1].Timestamp)
	}

	siteList, err := feedDao.GetRecentForSite(ctx, 1, 100)
	if err != nil {
		t.Fatalf("GetRecentForSite failed: %v", err)
	}
	if len(siteList) != 2 {
		t.Fatalf("expected 2 site entries, got %d", len(siteList))
	}
	if siteList[0].Timestamp != reading1.Timestamp {
		t.Fatalf("expected siteList[0] timestamp %f, got %f", reading1.Timestamp, siteList[0].Timestamp)
	}
	if siteList[1].Timestamp != reading0.Timestamp {
		t.Fatalf("expected siteList[1] timestamp %f, got %f", reading0.Timestamp, siteList[1].Timestamp)
	}
}
