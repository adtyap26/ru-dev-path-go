package redis

import (
	"context"
	"testing"
	"time"

	"redisolar-go/internal/models"
)

func capacityTestReadings() []models.MeterReading {
	now := time.Now().UTC()
	readings := make([]models.MeterReading, 10)
	for i := 0; i < 10; i++ {
		readings[i] = models.MeterReading{
			SiteID:      i,
			Timestamp:   float64(now.Unix()),
			WHUsed:      1.2,
			WHGenerated: float64(i),
			TempC:       22.0,
		}
	}
	return readings
}

func TestCapacity_Update(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	capacityDao := NewCapacityReportDao(base)
	ctx := context.Background()
	readings := capacityTestReadings()
	ks := testKeySchema()

	for _, reading := range readings {
		if err := capacityDao.Update(ctx, reading); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	key := ks.CapacityRankingKey()
	results, err := base.Client.ZRevRangeWithScores(ctx, key, 0, 20).Result()
	if err != nil {
		t.Fatalf("ZRevRangeWithScores failed: %v", err)
	}
	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
}

func TestCapacity_GetReport(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	capacityDao := NewCapacityReportDao(base)
	ctx := context.Background()
	readings := capacityTestReadings()

	for _, reading := range readings {
		if err := capacityDao.Update(ctx, reading); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	report, err := capacityDao.GetReport(ctx, 5)
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}

	highest := report.HighestCapacity
	lowest := report.LowestCapacity

	if len(highest) != 5 {
		t.Fatalf("expected 5 highest, got %d", len(highest))
	}
	if len(lowest) != 5 {
		t.Fatalf("expected 5 lowest, got %d", len(lowest))
	}
	if highest[0].Capacity <= highest[1].Capacity {
		t.Fatalf("highest not sorted descending: %f <= %f", highest[0].Capacity, highest[1].Capacity)
	}
	if lowest[0].Capacity >= lowest[1].Capacity {
		t.Fatalf("lowest not sorted ascending: %f >= %f", lowest[0].Capacity, lowest[1].Capacity)
	}
	if lowest[4].Capacity <= lowest[0].Capacity {
		t.Fatalf("lowest[4] should be > lowest[0]: %f <= %f", lowest[4].Capacity, lowest[0].Capacity)
	}
}

// Challenge #4
func TestCapacity_GetRank(t *testing.T) {
	t.Skip("Remove for Challenge #4")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	capacityDao := NewCapacityReportDao(base)
	ctx := context.Background()
	readings := capacityTestReadings()

	for _, reading := range readings {
		if err := capacityDao.Update(ctx, reading); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	rank0, err := capacityDao.GetRank(ctx, readings[0].SiteID)
	if err != nil {
		t.Fatalf("GetRank(0) failed: %v", err)
	}
	if rank0 != 9 {
		t.Fatalf("expected rank 9 for site %d, got %d", readings[0].SiteID, rank0)
	}

	rank9, err := capacityDao.GetRank(ctx, readings[9].SiteID)
	if err != nil {
		t.Fatalf("GetRank(9) failed: %v", err)
	}
	if rank9 != 0 {
		t.Fatalf("expected rank 0 for site %d, got %d", readings[9].SiteID, rank9)
	}
}
