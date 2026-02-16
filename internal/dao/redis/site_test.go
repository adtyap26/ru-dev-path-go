package redis

import (
	"context"
	"testing"

	"redisolar-go/internal/dao"
	"redisolar-go/internal/models"
)

func TestSite_DoesNotExist(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	siteDao := NewSiteDao(base)

	_, err := siteDao.FindByID(context.Background(), 0)
	if err != dao.ErrSiteNotFound {
		t.Fatalf("expected ErrSiteNotFound, got %v", err)
	}
}

func TestSite_Insert(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	siteDao := NewSiteDao(base)
	ctx := context.Background()

	site := models.Site{
		ID:         1,
		Capacity:   10.0,
		Panels:     100,
		Address:    "100 SE Pine St.",
		City:       "Portland",
		State:      "OR",
		PostalCode: "97202",
	}

	if err := siteDao.Insert(ctx, site); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	found, err := siteDao.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found != site {
		t.Fatalf("expected %+v, got %+v", site, found)
	}
}

func TestSite_InsertMany(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	siteDao := NewSiteDao(base)
	ctx := context.Background()

	site1 := models.Site{ID: 1, Capacity: 10.0, Panels: 100, Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202"}
	site2 := models.Site{ID: 2, Capacity: 25.0, Panels: 110, Address: "101 SW Ankeny", City: "Portland", State: "OR", PostalCode: "97203"}
	site3 := models.Site{ID: 3, Capacity: 100.0, Panels: 155, Address: "201 SE Burnside", City: "Portland", State: "OR", PostalCode: "97204"}

	if err := siteDao.InsertMany(ctx, site1, site2, site3); err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	for _, s := range []models.Site{site1, site2, site3} {
		found, err := siteDao.FindByID(ctx, s.ID)
		if err != nil {
			t.Fatalf("FindByID(%d) failed: %v", s.ID, err)
		}
		if found != s {
			t.Fatalf("expected %+v, got %+v", s, found)
		}
	}
}

func TestSite_FindByID(t *testing.T) {
	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	siteDao := NewSiteDao(base)
	ctx := context.Background()

	site := models.Site{ID: 1, Capacity: 10.0, Panels: 100, Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202"}

	if err := siteDao.Insert(ctx, site); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	found, err := siteDao.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found != site {
		t.Fatalf("expected %+v, got %+v", site, found)
	}
}

// Challenge #1
func TestSite_FindAll(t *testing.T) {
	t.Skip("Remove for Challenge #1")

	base := testBase(t)
	cleanupTestKeys(t, base.Client)
	siteDao := NewSiteDao(base)
	ctx := context.Background()

	site1 := models.Site{ID: 1, Capacity: 10.0, Panels: 100, Address: "100 SE Pine St.", City: "Portland", State: "OR", PostalCode: "97202"}
	site2 := models.Site{ID: 2, Capacity: 25.0, Panels: 110, Address: "101 SW Ankeny", City: "Portland", State: "OR", PostalCode: "97203"}
	site3 := models.Site{ID: 3, Capacity: 100.0, Panels: 155, Address: "201 SE Burnside", City: "Portland", State: "OR", PostalCode: "97204"}

	if err := siteDao.InsertMany(ctx, site1, site2, site3); err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	sites, err := siteDao.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	expected := map[int]models.Site{1: site1, 2: site2, 3: site3}
	if len(sites) != len(expected) {
		t.Fatalf("expected %d sites, got %d", len(expected), len(sites))
	}
	for _, s := range sites {
		exp, ok := expected[s.ID]
		if !ok {
			t.Fatalf("unexpected site ID %d", s.ID)
		}
		if s != exp {
			t.Fatalf("expected %+v, got %+v", exp, s)
		}
	}
}
