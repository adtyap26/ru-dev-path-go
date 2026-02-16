### Challenge #1

#### OVERVIEW

For your first programming challenge, you're going to implement `FindAll()` in `SiteDaoRedis`. Once you've successfully implemented this method, the map on the front end will be populated with all of the sample solar installation data.

First, open `internal/dao/redis/site.go`. You'll see a bare-bones implementation that looks like this:

```go
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
```

#### GET A FAILING TEST

Next, take a look at the tests in `internal/dao/redis/site_test.go`. You'll see one relevant test: `TestSite_FindAll`.

Remove the `t.Skip("Remove for Challenge #1")` line from this test.

Then run this test file from the command line:

```
$ go test ./internal/dao/redis/ -run TestSite -v
```

You should have one failing test.

#### HOW TO SOLVE IT

To correctly implement the `FindAll` method, you'll need to remember that we've stored site IDs in a set. You can see how to get the name of the set by studying the `Insert()` method of `SiteDaoRedis`.

Note: You might notice this pattern while studying the `InsertWithClient()` method:

```go
func (d *SiteDaoRedis) InsertWithClient(ctx context.Context, site models.Site, client goredis.Cmdable) error {
```

This is a performance optimization that lets the method use either the Redis client instance stored in `d.Client` or a passed-in pipeline/client object.

Go ahead and ignore this for now, and in your code use the Redis client instance stored in `d.Client`.

So, first, get the IDs of all of the sites from that set using `d.Client.SMembers()`. Then, for each site ID, get its hash using `d.Client.HGetAll()` and add the hash to the `siteHashes` slice.

Once you've populated the `siteHashes` variable, the following loop will convert each hash into a `Site` model using `models.SiteFromFlatMap()`:

```go
sites := make([]models.Site, 0, len(siteHashes))
for _, hash := range siteHashes {
    site, err := models.SiteFromFlatMap(hash)
    ...
}
```

Note: You'll need to convert the string site IDs from `SMembers` to integers using `strconv.Atoi()`. Add `"strconv"` to the imports.


### Challenge #2

For your second programming challenge, you're going to complete the `insertMetric()` method of the `MetricDaoRedis` struct. After you've done this, you can rerun the data loader to generate metrics.

And you'll also see those metrics charted on the metrics page of the front-end dashboard.

To start, open up `internal/dao/redis/metric.go`. Now scroll down to find the `insertMetric` method. It should look something like this:

```go
func (d *MetricDaoRedis) insertMetric(ctx context.Context, siteID int, value float64, unit models.MetricUnit, t time.Time, pipe goredis.Pipeliner) {
    metricKey := d.KeySchema.DayMetricKey(siteID, unit, t)
    minuteOfDay := getDayMinute(t)

    // START Challenge #2
    // ...
    _ = metricKey    // TODO: remove after implementing
    _ = minuteOfDay  // TODO: remove after implementing
    // END Challenge #2
}
```

As you can see, part of the method has already been implemented for you. This method receives a `goredis.Pipeliner` object and should run the operations necessary to store the given value as a metric for the site with the given `siteID` using the pipeline.

So, your task is to complete the logic necessary to:

1. Store the metric within a sorted set using the pipeline
2. Make sure the set expires within `MetricExpirationSeconds`

The member should be a string in the format `"%.2f:%d"` (value:minuteOfDay), and the score should be the `minuteOfDay` as a `float64`.

Refer to the go-redis documentation for `ZAdd()` if you need help. Use `goredis.Z{Score: ..., Member: ...}` for the member argument.

Don't forget to remove the `_ = metricKey` and `_ = minuteOfDay` placeholder lines after implementing.

#### TESTS

We've included tests to help you check your work. To see them, open up `internal/dao/redis/metric_test.go` and scroll down to the Challenge #2 tests. Remove the `t.Skip("Remove for Challenge #2")` lines from the three test functions.

You can run just this test suite from the command line as follows:

```
$ go test ./internal/dao/redis/ -run TestMetric -v
```

#### LOADING METRICS

Once you get these tests to pass, run the data loader.

Note: If your Redis deployment requires a username and/or password, make sure to set up the project's environment variables. See the README.md file for instructions.

You can run the data loader with make:

```
$ make load
```

After loading data completes, start the server and check your work on the front end.


### Challenge #3

Your third challenge is to write an optimized and correct implementation for the `Update` method in `SiteStatsDaoRedis`. Open up `internal/dao/redis/site_stats.go`, and have a look at the `UpdateWithPipeline()` method. It currently looks like this:

```go
func (d *SiteStatsDaoRedis) UpdateWithPipeline(ctx context.Context, reading models.MeterReading, pipe goredis.Pipeliner) error {
    t := reading.TimestampTime()
    key := d.KeySchema.SiteStatsKey(reading.SiteID, t)

    execute := false
    if pipe == nil {
        pipe = d.Client.Pipeline()
        execute = true
    }

    // START Challenge #3
    // ...
    _ = key     // TODO: remove after implementing
    _ = reading // TODO: remove after implementing
    // END Challenge #3

    if execute {
        _, err := pipe.Exec(ctx)
        return err
    }
    return nil
}
```

This method should update site statistics efficiently using a pipeline and Lua scripts. You need to:

1. Set the last reporting time using `pipe.HSet()` with `models.SiteStatsLastReportingTime` (use `time.Now().UTC().Format(time.RFC3339)` for the value)
2. Increment the count using `pipe.HIncrBy()` with `models.SiteStatsCount`
3. Set the key expiration using `pipe.Expire()` with `WeekSeconds`
4. Perform the compare-and-update operations using Lua scripts from the `scripts` package:
   - `scripts.UpdateIfGreater()` for `models.SiteStatsMaxWH` with `reading.WHGenerated`
   - `scripts.UpdateIfLess()` for `models.SiteStatsMinWH` with `reading.WHGenerated`
   - `scripts.UpdateIfGreater()` for `models.SiteStatsMaxCapacity` with `reading.CurrentCapacity()`

The Lua scripts in `internal/scripts/` provide the `UpdateIfGreater(ctx, pipe, key, field, value)` and `UpdateIfLess(ctx, pipe, key, field, value)` functions. These perform atomic compare-and-update operations, making the method immune to race conditions.

Don't forget to:
- Uncomment the `"redisolar-go/internal/scripts"` import at the top of the file
- Remove the `_ = key` and `_ = reading` placeholder lines

#### INTEGRATION WITH METER READING

You also need to integrate the site stats update into the meter reading flow. Open `internal/dao/redis/meter_reading.go` and:

1. Uncomment the `statsDao *SiteStatsDaoRedis` field in the struct
2. Uncomment `statsDao: NewSiteStatsDao(base)` in `NewMeterReadingDao()`
3. Uncomment the `d.statsDao.Update(ctx, reading)` call in `AddWithPipeline()`

#### TESTS

Run the test suite like so:

```
$ go test ./internal/dao/redis/ -run TestSiteStats -v
```

Once you get the test suite passing, try running the loader script a couple of times. If your Redis deployment is on a cloud server, you should notice improved performance from the pipeline-based approach.

Now, pat yourself on the back for performing an important optimization and for making the method immune to race conditions!


### Challenge #4

Open up the file `internal/dao/redis/capacity_report.go` and have a look at the `GetRank()` method. The method returns the ranking of a given site's capacity, where 0 represents the site with the greatest capacity.

```go
func (d *CapacityReportDaoRedis) GetRank(ctx context.Context, siteID int) (int64, error) {
    // START Challenge #4
    // ...
    _ = siteID // TODO: remove after implementing
    return 0, nil
    // END Challenge #4
}
```

Now open up the file `internal/dao/redis/capacity_report_test.go` and scroll down to the `TestCapacity_GetRank()` test function. Remove the `t.Skip("Remove for Challenge #4")` line, and run the test to ensure that it fails:

```
$ go test ./internal/dao/redis/ -run TestCapacity_GetRank -v
```

Now, go back to `internal/dao/redis/capacity_report.go`, and replace the stub with a call to the correct Redis sorted set command to get the site's rank.

Use `d.Client.ZRevRank()` with `d.KeySchema.CapacityRankingKey()` as the key and `strconv.Itoa(siteID)` as the member. Don't forget to remove the `_ = siteID` placeholder line and the `return 0, nil` stub.

You may want to refer to the [Redis Sorted Set command documentation](https://redis.io/commands#sorted-set). Don't forget to check your work by running the tests until they pass.


### Challenge #5

For Challenge #5, you're going to get the "Excess Capacity" filter working in the app.

To do that, you'll finish implementing the method `findByGeoWithCapacity()`, which you can find in the struct `SiteGeoDaoRedis` in `internal/dao/redis/site_geo.go`. Go ahead and open up that file, and scroll down to the first `START Challenge #5` comment.

In `findByGeoWithCapacity()`, there are two sections marked with `START Challenge #5` where you'll have to write some code.

#### STEP 1

Your first task is to get the sites matching the geo query. You need to populate the `locations` variable using `d.Client.GeoRadius()`.

```go
// START Challenge #5
// Step 1: Get site IDs matching the GEO query using GEORADIUS.
var locations []goredis.GeoLocation // TODO: populate using GeoRadius
_ = query                           // TODO: remove after implementing
// END Challenge #5
```

Use `d.Client.GeoRadius()` with `d.KeySchema.SiteGeoKey()`, the query's longitude and latitude, and a `&goredis.GeoRadiusQuery{Radius: query.Radius, Unit: string(query.RadiusUnit)}`.

#### STEP 2

For the second task, you need to use the pipeline to get each site's capacity score and filter sites with excess capacity.

```go
// START Challenge #5
// Step 2: Use a pipeline to get the capacity score (ZSCORE) for each site.
var filteredIDs []int // TODO: populate with site IDs that have excess capacity
_ = locations        // TODO: remove after implementing
_ = pipe             // TODO: remove after implementing
// END Challenge #5
```

For each location, call `pipe.ZScore(ctx, capacityKey, loc.Name)` where `capacityKey` is `d.KeySchema.CapacityRankingKey()`. Store the `*goredis.FloatCmd` results in a slice. After `pipe.Exec(ctx)`, iterate over the results and collect site IDs whose score is greater than `CapacityThreshold` (0.2) into `filteredIDs`.

Don't forget to remove the placeholder `_ = ...` lines and convert location names to integers with `strconv.Atoi(loc.Name)`.

To see how `filteredIDs` is used, study the pipeline fetch at the end of the method that retrieves site hashes.

#### TESTS

To test your work, open up `internal/dao/redis/site_geo_test.go`, and scroll down to the `TestSiteGeo_FindByGeoWithExcessCapacity()` function. Remove the `t.Skip("Remove for Challenge #5")` line. You can run this test from the command line:

```
$ go test ./internal/dao/redis/ -run TestSiteGeo_FindByGeoWithExcessCapacity -v
```

#### FINISHING UP

Once you get this method working, you'll be able to use the "Excess Capacity" filter on the front page of the app. Give it a go, and be proud of your work!


### Challenge #6

For Challenge #6, you're going to implement the Redis Streams insert that powers the real-time feed in the app.

Open `internal/dao/redis/feed.go` and find the `InsertWithPipeline()` method. It currently looks like this:

```go
func (d *FeedDaoRedis) InsertWithPipeline(ctx context.Context, reading models.MeterReading, pipe goredis.Pipeliner) error {
    execute := false
    if pipe == nil {
        pipe = d.Client.Pipeline()
        execute = true
    }

    data := models.MeterReadingToStreamMap(reading)

    // START Challenge #6
    // ...
    _ = data // TODO: remove after implementing
    // END Challenge #6

    if execute {
        _, err := pipe.Exec(ctx)
        return err
    }
    return nil
}
```

Your task is to add the meter reading data to two Redis Streams using XADD:

1. The **global feed stream** (`d.KeySchema.GlobalFeedKey()`) with `MaxLen: GlobalMaxFeedLength` and `Approx: true`
2. The **site-specific feed stream** (`d.KeySchema.FeedKey(reading.SiteID)`) with `MaxLen: SiteMaxFeedLength` and `Approx: true`

Use `pipe.XAdd()` with `&goredis.XAddArgs{Stream: ..., MaxLen: ..., Approx: true, Values: data}` for each stream.

Don't forget to remove the `_ = data` placeholder line.

#### TESTS

Open up `internal/dao/redis/feed_test.go` and remove the `t.Skip("Remove for Challenge #6")` line from `TestFeed_BasicInsertReturnsRecent`.

Run the test:

```
$ go test ./internal/dao/redis/ -run TestFeed -v
```


### Challenge #7

As far as I'm concerned, this is the most interesting challenge of the course. I hope you'll give it a try!

Your challenge is to implement the `Hit()` method on `SlidingWindowRateLimiter`. You'll find a bare-bones implementation in `internal/dao/redis/sliding_window_rate_limiter.go`:

```go
func (rl *SlidingWindowRateLimiter) Hit(ctx context.Context, name string) error {
    key := rl.keySchema.SlidingWindowRateLimiterKey(name, int(rl.windowSizeMs), rl.maxHits)

    // START Challenge #7
    // ...
    _ = key // TODO: remove after implementing
    return nil
    // END Challenge #7
}
```

A sliding window needs to record a timestamp for each request. You'll use a sorted set with three commands:

- **ZADD**
- **ZREMRANGEBYSCORE**
- **ZCARD**

Here's the algorithm:

1. Calculate the current time in milliseconds and the window start:
   ```go
   now := time.Now().UTC()
   nowMs := float64(now.UnixNano()) / 1e6
   windowStart := nowMs - rl.windowSizeMs
   ```

2. Create a unique member string (the element should be unique to avoid collisions):
   ```go
   member := fmt.Sprintf("%f-%f", nowMs, rand.Float64())
   ```

3. Use a pipeline with three commands:
   - **ZADD**: Add the member to the sorted set with `nowMs` as its score.
     ```go
     pipe.ZAdd(ctx, key, goredis.Z{Score: nowMs, Member: member})
     ```
   - **ZREMRANGEBYSCORE**: Remove all entries older than the window start. This is how we slide the window forward in time.
     ```go
     pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", windowStart))
     ```
   - **ZCARD**: Count the remaining entries in the sorted set.
     ```go
     cardCmd := pipe.ZCard(ctx, key)
     ```

4. Execute the pipeline and check the ZCARD result. If `count > rl.maxHits`, return `dao.ErrRateLimitExceeded`.

Don't forget to:
- Uncomment the `"fmt"`, `"math/rand"`, `"time"`, and `"redisolar-go/internal/dao"` imports at the top of the file
- Remove the `_ = key` placeholder and the `return nil` stub

#### KEY NAMING

Recall that all keys for this application live in `internal/keyschema/keyschema.go`. We've provided a method named `SlidingWindowRateLimiterKey` to generate the keys for your sliding window implementation.

Take a look at the method and study the structure of the key.

Notice that it has the form `[prefix]:limiter:[name]:[window_size_ms]:[max_hits]`. The key for the sliding window rate limiter doesn't need a minute block, like the fixed rate-limiter key. Instead, this key includes the window size in milliseconds in the name, to make it unique.

#### TESTS

We've included tests to help you check your work. Open up `internal/dao/redis/sliding_window_rate_limiter_test.go` and remove the `t.Skip("Remove for Challenge #7")` lines from all three test functions.

You can run just this test suite from the command line as follows:

```
$ go test ./internal/dao/redis/ -run TestSlidingWindow -v
```
