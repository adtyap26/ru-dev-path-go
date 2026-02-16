package redis

import (
	"context"
	"testing"
	"time"

	"redisolar-go/internal/models"
)

func generateMetricReadings(count int, now time.Time) []models.MeterReading {
	readings := make([]models.MeterReading, 0, count)
	t := now
	for i := 0; i < count; i++ {
		readings = append(readings, models.MeterReading{
			SiteID:      1,
			TempC:       float64(i),
			WHUsed:      float64(i),
			WHGenerated: float64(i),
			Timestamp:   float64(t.Unix()),
		})
		t = t.Add(-time.Minute)
	}
	return readings
}

func testInsertAndRetrieve(t *testing.T, metricDao *MetricDaoRedis, readings []models.MeterReading, limit int, now time.Time) {
	t.Helper()
	ctx := context.Background()

	for _, reading := range readings {
		if err := metricDao.Insert(ctx, reading); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	measurements, err := metricDao.GetRecent(ctx, 1, models.WHGenerated, now, limit)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	if len(measurements) != limit {
		t.Fatalf("expected %d measurements, got %d", limit, len(measurements))
	}

	i := limit
	for _, m := range measurements {
		if m.Value != float64(i-1) {
			t.Fatalf("expected value %.1f, got %.1f", float64(i-1), m.Value)
		}
		i--
	}
}

// Challenge #2
func TestMetric_Small(t *testing.T) {
	t.Skip("Remove for Challenge #2")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	metricDao := NewMetricDao(base)
	now := time.Now().UTC()
	readings := generateMetricReadings(72*60, now)
	testInsertAndRetrieve(t, metricDao, readings, 1, now)
}

func TestMetric_OneDay(t *testing.T) {
	t.Skip("Remove for Challenge #2")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	metricDao := NewMetricDao(base)
	now := time.Now().UTC()
	readings := generateMetricReadings(72*60, now)
	testInsertAndRetrieve(t, metricDao, readings, 60*24, now)
}

func TestMetric_MultipleDays(t *testing.T) {
	t.Skip("Remove for Challenge #2")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	metricDao := NewMetricDao(base)
	now := time.Now().UTC()
	readings := generateMetricReadings(72*60, now)
	testInsertAndRetrieve(t, metricDao, readings, 60*70, now)
}
