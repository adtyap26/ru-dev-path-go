package redis

import (
	"context"

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
	// START Challenge #1
	// Use SMEMBERS to get all site IDs from the site IDs key,
	// then use HGETALL for each site ID to get the site hash.
	// Finally, convert each hash to a Site model using models.SiteFromFlatMap.
	//
	// Hint: Use d.KeySchema.SiteIDsKey() and d.KeySchema.SiteHashKey(id)
	// Hint: Use d.Client.SMembers() and d.Client.HGetAll()
	siteHashes := make([]map[string]string, 0) // TODO: populate this
	// END Challenge #1

	sites := make([]models.Site, 0, len(siteHashes))
	for _, hash := range siteHashes {
		site, err := models.SiteFromFlatMap(hash)
		if err != nil {
			return nil, err
		}
		sites = append(sites, site)
	}
	return sites, nil
}
