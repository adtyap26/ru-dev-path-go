package keyschema

import (
	"testing"
	"time"

	"redisolar-go/internal/models"
)

func TestSiteHashKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.SiteHashKey(1)
	want := "ru102py-test:sites:info:1"
	if got != want {
		t.Errorf("SiteHashKey(1) = %q, want %q", got, want)
	}
}

func TestSiteIDsKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.SiteIDsKey()
	want := "ru102py-test:sites:ids"
	if got != want {
		t.Errorf("SiteIDsKey() = %q, want %q", got, want)
	}
}

func TestSiteGeoKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.SiteGeoKey()
	want := "ru102py-test:sites:geo"
	if got != want {
		t.Errorf("SiteGeoKey() = %q, want %q", got, want)
	}
}

func TestSiteStatsKey(t *testing.T) {
	ks := New("ru102py-test")
	day := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	got := ks.SiteStatsKey(1, day)
	want := "ru102py-test:sites:stats:2020-01-01:1"
	if got != want {
		t.Errorf("SiteStatsKey(1, day) = %q, want %q", got, want)
	}
}

func TestCapacityRankingKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.CapacityRankingKey()
	want := "ru102py-test:sites:capacity:ranking"
	if got != want {
		t.Errorf("CapacityRankingKey() = %q, want %q", got, want)
	}
}

func TestDayMetricKey(t *testing.T) {
	ks := New("ru102py-test")
	day := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	got := ks.DayMetricKey(1, models.WHUsed, day)
	want := "ru102py-test:metric:whU:2020-01-01:1"
	if got != want {
		t.Errorf("DayMetricKey(1, WHUsed, day) = %q, want %q", got, want)
	}
}

func TestGlobalFeedKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.GlobalFeedKey()
	want := "ru102py-test:sites:feed"
	if got != want {
		t.Errorf("GlobalFeedKey() = %q, want %q", got, want)
	}
}

func TestFeedKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.FeedKey(1)
	want := "ru102py-test:sites:feed:1"
	if got != want {
		t.Errorf("FeedKey(1) = %q, want %q", got, want)
	}
}

func TestTimeseriesKey(t *testing.T) {
	ks := New("ru102py-test")
	got := ks.TimeseriesKey(1, models.WHGenerated)
	want := "ru102py-test:sites:ts:1:whG"
	if got != want {
		t.Errorf("TimeseriesKey(1, WHGenerated) = %q, want %q", got, want)
	}
}
