package redis

import (
	"context"
	"testing"
	"time"

	"redisolar-go/internal/dao"
	"redisolar-go/internal/models"
)

var (
	portland  = &models.Coordinate{Lat: 45.523064, Lng: -122.676483}
	beaverton = &models.Coordinate{Lat: 45.485168, Lng: -122.804489}
)

func TestSiteGeo_DoesNotExist(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)

	_, err := geoDao.FindByID(context.Background(), 0)
	if err != dao.ErrSiteNotFound {
		t.Fatalf("expected ErrSiteNotFound, got %v", err)
	}
}

func TestSiteGeo_Insert(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)
	ctx := context.Background()

	site := models.Site{
		ID: 1, Capacity: 10.0, Panels: 100,
		Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202",
		Coordinate: portland,
	}

	if err := geoDao.Insert(ctx, site); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	found, err := geoDao.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.ID != site.ID || found.Address != site.Address {
		t.Fatalf("expected %+v, got %+v", site, found)
	}
}

func TestSiteGeo_InsertMany(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)
	ctx := context.Background()

	site1 := models.Site{ID: 1, Capacity: 10.0, Panels: 100, Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202", Coordinate: portland}
	site2 := models.Site{ID: 2, Capacity: 25.0, Panels: 110, Address: "101 SW Ankeny", City: "Portland", State: "OR", PostalCode: "97203", Coordinate: portland}
	site3 := models.Site{ID: 3, Capacity: 100.0, Panels: 155, Address: "201 SE Burnside", City: "Portland", State: "OR", PostalCode: "97204", Coordinate: portland}

	if err := geoDao.InsertMany(ctx, site1, site2, site3); err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	for _, s := range []models.Site{site1, site2, site3} {
		found, err := geoDao.FindByID(ctx, s.ID)
		if err != nil {
			t.Fatalf("FindByID(%d) failed: %v", s.ID, err)
		}
		if found.ID != s.ID {
			t.Fatalf("expected ID %d, got %d", s.ID, found.ID)
		}
	}
}

func TestSiteGeo_FindByID(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)
	ctx := context.Background()

	site := models.Site{
		ID: 1, Capacity: 10.0, Panels: 100,
		Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202",
		Coordinate: portland,
	}

	if err := geoDao.Insert(ctx, site); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	found, err := geoDao.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.ID != site.ID {
		t.Fatalf("expected ID %d, got %d", site.ID, found.ID)
	}
}

func TestSiteGeo_FindAll(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)
	ctx := context.Background()

	site1 := models.Site{ID: 1, Capacity: 10.0, Panels: 100, Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202", Coordinate: portland}
	site2 := models.Site{ID: 2, Capacity: 25.0, Panels: 110, Address: "101 SW Ankeny", City: "Portland", State: "OR", PostalCode: "97203", Coordinate: portland}
	site3 := models.Site{ID: 3, Capacity: 100.0, Panels: 155, Address: "201 SE Burnside", City: "Portland", State: "OR", PostalCode: "97204", Coordinate: portland}

	if err := geoDao.InsertMany(ctx, site1, site2, site3); err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	sites, err := geoDao.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(sites) != 3 {
		t.Fatalf("expected 3 sites, got %d", len(sites))
	}
}

func TestSiteGeo_FindByGeo(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)
	ctx := context.Background()

	site1 := models.Site{ID: 1, Capacity: 10.0, Panels: 100, Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202", Coordinate: portland}
	site2 := models.Site{ID: 2, Capacity: 25.0, Panels: 110, Address: "101 SW Ankeny", City: "Portland", State: "OR", PostalCode: "97203", Coordinate: portland}
	site3 := models.Site{ID: 3, Capacity: 100.0, Panels: 155, Address: "9585 SW Washington Sq.", City: "Beaverton", State: "OR", PostalCode: "97223", Coordinate: beaverton}

	if err := geoDao.InsertMany(ctx, site1, site2, site3); err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	query := models.GeoQuery{
		Coordinate: models.Coordinate{Lat: portland.Lat, Lng: portland.Lng},
		Radius:     1,
		RadiusUnit: models.GeoUnitMI,
	}

	sites, err := geoDao.FindByGeo(ctx, query)
	if err != nil {
		t.Fatalf("FindByGeo failed: %v", err)
	}
	if len(sites) != 2 {
		t.Fatalf("expected 2 sites near Portland, got %d", len(sites))
	}
}

// Challenge #5
func TestSiteGeo_FindByGeoWithExcessCapacity(t *testing.T) {
	t.Skip("Remove for Challenge #5")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	geoDao := NewSiteGeoDao(base)
	capacityDao := NewCapacityReportDao(base)
	ctx := context.Background()

	site1 := models.Site{
		ID: 1, Capacity: 10.0, Panels: 100,
		Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202",
		Coordinate: portland,
	}
	if err := geoDao.Insert(ctx, site1); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Check that this site is returned when we're not looking for excess capacity.
	query := models.GeoQuery{
		Coordinate: models.Coordinate{Lat: portland.Lat, Lng: portland.Lng},
		Radius:     1,
		RadiusUnit: models.GeoUnitMI,
	}
	sites, err := geoDao.FindByGeo(ctx, query)
	if err != nil {
		t.Fatalf("FindByGeo failed: %v", err)
	}
	if len(sites) != 1 {
		t.Fatalf("expected 1 site, got %d", len(sites))
	}

	// Simulate changing a meter reading with no excess capacity.
	now := time.Now()
	reading := models.MeterReading{
		SiteID:      site1.ID,
		WHUsed:      1.0,
		WHGenerated: 0.0,
		TempC:       10,
		Timestamp:   float64(now.Unix()),
	}
	if err := capacityDao.Update(ctx, reading); err != nil {
		t.Fatalf("CapacityDao.Update failed: %v", err)
	}

	// In this case, no sites are returned for an excess capacity query.
	excessQuery := models.GeoQuery{
		Coordinate:         models.Coordinate{Lat: portland.Lat, Lng: portland.Lng},
		Radius:             1,
		RadiusUnit:         models.GeoUnitMI,
		OnlyExcessCapacity: true,
	}
	sites, err = geoDao.FindByGeo(ctx, excessQuery)
	if err != nil {
		t.Fatalf("FindByGeo (excess) failed: %v", err)
	}
	if len(sites) != 0 {
		t.Fatalf("expected 0 sites with excess capacity, got %d", len(sites))
	}

	// Simulate changing a meter reading indicating excess capacity.
	reading = models.MeterReading{
		SiteID:      site1.ID,
		WHUsed:      1.0,
		WHGenerated: 2.0,
		TempC:       10,
		Timestamp:   float64(now.Unix()),
	}
	if err := capacityDao.Update(ctx, reading); err != nil {
		t.Fatalf("CapacityDao.Update failed: %v", err)
	}

	// Add more Sites -- none with excess capacity
	for i := 2; i < 20; i++ {
		site := models.Site{
			ID: i, Capacity: 10, Panels: 100,
			Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202",
			Coordinate: portland,
		}
		if err := geoDao.Insert(ctx, site); err != nil {
			t.Fatalf("Insert site %d failed: %v", i, err)
		}
		r := models.MeterReading{
			SiteID:      i,
			WHUsed:      float64(i),
			WHGenerated: 0,
			TempC:       10,
			Timestamp:   float64(now.Unix()),
		}
		if err := capacityDao.Update(ctx, r); err != nil {
			t.Fatalf("CapacityDao.Update site %d failed: %v", i, err)
		}
	}

	// In this case, one site is returned on the excess capacity query
	sites, err = geoDao.FindByGeo(ctx, excessQuery)
	if err != nil {
		t.Fatalf("FindByGeo (excess, final) failed: %v", err)
	}

	found := false
	for _, s := range sites {
		if s.ID == site1.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected site1 (ID=%d) in results, got %+v", site1.ID, sites)
	}
}
