package redis

import (
	"context"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"redisolar-go/internal/dao"
	"redisolar-go/internal/models"
)

type SiteDaoRedis struct {
	RedisDao
}

func NewSiteDao(base RedisDao) *SiteDaoRedis {
	return &SiteDaoRedis{RedisDao: base}
}

func (d *SiteDaoRedis) Insert(ctx context.Context, site models.Site) error {
	return d.InsertWithClient(ctx, site, d.Client)
}

func (d *SiteDaoRedis) InsertWithClient(ctx context.Context, site models.Site, client goredis.Cmdable) error {
	hashKey := d.KeySchema.SiteHashKey(site.ID)
	siteIDsKey := d.KeySchema.SiteIDsKey()
	flat := models.SiteToFlatMap(site)
	if err := client.HSet(ctx, hashKey, flat).Err(); err != nil {
		return err
	}
	return client.SAdd(ctx, siteIDsKey, site.ID).Err()
}

func (d *SiteDaoRedis) InsertMany(ctx context.Context, sites ...models.Site) error {
	for _, site := range sites {
		if err := d.Insert(ctx, site); err != nil {
			return err
		}
	}
	return nil
}

func (d *SiteDaoRedis) FindByID(ctx context.Context, siteID int) (models.Site, error) {
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

func (d *SiteDaoRedis) FindAll(ctx context.Context) ([]models.Site, error) {
	siteIDs, err := d.Client.SMembers(ctx, d.KeySchema.SiteIDsKey()).Result()
	if err != nil {
		return nil, err
	}

	sites := make([]models.Site, 0, len(siteIDs))
	for _, idStr := range siteIDs {
		id, err := strconv.Atoi(idStr)
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
