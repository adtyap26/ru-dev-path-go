package redis

import (
	"context"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/dao"
	"redisolar-go/internal/models"
)

const CapacityThreshold = 0.2

type SiteGeoDaoRedis struct {
	RedisDao
}

func NewSiteGeoDao(base RedisDao) *SiteGeoDaoRedis {
	return &SiteGeoDaoRedis{RedisDao: base}
}

func (d *SiteGeoDaoRedis) Insert(ctx context.Context, site models.Site) error {
	return d.InsertWithClient(ctx, site, d.Client)
}

func (d *SiteGeoDaoRedis) InsertWithClient(ctx context.Context, site models.Site, client goredis.Cmdable) error {
	hashKey := d.KeySchema.SiteHashKey(site.ID)
	flat := models.SiteToFlatMap(site)
	if err := client.HSet(ctx, hashKey, flat).Err(); err != nil {
		return err
	}

	if site.Coordinate == nil {
		return nil
	}

	return client.GeoAdd(ctx, d.KeySchema.SiteGeoKey(), &goredis.GeoLocation{
		Name:      strconv.Itoa(site.ID),
		Longitude: site.Coordinate.Lng,
		Latitude:  site.Coordinate.Lat,
	}).Err()
}

func (d *SiteGeoDaoRedis) InsertMany(ctx context.Context, sites ...models.Site) error {
	for _, site := range sites {
		if err := d.Insert(ctx, site); err != nil {
			return err
		}
	}
	return nil
}

func (d *SiteGeoDaoRedis) FindByID(ctx context.Context, siteID int) (models.Site, error) {
	hashKey := d.KeySchema.SiteHashKey(siteID)
	result, err := d.Client.HGetAll(ctx, hashKey).Result()
	if err != nil {
		return models.Site{}, err
	}
	if len(result) == 0 {
		return models.Site{}, dao.ErrSiteNotFound
	}
	return models.SiteFromFlatMap(result)
}

func (d *SiteGeoDaoRedis) FindByGeo(ctx context.Context, query models.GeoQuery) ([]models.Site, error) {
	if query.OnlyExcessCapacity {
		return d.findByGeoWithCapacity(ctx, query)
	}
	return d.findByGeo(ctx, query)
}

func (d *SiteGeoDaoRedis) findByGeo(ctx context.Context, query models.GeoQuery) ([]models.Site, error) {
	locations, err := d.Client.GeoRadius(ctx, d.KeySchema.SiteGeoKey(),
		query.Coordinate.Lng, query.Coordinate.Lat,
		&goredis.GeoRadiusQuery{
			Radius: query.Radius,
			Unit:   string(query.RadiusUnit),
		}).Result()
	if err != nil {
		return nil, err
	}

	sites := make([]models.Site, 0, len(locations))
	for _, loc := range locations {
		id, err := strconv.Atoi(loc.Name)
		if err != nil {
			continue
		}
		result, err := d.Client.HGetAll(ctx, d.KeySchema.SiteHashKey(id)).Result()
		if err != nil {
			return nil, err
		}
		site, err := models.SiteFromFlatMap(result)
		if err != nil {
			return nil, err
		}
		sites = append(sites, site)
	}
	return sites, nil
}

func (d *SiteGeoDaoRedis) findByGeoWithCapacity(ctx context.Context, query models.GeoQuery) ([]models.Site, error) {
	locations, err := d.Client.GeoRadius(ctx, d.KeySchema.SiteGeoKey(),
		query.Coordinate.Lng, query.Coordinate.Lat,
		&goredis.GeoRadiusQuery{
			Radius: query.Radius,
			Unit:   string(query.RadiusUnit),
		}).Result()
	if err != nil {
		return nil, err
	}

	capacityKey := d.KeySchema.CapacityRankingKey()
	pipe := d.Client.Pipeline()
	cmds := make([]*goredis.FloatCmd, len(locations))
	for i, loc := range locations {
		cmds[i] = pipe.ZScore(ctx, capacityKey, loc.Name)
	}
	_, _ = pipe.Exec(ctx)

	// Collect site IDs with capacity above threshold
	var filteredIDs []int
	for i, loc := range locations {
		score, err := cmds[i].Result()
		if err != nil {
			continue
		}
		if score > CapacityThreshold {
			id, err := strconv.Atoi(loc.Name)
			if err != nil {
				continue
			}
			filteredIDs = append(filteredIDs, id)
		}
	}

	// Fetch site hashes with pipeline
	pipe2 := d.Client.Pipeline()
	hashCmds := make([]*goredis.MapStringStringCmd, len(filteredIDs))
	for i, id := range filteredIDs {
		hashCmds[i] = pipe2.HGetAll(ctx, d.KeySchema.SiteHashKey(id))
	}
	_, _ = pipe2.Exec(ctx)

	sites := make([]models.Site, 0, len(filteredIDs))
	for _, cmd := range hashCmds {
		result, err := cmd.Result()
		if err != nil || len(result) == 0 {
			continue
		}
		site, err := models.SiteFromFlatMap(result)
		if err != nil {
			continue
		}
		sites = append(sites, site)
	}
	return sites, nil
}

func (d *SiteGeoDaoRedis) FindAll(ctx context.Context) ([]models.Site, error) {
	siteIDs, err := d.Client.ZRange(ctx, d.KeySchema.SiteGeoKey(), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	pipe := d.Client.Pipeline()
	cmds := make([]*goredis.MapStringStringCmd, len(siteIDs))
	for i, idStr := range siteIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		cmds[i] = pipe.HGetAll(ctx, d.KeySchema.SiteHashKey(id))
	}
	_, _ = pipe.Exec(ctx)

	sites := make([]models.Site, 0, len(siteIDs))
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		result, err := cmd.Result()
		if err != nil || len(result) == 0 {
			continue
		}
		site, err := models.SiteFromFlatMap(result)
		if err != nil {
			continue
		}
		sites = append(sites, site)
	}
	return sites, nil
}
