package datagen

import (
	"context"
	"math/rand"
	"time"

	goredis "github.com/redis/go-redis/v9"

	redisdao "redisolar-go/internal/dao/redis"
	"redisolar-go/internal/keyschema"
	"redisolar-go/internal/models"
)

const maxTemperatureC = 30.0

type SampleDataGenerator struct {
	client     *goredis.Client
	sites      []models.Site
	minuteDays int
	keySchema  *keyschema.KeySchema
	readings   [][]models.MeterReading
}

func NewSampleDataGenerator(client *goredis.Client, sites []models.Site, days int, ks *keyschema.KeySchema) *SampleDataGenerator {
	minuteDays := days * 3 * 60
	readings := make([][]models.MeterReading, len(sites))
	for i := range readings {
		readings[i] = make([]models.MeterReading, minuteDays)
	}
	return &SampleDataGenerator{
		client:     client,
		sites:      sites,
		minuteDays: minuteDays,
		keySchema:  ks,
		readings:   readings,
	}
}

func (g *SampleDataGenerator) Size() int {
	return g.minuteDays * len(g.sites)
}

func getMaxMinuteWHGenerated(capacity float64) float64 {
	return capacity * 1000 / 24 / 60
}

func getInitialMinuteWHUsed(maxCapacity float64) float64 {
	if rand.Float64() > 0.5 {
		return maxCapacity + 0.1
	}
	return maxCapacity - 0.1
}

func getNextValue(maximum, current float64) float64 {
	stepSize := 0.1 * maximum
	if rand.Intn(2) == 0 {
		return current + stepSize
	}
	if current-stepSize < 0.0 {
		return 0.0
	}
	return current - stepSize
}

func (g *SampleDataGenerator) Generate(ctx context.Context, pipe goredis.Pipeliner) int {
	meterReadingDao := redisdao.NewMeterReadingDao(redisdao.NewRedisDao(g.client, g.keySchema))

	for sIdx, site := range g.sites {
		maxCap := getMaxMinuteWHGenerated(site.Capacity)
		currentCap := getNextValue(maxCap, maxCap)
		currentTemp := getNextValue(maxTemperatureC, maxTemperatureC)
		currentUsage := getInitialMinuteWHUsed(maxCap)
		currentTime := time.Now().UTC().Add(-time.Duration(g.minuteDays) * time.Minute)

		for i := 0; i < g.minuteDays; i++ {
			reading := models.MeterReading{
				SiteID:      site.ID,
				Timestamp:   float64(currentTime.Unix()),
				WHUsed:      currentUsage,
				WHGenerated: currentCap,
				TempC:       currentTemp,
			}
			g.readings[sIdx][i] = reading

			currentTime = currentTime.Add(time.Minute)
			currentTemp = getNextValue(currentTemp, currentTemp)
			currentCap = getNextValue(currentCap, currentCap)
			currentUsage = getNextValue(currentUsage, currentUsage)
		}
	}

	count := 0
	for i := 0; i < g.minuteDays; i++ {
		for j := 0; j < len(g.sites); j++ {
			reading := g.readings[j][i]
			meterReadingDao.Add(ctx, reading)
			count++
		}
	}
	return count
}
